package reporter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/reporter"
	"github.com/mrz1836/go-fortress/guardian/validator"
)

// TestJSONLReporter_Name tests the reporter name.
func TestJSONLReporter_Name(t *testing.T) {
	t.Parallel()

	r := reporter.NewJSONLReporter()
	assert.Equal(t, "jsonl", r.Name())
}

// TestJSONLReporter_Write tests writing JSONL output.
func TestJSONLReporter_Write(t *testing.T) {
	t.Parallel()

	now := time.Now()
	report := &reporter.Report{
		Version:   "1.0.0",
		StartTime: now,
		EndTime:   now.Add(5 * time.Second),
		Duration:  5 * time.Second,
		Mode:      reporter.ModeVerify,
		StaticResults: &reporter.StaticResults{
			Findings:      []validator.Finding{{Message: "test"}},
			ValidatorsRun: []string{"actionlint"},
			Duration:      100 * time.Millisecond,
		},
		ScenarioResults: []reporter.ScenarioResult{
			{
				ScenarioID: "TEST-001",
				Status:     reporter.ResultPass,
				Duration:   1 * time.Second,
				ExitCode:   0,
			},
			{
				ScenarioID: "TEST-002",
				Status:     reporter.ResultFail,
				Duration:   2 * time.Second,
				ExitCode:   1,
				Error:      "expected success",
			},
		},
		Summary: reporter.ReportSummary{
			PassedScenarios:  1,
			FailedScenarios:  1,
			SkippedScenarios: 0,
		},
	}

	r := reporter.NewJSONLReporter()
	var buf bytes.Buffer
	err := r.Write(context.Background(), report, &buf)
	require.NoError(t, err)

	// Parse the JSONL output
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 5) // run_start, static_complete, 2 scenarios, run_end

	// Verify run_start event
	var runStart map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &runStart))
	assert.Equal(t, "run_start", runStart["type"])
	assert.Equal(t, "1.0.0", runStart["version"])
	assert.Equal(t, "verify", runStart["mode"])

	// Verify static_complete event
	var staticComplete map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &staticComplete))
	assert.Equal(t, "static_complete", staticComplete["type"])
	assert.InDelta(t, float64(1), staticComplete["findings"], 0.0001)

	// Verify scenario events
	var scenario1 map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(lines[2]), &scenario1))
	assert.Equal(t, "scenario", scenario1["type"])
	assert.Equal(t, "TEST-001", scenario1["id"])
	assert.Equal(t, "pass", scenario1["status"])

	var scenario2 map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(lines[3]), &scenario2))
	assert.Equal(t, "scenario", scenario2["type"])
	assert.Equal(t, "TEST-002", scenario2["id"])
	assert.Equal(t, "fail", scenario2["status"])
	assert.Equal(t, "expected success", scenario2["error"])

	// Verify run_end event
	var runEnd map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(lines[4]), &runEnd))
	assert.Equal(t, "run_end", runEnd["type"])
	assert.InDelta(t, float64(1), runEnd["passed"], 0.0001)
	assert.InDelta(t, float64(1), runEnd["failed"], 0.0001)
}

// TestJSONLReporter_Write_NoStaticResults tests writing without static results.
func TestJSONLReporter_Write_NoStaticResults(t *testing.T) {
	t.Parallel()

	now := time.Now()
	report := &reporter.Report{
		Version:         "1.0.0",
		StartTime:       now,
		EndTime:         now.Add(1 * time.Second),
		Mode:            reporter.ModeTest,
		StaticResults:   nil, // No static results
		ScenarioResults: []reporter.ScenarioResult{},
		Summary:         reporter.ReportSummary{},
	}

	r := reporter.NewJSONLReporter()
	var buf bytes.Buffer
	err := r.Write(context.Background(), report, &buf)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 2) // run_start, run_end (no static_complete)
}

// TestJSONLReporter_WriteFile tests writing to a file.
func TestJSONLReporter_WriteFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "results.jsonl")

	now := time.Now()
	report := &reporter.Report{
		Version:   "1.0.0",
		StartTime: now,
		EndTime:   now,
		Mode:      reporter.ModeVerify,
		Summary:   reporter.ReportSummary{},
	}

	r := reporter.NewJSONLReporter()
	err := r.WriteFile(context.Background(), report, outPath)
	require.NoError(t, err)

	// Verify file was created
	data, err := os.ReadFile(outPath) //nolint:gosec // test file path from t.TempDir()
	require.NoError(t, err)
	assert.Contains(t, string(data), "run_start")
}

// TestJSONLReporter_WriteFile_Error tests write error handling.
func TestJSONLReporter_WriteFile_Error(t *testing.T) {
	t.Parallel()

	// Try to write to a non-existent directory
	report := &reporter.Report{
		Version: "1.0.0",
	}

	r := reporter.NewJSONLReporter()
	err := r.WriteFile(context.Background(), report, "/nonexistent/path/results.jsonl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating file")
}
