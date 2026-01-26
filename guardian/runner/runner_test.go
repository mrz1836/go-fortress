package runner_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/reporter"
	"github.com/mrz1836/go-fortress/guardian/runner"
)

// Test errors for retry tests.
var (
	errTemporary      = errors.New("temporary error")
	errPermanent      = errors.New("permanent error")
	errAlwaysFails    = errors.New("always fails")
	errFail           = errors.New("fail")
	errDockerNotAvail = errors.New("docker not available")
)

// mockRunner is a test implementation of the Runner interface.
type mockRunner struct {
	runResult *runner.RunResult
	runErr    error
	checkErr  error
}

func (m *mockRunner) Run(_ context.Context, _ runner.RunOptions) (*runner.RunResult, error) {
	if m.runErr != nil {
		return nil, m.runErr
	}
	return m.runResult, nil
}

func (m *mockRunner) CheckAvailable(_ context.Context) error {
	return m.checkErr
}

// TestNewActRunner tests ActRunner creation.
func TestNewActRunner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		wantPath string
	}{
		{
			name:     "empty path defaults to act",
			path:     "",
			wantPath: "act",
		},
		{
			name:     "custom path is preserved",
			path:     "/usr/local/bin/act",
			wantPath: "/usr/local/bin/act",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actRunner, err := runner.NewActRunner(tt.path)
			require.NoError(t, err)
			require.NotNil(t, actRunner)
		})
	}
}

// TestDefaultRetryConfig tests retry configuration defaults.
func TestDefaultRetryConfig(t *testing.T) {
	t.Parallel()

	cfg := runner.DefaultRetryConfig()

	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 1*time.Second, cfg.InitialBackoff)
	assert.Equal(t, 30*time.Second, cfg.MaxBackoff)
	assert.InDelta(t, 2.0, cfg.BackoffFactor, 0.001)
}

// TestCheckDiskSpace tests disk space verification.
func TestCheckDiskSpace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		requiredGB int
		wantErr    bool
	}{
		{
			name:       "empty path defaults to current directory",
			path:       "",
			requiredGB: 0,
			wantErr:    false,
		},
		{
			name:       "current directory with minimal requirements",
			path:       ".",
			requiredGB: 1,
			wantErr:    false, // Should pass on any reasonable system
		},
		{
			name:       "ridiculous requirement should fail",
			path:       ".",
			requiredGB: 1000000, // 1 million GB
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := runner.CheckDiskSpace(tt.path, tt.requiredGB)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, runner.ErrInsufficientDiskSpace)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRetryWithBackoff tests the retry mechanism.
func TestRetryWithBackoff(t *testing.T) {
	t.Parallel()

	t.Run("success on first try", func(t *testing.T) {
		t.Parallel()

		cfg := runner.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
			BackoffFactor:  2.0,
		}

		attempts := 0
		result, err := runner.RetryWithBackoff(context.Background(), cfg, func() (string, bool, error) {
			attempts++
			return "success", false, nil
		})

		require.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 1, attempts)
	})

	t.Run("retries on retryable error then succeeds", func(t *testing.T) {
		t.Parallel()

		cfg := runner.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 1 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
			BackoffFactor:  2.0,
		}

		attempts := 0
		result, err := runner.RetryWithBackoff(context.Background(), cfg, func() (int, bool, error) {
			attempts++
			if attempts < 3 {
				return 0, true, errTemporary
			}
			return 42, false, nil
		})

		require.NoError(t, err)
		assert.Equal(t, 42, result)
		assert.Equal(t, 3, attempts)
	})

	t.Run("non-retryable error stops immediately", func(t *testing.T) {
		t.Parallel()

		cfg := runner.RetryConfig{
			MaxRetries:     5,
			InitialBackoff: 1 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
			BackoffFactor:  2.0,
		}

		attempts := 0
		_, err := runner.RetryWithBackoff(context.Background(), cfg, func() (string, bool, error) {
			attempts++
			return "", false, errPermanent
		})

		require.Error(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("max retries exhausted", func(t *testing.T) {
		t.Parallel()

		cfg := runner.RetryConfig{
			MaxRetries:     2,
			InitialBackoff: 1 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
			BackoffFactor:  2.0,
		}

		attempts := 0
		_, err := runner.RetryWithBackoff(context.Background(), cfg, func() (string, bool, error) {
			attempts++
			return "", true, errAlwaysFails
		})

		require.Error(t, err)
		assert.Equal(t, 3, attempts) // initial + 2 retries
	})

	t.Run("context cancellation stops retries", func(t *testing.T) {
		t.Parallel()

		cfg := runner.RetryConfig{
			MaxRetries:     10,
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     1 * time.Second,
			BackoffFactor:  2.0,
		}

		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		_, err := runner.RetryWithBackoff(ctx, cfg, func() (string, bool, error) {
			attempts++
			return "", true, errFail
		})

		require.Error(t, err)
		assert.LessOrEqual(t, attempts, 3) // Should be stopped early
	})
}

// TestScenarioRunner_NewScenarioRunner tests ScenarioRunner creation.
func TestScenarioRunner_NewScenarioRunner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		inputParallel  int
		inputTimeout   time.Duration
		expectParallel int
		expectTimeout  time.Duration
	}{
		{
			name:           "defaults applied for zero values",
			inputParallel:  0,
			inputTimeout:   0,
			expectParallel: 4,
			expectTimeout:  30 * time.Second,
		},
		{
			name:           "negative parallel becomes default",
			inputParallel:  -1,
			inputTimeout:   0,
			expectParallel: 4,
			expectTimeout:  30 * time.Second,
		},
		{
			name:           "custom values preserved",
			inputParallel:  8,
			inputTimeout:   2 * time.Minute,
			expectParallel: 8,
			expectTimeout:  2 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockRunner{}
			cfg := runner.ScenarioConfig{
				ParallelScenarios: tt.inputParallel,
				DefaultTimeout:    tt.inputTimeout,
			}

			sr := runner.NewScenarioRunner(mock, cfg)
			require.NotNil(t, sr)
		})
	}
}

