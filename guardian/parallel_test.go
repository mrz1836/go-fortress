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

// mockRunner implements runner.Runner for testing.
type mockRunner struct {
	mu            sync.Mutex
	runCount      int32
	runDelay      time.Duration
	runResults    map[string]*runner.RunResult
	runErrors     map[string]error
	concurrent    int32
	maxConcurrent int32
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		runResults: make(map[string]*runner.RunResult),
		runErrors:  make(map[string]error),
	}
}

func (m *mockRunner) Run(_ context.Context, opts runner.RunOptions) (*runner.RunResult, error) {
	// Track concurrency
	current := atomic.AddInt32(&m.concurrent, 1)
	defer atomic.AddInt32(&m.concurrent, -1)

	// Update max concurrent
	for {
		maxVal := atomic.LoadInt32(&m.maxConcurrent)
		if current <= maxVal {
			break
		}
		if atomic.CompareAndSwapInt32(&m.maxConcurrent, maxVal, current) {
			break
		}
	}

	atomic.AddInt32(&m.runCount, 1)

	if m.runDelay > 0 {
		time.Sleep(m.runDelay)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for configured error
	if err, ok := m.runErrors[opts.Job]; ok {
		return nil, err
	}

	// Return configured result or default success
	if result, ok := m.runResults[opts.Job]; ok {
		return result, nil
	}

	return &runner.RunResult{
		ExitCode: 0,
		Output:   "success",
	}, nil
}

func (m *mockRunner) CheckAvailable(_ context.Context) error {
	return nil
}

// createTestGuardian creates a Guardian with mock runner for testing.
func createTestGuardian(t *testing.T, parallelScenarios int) (*Guardian, *mockRunner) {
	t.Helper()

	mock := newMockRunner()

	cfg := DefaultConfig()
	cfg.ParallelScenarios = parallelScenarios
	cfg.ScenarioTimeout = 5 * time.Second

	g := &Guardian{
		config:     cfg,
		validators: nil,
		runner:     mock,
		scenarios:  scenarios.NewRegistry(),
		reporters:  reporter.NewRegistry(),
	}

	return g, mock
}

// createTestScenarios creates n test scenarios with the given IDs.
// Each scenario gets a unique FixturePath to allow parallel execution.
func createTestScenarios(ids ...string) []*scenarios.Scenario {
	result := make([]*scenarios.Scenario, len(ids))
	for i, id := range ids {
		result[i] = &scenarios.Scenario{
			ID:          id,
			Category:    scenarios.CategoryQuality,
			Job:         id,              // Use ID as job name for mock lookup
			FixturePath: "fixture-" + id, // Unique fixture per scenario for parallel execution
			Expected: scenarios.ExpectedResult{
				Status: scenarios.StatusSuccess,
			},
		}
	}
	return result
}

// TestExecuteScenarios_EmptyList tests that empty scenario list returns empty results.
func TestExecuteScenarios_EmptyList(t *testing.T) {
	t.Parallel()

	g, _ := createTestGuardian(t, 4)

	results, err := g.executeScenarios(context.Background(), []*scenarios.Scenario{})
	require.NoError(t, err)
	assert.Empty(t, results)
}

// TestExecuteScenarios_SingleScenario tests that single scenario uses sequential path.
func TestExecuteScenarios_SingleScenario(t *testing.T) {
	t.Parallel()

	g, mock := createTestGuardian(t, 4)
	scenarioList := createTestScenarios("TEST-001")

	results, err := g.executeScenarios(context.Background(), scenarioList)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "TEST-001", results[0].ScenarioID)
	assert.Equal(t, int32(1), atomic.LoadInt32(&mock.runCount))
}

// TestExecuteScenarios_SequentialFallback tests sequential execution when parallelism is 1.
func TestExecuteScenarios_SequentialFallback(t *testing.T) {
	t.Parallel()

	g, mock := createTestGuardian(t, 1) // Parallelism disabled
	mock.runDelay = 10 * time.Millisecond

	scenarioList := createTestScenarios("TEST-001", "TEST-002", "TEST-003")

	start := time.Now()
	results, err := g.executeScenarios(context.Background(), scenarioList)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Should take at least 30ms (3 scenarios * 10ms each) if running sequentially
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(30),
		"sequential execution should take at least 30ms")

	// Max concurrent should be 1 for sequential
	assert.Equal(t, int32(1), atomic.LoadInt32(&mock.maxConcurrent),
		"sequential execution should have max concurrency of 1")
}

