package runner

import (
	"context"
	"time"
)

// Runner executes GitHub Actions workflows locally.
type Runner interface {
	// Run executes a workflow using act.
	// Returns the execution result including output and exit code.
	Run(ctx context.Context, opts RunOptions) (*RunResult, error)

	// CheckAvailable verifies that act is installed and Docker is running.
	CheckAvailable(ctx context.Context) error
}

// RunOptions configures a workflow execution.
type RunOptions struct {
	// WorkingDir is the repository root for execution.
	WorkingDir string

	// WorkflowFile is the path to the workflow YAML file.
	WorkflowFile string

	// Job is the specific job to run (optional).
	Job string

	// EventFile is the path to the event payload JSON (optional).
	EventFile string

	// SecretsFile is the path to the secrets file (optional).
	SecretsFile string

	// Env is additional environment variables.
	Env map[string]string

	// Timeout is the maximum execution time.
	Timeout time.Duration

	// Verbose enables detailed output.
	Verbose bool

	// KeepContainer preserves the container after execution.
	KeepContainer bool
}

// RunResult contains the outcome of a workflow execution.
type RunResult struct {
	// ExitCode is the process exit code.
	ExitCode int

	// Output is the combined stdout/stderr.
	Output string

	// Duration is the execution time.
	Duration time.Duration

	// ContainerID is the container used (if KeepContainer was true).
	ContainerID string
}
