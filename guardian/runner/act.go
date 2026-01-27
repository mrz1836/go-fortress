// Package runner provides workflow execution using act.
package runner

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// ErrInsufficientDiskSpace is returned when there is not enough disk space for container images.
var ErrInsufficientDiskSpace = errors.New("insufficient disk space")

// ErrDangerousEnvVar is returned when a dangerous environment variable is detected.
var ErrDangerousEnvVar = errors.New("dangerous environment variable")

// dangerousEnvVars contains environment variable names that could be used for injection attacks.
// These variables affect dynamic linker, interpreter paths, or other security-sensitive settings.
//
//nolint:gochecknoglobals // security configuration is intentionally global
var dangerousEnvVars = map[string]bool{
	// Dynamic linker injection
	"LD_PRELOAD":            true,
	"LD_LIBRARY_PATH":       true,
	"LD_AUDIT":              true,
	"LD_DEBUG":              true,
	"LD_PROFILE":            true,
	"DYLD_INSERT_LIBRARIES": true, // macOS equivalent of LD_PRELOAD
	"DYLD_LIBRARY_PATH":     true, // macOS library path

	// Interpreter path injection
	"PYTHONPATH":        true,
	"PYTHONSTARTUP":     true,
	"PYTHONHOME":        true,
	"NODE_PATH":         true,
	"NODE_OPTIONS":      true,
	"RUBYLIB":           true,
	"RUBYOPT":           true,
	"PERL5LIB":          true,
	"PERL5OPT":          true,
	"CLASSPATH":         true,
	"JAVA_TOOL_OPTIONS": true,

	// Shell injection
	"BASH_ENV":       true,
	"ENV":            true, // Used by sh
	"ZDOTDIR":        true,
	"PROMPT_COMMAND": true,

	// Git hooks (could execute arbitrary code)
	"GIT_EXEC_PATH":   true,
	"GIT_SSH_COMMAND": true,

	// Other dangerous variables
	"IFS": true, // Input Field Separator - can affect shell parsing
}

// validEnvVarNamePattern matches valid environment variable names.
// Names must start with a letter or underscore and contain only alphanumeric chars and underscores.
var validEnvVarNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// ValidateEnvVar checks if an environment variable name is safe to use.
// Returns an error if the variable name is dangerous or invalid.
func ValidateEnvVar(name string) error {
	// Check for valid format
	if !validEnvVarNamePattern.MatchString(name) {
		return fmt.Errorf("%w: invalid variable name format: %s", ErrDangerousEnvVar, name)
	}

	// Check against dangerous variable list
	if dangerousEnvVars[name] {
		return fmt.Errorf("%w: %s is a security-sensitive variable", ErrDangerousEnvVar, name)
	}

	return nil
}

// ValidateEnvVars checks all environment variable names in a map.
// Returns an error containing all invalid variable names if any are found.
func ValidateEnvVars(env map[string]string) error {
	var invalidVars []string

	for name := range env {
		if err := ValidateEnvVar(name); err != nil {
			invalidVars = append(invalidVars, name)
		}
	}

	if len(invalidVars) > 0 {
		return fmt.Errorf("%w: rejected variables: %s", ErrDangerousEnvVar, strings.Join(invalidVars, ", "))
	}

	return nil
}

// ActRunner implements Runner using nektos/act.
type ActRunner struct {
	path string
}

// RetryConfig configures retry behavior for operations.
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

// NewActRunner creates a new act-based runner.
func NewActRunner(path string) (*ActRunner, error) {
	if path == "" {
		path = "act"
	}

	return &ActRunner{path: path}, nil
}

// DefaultRetryConfig returns sensible defaults for retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
	}
}

// CheckAvailable verifies that act is installed and Docker is running.
func (r *ActRunner) CheckAvailable(ctx context.Context) error {
	// Check act is installed
	if err := r.checkActInstalled(ctx); err != nil {
		return fmt.Errorf("act not available: %w", err)
	}

	// Check Docker is running
	if err := r.checkDockerRunning(ctx); err != nil {
		return fmt.Errorf("docker not available: %w", err)
	}

	return nil
}

// Run executes a workflow using act.
func (r *ActRunner) Run(ctx context.Context, opts RunOptions) (*RunResult, error) {
	start := time.Now()

	// Make EventFile path absolute before changing working directory
	if opts.EventFile != "" && !filepath.IsAbs(opts.EventFile) {
		cwd, err := os.Getwd()
		if err == nil {
			opts.EventFile = filepath.Join(cwd, opts.EventFile)
		}
	}

	args := r.BuildArgs(opts)

	cmd := exec.CommandContext(ctx, r.path, args...) //nolint:gosec // r.path is validated act binary
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	}

	// Set environment variables with security validation
	if len(opts.Env) > 0 {
		// Validate all environment variable names before applying
		if err := ValidateEnvVars(opts.Env); err != nil {
			return nil, fmt.Errorf("environment validation failed: %w", err)
		}

		cmd.Env = cmd.Environ()

		for k, v := range opts.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &RunResult{
		Duration: time.Since(start),
		Output:   stdout.String() + stderr.String(),
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("running act: %w", err)
		}
	}

	// Extract container ID if keep-container was used
	if opts.KeepContainer {
		result.ContainerID = ExtractContainerID(result.Output)
	}

	return result, nil
}