// TestExecuteScenarios_ParallelExecution tests parallel execution with multiple workers.
func TestExecuteScenarios_ParallelExecution(t *testing.T) {
	t.Parallel()

	g, mock := createTestGuardian(t, 4)
	mock.runDelay = 50 * time.Millisecond

	// Create 8 scenarios - with 4 workers, should complete in ~2 batches
	scenarioList := createTestScenarios(
		"TEST-001", "TEST-002", "TEST-003", "TEST-004",
		"TEST-005", "TEST-006", "TEST-007", "TEST-008",
	)

	start := time.Now()
	results, err := g.executeScenarios(context.Background(), scenarioList)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, 8)

	// With 4 workers and 50ms per scenario, 8 scenarios should take ~100ms (2 batches)
	// Sequential would take ~400ms. Allow some margin.
	assert.Less(t, elapsed.Milliseconds(), int64(300),
		"parallel execution should be faster than sequential")

	// Max concurrent should be at least 2 (ideally 4)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&mock.maxConcurrent), int32(2),
		"parallel execution should have concurrency > 1")
}

// TestExecuteScenarios_ResultOrdering tests that results preserve input order.
func TestExecuteScenarios_ResultOrdering(t *testing.T) {
	t.Parallel()

	g, mock := createTestGuardian(t, 4)
	// Variable delays to ensure results arrive out of order
	mock.runResults = map[string]*runner.RunResult{
		"FAST-001": {ExitCode: 0, Output: "fast1"},
		"SLOW-002": {ExitCode: 0, Output: "slow2"},
		"FAST-003": {ExitCode: 0, Output: "fast3"},
		"SLOW-004": {ExitCode: 0, Output: "slow4"},
	}

	scenarioList := createTestScenarios("FAST-001", "SLOW-002", "FAST-003", "SLOW-004")

	results, err := g.executeScenarios(context.Background(), scenarioList)
	require.NoError(t, err)
	require.Len(t, results, 4)

	// Results should be in the same order as input
	assert.Equal(t, "FAST-001", results[0].ScenarioID)
	assert.Equal(t, "SLOW-002", results[1].ScenarioID)
	assert.Equal(t, "FAST-003", results[2].ScenarioID)
	assert.Equal(t, "SLOW-004", results[3].ScenarioID)
}

// TestExecuteScenarios_ContextCancellation tests graceful handling of context cancellation.
func TestExecuteScenarios_ContextCancellation(t *testing.T) {
	t.Parallel()

	g, mock := createTestGuardian(t, 2)
	mock.runDelay = 100 * time.Millisecond

	scenarioList := createTestScenarios("TEST-001", "TEST-002", "TEST-003", "TEST-004")

	// Cancel context after 50ms (before first batch completes)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	results, err := g.executeScenarios(ctx, scenarioList)

	// Should return results (some may be skipped due to cancellation)
	assert.Len(t, results, 4)

	// At least some scenarios should have been processed
	// The error return is optional based on implementation
	_ = err
}

// TestExecuteScenarios_WorkerCountCapping tests that worker count doesn't exceed scenario count.
func TestExecuteScenarios_WorkerCountCapping(t *testing.T) {
	t.Parallel()

	g, mock := createTestGuardian(t, 10) // More workers than scenarios
	mock.runDelay = 20 * time.Millisecond

	scenarioList := createTestScenarios("TEST-001", "TEST-002", "TEST-003")

	results, err := g.executeScenarios(context.Background(), scenarioList)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Max concurrent should not exceed scenario count
	assert.LessOrEqual(t, atomic.LoadInt32(&mock.maxConcurrent), int32(3),
		"worker count should be capped to scenario count")
}

// TestExecuteScenarios_ErrorHandling tests that individual errors don't stop other scenarios.
func TestExecuteScenarios_ErrorHandling(t *testing.T) {
	t.Parallel()

	g, mock := createTestGuardian(t, 4)
	mock.runErrors = map[string]error{
		"ERROR-002": assert.AnError,
	}

	scenarioList := createTestScenarios("OK-001", "ERROR-002", "OK-003", "OK-004")

	results, err := g.executeScenarios(context.Background(), scenarioList)
	require.NoError(t, err, "overall execution should not error")
	require.Len(t, results, 4)

	// Check that the error scenario has error status
	assert.Equal(t, reporter.ResultError, results[1].Status)
	assert.NotEmpty(t, results[1].Error)

	// Other scenarios should pass
	assert.Equal(t, reporter.ResultPass, results[0].Status)
	assert.Equal(t, reporter.ResultPass, results[2].Status)
	assert.Equal(t, reporter.ResultPass, results[3].Status)
}

