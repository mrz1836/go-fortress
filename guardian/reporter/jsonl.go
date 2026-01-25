package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// JSONLReporter writes JSONL (JSON Lines) format output.
type JSONLReporter struct{}

// NewJSONLReporter creates a new JSONL reporter.
func NewJSONLReporter() *JSONLReporter {
	return &JSONLReporter{}
}

// Name returns the reporter identifier.
func (r *JSONLReporter) Name() string {
	return "jsonl"
}

// Write outputs the report in JSONL format.
func (r *JSONLReporter) Write(_ context.Context, report *Report, w io.Writer) error {
	enc := json.NewEncoder(w)

	// Write run start event
	if err := enc.Encode(jsonlEvent{
		Type:      "run_start",
		Timestamp: report.StartTime,
		Version:   report.Version,
		Mode:      string(report.Mode),
	}); err != nil {
		return fmt.Errorf("encoding run_start: %w", err)
	}

	// Write static validation complete event
	if report.StaticResults != nil {
		if err := enc.Encode(jsonlEvent{
			Type:       "static_complete",
			Timestamp:  report.StartTime.Add(report.StaticResults.Duration),
			Findings:   len(report.StaticResults.Findings),
			DurationMs: report.StaticResults.Duration.Milliseconds(),
		}); err != nil {
			return fmt.Errorf("encoding static_complete: %w", err)
		}
	}

	// Write scenario events
	for _, result := range report.ScenarioResults {
		event := jsonlEvent{
			Type:       "scenario",
			ID:         result.ScenarioID,
			Status:     string(result.Status),
			DurationMs: result.Duration.Milliseconds(),
			ExitCode:   result.ExitCode,
		}
		if result.Error != "" {
			event.Error = result.Error
		}
		if err := enc.Encode(event); err != nil {
			return fmt.Errorf("encoding scenario %s: %w", result.ScenarioID, err)
		}
	}

	// Write run end event
	if err := enc.Encode(jsonlEvent{
		Type:      "run_end",
		Timestamp: report.EndTime,
		Passed:    report.Summary.PassedScenarios,
		Failed:    report.Summary.FailedScenarios,
		Skipped:   report.Summary.SkippedScenarios,
	}); err != nil {
		return fmt.Errorf("encoding run_end: %w", err)
	}

	return nil
}

// WriteFile outputs the report to a file in JSONL format.
func (r *JSONLReporter) WriteFile(ctx context.Context, report *Report, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return r.Write(ctx, report, f)
}

// jsonlEvent represents a single line in the JSONL output.
type jsonlEvent struct {
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
	Version    string    `json:"version,omitempty"`
	Mode       string    `json:"mode,omitempty"`
	ID         string    `json:"id,omitempty"`
	Status     string    `json:"status,omitempty"`
	Findings   int       `json:"findings,omitempty"`
	DurationMs int64     `json:"duration_ms,omitempty"`
	ExitCode   int       `json:"exit_code,omitempty"`
	Error      string    `json:"error,omitempty"`
	Passed     int       `json:"passed,omitempty"`
	Failed     int       `json:"failed,omitempty"`
	Skipped    int       `json:"skipped,omitempty"`
}
