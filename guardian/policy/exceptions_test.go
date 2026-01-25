package policy_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/policy"
	"github.com/mrz1836/go-fortress/guardian/validator"
)

// TestLoadExceptionConfig tests loading exception configuration from YAML files.
func TestLoadExceptionConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yaml        string
		expectCount int
		expectError bool
		createFile  bool
	}{
		{
			name: "valid config with exceptions",
			yaml: `
exceptions:
  - policy: sha-pinned-actions
    path: .github/workflows/test.yml
    reason: "Legacy workflow, migrating soon"
    created_at: 2025-01-01T00:00:00Z
`,
			expectCount: 1,
			createFile:  true,
		},
		{
			name: "empty config",
			yaml: `
exceptions: []
`,
			expectCount: 0,
			createFile:  true,
		},
		{
			name:        "missing file returns empty config",
			createFile:  false,
			expectCount: 0,
		},
		{
			name:        "invalid yaml",
			yaml:        `{{{invalid`,
			expectError: true,
			createFile:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "guardian.yaml")

			if tt.createFile {
				require.NoError(t, os.WriteFile(configPath, []byte(tt.yaml), 0o600))
			}

			config, err := policy.LoadExceptionConfig(configPath)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)
			assert.Len(t, config.Exceptions, tt.expectCount)
		})
	}
}

// TestException_Matches tests if exceptions match findings.
func TestException_Matches(t *testing.T) {
	t.Parallel()

	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name      string
		exception policy.Exception
		finding   validator.Finding
		matches   bool
	}{
		{
			name: "matches policy without prefix",
			exception: policy.Exception{
				Policy:    "sha-pinned-actions",
				CreatedAt: now,
			},
			finding: validator.Finding{
				RuleID: "policy/sha-pinned-actions",
				File:   "test.yml",
			},
			matches: true,
		},
		{
			name: "policy mismatch",
			exception: policy.Exception{
				Policy:    "explicit-permissions",
				CreatedAt: now,
			},
			finding: validator.Finding{
				RuleID: "policy/sha-pinned-actions",
				File:   "test.yml",
			},
			matches: false,
		},
		{
			name: "path matches exactly",
			exception: policy.Exception{
				Policy:    "sha-pinned-actions",
				Path:      "workflow.yml",
				CreatedAt: now,
			},
			finding: validator.Finding{
				RuleID: "policy/sha-pinned-actions",
				File:   "workflow.yml",
			},
			matches: true,
		},
		{
			name: "path matches with glob",
			exception: policy.Exception{
				Policy:    "sha-pinned-actions",
				Path:      "*.yml",
				CreatedAt: now,
			},
			finding: validator.Finding{
				RuleID: "policy/sha-pinned-actions",
				File:   "workflow.yml",
			},
			matches: true,
		},
		{
			name: "path doesn't match",
			exception: policy.Exception{
				Policy:    "sha-pinned-actions",
				Path:      "other.yml",
				CreatedAt: now,
			},
			finding: validator.Finding{
				RuleID: "policy/sha-pinned-actions",
				File:   "workflow.yml",
			},
			matches: false,
		},
		{
			name: "expired exception doesn't match",
			exception: policy.Exception{
				Policy:    "sha-pinned-actions",
				Expires:   &past,
				CreatedAt: now.Add(-48 * time.Hour),
			},
			finding: validator.Finding{
				RuleID: "policy/sha-pinned-actions",
				File:   "test.yml",
			},
			matches: false,
		},
		{
			name: "non-expired exception matches",
			exception: policy.Exception{
				Policy:    "sha-pinned-actions",
				Expires:   &future,
				CreatedAt: now,
			},
			finding: validator.Finding{
				RuleID: "policy/sha-pinned-actions",
				File:   "test.yml",
			},
			matches: true,
		},
		{
			name: "no expiration always matches",
			exception: policy.Exception{
				Policy:    "sha-pinned-actions",
				CreatedAt: now,
			},
			finding: validator.Finding{
				RuleID: "policy/sha-pinned-actions",
				File:   "test.yml",
			},
			matches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.exception.Matches(&tt.finding)
			assert.Equal(t, tt.matches, result)
		})
	}
}