// pullImage attempts to pull a Docker image once, returning retry decision.
func pullImage(ctx context.Context, image string) (struct{}, bool, error) {
	cmd := exec.CommandContext(ctx, "docker", "pull", image)
	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if runErr != nil {
		errMsg := stderr.String()
		// Retry on network-related errors
		shouldRetry := strings.Contains(errMsg, "timeout") ||
			strings.Contains(errMsg, "connection refused") ||
			strings.Contains(errMsg, "network") ||
			strings.Contains(errMsg, "temporary failure") ||
			strings.Contains(errMsg, "503") ||
			strings.Contains(errMsg, "rate limit")

		return struct{}{}, shouldRetry, fmt.Errorf("pulling image %s: %w (%s)", image, runErr, errMsg)
	}

	return struct{}{}, false, nil
}

// PullImageWithRetry pulls a Docker image with exponential backoff retry.
func (r *ActRunner) PullImageWithRetry(ctx context.Context, image string) error {
	cfg := DefaultRetryConfig()

	_, err := RetryWithBackoff(ctx, cfg, func() (struct{}, bool, error) {
		return pullImage(ctx, image)
	})

	return err
}

// EnsureImageAvailable ensures a Docker image is available, pulling if necessary.
func (r *ActRunner) EnsureImageAvailable(ctx context.Context, image string) error {
	// Check if image exists locally
	cmd := exec.CommandContext(ctx, "docker", "image", "inspect", image)
	if err := cmd.Run(); err == nil {
		return nil // Image exists
	}

	// Pull with retry
	return r.PullImageWithRetry(ctx, image)
}

// BuildArgs constructs the command line arguments for act.
func (r *ActRunner) BuildArgs(opts RunOptions) []string {
	args := []string{}

	// Workflow file - prepend .github/workflows/ if not already present
	if opts.WorkflowFile != "" {
		workflowPath := opts.WorkflowFile
		if !strings.HasPrefix(workflowPath, ".github/workflows/") && !strings.HasPrefix(workflowPath, "/") {
			workflowPath = ".github/workflows/" + workflowPath
		}
		args = append(args, "--workflows", workflowPath)
	}

	// Specific job
	if opts.Job != "" {
		args = append(args, "--job", opts.Job)
	}

	// Event payload
	if opts.EventFile != "" {
		args = append(args, "--eventpath", opts.EventFile)
	}

	// Secrets file
	if opts.SecretsFile != "" {
		args = append(args, "--secret-file", opts.SecretsFile)
	}

	// Platform mapping for consistent behavior
	args = append(args, "--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest")
	args = append(args, "--platform", "ubuntu-22.04=catthehacker/ubuntu:act-22.04")
	args = append(args, "--platform", "ubuntu-24.04=catthehacker/ubuntu:act-latest")

	// Verbose output
	if opts.Verbose {
		args = append(args, "--verbose")
	}

	// Keep container for debugging
	if opts.KeepContainer {
		args = append(args, "--reuse")
	}

	return args
}

// checkActInstalled verifies act is available.
func (r *ActRunner) checkActInstalled(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, r.path, "--version") //nolint:gosec // r.path is validated act binary
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("act not found at %s: %w", r.path, err)
	}

	return nil
}

// checkDockerRunning verifies Docker daemon is accessible.
func (r *ActRunner) checkDockerRunning(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker not running: %w", err)
	}

	return nil
}

// ExtractContainerID extracts container ID from act output.
func ExtractContainerID(output string) string {
	// Act outputs container information in various formats
	// Look for common patterns
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "container_id=") {
			parts := strings.Split(line, "container_id=")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

// CheckDiskSpace verifies sufficient disk space for container images.
// requiredGB is the minimum required disk space in gigabytes.
// path is the directory to check (typically Docker's data directory or current working dir).
func CheckDiskSpace(path string, requiredGB int) error {
	if path == "" {
		path = "."
	}

	var stat syscall.Statfs_t

	if err := syscall.Statfs(path, &stat); err != nil {
		// If we can't check, allow to proceed and let Docker handle it
		return nil //nolint:nilerr // intentional - allow to proceed when check fails
	}

	// Available bytes = available blocks * block size
	// #nosec G115 -- block size is always positive and fits in uint64
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	// #nosec G115 -- requiredGB is validated to be positive
	requiredBytes := uint64(requiredGB) * 1024 * 1024 * 1024

	if availableBytes < requiredBytes {
		availableGB := float64(availableBytes) / (1024 * 1024 * 1024)

		return fmt.Errorf("%w: %.1f GB available, %d GB required", ErrInsufficientDiskSpace, availableGB, requiredGB)
	}

	return nil
}

// RetryWithBackoff executes a function with exponential backoff on failure.
// The function should return (result, shouldRetry, error).
func RetryWithBackoff[T any](ctx context.Context, cfg RetryConfig, operation func() (T, bool, error)) (T, error) {
	var result T

	var lastErr error

	backoff := cfg.InitialBackoff

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		var shouldRetry bool

		result, shouldRetry, lastErr = operation()
		if lastErr == nil {
			return result, nil
		}

		if !shouldRetry || attempt >= cfg.MaxRetries {
			return result, lastErr
		}

		// Add jitter to prevent thundering herd using crypto/rand for secure randomness
		maxJitter := int64(backoff / 4)
		jitterBig, _ := rand.Int(rand.Reader, big.NewInt(maxJitter))
		jitter := time.Duration(jitterBig.Int64())
		sleepDuration := backoff + jitter

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(sleepDuration):
		}

		// Increase backoff for next attempt
		backoff = time.Duration(float64(backoff) * cfg.BackoffFactor)
		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}
	}

	return result, lastErr
}
