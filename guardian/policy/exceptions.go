package policy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

// Sentinel errors for exception validation.
var (
	errPolicyRequired    = errors.New("policy is required")
	errReasonRequired    = errors.New("reason is required")
	errCreatedAtRequired = errors.New("created_at is required")
)

// Exception allows bypassing a policy for specific files or patterns.
type Exception struct {
	// Policy is the ID of the policy to bypass.
	Policy string `yaml:"policy" json:"policy"`

	// Path is a glob pattern matching files to exempt.
	// Example: ".github/workflows/test.yml"
	Path string `yaml:"path" json:"path"`

	// Reason documents why this exception exists.
	Reason string `yaml:"reason" json:"reason"`

	// Expires is when this exception should be reviewed.
	// Optional; exceptions without expiration are permanent.
	Expires *time.Time `yaml:"expires,omitempty" json:"expires,omitempty"`

	// ApprovedBy records who approved this exception.
	ApprovedBy string `yaml:"approved_by,omitempty" json:"approved_by,omitempty"`

	// CreatedAt is when the exception was created.
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`
}

// ExceptionConfig is the structure of .github/guardian.yaml.
type ExceptionConfig struct {
	// Exceptions lists all configured policy exemptions.
	Exceptions []Exception `yaml:"exceptions" json:"exceptions"`
}

// LoadExceptionConfig loads exceptions from a YAML file.
func LoadExceptionConfig(path string) (*ExceptionConfig, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is from trusted config
	if err != nil {
		if os.IsNotExist(err) {
			return &ExceptionConfig{}, nil
		}
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var config ExceptionConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	return &config, nil
}

// Matches checks if an exception applies to a finding.
func (e *Exception) Matches(finding *validator.Finding) bool {
	// Check policy ID matches
	policyID := finding.RuleID
	// Strip "policy/" prefix if present
	if len(policyID) > 7 && policyID[:7] == "policy/" {
		policyID = policyID[7:]
	}
	if e.Policy != policyID {
		return false
	}

	// Check if exception has expired
	if e.Expires != nil && time.Now().After(*e.Expires) {
		return false
	}

	// Check path matches
	if e.Path != "" {
		matched, err := filepath.Match(e.Path, finding.File)
		if err != nil || !matched {
			// Try with basename
			matched, _ = filepath.Match(e.Path, filepath.Base(finding.File))
			if !matched {
				return false
			}
		}
	}

	return true
}

// IsExpired checks if the exception has expired.
func (e *Exception) IsExpired() bool {
	if e.Expires == nil {
		return false
	}
	return time.Now().After(*e.Expires)
}

// Validate checks if the exception configuration is valid.
func (e *Exception) Validate() error {
	if e.Policy == "" {
		return errPolicyRequired
	}
	if e.Reason == "" {
		return errReasonRequired
	}
	if e.CreatedAt.IsZero() {
		return errCreatedAtRequired
	}
	return nil
}

// AuditEntry represents an exception usage for auditing.
type AuditEntry struct {
	Timestamp time.Time
	Policy    string
	Path      string
	Reason    string
	ExpiresAt *time.Time
}

// AuditLog tracks exception usage.
type AuditLog struct {
	entries []AuditEntry
}

// NewAuditLog creates a new audit log.
func NewAuditLog() *AuditLog {
	return &AuditLog{
		entries: []AuditEntry{},
	}
}

// Record logs an exception usage.
func (l *AuditLog) Record(exc *Exception, file string) {
	l.entries = append(l.entries, AuditEntry{
		Timestamp: time.Now(),
		Policy:    exc.Policy,
		Path:      file,
		Reason:    exc.Reason,
		ExpiresAt: exc.Expires,
	})
}

// Entries returns all recorded audit entries.
func (l *AuditLog) Entries() []AuditEntry {
	return l.entries
}

// ExpiringSoon returns exceptions that will expire within the given duration.
func (c *ExceptionConfig) ExpiringSoon(within time.Duration) []Exception {
	var expiring []Exception
	deadline := time.Now().Add(within)

	for _, exc := range c.Exceptions {
		if exc.Expires != nil && exc.Expires.Before(deadline) && !exc.IsExpired() {
			expiring = append(expiring, exc)
		}
	}

	return expiring
}
