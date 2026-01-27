package guardian_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian"
)

// TestGuardian_RunVerify_RunnerStatus tests that RunVerify populates RunnerStatus correctly.
func TestGuardian_RunVerify_RunnerStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create guardian with default config (runner will likely fail if Docker not available)
	cfg := guardian.DefaultConfig()
	cfg.WorkflowsDir = ".github/workflows" // Use actual workflow dir

	g, err := guardian.New(ctx, cfg)
	require.NoError(t, err, "guardian.New should not error even if runner unavailable")

	// Run verify - should succeed regardless of Docker availability
	report, err := g.RunVerify(ctx)
	require.NoError(t, err, "RunVerify should not error even if runner unavailable")

	// RunnerStatus should always be populated
	require.NotNil(t, report.RunnerStatus, "RunnerStatus should be populated")
	assert.GreaterOrEqual(t, report.RunnerStatus.RegisteredCount, 0,
		"RegisteredCount should be set")

	// If runner not available, message should explain why
	if !report.RunnerStatus.Available {
		assert.NotEmpty(t, report.RunnerStatus.UnavailableMsg,
			"UnavailableMsg should explain why runner is unavailable")
		assert.Equal(t, 0, report.RunnerStatus.ExecutedCount,
			"ExecutedCount should be 0 when runner unavailable")
	}
}

// TestGuardian_RunTest_RunnerStatus tests that RunTest populates RunnerStatus correctly.
func TestGuardian_RunTest_RunnerStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cfg := guardian.DefaultConfig()
	cfg.WorkflowsDir = ".github/workflows"

	g, err := guardian.New(ctx, cfg)
	require.NoError(t, err)

	report, err := g.RunTest(ctx)
	require.NoError(t, err, "RunTest should not error even if runner unavailable")

	require.NotNil(t, report.RunnerStatus, "RunnerStatus should be populated")

	if !report.RunnerStatus.Available {
		assert.NotEmpty(t, report.RunnerStatus.UnavailableMsg,
			"UnavailableMsg should explain why runner is unavailable")
	}
}

// TestGuardian_RunStatic_NoRunnerRequired tests that static validation works without runner.
func TestGuardian_RunStatic_NoRunnerRequired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cfg := guardian.DefaultConfig()
	cfg.WorkflowsDir = ".github/workflows"

	g, err := guardian.New(ctx, cfg)
	require.NoError(t, err)

	// RunStatic should always succeed regardless of runner availability
	results, err := g.RunStatic(ctx)
	require.NoError(t, err, "RunStatic should succeed without runner")
	require.NotNil(t, results)
}

// TestGuardian_New_CapturesRunnerError tests that New captures runner initialization errors.
func TestGuardian_New_CapturesRunnerError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use an invalid act path to force runner initialization failure
	cfg := guardian.DefaultConfig()
	cfg.ActPath = "/nonexistent/path/to/act"
	cfg.WorkflowsDir = ".github/workflows"

	g, err := guardian.New(ctx, cfg)

	// New should succeed even if runner init fails
	require.NoError(t, err, "guardian.New should not error even with invalid act path")
	require.NotNil(t, g)

	// When we run verify, the error should be captured in RunnerStatus
	report, err := g.RunVerify(ctx)
	require.NoError(t, err)
	require.NotNil(t, report.RunnerStatus)

	// Runner should be unavailable with an error message
	assert.False(t, report.RunnerStatus.Available,
		"Runner should be unavailable with invalid act path")
	assert.NotEmpty(t, report.RunnerStatus.UnavailableMsg,
		"UnavailableMsg should contain the error from runner init")
}
