package guardian

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/reporter"
	"github.com/mrz1836/go-fortress/guardian/runner"
	"github.com/mrz1836/go-fortress/guardian/scenarios"
)

// sequencedRunner returns a different RunResult per call, in order.
// Used to simulate "first attempt is a flaky fast-fail, second attempt passes".
type sequencedRunner struct {
	mu      sync.Mutex
	results []*runner.RunResult
	delays  []time.Duration
	callIdx int32
}

func (s *sequencedRunner) Run(_ context.Context, _ runner.RunOptions) (*runner.RunResult, error) {
	n := int(atomic.AddInt32(&s.callIdx, 1)) - 1
	if n >= len(s.results) {
		n = len(s.results) - 1
	}

	s.mu.Lock()
	delay := time.Duration(0)
	if n < len(s.delays) {
		delay = s.delays[n]
	}
	result := s.results[n]
	s.mu.Unlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	return result, nil
}

func (s *sequencedRunner) CheckAvailable(_ context.Context) error { return nil }

// flakeRetryGuardian wires up a Guardian instance with the given runner and
// tightened retry timings to keep tests fast.
func flakeRetryGuardian(r runner.Runner, retries int) *Guardian {
	cfg := DefaultConfig()
	cfg.ParallelScenarios = 1
	cfg.ScenarioTimeout = 5 * time.Second
	cfg.FlakeRetries = retries
	cfg.FlakeFastFailThreshold = 1 * time.Second
	cfg.FlakeRetryBackoff = 0 // no wait in tests

	return &Guardian{
		config:    cfg,
		runner:    r,
		scenarios: scenarios.NewRegistry(),
		reporters: reporter.NewRegistry(),
	}
}

// scenarioWithPatterns returns a fixture-failure scenario that expects log
// output to contain "needle".
func scenarioWithPatterns() *scenarios.Scenario {
	return &scenarios.Scenario{
		ID:          "FLAKE-001",
		Category:    scenarios.CategoryQuality,
		FixturePath: "fixture-flake",
		Expected: scenarios.ExpectedResult{
			Status:      scenarios.StatusFailure,
			LogPatterns: []string{"needle"},
		},
		Timeout: 5 * time.Second,
	}
}

// TestExecuteScenario_FlakeRetry_PassesAfterRetry verifies that a scenario
// matching the flake signature (exit 1, no output, sub-threshold duration) is
// retried, and the retry result is what's returned.
func TestExecuteScenario_FlakeRetry_PassesAfterRetry(t *testing.T) {
	t.Parallel()

	r := &sequencedRunner{
		results: []*runner.RunResult{
			{ExitCode: 1, Output: ""},                 // attempt 1: flaky fast-fail
			{ExitCode: 1, Output: "found the needle"}, // attempt 2: real run
		},
	}
	g := flakeRetryGuardian(r, 1)

	result, err := g.executeScenario(context.Background(), scenarioWithPatterns(), ScenarioOptions{})
	require.NoError(t, err)
	assert.Equal(t, reporter.ResultPass, result.Status, "scenario should pass on retry")
	assert.Equal(t, int32(2), atomic.LoadInt32(&r.callIdx), "runner should be called twice")
}

// TestExecuteScenario_FlakeRetry_NoRetryWhenPatternMatched verifies that a
// scenario with even one matched pattern is NOT considered flake — real output
// was captured, so the failure is genuine.
func TestExecuteScenario_FlakeRetry_NoRetryWhenPatternMatched(t *testing.T) {
	t.Parallel()

	// Output contains the pattern (so it matches), but ALSO an excluded pattern
	// — making this a real failure (validateResult records the exclude hit in
	// MissingPatterns even though MatchedPatterns has the regular pattern).
	r := &sequencedRunner{
		results: []*runner.RunResult{
			{ExitCode: 1, Output: "found the needle but also FORBIDDEN"},
		},
	}
	g := flakeRetryGuardian(r, 3)

	s := scenarioWithPatterns()
	s.Expected.ExcludePatterns = []string{"FORBIDDEN"}

	result, err := g.executeScenario(context.Background(), s, ScenarioOptions{})
	require.NoError(t, err)
	assert.Equal(t, reporter.ResultFail, result.Status)
	assert.Equal(t, int32(1), atomic.LoadInt32(&r.callIdx), "should not retry when output contains real content")
}

// TestExecuteScenario_FlakeRetry_NoRetryWhenSlow verifies the duration gate:
// even with missing patterns and exit-code match, a slow failure indicates act
// actually ran the workflow, so it's a real bug, not flake.
func TestExecuteScenario_FlakeRetry_NoRetryWhenSlow(t *testing.T) {
	t.Parallel()

	r := &sequencedRunner{
		results: []*runner.RunResult{
			{ExitCode: 1, Output: ""}, // no patterns matched
		},
		delays: []time.Duration{
			1500 * time.Millisecond, // > FlakeFastFailThreshold (1s)
		},
	}
	g := flakeRetryGuardian(r, 3)

	result, err := g.executeScenario(context.Background(), scenarioWithPatterns(), ScenarioOptions{})
	require.NoError(t, err)
	assert.Equal(t, reporter.ResultFail, result.Status)
	assert.Equal(t, int32(1), atomic.LoadInt32(&r.callIdx), "slow failures should not be retried")
}

// TestExecuteScenario_FlakeRetry_DisabledByZero verifies that setting
// FlakeRetries = 0 is a hard kill switch — no retry under any conditions.
func TestExecuteScenario_FlakeRetry_DisabledByZero(t *testing.T) {
	t.Parallel()

	r := &sequencedRunner{
		results: []*runner.RunResult{
			{ExitCode: 1, Output: ""}, // perfect flake signature
		},
	}
	g := flakeRetryGuardian(r, 0)

	result, err := g.executeScenario(context.Background(), scenarioWithPatterns(), ScenarioOptions{})
	require.NoError(t, err)
	assert.Equal(t, reporter.ResultFail, result.Status)
	assert.Equal(t, int32(1), atomic.LoadInt32(&r.callIdx), "no retry when disabled")
}

// TestExecuteScenario_FlakeRetry_PersistentFlake verifies that if every attempt
// is a fast-fail, we stop after FlakeRetries and report the final result.
func TestExecuteScenario_FlakeRetry_PersistentFlake(t *testing.T) {
	t.Parallel()

	r := &sequencedRunner{
		results: []*runner.RunResult{
			{ExitCode: 1, Output: ""},
			{ExitCode: 1, Output: ""},
			{ExitCode: 1, Output: ""},
		},
	}
	g := flakeRetryGuardian(r, 2) // 1 initial + 2 retries = 3 total

	result, err := g.executeScenario(context.Background(), scenarioWithPatterns(), ScenarioOptions{})
	require.NoError(t, err)
	assert.Equal(t, reporter.ResultFail, result.Status)
	assert.Equal(t, int32(3), atomic.LoadInt32(&r.callIdx), "should attempt initial + 2 retries")
}
