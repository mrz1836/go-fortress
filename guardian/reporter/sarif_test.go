package reporter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/reporter"
	"github.com/mrz1836/go-fortress/guardian/validator"
)

// TestSARIFReporter_Name tests the reporter name.
func TestSARIFReporter_Name(t *testing.T) {
	t.Parallel()

	r := reporter.NewSARIFReporter()
	assert.Equal(t, "sarif", r.Name())
}

// TestSARIFReporter_Write tests writing SARIF output.
func TestSARIFReporter_Write(t *testing.T) {
	t.Parallel()

	now := time.Now()
	report := &reporter.Report{
		Version:   "1.0.0",
		StartTime: now,
		EndTime:   now.Add(5 * time.Second),
		Mode:      reporter.ModeVerify,
		StaticResults: &reporter.StaticResults{
			Findings: []validator.Finding{
				{
					RuleID:      "policy/sha-pinned-actions",
					Severity:    validator.SeverityError,
					Message:     "action is not pinned",
					File:        ".github/workflows/test.yml",
					Line:        10,
					Column:      5,
					EndLine:     10,
					EndColumn:   30,
					Suggestion:  "Pin to a full SHA",
					Fingerprint: "abc123",
				},
				{
					RuleID:   "policy/explicit-permissions",
					Severity: validator.SeverityWarning,
					Message:  "missing permissions",
					File:     ".github/workflows/test.yml",
					Line:     1,
				},
			},
			ValidatorsRun: []string{"actionlint", "policy"},
		},
		Summary: reporter.ReportSummary{
			TotalFindings: 2,
		},
	}

	r := reporter.NewSARIFReporter()
	var buf bytes.Buffer
	err := r.Write(context.Background(), report, &buf)
	require.NoError(t, err)

	// Verify SARIF structure
	var sarif map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &sarif))

	assert.Equal(t, "2.1.0", sarif["version"])
	// Schema URL may vary by library version, just verify it exists
	assert.NotEmpty(t, sarif["$schema"])

	runs := sarif["runs"].([]interface{})
	require.Len(t, runs, 1)

	run := runs[0].(map[string]interface{})
	tool := run["tool"].(map[string]interface{})
	driver := tool["driver"].(map[string]interface{})

	assert.Equal(t, "Fortress Guardian", driver["name"])
	assert.Equal(t, "1.0.0", driver["version"])

	// Check results
	results := run["results"].([]interface{})
	require.Len(t, results, 2)

	// First result
	result1 := results[0].(map[string]interface{})
	assert.Equal(t, "policy/sha-pinned-actions", result1["ruleId"])
	assert.Equal(t, "error", result1["level"])

	// Check location
	locations := result1["locations"].([]interface{})
	require.Len(t, locations, 1)
	loc := locations[0].(map[string]interface{})
	physLoc := loc["physicalLocation"].(map[string]interface{})
	artifactLoc := physLoc["artifactLocation"].(map[string]interface{})
	assert.Equal(t, ".github/workflows/test.yml", artifactLoc["uri"])

	region := physLoc["region"].(map[string]interface{})
	assert.InDelta(t, float64(10), region["startLine"], 0.0001)
	assert.InDelta(t, float64(5), region["startColumn"], 0.0001)
	assert.InDelta(t, float64(10), region["endLine"], 0.0001)
	assert.InDelta(t, float64(30), region["endColumn"], 0.0001)

	// Check fingerprint
	partialFingerprints := result1["partialFingerprints"].(map[string]interface{})
	assert.Equal(t, "abc123", partialFingerprints["guardian/fingerprint"])

	// Second result
	result2 := results[1].(map[string]interface{})
	assert.Equal(t, "policy/explicit-permissions", result2["ruleId"])
	assert.Equal(t, "warning", result2["level"])
}

// TestSARIFReporter_Write_NoFindings tests writing with no findings.
func TestSARIFReporter_Write_NoFindings(t *testing.T) {
	t.Parallel()

	now := time.Now()
	report := &reporter.Report{
		Version:   "1.0.0",
		StartTime: now,
		EndTime:   now,
		Mode:      reporter.ModeVerify,
		StaticResults: &reporter.StaticResults{
			Findings:      []validator.Finding{},
			ValidatorsRun: []string{"actionlint"},
		},
	}

	r := reporter.NewSARIFReporter()
	var buf bytes.Buffer
	err := r.Write(context.Background(), report, &buf)
	require.NoError(t, err)

	// Verify valid SARIF even with no results
	var sarif map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &sarif))
	assert.Equal(t, "2.1.0", sarif["version"])
}

