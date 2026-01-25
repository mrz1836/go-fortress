package reporter_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/reporter"
)

// mockReporter is a test reporter implementation.
type mockReporter struct {
	name string
}

func (m *mockReporter) Name() string {
	return m.name
}

func (m *mockReporter) Write(_ context.Context, _ *reporter.Report, _ io.Writer) error {
	return nil
}

func (m *mockReporter) WriteFile(_ context.Context, _ *reporter.Report, _ string) error {
	return nil
}

// TestReporterRegistry_NewRegistry tests creating a new registry.
func TestReporterRegistry_NewRegistry(t *testing.T) {
	t.Parallel()

	reg := reporter.NewRegistry()
	require.NotNil(t, reg)

	// Empty registry should return empty slice
	all := reg.All()
	assert.Empty(t, all)
}

// TestReporterRegistry_RegisterAndGet tests registering and retrieving reporters.
func TestReporterRegistry_RegisterAndGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		reporters []*mockReporter
		getNames  []string
		expectOK  []bool
	}{
		{
			name: "register single reporter",
			reporters: []*mockReporter{
				{name: "test-reporter"},
			},
			getNames: []string{"test-reporter", "nonexistent"},
			expectOK: []bool{true, false},
		},
		{
			name: "register multiple reporters",
			reporters: []*mockReporter{
				{name: "jsonl"},
				{name: "sarif"},
				{name: "terminal"},
			},
			getNames: []string{"jsonl", "sarif", "terminal", "html"},
			expectOK: []bool{true, true, true, false},
		},
		{
			name: "re-register same reporter replaces it",
			reporters: []*mockReporter{
				{name: "same-name"},
				{name: "same-name"},
			},
			getNames: []string{"same-name"},
			expectOK: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := reporter.NewRegistry()

			for _, r := range tt.reporters {
				reg.Register(r)
			}

			for i, name := range tt.getNames {
				rep, ok := reg.Get(name)
				assert.Equal(t, tt.expectOK[i], ok, "Get(%q) ok mismatch", name)
				if ok {
					assert.NotNil(t, rep)
					assert.Equal(t, name, rep.Name())
				} else {
					assert.Nil(t, rep)
				}
			}
		})
	}
}

// TestReporterRegistry_All tests that All returns reporters in registration order.
func TestReporterRegistry_All(t *testing.T) {
	t.Parallel()

	reg := reporter.NewRegistry()

	// Register reporters in specific order
	reporters := []*mockReporter{
		{name: "first"},
		{name: "second"},
		{name: "third"},
	}

	for _, r := range reporters {
		reg.Register(r)
	}

	all := reg.All()
	require.Len(t, all, 3)

	// Verify order is preserved
	assert.Equal(t, "first", all[0].Name())
	assert.Equal(t, "second", all[1].Name())
	assert.Equal(t, "third", all[2].Name())
}

// TestReporterRegistry_All_ReregisterPreservesOrder tests that re-registering doesn't change order.
func TestReporterRegistry_All_ReregisterPreservesOrder(t *testing.T) {
	t.Parallel()

	reg := reporter.NewRegistry()

	reg.Register(&mockReporter{name: "a"})
	reg.Register(&mockReporter{name: "b"})
	reg.Register(&mockReporter{name: "c"})
	reg.Register(&mockReporter{name: "b"}) // Re-register b

	all := reg.All()
	require.Len(t, all, 3)

	// Order should still be a, b, c (not a, c, b)
	assert.Equal(t, "a", all[0].Name())
	assert.Equal(t, "b", all[1].Name())
	assert.Equal(t, "c", all[2].Name())
}

// TestIsCI tests CI environment detection.
func TestIsCI(t *testing.T) {
	// Not using t.Parallel() because we're modifying env vars

	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "no CI env vars returns false",
			envVars:  map[string]string{},
			expected: false,
		},
		{
			name:     "CI=true returns true",
			envVars:  map[string]string{"CI": "true"},
			expected: true,
		},
		{
			name:     "CI=1 returns true",
			envVars:  map[string]string{"CI": "1"},
			expected: true,
		},
		{
			name:     "GITHUB_ACTIONS=true returns true",
			envVars:  map[string]string{"GITHUB_ACTIONS": "true"},
			expected: true,
		},
		{
			name:     "GITLAB_CI=true returns true",
			envVars:  map[string]string{"GITLAB_CI": "true"},
			expected: true,
		},
		{
			name:     "CIRCLECI=true returns true",
			envVars:  map[string]string{"CIRCLECI": "true"},
			expected: true,
		},
		{
			name:     "TRAVIS=true returns true",
			envVars:  map[string]string{"TRAVIS": "true"},
			expected: true,
		},
		{
			name:     "JENKINS_URL set returns true",
			envVars:  map[string]string{"JENKINS_URL": "http://jenkins.example.com"},
			expected: true,
		},
		{
			name:     "BUILDKITE=true returns true",
			envVars:  map[string]string{"BUILDKITE": "true"},
			expected: true,
		},
	}

	// Save original env
	ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "CIRCLECI", "TRAVIS", "JENKINS_URL", "BUILDKITE"}
	originalEnv := make(map[string]string)
	for _, key := range ciVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Cleanup
	defer func() {
		for key, val := range originalEnv {
			if val == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, val)
			}
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all CI env vars
			for _, key := range ciVars {
				_ = os.Unsetenv(key)
			}

			// Set test env vars
			for key, val := range tt.envVars {
				require.NoError(t, os.Setenv(key, val))
			}

			result := reporter.IsCI()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsGitHubActions tests GitHub Actions environment detection.
func TestIsGitHubActions(t *testing.T) {
	// Not using t.Parallel() because we're modifying env vars

	tests := []struct {
		name     string
		envVar   string
		expected bool
	}{
		{
			name:     "not set returns false",
			envVar:   "",
			expected: false,
		},
		{
			name:     "true returns true",
			envVar:   "true",
			expected: true,
		},
		{
			name:     "false returns false",
			envVar:   "false",
			expected: false,
		},
		{
			name:     "1 returns false (must be exactly 'true')",
			envVar:   "1",
			expected: false,
		},
	}

	// Save original env
	originalEnv := os.Getenv("GITHUB_ACTIONS")
	defer func() {
		if originalEnv == "" {
			_ = os.Unsetenv("GITHUB_ACTIONS")
		} else {
			_ = os.Setenv("GITHUB_ACTIONS", originalEnv)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar == "" {
				_ = os.Unsetenv("GITHUB_ACTIONS")
			} else {
				require.NoError(t, os.Setenv("GITHUB_ACTIONS", tt.envVar))
			}

			result := reporter.IsGitHubActions()
			assert.Equal(t, tt.expected, result)
		})
	}
}