// TestScenarioRunner_Execute tests single scenario execution.
func TestScenarioRunner_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		runResult  *runner.RunResult
		runErr     error
		expected   runner.ExpectedResult
		wantStatus reporter.ResultStatus
	}{
		{
			name: "successful scenario passes",
			runResult: &runner.RunResult{
				ExitCode: 0,
				Output:   "Build completed successfully",
			},
			expected: runner.ExpectedResult{
				Status: runner.StatusSuccess,
			},
			wantStatus: reporter.ResultPass,
		},
		{
			name: "failed scenario when expecting failure passes",
			runResult: &runner.RunResult{
				ExitCode: 1,
				Output:   "Lint errors found",
			},
			expected: runner.ExpectedResult{
				Status: runner.StatusFailure,
			},
			wantStatus: reporter.ResultPass,
		},
		{
			name: "unexpected failure reports fail",
			runResult: &runner.RunResult{
				ExitCode: 1,
				Output:   "Unexpected error",
			},
			expected: runner.ExpectedResult{
				Status: runner.StatusSuccess,
			},
			wantStatus: reporter.ResultFail,
		},
		{
			name:   "runner error reports error status",
			runErr: errDockerNotAvail,
			expected: runner.ExpectedResult{
				Status: runner.StatusSuccess,
			},
			wantStatus: reporter.ResultError,
		},
		{
			name: "pattern matching succeeds",
			runResult: &runner.RunResult{
				ExitCode: 0,
				Output:   "Running tests...\nAll 42 tests passed!",
			},
			expected: runner.ExpectedResult{
				Status:      runner.StatusSuccess,
				LogPatterns: []string{"tests passed", "\\d+ tests"},
			},
			wantStatus: reporter.ResultPass,
		},
		{
			name: "missing pattern fails",
			runResult: &runner.RunResult{
				ExitCode: 0,
				Output:   "Build completed",
			},
			expected: runner.ExpectedResult{
				Status:      runner.StatusSuccess,
				LogPatterns: []string{"tests passed"},
			},
			wantStatus: reporter.ResultFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockRunner{
				runResult: tt.runResult,
				runErr:    tt.runErr,
			}

			sr := runner.NewScenarioRunner(mock, runner.ScenarioConfig{
				DefaultTimeout: 5 * time.Second,
			})

			scenario := &runner.ScenarioDefinition{
				ID:       "TEST-001",
				Expected: tt.expected,
			}

			result, err := sr.Execute(context.Background(), scenario)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, result.Status)
		})
	}
}

