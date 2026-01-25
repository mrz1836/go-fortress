//go:build mage

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/go-fortress/guardian"
	"github.com/mrz1836/go-fortress/guardian/reporter"
)

// Sentinel errors for CI commands.
var (
	errStaticValidationFailed = errors.New("static validation found errors")
	errTestsFailed            = errors.New("tests failed")
	errVerificationFailed     = errors.New("verification failed")
	errNameRequired           = errors.New("name parameter is required: magex ci:scenario name=LINT-001")
	errScenarioFailed         = errors.New("scenario failed")
)

// Ci is the namespace for CI testing commands.
type Ci mg.Namespace

// Static runs static validation only (no Docker required).
// Target execution time: < 2 seconds.
func (Ci) Static(ctx context.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, "Running static validation...")

	cfg := guardian.LoadFromEnv()
	g, err := guardian.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("creating guardian: %w", err)
	}

	results, err := g.RunStatic(ctx)
	if err != nil {
		return fmt.Errorf("running static validation: %w", err)
	}

	// Print results
	termReporter := reporter.NewTerminalReporter()
	report := &reporter.Report{
		Mode:          reporter.ModeStatic,
		StaticResults: results,
	}
	if err := termReporter.Write(ctx, report, os.Stdout); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	// Fail if there are errors
	for _, f := range results.Findings {
		if f.Severity == "error" {
			return errStaticValidationFailed
		}
	}

	return nil
}

// Test runs quick validation scenarios.
// Includes static validation plus fast failure scenarios.
// Target execution time: < 60 seconds.
func (Ci) Test(ctx context.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, "Running CI tests...")

	cfg := guardian.LoadFromEnv()
	g, err := guardian.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("creating guardian: %w", err)
	}

	report, err := g.RunTest(ctx)
	if err != nil {
		return fmt.Errorf("running tests: %w", err)
	}

	// Print results
	termReporter := reporter.NewTerminalReporter()
	if err := termReporter.Write(ctx, report, os.Stdout); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	// Write JSONL report
	jsonlPath := filepath.Join(cfg.OutputDir, cfg.JSONLOutput)
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}
	jsonlReporter := reporter.NewJSONLReporter()
	if err := jsonlReporter.WriteFile(ctx, report, jsonlPath); err != nil {
		return fmt.Errorf("writing JSONL report: %w", err)
	}

	// Check for failures
	if report.Summary.FailedScenarios > 0 || report.Summary.ErrorScenarios > 0 {
		return fmt.Errorf("%w: %d failed, %d errors",
			errTestsFailed, report.Summary.FailedScenarios, report.Summary.ErrorScenarios)
	}

	return nil
}

// Verify runs comprehensive validation.
// Includes all scenarios for pre-merge verification.
// Target execution time: < 5 minutes.
func (Ci) Verify(ctx context.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, "Running full CI verification...")

	cfg := guardian.LoadFromEnv()
	g, err := guardian.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("creating guardian: %w", err)
	}

	report, err := g.RunVerify(ctx)
	if err != nil {
		return fmt.Errorf("running verification: %w", err)
	}

	// Print results
	termReporter := reporter.NewTerminalReporter()
	if err := termReporter.Write(ctx, report, os.Stdout); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	// Write JSONL report
	jsonlPath := filepath.Join(cfg.OutputDir, cfg.JSONLOutput)
	jsonlReporter := reporter.NewJSONLReporter()
	if err := jsonlReporter.WriteFile(ctx, report, jsonlPath); err != nil {
		return fmt.Errorf("writing JSONL report: %w", err)
	}

	// Write SARIF report
	sarifPath := filepath.Join(cfg.OutputDir, cfg.SARIFOutput)
	sarifReporter := reporter.NewSARIFReporter()
	if err := sarifReporter.WriteFile(ctx, report, sarifPath); err != nil {
		return fmt.Errorf("writing SARIF report: %w", err)
	}

	// Write GitHub annotations if in CI
	if reporter.IsGitHubActions() {
		annoReporter := reporter.NewAnnotationsReporter()
		if err := annoReporter.Write(ctx, report, os.Stdout); err != nil {
			return fmt.Errorf("writing annotations: %w", err)
		}
	}

	// Check for failures
	if report.Summary.FailedScenarios > 0 || report.Summary.ErrorScenarios > 0 {
		return fmt.Errorf("%w: %d failed, %d errors",
			errVerificationFailed, report.Summary.FailedScenarios, report.Summary.ErrorScenarios)
	}

	return nil
}

