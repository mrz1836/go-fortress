// Package guardian provides a CI validation framework for GitHub Actions workflows.
//
// Fortress Guardian enables local reproduction of CI failures via nektos/act,
// performs static validation using actionlint, enforces security policies as Go code,
// and integrates with MAGE-X through the ci: command namespace.
//
// # Quick Start
//
// Create a Guardian instance with default configuration:
//
//	g, err := guardian.New(guardian.DefaultConfig())
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Run static validation (no Docker required):
//
//	results, err := g.RunStatic(ctx)
//
// Run comprehensive verification:
//
//	report, err := g.RunVerify(ctx)
package guardian

import (
	"context"
	"fmt"
	"time"

	"github.com/mrz1836/go-fortress/guardian/policy"
	"github.com/mrz1836/go-fortress/guardian/reporter"
	"github.com/mrz1836/go-fortress/guardian/runner"
	"github.com/mrz1836/go-fortress/guardian/scenarios"
	"github.com/mrz1836/go-fortress/guardian/validator"
)

// Guardian is the main entry point for CI validation.
type Guardian struct {
	config     *Config
	validators *validator.Registry
	runner     runner.Runner
	policy     *policy.Engine
	reporters  *reporter.Registry
	scenarios  *scenarios.Registry
}

// New creates a Guardian instance with the given configuration.
func New(cfg *Config) (*Guardian, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	g := &Guardian{
		config:     cfg,
		validators: validator.NewRegistry(),
		reporters:  reporter.NewRegistry(),
		scenarios:  scenarios.NewRegistry(),
	}

	// Initialize policy engine
	policyEngine, err := policy.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("creating policy engine: %w", err)
	}
	g.policy = policyEngine

	// Load policy exceptions if config file exists
	if cfg.ExceptionsFile != "" {
		if err := g.policy.LoadExceptions(context.Background(), cfg.ExceptionsFile); err != nil {
			// Non-fatal: exceptions file may not exist yet
			_ = err
		}
	}

	// Initialize runner (may fail if Docker unavailable)
	r, err := runner.NewActRunner(cfg.ActPath)
	if err != nil {
		// Runner initialization failure is non-fatal; static validation still works
		g.runner = nil
	} else {
		g.runner = r
	}

	// Register default validators
	g.registerDefaultValidators()

	// Register default reporters
	g.registerDefaultReporters()

	// Register default scenarios
	g.registerDefaultScenarios()

	return g, nil
}

// registerDefaultValidators sets up the standard validators.
func (g *Guardian) registerDefaultValidators() {
	// Actionlint validator
	g.validators.Register(validator.NewActionlintValidator(g.config.ActionlintPath))

	// Schema validator for action.yml files
	g.validators.Register(validator.NewSchemaValidator())

	// Deprecation validator
	g.validators.Register(validator.NewDeprecationValidator())

	// Env validator for .env.base files
	g.validators.Register(validator.NewEnvValidator())
}

// registerDefaultReporters sets up the standard reporters.
func (g *Guardian) registerDefaultReporters() {
	g.reporters.Register(reporter.NewTerminalReporter())
	g.reporters.Register(reporter.NewJSONLReporter())
	g.reporters.Register(reporter.NewSARIFReporter())
	g.reporters.Register(reporter.NewAnnotationsReporter())
}

// registerDefaultScenarios sets up the built-in test scenarios.
func (g *Guardian) registerDefaultScenarios() {
	scenarios.RegisterAll(g.scenarios)
}

