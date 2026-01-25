package runner

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/mrz1836/go-fortress/guardian/reporter"
)

// ScenarioRunner executes test scenarios against the act runner.
type ScenarioRunner struct {
	runner Runner
	config ScenarioConfig
}

// ScenarioConfig configures scenario execution.
type ScenarioConfig struct {
	// ParallelScenarios is the max concurrent scenario execution.
	ParallelScenarios int

	// DefaultTimeout is the default timeout for scenarios.
	DefaultTimeout time.Duration

	// Verbose enables detailed output for all scenarios.
	Verbose bool

	// KeepContainers preserves containers for all scenarios.
	KeepContainers bool
}

// NewScenarioRunner creates a new scenario runner.
func NewScenarioRunner(runner Runner, config ScenarioConfig) *ScenarioRunner {
	if config.ParallelScenarios <= 0 {
		config.ParallelScenarios = 4
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 30 * time.Second
	}

	return &ScenarioRunner{
		runner: runner,
		config: config,
	}
}

// ScenarioDefinition defines a scenario to execute.
type ScenarioDefinition struct {
	ID          string
	FixturePath string
	Workflow    string
	Job         string
	EventFile   string
	Expected    ExpectedResult
	Timeout     time.Duration
}

// ExpectedResult defines what a scenario should produce.
type ExpectedResult struct {
	Status          ExpectedStatus
	LogPatterns     []string
	ExcludePatterns []string
	ExitCode        *int
}

// ExpectedStatus represents expected workflow outcome.
type ExpectedStatus string

const (
	StatusSuccess ExpectedStatus = "success"
	StatusFailure ExpectedStatus = "failure"
)

