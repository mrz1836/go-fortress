package validator

// Finding represents a single issue detected during validation.
type Finding struct {
	// RuleID identifies the rule that triggered this finding.
	// Format: "guardian/<rule-name>" or "actionlint/<kind>".
	RuleID string `json:"rule_id"`

	// Severity indicates the importance of this finding.
	Severity Severity `json:"severity"`

	// Message describes the issue found.
	Message string `json:"message"`

	// File is the path to the file containing the issue.
	// Path is relative to repository root.
	File string `json:"file"`

	// Line is the 1-based line number where the issue occurs.
	Line int `json:"line"`

	// Column is the 1-based column number where the issue starts.
	// May be 0 if not applicable.
	Column int `json:"column,omitempty"`

	// EndLine is the 1-based line where the issue ends.
	// Optional; same as Line if not specified.
	EndLine int `json:"end_line,omitempty"`

	// EndColumn is the 1-based column where the issue ends.
	EndColumn int `json:"end_column,omitempty"`

	// Source identifies what detected this finding.
	Source FindingSource `json:"source"`

	// Suggestion is an optional fix recommendation.
	Suggestion string `json:"suggestion,omitempty"`

	// Fingerprint is a hash for deduplication.
	Fingerprint string `json:"fingerprint,omitempty"`
}

// Severity levels for findings.
type Severity string

// Severity constants define finding severity levels.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityNote    Severity = "note"
	SeverityInfo    Severity = "info"
)

// FindingSource identifies what tool generated the finding.
type FindingSource string

// FindingSource constants define available finding sources.
const (
	SourceActionlint  FindingSource = "actionlint"
	SourcePolicy      FindingSource = "policy"
	SourceSchema      FindingSource = "schema"
	SourceDeprecation FindingSource = "deprecation"
	SourceEnv         FindingSource = "env"
	SourceValidator   FindingSource = "validator"
)