// Scenario runs a single scenario by ID.
// Parameters: name=<ID> verbose=true keep=true timeout=<duration>
func (Ci) Scenario(ctx context.Context) error {
	args := getMageArgs()
	params := parseArgs(args)

	name := params["name"]
	if name == "" {
		return errNameRequired
	}

	cfg := guardian.LoadFromEnv()

	// Override config from parameters
	if params["verbose"] == "true" {
		cfg.Verbose = true
	}
	if params["keep"] == "true" {
		cfg.KeepContainers = true
	}

	g, err := guardian.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("creating guardian: %w", err)
	}

	opts := guardian.ScenarioOptions{
		Verbose:       cfg.Verbose,
		KeepContainer: cfg.KeepContainers,
	}

	_, _ = fmt.Fprintf(os.Stdout, "Running scenario %s...\n", name)

	result, err := g.RunScenario(ctx, name, opts)
	if err != nil {
		return fmt.Errorf("running scenario: %w", err)
	}

	// Print result
	_, _ = fmt.Fprintf(os.Stdout, "\nScenario: %s\n", result.ScenarioID)
	_, _ = fmt.Fprintf(os.Stdout, "Status: %s\n", result.Status)
	_, _ = fmt.Fprintf(os.Stdout, "Duration: %s\n", result.Duration)

	if result.Error != "" {
		_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", result.Error)
	}

	if len(result.MatchedPatterns) > 0 {
		_, _ = fmt.Fprintf(os.Stdout, "Matched patterns: %v\n", result.MatchedPatterns)
	}

	if len(result.MissingPatterns) > 0 {
		_, _ = fmt.Fprintf(os.Stdout, "Missing patterns: %v\n", result.MissingPatterns)
	}

	if result.Status != reporter.ResultPass {
		return fmt.Errorf("%w: %s", errScenarioFailed, name)
	}

	return nil
}

// List shows all available scenarios.
// Parameters: filter=<category>
func (Ci) List(ctx context.Context) error {
	args := getMageArgs()
	params := parseArgs(args)

	cfg := guardian.LoadFromEnv()
	g, err := guardian.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("creating guardian: %w", err)
	}

	filter := guardian.ScenarioFilter{
		Category: params["filter"],
	}

	scenarios, err := g.ListScenarios(ctx, filter)
	if err != nil {
		return fmt.Errorf("listing scenarios: %w", err)
	}

	// Convert to list items for terminal reporter
	items := make([]reporter.ScenarioListItem, 0, len(scenarios))
	for _, s := range scenarios {
		items = append(items, reporter.ScenarioListItem{
			ID:             s.ID,
			Category:       s.Category,
			Description:    s.Description,
			ExpectedStatus: s.ExpectedStatus,
			Tags:           s.Tags,
			Disabled:       s.Disabled,
		})
	}

	// Use formatted terminal output
	termReporter := reporter.NewTerminalReporter()
	termReporter.WriteScenarioList(os.Stdout, items, params["filter"])

	return nil
}

// getMageArgs returns arguments passed via MAGE_ARGS environment variable.
func getMageArgs() []string {
	if args := os.Getenv("MAGE_ARGS"); args != "" {
		return strings.Fields(args)
	}
	// Also check for command line args after '--'
	return os.Args[1:]
}

// parseArgs parses key=value arguments into a map.
func parseArgs(args []string) map[string]string {
	params := make(map[string]string)
	for _, arg := range args {
		if parts := strings.SplitN(arg, "=", 2); len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
	}
	return params
}