// RunStatic performs static validation only (no Docker required).
// Returns findings from all validators and policy checks.
// Target execution time: < 5 seconds.
func (g *Guardian) RunStatic(ctx context.Context) (*reporter.StaticResults, error) {
	start := time.Now()

	results := &reporter.StaticResults{
		Findings:      []validator.Finding{},
		ValidatorsRun: []string{},
	}

	// Find all workflow files
	workflows, err := g.findWorkflows()
	if err != nil {
		return nil, fmt.Errorf("finding workflows: %w", err)
	}

	// Run all validators on each workflow
	for _, wf := range workflows {
		findings, err := g.validators.ValidateAll(ctx, wf)
		if err != nil {
			return nil, fmt.Errorf("validating %s: %w", wf, err)
		}
		results.Findings = append(results.Findings, findings...)
	}

	// Record which validators ran
	for _, v := range g.validators.All() {
		results.ValidatorsRun = append(results.ValidatorsRun, v.Name())
	}

	// Run policy checks
	for _, wf := range workflows {
		workflow, err := policy.ParseWorkflow(wf)
		if err != nil {
			// Skip files that can't be parsed as workflows
			continue
		}
		policyFindings, err := g.policy.Evaluate(ctx, workflow)
		if err != nil {
			return nil, fmt.Errorf("evaluating policies for %s: %w", wf, err)
		}
		results.Findings = append(results.Findings, policyFindings...)
	}

	results.Duration = time.Since(start)
	return results, nil
}

// RunTest executes quick validation scenarios.
// Includes static validation plus fast failure scenarios.
// Target execution time: < 60 seconds.
func (g *Guardian) RunTest(ctx context.Context) (*reporter.Report, error) {
	start := time.Now()

	report := &reporter.Report{
		Version:   Version,
		StartTime: start,
		Mode:      reporter.ModeTest,
	}

	// Run static validation first
	staticResults, err := g.RunStatic(ctx)
	if err != nil {
		return nil, fmt.Errorf("static validation: %w", err)
	}
	report.StaticResults = staticResults

	// Run fast scenarios if runner is available
	if g.runner != nil {
		if err := g.runner.CheckAvailable(ctx); err == nil {
			fastScenarios := g.scenarios.ByTags([]string{"fast"})
			scenarioResults, err := g.executeScenarios(ctx, fastScenarios)
			if err != nil {
				return nil, fmt.Errorf("executing scenarios: %w", err)
			}
			report.ScenarioResults = scenarioResults
		}
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)
	report.Summary = g.calculateSummary(report)

	return report, nil
}

// RunVerify executes comprehensive validation.
// Includes all scenarios for pre-merge verification.
// Target execution time: < 5 minutes.
func (g *Guardian) RunVerify(ctx context.Context) (*reporter.Report, error) {
	start := time.Now()

	report := &reporter.Report{
		Version:   Version,
		StartTime: start,
		Mode:      reporter.ModeVerify,
	}

	// Run static validation first
	staticResults, err := g.RunStatic(ctx)
	if err != nil {
		return nil, fmt.Errorf("static validation: %w", err)
	}
	report.StaticResults = staticResults

	// Run all scenarios if runner is available
	if g.runner != nil {
		if err := g.runner.CheckAvailable(ctx); err == nil {
			allScenarios := g.scenarios.All()
			scenarioResults, err := g.executeScenarios(ctx, allScenarios)
			if err != nil {
				return nil, fmt.Errorf("executing scenarios: %w", err)
			}
			report.ScenarioResults = scenarioResults
		}
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)
	report.Summary = g.calculateSummary(report)

	return report, nil
}

// RunScenario executes a single scenario by ID.
// Used for debugging specific CI behaviors.
func (g *Guardian) RunScenario(ctx context.Context, id string, opts ScenarioOptions) (*reporter.ScenarioResult, error) {
	scenario, ok := g.scenarios.Get(id)
	if !ok {
		return nil, fmt.Errorf("scenario not found: %s", id)
	}

	if g.runner == nil {
		return nil, fmt.Errorf("runner not available: Docker may not be running")
	}

	if err := g.runner.CheckAvailable(ctx); err != nil {
		return nil, fmt.Errorf("runner not available: %w", err)
	}

	return g.executeScenario(ctx, scenario, opts)
}