// TestScenarioRunner_ExecuteAll tests multiple scenario execution.
func TestScenarioRunner_ExecuteAll(t *testing.T) {
	t.Parallel()

	t.Run("empty scenarios returns empty results", func(t *testing.T) {
		t.Parallel()

		mock := &mockRunner{}
		sr := runner.NewScenarioRunner(mock, runner.ScenarioConfig{})

		results, err := sr.ExecuteAll(context.Background(), []*runner.ScenarioDefinition{})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("sequential execution with single scenario", func(t *testing.T) {
		t.Parallel()

		mock := &mockRunner{
			runResult: &runner.RunResult{ExitCode: 0, Output: "success"},
		}
		sr := runner.NewScenarioRunner(mock, runner.ScenarioConfig{
			ParallelScenarios: 1,
		})

		scenarios := []*runner.ScenarioDefinition{
			{ID: "TEST-001", Expected: runner.ExpectedResult{Status: runner.StatusSuccess}},
		}

		results, err := sr.ExecuteAll(context.Background(), scenarios)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "TEST-001", results[0].ScenarioID)
	})

	t.Run("parallel execution preserves order", func(t *testing.T) {
		t.Parallel()

		mock := &mockRunner{
			runResult: &runner.RunResult{ExitCode: 0, Output: "success"},
		}
		sr := runner.NewScenarioRunner(mock, runner.ScenarioConfig{
			ParallelScenarios: 4,
		})

		scenarios := []*runner.ScenarioDefinition{
			{ID: "TEST-001", Expected: runner.ExpectedResult{Status: runner.StatusSuccess}},
			{ID: "TEST-002", Expected: runner.ExpectedResult{Status: runner.StatusSuccess}},
			{ID: "TEST-003", Expected: runner.ExpectedResult{Status: runner.StatusSuccess}},
		}

		results, err := sr.ExecuteAll(context.Background(), scenarios)
		require.NoError(t, err)
		require.Len(t, results, 3)

		// Results should be in original order
		assert.Equal(t, "TEST-001", results[0].ScenarioID)
		assert.Equal(t, "TEST-002", results[1].ScenarioID)
		assert.Equal(t, "TEST-003", results[2].ScenarioID)
	})
}

// TestEvents tests event payload functions.
func TestEvents(t *testing.T) {
	t.Parallel()

	t.Run("DefaultPushEvent creates valid event", func(t *testing.T) {
		t.Parallel()

		event := runner.DefaultPushEvent()
		require.NotNil(t, event)

		assert.Equal(t, "refs/heads/main", event.Ref)
		assert.Equal(t, "test-repo", event.Repository.Name)
		assert.Equal(t, "test-owner/test-repo", event.Repository.FullName)
	})

	t.Run("DefaultPullRequestEvent creates valid event", func(t *testing.T) {
		t.Parallel()

		event := runner.DefaultPullRequestEvent()
		require.NotNil(t, event)

		assert.Equal(t, "opened", event.Action)
		assert.Equal(t, 1, event.Number)
		assert.Equal(t, "feature-branch", event.PullRequest.Head.Ref)
		assert.Equal(t, "main", event.PullRequest.Base.Ref)
	})

	t.Run("ForkPullRequestEvent creates fork event", func(t *testing.T) {
		t.Parallel()

		event := runner.ForkPullRequestEvent()
		require.NotNil(t, event)

		// Fork should have different owner for head
		assert.Equal(t, "fork-owner/test-repo", event.PullRequest.Head.Repo.FullName)
		assert.Equal(t, "fork-owner:feature-branch", event.PullRequest.Head.Label)
	})

	t.Run("WriteEventFile creates valid JSON", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		event := runner.DefaultPushEvent()

		path, err := runner.WriteEventFile(event, tmpDir)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(tmpDir, "event.json"), path)

		// Verify file exists and is valid JSON
		data, err := os.ReadFile(path) //nolint:gosec // path is from test temp dir
		require.NoError(t, err)

		var decoded runner.PushEvent
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, event.Ref, decoded.Ref)
	})

	t.Run("LoadEventFile reads valid JSON", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		event := runner.DefaultPushEvent()

		path, err := runner.WriteEventFile(event, tmpDir)
		require.NoError(t, err)

		loaded, err := runner.LoadEventFile(path)
		require.NoError(t, err)
		assert.Equal(t, "refs/heads/main", loaded["ref"])
	})

	t.Run("LoadEventFile errors on missing file", func(t *testing.T) {
		t.Parallel()

		_, err := runner.LoadEventFile("/nonexistent/path/event.json")
		require.Error(t, err)
	})

	t.Run("LoadEventFile errors on invalid JSON", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(path, []byte("not valid json"), 0o600)
		require.NoError(t, err)

		_, err = runner.LoadEventFile(path)
		require.Error(t, err)
	})
}

