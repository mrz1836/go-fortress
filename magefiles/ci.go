//go:build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-fortress/guardian"
	"github.com/mrz1836/go-fortress/guardian/reporter"
)

// Ci is the namespace for CI testing commands.
type Ci mg.Namespace

// Static runs static validation only (no Docker required).
// Target execution time: < 2 seconds.
func (Ci) Static(ctx context.Context) error {
	fmt.Println("Running static validation...")

	cfg := guardian.LoadFromEnv()
	g, err := guardian.New(cfg)
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
			return fmt.Errorf("static validation found errors")
		}
	}

	return nil
}

// Test runs quick validation scenarios.
// Includes static validation plus fast failure scenarios.
// Target execution time: < 60 seconds.
func (Ci) Test(ctx context.Context) error {
	fmt.Println("Running CI tests...")

	cfg := guardian.LoadFromEnv()
	g, err := guardian.New(cfg)
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
		return fmt.Errorf("tests failed: %d failed, %d errors",
			report.Summary.FailedScenarios, report.Summary.ErrorScenarios)
	}

	return nil
}

// Verify runs comprehensive validation.
// Includes all scenarios for pre-merge verification.
// Target execution time: < 5 minutes.
func (Ci) Verify(ctx context.Context) error {
	fmt.Println("Running full CI verification...")

	cfg := guardian.LoadFromEnv()
	g, err := guardian.New(cfg)
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
		return fmt.Errorf("verification failed: %d failed, %d errors",
			report.Summary.FailedScenarios, report.Summary.ErrorScenarios)
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
		return fmt.Errorf("name parameter is required: magex ci:scenario name=LINT-001")
	}

	cfg := guardian.LoadFromEnv()

	// Override config from parameters
	if params["verbose"] == "true" {
		cfg.Verbose = true
	}
	if params["keep"] == "true" {
		cfg.KeepContainers = true
	}

	g, err := guardian.New(cfg)
	if err != nil {
		return fmt.Errorf("creating guardian: %w", err)
	}

	opts := guardian.ScenarioOptions{
		Verbose:       cfg.Verbose,
		KeepContainer: cfg.KeepContainers,
	}

	fmt.Printf("Running scenario %s...\n", name)

	result, err := g.RunScenario(ctx, name, opts)
	if err != nil {
		return fmt.Errorf("running scenario: %w", err)
	}

	// Print result
	fmt.Printf("\nScenario: %s\n", result.ScenarioID)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Duration: %s\n", result.Duration)
	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
	}
	if len(result.MatchedPatterns) > 0 {
		fmt.Printf("Matched patterns: %v\n", result.MatchedPatterns)
	}
	if len(result.MissingPatterns) > 0 {
		fmt.Printf("Missing patterns: %v\n", result.MissingPatterns)
	}

	if result.Status != reporter.ResultPass {
		return fmt.Errorf("scenario %s failed", name)
	}

	return nil
}

// List shows all available scenarios.
// Parameters: filter=<category>
func (Ci) List(ctx context.Context) error {
	args := getMageArgs()
	params := parseArgs(args)

	cfg := guardian.LoadFromEnv()
	g, err := guardian.New(cfg)
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

	// Group by category
	byCategory := make(map[string][]string)
	for _, s := range scenarios {
		byCategory[s.Category] = append(byCategory[s.Category],
			fmt.Sprintf("  %s - %s", s.ID, s.Description))
	}

	fmt.Printf("Available CI Scenarios (%d total)\n\n", len(scenarios))

	for category, items := range byCategory {
		fmt.Printf("%s (%d):\n", category, len(items))
		for _, item := range items {
			fmt.Println(item)
		}
		fmt.Println()
	}

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