// TestException_IsExpired tests expiration checking.
func TestException_IsExpired(t *testing.T) {
	t.Parallel()

	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name    string
		expires *time.Time
		expired bool
	}{
		{
			name:    "no expiration is not expired",
			expires: nil,
			expired: false,
		},
		{
			name:    "future expiration is not expired",
			expires: &future,
			expired: false,
		},
		{
			name:    "past expiration is expired",
			expires: &past,
			expired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exception := policy.Exception{
				Policy:    "test",
				Expires:   tt.expires,
				CreatedAt: now,
			}

			result := exception.IsExpired()
			assert.Equal(t, tt.expired, result)
		})
	}
}

// TestException_Validate tests exception validation.
func TestException_Validate(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name        string
		exception   policy.Exception
		expectError bool
	}{
		{
			name: "valid exception",
			exception: policy.Exception{
				Policy:    "sha-pinned-actions",
				Reason:    "Legacy workflow",
				CreatedAt: now,
			},
			expectError: false,
		},
		{
			name: "missing policy",
			exception: policy.Exception{
				Reason:    "Legacy workflow",
				CreatedAt: now,
			},
			expectError: true,
		},
		{
			name: "missing reason",
			exception: policy.Exception{
				Policy:    "sha-pinned-actions",
				CreatedAt: now,
			},
			expectError: true,
		},
		{
			name: "missing created_at",
			exception: policy.Exception{
				Policy: "sha-pinned-actions",
				Reason: "Legacy workflow",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.exception.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAuditLog tests audit logging functionality.
func TestAuditLog(t *testing.T) {
	t.Parallel()

	log := policy.NewAuditLog()
	require.NotNil(t, log)

	// Initially empty
	assert.Empty(t, log.Entries())

	// Record an exception usage
	now := time.Now()
	exc := &policy.Exception{
		Policy:    "sha-pinned-actions",
		Reason:    "Test exception",
		CreatedAt: now,
	}
	log.Record(exc, "workflow.yml")

	// Should have one entry
	entries := log.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, "sha-pinned-actions", entries[0].Policy)
	assert.Equal(t, "workflow.yml", entries[0].Path)
	assert.Equal(t, "Test exception", entries[0].Reason)

	// Record another
	log.Record(exc, "other.yml")
	assert.Len(t, log.Entries(), 2)
}

// TestExceptionConfig_ExpiringSoon tests finding soon-to-expire exceptions.
func TestExceptionConfig_ExpiringSoon(t *testing.T) {
	t.Parallel()

	now := time.Now()
	in1Day := now.Add(24 * time.Hour)
	in10Days := now.Add(10 * 24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	config := &policy.ExceptionConfig{
		Exceptions: []policy.Exception{
			{
				Policy:    "policy-a",
				Reason:    "Expiring soon",
				CreatedAt: now,
				Expires:   &in1Day,
			},
			{
				Policy:    "policy-b",
				Reason:    "Expiring later",
				CreatedAt: now,
				Expires:   &in10Days,
			},
			{
				Policy:    "policy-c",
				Reason:    "Already expired",
				CreatedAt: now.Add(-48 * time.Hour),
				Expires:   &past,
			},
			{
				Policy:    "policy-d",
				Reason:    "No expiration",
				CreatedAt: now,
			},
		},
	}

	// Check what expires within 7 days
	expiring := config.ExpiringSoon(7 * 24 * time.Hour)
	require.Len(t, expiring, 1)
	assert.Equal(t, "policy-a", expiring[0].Policy)

	// Check what expires within 14 days
	expiring = config.ExpiringSoon(14 * 24 * time.Hour)
	require.Len(t, expiring, 2)
}