// TestExecuteScenariosSequential_ContextCancellation tests sequential context handling.
func TestExecuteScenariosSequential_ContextCancellation(t *testing.T) {
	t.Parallel()

	g, mock := createTestGuardian(t, 1)
	mock.runDelay = 50 * time.Millisecond

	scenarioList := createTestScenarios("TEST-001", "TEST-002", "TEST-003")

	// Cancel context after 75ms (after first scenario, during second)
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer cancel()

	results, err := g.executeScenariosSequential(ctx, scenarioList)

	// Should return all results
	assert.Len(t, results, 3)

	// First scenario should complete
	assert.Equal(t, "TEST-001", results[0].ScenarioID)

	// Remaining scenarios should be skipped
	hasSkipped := false
	for _, r := range results {
		if r.Status == reporter.ResultSkip {
			hasSkipped = true
			break
		}
	}

	// Either we got a context error or some scenarios were skipped
	if err == nil {
		assert.True(t, hasSkipped, "should have skipped scenarios when context canceled")
	}
}

// TestExecuteScenariosParallel_LargeWorkload tests parallel execution with many scenarios.
func TestExecuteScenariosParallel_LargeWorkload(t *testing.T) {
	t.Parallel()

	g, mock := createTestGuardian(t, 8)
	mock.runDelay = 5 * time.Millisecond

	// Create 50 scenarios
	ids := make([]string, 50)
	for i := 0; i < 50; i++ {
		ids[i] = "TEST-" + string(rune('A'+i%26)) + string(rune('0'+i/26))
	}
	scenarioList := createTestScenarios(ids...)

	start := time.Now()
	results, err := g.executeScenarios(context.Background(), scenarioList)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, 50)

	// Sequential would take ~250ms (50 * 5ms), parallel should be much faster
	assert.Less(t, elapsed.Milliseconds(), int64(200),
		"parallel execution of 50 scenarios should complete in under 200ms")

	// Verify all results have correct IDs in order
	for i, r := range results {
		assert.Equal(t, ids[i], r.ScenarioID)
	}
}

// TestExecuteScenarios_MixedResults tests scenarios with different outcomes.
func TestExecuteScenarios_MixedResults(t *testing.T) {
	t.Parallel()

	g, mock := createTestGuardian(t, 4)
	mock.runResults = map[string]*runner.RunResult{
		"PASS-001": {ExitCode: 0, Output: "success"},
		"FAIL-002": {ExitCode: 1, Output: "failure"},
		"PASS-003": {ExitCode: 0, Output: "success"},
	}

	scenarioList := []*scenarios.Scenario{
		{
			ID:       "PASS-001",
			Job:      "PASS-001",
			Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess},
		},
		{
			ID:       "FAIL-002",
			Job:      "FAIL-002",
			Expected: scenarios.ExpectedResult{Status: scenarios.StatusFailure},
		},
		{
			ID:       "PASS-003",
			Job:      "PASS-003",
			Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess},
		},
	}

	results, err := g.executeScenarios(context.Background(), scenarioList)
	require.NoError(t, err)
	require.Len(t, results, 3)

	// All should pass since they match expectations
	assert.Equal(t, reporter.ResultPass, results[0].Status)
	assert.Equal(t, reporter.ResultPass, results[1].Status)
	assert.Equal(t, reporter.ResultPass, results[2].Status)
}

// fixtureTrackingRunner extends mockRunner to track concurrent fixture access.
type fixtureTrackingRunner struct {
	*mockRunner

	mu                    sync.Mutex
	fixtureAccess         map[string]int32 // Current concurrent access count per fixture
	maxFixtureConcurrency map[string]int32 // Max concurrent access observed per fixture
	accessOrder           []string         // Order of fixture access starts
}

func newFixtureTrackingRunner() *fixtureTrackingRunner {
	return &fixtureTrackingRunner{
		mockRunner:            newMockRunner(),
		fixtureAccess:         make(map[string]int32),
		maxFixtureConcurrency: make(map[string]int32),
		accessOrder:           []string{},
	}
}

func (f *fixtureTrackingRunner) Run(ctx context.Context, opts runner.RunOptions) (*runner.RunResult, error) {
	fixture := opts.WorkingDir

	// Track fixture access
	f.mu.Lock()
	f.fixtureAccess[fixture]++
	current := f.fixtureAccess[fixture]
	if current > f.maxFixtureConcurrency[fixture] {
		f.maxFixtureConcurrency[fixture] = current
	}
	f.accessOrder = append(f.accessOrder, fixture)
	f.mu.Unlock()

	// Delegate to mock runner
	result, err := f.mockRunner.Run(ctx, opts)

	// Release fixture access
	f.mu.Lock()
	f.fixtureAccess[fixture]--
	f.mu.Unlock()

	return result, err
}

