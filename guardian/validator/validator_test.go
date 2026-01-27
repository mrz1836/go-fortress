package validator_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/validator"
)

var errValidationFailed = errors.New("validation failed")

// Test errors for error reporting tests.
var (
	errOutOfMemory       = errors.New("validator crashed: out of memory")
	errConnectionTimeout = errors.New("connection timeout")
	errParseError        = errors.New("parse error")
)

// mockValidator is a test validator implementation.
type mockValidator struct {
	name     string
	findings []validator.Finding
	err      error
}

func (m *mockValidator) Name() string {
	return m.name
}

func (m *mockValidator) Validate(_ context.Context, _ string) ([]validator.Finding, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.findings, nil
}

// TestRegistry_NewRegistry tests creating a new registry.
func TestRegistry_NewRegistry(t *testing.T) {
	t.Parallel()

	reg := validator.NewRegistry()
	require.NotNil(t, reg)

	// Empty registry should return empty slice
	all := reg.All()
	assert.Empty(t, all)
}

// TestRegistry_RegisterAndGet tests registering and retrieving validators.
func TestRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		validators []*mockValidator
		getNames   []string
		expectOK   []bool
	}{
		{
			name: "register single validator",
			validators: []*mockValidator{
				{name: "test-validator"},
			},
			getNames: []string{"test-validator", "nonexistent"},
			expectOK: []bool{true, false},
		},
		{
			name: "register multiple validators",
			validators: []*mockValidator{
				{name: "validator-a"},
				{name: "validator-b"},
				{name: "validator-c"},
			},
			getNames: []string{"validator-a", "validator-b", "validator-c", "validator-d"},
			expectOK: []bool{true, true, true, false},
		},
		{
			name: "re-register same validator replaces it",
			validators: []*mockValidator{
				{name: "same-name", findings: []validator.Finding{{Message: "first"}}},
				{name: "same-name", findings: []validator.Finding{{Message: "second"}}},
			},
			getNames: []string{"same-name"},
			expectOK: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := validator.NewRegistry()

			for _, v := range tt.validators {
				reg.Register(v)
			}

			for i, name := range tt.getNames {
				v, ok := reg.Get(name)
				assert.Equal(t, tt.expectOK[i], ok, "Get(%q) ok mismatch", name)
				if ok {
					assert.NotNil(t, v)
					assert.Equal(t, name, v.Name())
				} else {
					assert.Nil(t, v)
				}
			}
		})
	}
}

// TestRegistry_All tests that All returns validators in registration order.
func TestRegistry_All(t *testing.T) {
	t.Parallel()

	reg := validator.NewRegistry()

	// Register validators in specific order
	validators := []*mockValidator{
		{name: "first"},
		{name: "second"},
		{name: "third"},
	}

	for _, v := range validators {
		reg.Register(v)
	}

	all := reg.All()
	require.Len(t, all, 3)

	// Verify order is preserved
	assert.Equal(t, "first", all[0].Name())
	assert.Equal(t, "second", all[1].Name())
	assert.Equal(t, "third", all[2].Name())
}

// TestRegistry_All_ReregisterPreservesOrder tests that re-registering doesn't change order.
func TestRegistry_All_ReregisterPreservesOrder(t *testing.T) {
	t.Parallel()

	reg := validator.NewRegistry()

	reg.Register(&mockValidator{name: "a"})
	reg.Register(&mockValidator{name: "b"})
	reg.Register(&mockValidator{name: "c"})
	reg.Register(&mockValidator{name: "b"}) // Re-register b

	all := reg.All()
	require.Len(t, all, 3)

	// Order should still be a, b, c (not a, c, b)
	assert.Equal(t, "a", all[0].Name())
	assert.Equal(t, "b", all[1].Name())
	assert.Equal(t, "c", all[2].Name())
}

// TestRegistry_ValidateAll tests running all validators against a file.
func TestRegistry_ValidateAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		validators       []*mockValidator
		expectedFindings int
	}{
		{
			name:             "no validators",
			validators:       []*mockValidator{},
			expectedFindings: 0,
		},
		{
			name: "single validator with findings",
			validators: []*mockValidator{
				{
					name: "validator-1",
					findings: []validator.Finding{
						{Message: "finding 1"},
						{Message: "finding 2"},
					},
				},
			},
			expectedFindings: 2,
		},
		{
			name: "multiple validators aggregate findings",
			validators: []*mockValidator{
				{
					name:     "validator-a",
					findings: []validator.Finding{{Message: "a1"}, {Message: "a2"}},
				},
				{
					name:     "validator-b",
					findings: []validator.Finding{{Message: "b1"}},
				},
				{
					name:     "validator-c",
					findings: []validator.Finding{{Message: "c1"}, {Message: "c2"}, {Message: "c3"}},
				},
			},
			expectedFindings: 6,
		},
		{
			name: "validator error creates error finding, others continue",
			validators: []*mockValidator{
				{
					name:     "validator-ok",
					findings: []validator.Finding{{Message: "ok"}},
				},
				{
					name: "validator-error",
					err:  errValidationFailed,
				},
				{
					name:     "validator-also-ok",
					findings: []validator.Finding{{Message: "also ok"}},
				},
			},
			expectedFindings: 3, // 2 from working validators + 1 error finding from failed validator
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := validator.NewRegistry()
			for _, v := range tt.validators {
				reg.Register(v)
			}

			ctx := context.Background()
			findings, err := reg.ValidateAll(ctx, "test-workflow.yml")

			require.NoError(t, err)
			assert.Len(t, findings, tt.expectedFindings)
		})
	}
}