// Execute runs a single scenario.
func (r *ScenarioRunner) Execute(ctx context.Context, scenario *ScenarioDefinition) (*reporter.ScenarioResult, error) {
	timeout := scenario.Timeout
	if timeout == 0 {
		timeout = r.config.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	opts := RunOptions{
		WorkingDir:    scenario.FixturePath,
		WorkflowFile:  scenario.Workflow,
		Job:           scenario.Job,
		EventFile:     scenario.EventFile,
		Timeout:       timeout,
		Verbose:       r.config.Verbose,
		KeepContainer: r.config.KeepContainers,
	}

	start := time.Now()
	runResult, err := r.runner.Run(ctx, opts)
	duration := time.Since(start)

	result := &reporter.ScenarioResult{
		ScenarioID: scenario.ID,
		Duration:   duration,
	}

	if err != nil {
		result.Status = reporter.ResultError
		result.Error = err.Error()
		return result, nil
	}

	result.ExitCode = runResult.ExitCode
	result.Output = runResult.Output

	// Validate against expected results
	result.Status, result.MatchedPatterns, result.MissingPatterns = validateResult(
		runResult, &scenario.Expected,
	)

	return result, nil
}

// ExecuteAll runs multiple scenarios with parallelism control.
// Respects the ParallelScenarios config setting for concurrent execution.
func (r *ScenarioRunner) ExecuteAll(ctx context.Context, scenarios []*ScenarioDefinition) ([]reporter.ScenarioResult, error) {
	if len(scenarios) == 0 {
		return []reporter.ScenarioResult{}, nil
	}

	// If parallelism is 1 or we have only 1 scenario, run sequentially
	if r.config.ParallelScenarios <= 1 || len(scenarios) == 1 {
		return r.executeSequential(ctx, scenarios)
	}

	return r.executeParallel(ctx, scenarios)
}

// executeSequential runs scenarios one at a time.
func (r *ScenarioRunner) executeSequential(ctx context.Context, scenarios []*ScenarioDefinition) ([]reporter.ScenarioResult, error) {
	results := make([]reporter.ScenarioResult, 0, len(scenarios))

	for _, s := range scenarios {
		select {
		case <-ctx.Done():
			// Context cancelled, skip remaining scenarios
			for _, remaining := range scenarios[len(results):] {
				results = append(results, reporter.ScenarioResult{
					ScenarioID: remaining.ID,
					Status:     reporter.ResultSkip,
					Error:      "context cancelled",
				})
			}
			return results, ctx.Err()
		default:
		}

		result, err := r.Execute(ctx, s)
		if err != nil {
			results = append(results, reporter.ScenarioResult{
				ScenarioID: s.ID,
				Status:     reporter.ResultError,
				Error:      err.Error(),
			})
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// executeParallel runs scenarios concurrently with a worker pool.
func (r *ScenarioRunner) executeParallel(ctx context.Context, scenarios []*ScenarioDefinition) ([]reporter.ScenarioResult, error) {
	type indexedResult struct {
		index  int
		result reporter.ScenarioResult
	}

	// Create work channel and results channel
	work := make(chan struct {
		index    int
		scenario *ScenarioDefinition
	}, len(scenarios))
	results := make(chan indexedResult, len(scenarios))

	// Start worker pool
	workerCount := r.config.ParallelScenarios
	if workerCount > len(scenarios) {
		workerCount = len(scenarios)
	}

	// Create a child context for workers
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start workers
	for i := 0; i < workerCount; i++ {
		go func() {
			for item := range work {
				select {
				case <-workerCtx.Done():
					results <- indexedResult{
						index: item.index,
						result: reporter.ScenarioResult{
							ScenarioID: item.scenario.ID,
							Status:     reporter.ResultSkip,
							Error:      "context cancelled",
						},
					}
					continue
				default:
				}

				result, err := r.Execute(workerCtx, item.scenario)
				if err != nil {
					results <- indexedResult{
						index: item.index,
						result: reporter.ScenarioResult{
							ScenarioID: item.scenario.ID,
							Status:     reporter.ResultError,
							Error:      err.Error(),
						},
					}
					continue
				}
				results <- indexedResult{
					index:  item.index,
					result: *result,
				}
			}
		}()
	}

	// Send work to workers
	for i, s := range scenarios {
		work <- struct {
			index    int
			scenario *ScenarioDefinition
		}{index: i, scenario: s}
	}
	close(work)

	// Collect results in order
	orderedResults := make([]reporter.ScenarioResult, len(scenarios))
	for i := 0; i < len(scenarios); i++ {
		r := <-results
		orderedResults[r.index] = r.result
	}

	return orderedResults, nil
}

// validateResult checks if the run result matches expectations.
func validateResult(result *RunResult, expected *ExpectedResult) (reporter.ResultStatus, []string, []string) {
	var matchedPatterns []string
	var missingPatterns []string

	// Check status (exit code based)
	actualSuccess := result.ExitCode == 0
	expectedSuccess := expected.Status == StatusSuccess

	if actualSuccess != expectedSuccess {
		if expectedSuccess {
			return reporter.ResultFail, nil, []string{fmt.Sprintf("expected success but got exit code %d", result.ExitCode)}
		}
		return reporter.ResultFail, nil, []string{fmt.Sprintf("expected failure but got exit code %d", result.ExitCode)}
	}

	// Check specific exit code if provided
	if expected.ExitCode != nil && result.ExitCode != *expected.ExitCode {
		return reporter.ResultFail, nil, []string{fmt.Sprintf("expected exit code %d but got %d", *expected.ExitCode, result.ExitCode)}
	}

	// Check log patterns
	for _, pattern := range expected.LogPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			missingPatterns = append(missingPatterns, fmt.Sprintf("invalid pattern: %s", pattern))
			continue
		}

		if re.MatchString(result.Output) {
			matchedPatterns = append(matchedPatterns, pattern)
		} else {
			missingPatterns = append(missingPatterns, pattern)
		}
	}

	// Check exclude patterns
	for _, pattern := range expected.ExcludePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		if re.MatchString(result.Output) {
			missingPatterns = append(missingPatterns, fmt.Sprintf("excluded pattern found: %s", pattern))
		}
	}

	// Determine final status
	if len(missingPatterns) > 0 {
		return reporter.ResultFail, matchedPatterns, missingPatterns
	}

	return reporter.ResultPass, matchedPatterns, nil
}