// TestSARIFReporter_Write_NoStaticResults tests writing without static results.
func TestSARIFReporter_Write_NoStaticResults(t *testing.T) {
	t.Parallel()

	now := time.Now()
	report := &reporter.Report{
		Version:       "1.0.0",
		StartTime:     now,
		EndTime:       now,
		Mode:          reporter.ModeTest,
		StaticResults: nil,
	}

	r := reporter.NewSARIFReporter()
	var buf bytes.Buffer
	err := r.Write(context.Background(), report, &buf)
	require.NoError(t, err)

	// Should still produce valid SARIF
	var sarif map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &sarif))
	assert.Equal(t, "2.1.0", sarif["version"])
}

// TestSARIFReporter_Write_SeverityMapping tests severity to level mapping.
func TestSARIFReporter_Write_SeverityMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity validator.Severity
		level    string
	}{
		{validator.SeverityError, "error"},
		{validator.SeverityWarning, "warning"},
		{validator.SeverityNote, "note"},
		{validator.SeverityInfo, "none"},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			t.Parallel()

			now := time.Now()
			report := &reporter.Report{
				Version:   "1.0.0",
				StartTime: now,
				EndTime:   now,
				Mode:      reporter.ModeVerify,
				StaticResults: &reporter.StaticResults{
					Findings: []validator.Finding{
						{
							RuleID:   "test-rule",
							Severity: tt.severity,
							Message:  "test message",
							File:     "test.yml",
							Line:     1,
						},
					},
				},
			}

			r := reporter.NewSARIFReporter()
			var buf bytes.Buffer
			err := r.Write(context.Background(), report, &buf)
			require.NoError(t, err)

			var sarif map[string]interface{}
			require.NoError(t, json.Unmarshal(buf.Bytes(), &sarif))

			runs := sarif["runs"].([]interface{})
			run := runs[0].(map[string]interface{})
			results := run["results"].([]interface{})
			result := results[0].(map[string]interface{})

			assert.Equal(t, tt.level, result["level"])
		})
	}
}

// TestSARIFReporter_WriteFile tests writing to a file.
func TestSARIFReporter_WriteFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "results.sarif")

	now := time.Now()
	report := &reporter.Report{
		Version:   "1.0.0",
		StartTime: now,
		EndTime:   now,
		Mode:      reporter.ModeVerify,
	}

	r := reporter.NewSARIFReporter()
	err := r.WriteFile(context.Background(), report, outPath)
	require.NoError(t, err)

	// Verify file was created with valid SARIF
	data, err := os.ReadFile(outPath) //nolint:gosec // test file path from t.TempDir()
	require.NoError(t, err)

	var sarif map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &sarif))
	assert.Equal(t, "2.1.0", sarif["version"])
}

// TestSARIFReporter_WriteFile_Error tests write error handling.
func TestSARIFReporter_WriteFile_Error(t *testing.T) {
	t.Parallel()

	report := &reporter.Report{
		Version: "1.0.0",
	}

	r := reporter.NewSARIFReporter()
	err := r.WriteFile(context.Background(), report, "/nonexistent/path/results.sarif")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating file")
}

// TestSARIFReporter_Write_DuplicateRules tests that duplicate rules are only added once.
func TestSARIFReporter_Write_DuplicateRules(t *testing.T) {
	t.Parallel()

	now := time.Now()
	report := &reporter.Report{
		Version:   "1.0.0",
		StartTime: now,
		EndTime:   now,
		Mode:      reporter.ModeVerify,
		StaticResults: &reporter.StaticResults{
			Findings: []validator.Finding{
				{
					RuleID:   "same-rule",
					Severity: validator.SeverityError,
					Message:  "first occurrence",
					File:     "file1.yml",
					Line:     1,
				},
				{
					RuleID:   "same-rule",
					Severity: validator.SeverityError,
					Message:  "second occurrence",
					File:     "file2.yml",
					Line:     5,
				},
			},
		},
	}

	r := reporter.NewSARIFReporter()
	var buf bytes.Buffer
	err := r.Write(context.Background(), report, &buf)
	require.NoError(t, err)

	var sarif map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &sarif))

	runs := sarif["runs"].([]interface{})
	run := runs[0].(map[string]interface{})

	// Should have 2 results but only 1 rule
	results := run["results"].([]interface{})
	assert.Len(t, results, 2)

	tool := run["tool"].(map[string]interface{})
	driver := tool["driver"].(map[string]interface{})
	rules := driver["rules"].([]interface{})
	assert.Len(t, rules, 1)
}