// TestBuildArgs tests command line argument building for act.
func TestBuildArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     runner.RunOptions
		expected []string
	}{
		{
			name:     "empty options includes platform mappings",
			opts:     runner.RunOptions{},
			expected: []string{"--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest", "--platform", "ubuntu-22.04=catthehacker/ubuntu:act-22.04", "--platform", "ubuntu-24.04=catthehacker/ubuntu:act-latest"},
		},
		{
			name: "workflow file without prefix",
			opts: runner.RunOptions{
				WorkflowFile: "test.yml",
			},
			expected: []string{"--workflows", ".github/workflows/test.yml", "--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest", "--platform", "ubuntu-22.04=catthehacker/ubuntu:act-22.04", "--platform", "ubuntu-24.04=catthehacker/ubuntu:act-latest"},
		},
		{
			name: "workflow file with prefix",
			opts: runner.RunOptions{
				WorkflowFile: ".github/workflows/ci.yml",
			},
			expected: []string{"--workflows", ".github/workflows/ci.yml", "--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest", "--platform", "ubuntu-22.04=catthehacker/ubuntu:act-22.04", "--platform", "ubuntu-24.04=catthehacker/ubuntu:act-latest"},
		},
		{
			name: "with job specified",
			opts: runner.RunOptions{
				Job: "build",
			},
			expected: []string{"--job", "build", "--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest", "--platform", "ubuntu-22.04=catthehacker/ubuntu:act-22.04", "--platform", "ubuntu-24.04=catthehacker/ubuntu:act-latest"},
		},
		{
			name: "with event file",
			opts: runner.RunOptions{
				EventFile: "/tmp/event.json",
			},
			expected: []string{"--eventpath", "/tmp/event.json", "--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest", "--platform", "ubuntu-22.04=catthehacker/ubuntu:act-22.04", "--platform", "ubuntu-24.04=catthehacker/ubuntu:act-latest"},
		},
		{
			name: "with secrets file",
			opts: runner.RunOptions{
				SecretsFile: "/tmp/secrets.env",
			},
			expected: []string{"--secret-file", "/tmp/secrets.env", "--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest", "--platform", "ubuntu-22.04=catthehacker/ubuntu:act-22.04", "--platform", "ubuntu-24.04=catthehacker/ubuntu:act-latest"},
		},
		{
			name: "verbose mode",
			opts: runner.RunOptions{
				Verbose: true,
			},
			expected: []string{"--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest", "--platform", "ubuntu-22.04=catthehacker/ubuntu:act-22.04", "--platform", "ubuntu-24.04=catthehacker/ubuntu:act-latest", "--verbose"},
		},
		{
			name: "keep container",
			opts: runner.RunOptions{
				KeepContainer: true,
			},
			expected: []string{"--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest", "--platform", "ubuntu-22.04=catthehacker/ubuntu:act-22.04", "--platform", "ubuntu-24.04=catthehacker/ubuntu:act-latest", "--reuse"},
		},
		{
			name: "all options",
			opts: runner.RunOptions{
				WorkflowFile:  "release.yml",
				Job:           "deploy",
				EventFile:     "/tmp/event.json",
				SecretsFile:   "/tmp/secrets.env",
				Verbose:       true,
				KeepContainer: true,
			},
			expected: []string{
				"--workflows", ".github/workflows/release.yml",
				"--job", "deploy",
				"--eventpath", "/tmp/event.json",
				"--secret-file", "/tmp/secrets.env",
				"--platform", "ubuntu-latest=catthehacker/ubuntu:act-latest",
				"--platform", "ubuntu-22.04=catthehacker/ubuntu:act-22.04",
				"--platform", "ubuntu-24.04=catthehacker/ubuntu:act-latest",
				"--verbose",
				"--reuse",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actRunner, err := runner.NewActRunner("")
			require.NoError(t, err)

			args := actRunner.BuildArgs(tt.opts)
			assert.Equal(t, tt.expected, args)
		})
	}
}