// ListScenarios returns all available scenarios.
// Supports filtering by category or tags.
func (g *Guardian) ListScenarios(_ context.Context, filter ScenarioFilter) ([]scenarios.Info, error) {
	var result []scenarios.Info

	allScenarios := g.scenarios.All()
	for _, s := range allScenarios {
		// Apply category filter
		if filter.Category != "" && string(s.Category) != filter.Category {
			continue
		}

		// Apply tag filter
		if len(filter.Tags) > 0 && !hasAllTags(s.Tags, filter.Tags) {
			continue
		}

		result = append(result, scenarios.Info{
			ID:             s.ID,
			Category:       string(s.Category),
			Description:    s.Description,
			ExpectedStatus: string(s.Expected.Status),
			Tags:           s.Tags,
		})
	}

	return result, nil
}

// ScenarioOptions configures single scenario execution.
type ScenarioOptions struct {
	Verbose       bool
	KeepContainer bool
	Timeout       time.Duration
}

// ScenarioFilter configures scenario listing.
type ScenarioFilter struct {
	Category        string
	Tags            []string
	IncludeDisabled bool
}

// findWorkflows returns all workflow files in the configured directory.
func (g *Guardian) findWorkflows() ([]string, error) {
	return validator.FindWorkflowFiles(g.config.WorkflowsDir)
}

// executeScenarios runs multiple scenarios with parallelism control.
func (g *Guardian) executeScenarios(ctx context.Context, scenarioList []*scenarios.Scenario) ([]reporter.ScenarioResult, error) {
	results := make([]reporter.ScenarioResult, 0, len(scenarioList))

	// For now, run sequentially; parallel execution will be added in Phase 5 (US3)
	for _, s := range scenarioList {
		result, err := g.executeScenario(ctx, s, ScenarioOptions{})
		if err != nil {
			// Record error but continue with other scenarios
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

// executeScenario runs a single scenario and validates results.
func (g *Guardian) executeScenario(ctx context.Context, s *scenarios.Scenario, opts ScenarioOptions) (*reporter.ScenarioResult, error) {
	timeout := s.Timeout
	if opts.Timeout > 0 {
		timeout = opts.Timeout
	}
	if timeout == 0 {
		timeout = g.config.ScenarioTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	runOpts := runner.RunOptions{
		WorkingDir:    s.FixturePath,
		WorkflowFile:  s.Workflow,
		Job:           s.Job,
		EventFile:     s.EventFile,
		Timeout:       timeout,
		Verbose:       opts.Verbose,
		KeepContainer: opts.KeepContainer,
	}

	start := time.Now()
	runResult, err := g.runner.Run(ctx, runOpts)
	duration := time.Since(start)

	result := &reporter.ScenarioResult{
		ScenarioID: s.ID,
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
	result.Status, result.MatchedPatterns, result.MissingPatterns = s.Validate(runResult)

	return result, nil
}

// calculateSummary computes aggregate statistics for a report.
func (g *Guardian) calculateSummary(report *reporter.Report) reporter.ReportSummary {
	summary := reporter.ReportSummary{
		FindingsByLevel: make(map[validator.Severity]int),
	}

	// Count findings
	if report.StaticResults != nil {
		summary.TotalFindings = len(report.StaticResults.Findings)
		for _, f := range report.StaticResults.Findings {
			summary.FindingsByLevel[f.Severity]++
		}
	}

	// Count scenarios
	summary.TotalScenarios = len(report.ScenarioResults)
	for _, r := range report.ScenarioResults {
		switch r.Status {
		case reporter.ResultPass:
			summary.PassedScenarios++
		case reporter.ResultFail:
			summary.FailedScenarios++
		case reporter.ResultSkip:
			summary.SkippedScenarios++
		case reporter.ResultError:
			summary.ErrorScenarios++
		}
	}

	return summary
}

// hasAllTags checks if the scenario has all required tags.
func hasAllTags(scenarioTags, requiredTags []string) bool {
	tagSet := make(map[string]bool)
	for _, t := range scenarioTags {
		tagSet[t] = true
	}
	for _, t := range requiredTags {
		if !tagSet[t] {
			return false
		}
	}
	return true
}

// Version is the Guardian version.
const Version = "1.0.0"
