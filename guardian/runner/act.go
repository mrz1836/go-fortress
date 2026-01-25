package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ActRunner implements Runner using nektos/act.
type ActRunner struct {
	path string
}

// NewActRunner creates a new act-based runner.
func NewActRunner(path string) (*ActRunner, error) {
	if path == "" {
		path = "act"
	}

	return &ActRunner{path: path}, nil
}

// CheckAvailable verifies that act is installed and Docker is running.
func (r *ActRunner) CheckAvailable(ctx context.Context) error {
	// Check act is installed
	if err := r.checkActInstalled(ctx); err != nil {
		return fmt.Errorf("act not available: %w", err)
	}

	// Check Docker is running
	if err := r.checkDockerRunning(ctx); err != nil {
		return fmt.Errorf("Docker not available: %w", err)
	}

	return nil
}

// Run executes a workflow using act.
func (r *ActRunner) Run(ctx context.Context, opts RunOptions) (*RunResult, error) {
	start := time.Now()

	args := r.buildArgs(opts)

	cmd := exec.CommandContext(ctx, r.path, args...)
	if opts.WorkingDir != "" {
		cmd.Dir = opts.WorkingDir
	}

	// Set environment variables
	if len(opts.Env) > 0 {
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
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("running act: %w", err)
		}
	}

	// Extract container ID if keep-container was used
	if opts.KeepContainer {
		result.ContainerID = extractContainerID(result.Output)
	}

	return result, nil
}

// buildArgs constructs the command line arguments for act.
func (r *ActRunner) buildArgs(opts RunOptions) []string {
	args := []string{}

	// Workflow file
	if opts.WorkflowFile != "" {
		args = append(args, "--workflows", opts.WorkflowFile)
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
	cmd := exec.CommandContext(ctx, r.path, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("act not found at %s: %w", r.path, err)
	}
	return nil
}

// checkDockerRunning verifies Docker daemon is accessible.
func (r *ActRunner) checkDockerRunning(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker not running: %w", err)
	}
	return nil
}

// extractContainerID extracts container ID from act output.
func extractContainerID(output string) string {
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
func CheckDiskSpace(requiredGB int) error {
	// This is a simplified check; in production you might use syscall.Statfs
	// For now, we'll skip this check and let Docker fail naturally
	_ = requiredGB
	return nil
}