// TestFindWorkflowFiles tests finding workflow YAML files.
func TestFindWorkflowFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		files         map[string]string // relative path -> content
		expectedCount int
		expectedFiles []string // relative paths expected
	}{
		{
			name: "finds yml files",
			files: map[string]string{
				"workflow1.yml": "name: Test 1",
				"workflow2.yml": "name: Test 2",
			},
			expectedCount: 2,
		},
		{
			name: "finds yaml files",
			files: map[string]string{
				"workflow1.yaml": "name: Test 1",
				"workflow2.yaml": "name: Test 2",
			},
			expectedCount: 2,
		},
		{
			name: "finds both yml and yaml",
			files: map[string]string{
				"a.yml":  "name: A",
				"b.yaml": "name: B",
			},
			expectedCount: 2,
		},
		{
			name: "ignores non-workflow files",
			files: map[string]string{
				"workflow.yml": "name: Workflow",
				"readme.md":    "# Readme",
				"script.sh":    "#!/bin/bash",
				"config.json":  "{}",
			},
			expectedCount: 1,
		},
		{
			name: "case insensitive extension",
			files: map[string]string{
				"a.YML":  "name: A",
				"b.YAML": "name: B",
			},
			expectedCount: 2,
		},
		{
			name:          "empty directory",
			files:         map[string]string{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory with test files
			tmpDir := t.TempDir()

			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				err := os.WriteFile(fullPath, []byte(content), 0o600)
				require.NoError(t, err)
			}

			files, err := validator.FindWorkflowFiles(tmpDir)
			require.NoError(t, err)
			assert.Len(t, files, tt.expectedCount)
		})
	}
}

// TestFindWorkflowFiles_MissingDir tests behavior with missing directory.
func TestFindWorkflowFiles_MissingDir(t *testing.T) {
	t.Parallel()

	// Non-existent directory should return empty slice, not error
	files, err := validator.FindWorkflowFiles("/nonexistent/path/to/workflows")
	require.NoError(t, err)
	assert.Empty(t, files)
}

// TestFindWorkflowFiles_Subdirectories tests finding files in subdirectories.
func TestFindWorkflowFiles_Subdirectories(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create nested structure
	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0o750))

	// Create files at different levels
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "root.yml"), []byte("name: Root"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "nested.yml"), []byte("name: Nested"), 0o600))

	files, err := validator.FindWorkflowFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

// TestRegistry_ValidateAll_ErrorReporting tests that validator errors are reported as findings.
func TestRegistry_ValidateAll_ErrorReporting(t *testing.T) {
	t.Parallel()

	reg := validator.NewRegistry()

	// Register a validator that returns an error
	reg.Register(&mockValidator{
		name: "failing-validator",
		err:  errOutOfMemory,
	})

	// Register a working validator
	reg.Register(&mockValidator{
		name: "working-validator",
		findings: []validator.Finding{
			{Message: "normal finding", RuleID: "test/rule"},
		},
	})

	ctx := context.Background()
	findings, err := reg.ValidateAll(ctx, "test-workflow.yml")

	// ValidateAll should not return an error
	require.NoError(t, err)

	// Should have 2 findings: 1 error finding + 1 normal finding
	require.Len(t, findings, 2)

	// Find the error finding
	var errorFinding *validator.Finding
	var normalFinding *validator.Finding
	for i := range findings {
		if findings[i].Source == validator.SourceValidator {
			errorFinding = &findings[i]
		} else {
			normalFinding = &findings[i]
		}
	}

	// Verify the error finding
	require.NotNil(t, errorFinding, "should have an error finding from failed validator")
	assert.Equal(t, "validator/failing-validator-error", errorFinding.RuleID)
	assert.Equal(t, validator.SeverityError, errorFinding.Severity)
	assert.Contains(t, errorFinding.Message, "validator failed")
	assert.Contains(t, errorFinding.Message, errOutOfMemory.Error())
	assert.Equal(t, "test-workflow.yml", errorFinding.File)
	assert.Equal(t, validator.SourceValidator, errorFinding.Source)

	// Verify the normal finding still comes through
	require.NotNil(t, normalFinding, "should have the normal finding from working validator")
	assert.Equal(t, "test/rule", normalFinding.RuleID)
}

// TestRegistry_ValidateAll_AllValidatorsFail tests when all validators fail.
func TestRegistry_ValidateAll_AllValidatorsFail(t *testing.T) {
	t.Parallel()

	reg := validator.NewRegistry()

	// Register multiple failing validators
	reg.Register(&mockValidator{
		name: "validator-1",
		err:  errConnectionTimeout,
	})
	reg.Register(&mockValidator{
		name: "validator-2",
		err:  errParseError,
	})

	ctx := context.Background()
	findings, err := reg.ValidateAll(ctx, "workflow.yml")

	// Should not return an error
	require.NoError(t, err)

	// Should have 2 error findings
	require.Len(t, findings, 2)

	// All findings should be error severity
	for _, f := range findings {
		assert.Equal(t, validator.SeverityError, f.Severity)
		assert.Equal(t, validator.SourceValidator, f.Source)
		assert.Contains(t, f.Message, "validator failed")
	}
}