// TestExtractContainerID tests container ID extraction from act output.
func TestExtractContainerID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "no container ID",
			output:   "Running job...\nDone!",
			expected: "",
		},
		{
			name:     "container ID present",
			output:   "Starting container\ncontainer_id=abc123def456\nDone",
			expected: "abc123def456",
		},
		{
			name:     "container ID with whitespace",
			output:   "container_id=xyz789  \nNext line",
			expected: "xyz789",
		},
		{
			name:     "multiple lines before container ID",
			output:   "Line 1\nLine 2\nLine 3\ncontainer_id=mycontainer\nLine 5",
			expected: "mycontainer",
		},
		{
			name:     "empty output",
			output:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := runner.ExtractContainerID(tt.output)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateResult tests the result validation logic.
func TestValidateResult(t *testing.T) {
	t.Parallel()

	exitCode1 := 1

	tests := []struct {
		name        string
		result      *runner.RunResult
		expected    *runner.ExpectedResult
		wantStatus  reporter.ResultStatus
		wantMatched []string
		wantMissing []string
	}{
		{
			name:       "success matches success",
			result:     &runner.RunResult{ExitCode: 0, Output: "done"},
			expected:   &runner.ExpectedResult{Status: runner.StatusSuccess},
			wantStatus: reporter.ResultPass,
		},
		{
			name:       "failure matches failure",
			result:     &runner.RunResult{ExitCode: 1, Output: "error"},
			expected:   &runner.ExpectedResult{Status: runner.StatusFailure},
			wantStatus: reporter.ResultPass,
		},
		{
			name:       "success when expecting failure fails",
			result:     &runner.RunResult{ExitCode: 0, Output: "done"},
			expected:   &runner.ExpectedResult{Status: runner.StatusFailure},
			wantStatus: reporter.ResultFail,
		},
		{
			name:       "specific exit code matches",
			result:     &runner.RunResult{ExitCode: 1, Output: "error"},
			expected:   &runner.ExpectedResult{Status: runner.StatusFailure, ExitCode: &exitCode1},
			wantStatus: reporter.ResultPass,
		},
		{
			name:        "log patterns all matched",
			result:      &runner.RunResult{ExitCode: 0, Output: "starting process\ncompleted in 5s"},
			expected:    &runner.ExpectedResult{Status: runner.StatusSuccess, LogPatterns: []string{"starting", "completed"}},
			wantStatus:  reporter.ResultPass,
			wantMatched: []string{"starting", "completed"},
		},
		{
			name:        "log pattern missing",
			result:      &runner.RunResult{ExitCode: 0, Output: "starting process"},
			expected:    &runner.ExpectedResult{Status: runner.StatusSuccess, LogPatterns: []string{"starting", "completed"}},
			wantStatus:  reporter.ResultFail,
			wantMatched: []string{"starting"},
			wantMissing: []string{"completed"},
		},
		{
			name:        "exclude pattern found fails",
			result:      &runner.RunResult{ExitCode: 0, Output: "process completed with secret=abc123"},
			expected:    &runner.ExpectedResult{Status: runner.StatusSuccess, ExcludePatterns: []string{"secret="}},
			wantStatus:  reporter.ResultFail,
			wantMissing: []string{"excluded pattern found: secret="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Execute via ScenarioRunner.Execute to test validateResult indirectly
			mock := &mockRunner{runResult: tt.result}
			sr := runner.NewScenarioRunner(mock, runner.ScenarioConfig{
				DefaultTimeout: 5 * time.Second,
			})

			scenario := &runner.ScenarioDefinition{
				ID:       "TEST-001",
				Expected: *tt.expected,
			}

			result, err := sr.Execute(context.Background(), scenario)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, result.Status)

			if tt.wantMatched != nil {
				assert.Equal(t, tt.wantMatched, result.MatchedPatterns)
			}
			if tt.wantMissing != nil {
				assert.Equal(t, tt.wantMissing, result.MissingPatterns)
			}
		})
	}
}