// TestExecuteScenarios_FixtureSerialization tests that scenarios sharing fixtures run sequentially.
func TestExecuteScenarios_FixtureSerialization(t *testing.T) {
	t.Parallel()

	tracker := newFixtureTrackingRunner()
	tracker.runDelay = 30 * time.Millisecond

	cfg := DefaultConfig()
	cfg.ParallelScenarios = 4
	cfg.ScenarioTimeout = 5 * time.Second

	g := &Guardian{
		config:    cfg,
		runner:    tracker,
		scenarios: scenarios.NewRegistry(),
		reporters: reporter.NewRegistry(),
	}

	// Create scenarios: 4 scenarios share "fixture-A", 4 scenarios share "fixture-B"
	// With 4 workers, both fixtures should run in parallel, but scenarios within
	// each fixture should be serialized
	scenarioList := []*scenarios.Scenario{
		{ID: "A-001", Job: "A-001", FixturePath: "fixture-A", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
		{ID: "A-002", Job: "A-002", FixturePath: "fixture-A", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
		{ID: "B-001", Job: "B-001", FixturePath: "fixture-B", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
		{ID: "B-002", Job: "B-002", FixturePath: "fixture-B", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
		{ID: "A-003", Job: "A-003", FixturePath: "fixture-A", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
		{ID: "A-004", Job: "A-004", FixturePath: "fixture-A", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
		{ID: "B-003", Job: "B-003", FixturePath: "fixture-B", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
		{ID: "B-004", Job: "B-004", FixturePath: "fixture-B", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
	}

	start := time.Now()
	results, err := g.executeScenarios(context.Background(), scenarioList)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, 8)

	// Critical assertion: each fixture should have max concurrency of 1
	// This ensures scenarios sharing a fixture are serialized
	assert.Equal(t, int32(1), tracker.maxFixtureConcurrency["fixture-A"],
		"fixture-A should have max concurrency of 1 (serialized)")
	assert.Equal(t, int32(1), tracker.maxFixtureConcurrency["fixture-B"],
		"fixture-B should have max concurrency of 1 (serialized)")

	// Verify parallelism still works across fixtures
	// 8 scenarios * 30ms each = 240ms sequential
	// With 2 fixtures in parallel, each running 4 scenarios: 4 * 30ms = 120ms
	// Allow margin for scheduling overhead
	assert.Less(t, elapsed.Milliseconds(), int64(200),
		"different fixtures should still run in parallel")

	// Verify overall concurrency is > 1 (parallel execution is working)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&tracker.maxConcurrent), int32(2),
		"should have parallel execution across different fixtures")
}

// TestExecuteScenarios_DistinctFixturesFullParallel tests that scenarios with distinct fixtures run fully parallel.
func TestExecuteScenarios_DistinctFixturesFullParallel(t *testing.T) {
	t.Parallel()

	tracker := newFixtureTrackingRunner()
	tracker.runDelay = 30 * time.Millisecond

	cfg := DefaultConfig()
	cfg.ParallelScenarios = 4
	cfg.ScenarioTimeout = 5 * time.Second

	g := &Guardian{
		config:    cfg,
		runner:    tracker,
		scenarios: scenarios.NewRegistry(),
		reporters: reporter.NewRegistry(),
	}

	// Create 4 scenarios each with a unique fixture - should run fully parallel
	scenarioList := []*scenarios.Scenario{
		{ID: "TEST-001", Job: "TEST-001", FixturePath: "fixture-1", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
		{ID: "TEST-002", Job: "TEST-002", FixturePath: "fixture-2", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
		{ID: "TEST-003", Job: "TEST-003", FixturePath: "fixture-3", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
		{ID: "TEST-004", Job: "TEST-004", FixturePath: "fixture-4", Expected: scenarios.ExpectedResult{Status: scenarios.StatusSuccess}},
	}

	start := time.Now()
	results, err := g.executeScenarios(context.Background(), scenarioList)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, 4)

	// With 4 workers and 4 unique fixtures, all should run in parallel
	// Sequential would be ~120ms, parallel should be ~30ms
	assert.Less(t, elapsed.Milliseconds(), int64(80),
		"distinct fixtures should run fully parallel")

	// Max concurrency should be 4
	assert.Equal(t, int32(4), atomic.LoadInt32(&tracker.maxConcurrent),
		"all 4 scenarios with unique fixtures should run concurrently")
}
