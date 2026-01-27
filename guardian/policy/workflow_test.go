package policy_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-fortress/guardian/policy"
)

// TestParseWorkflow tests parsing workflow from file.
func TestParseWorkflow(t *testing.T) {
	t.Parallel()

	yaml := `
name: Test Workflow
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`

	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "test.yml")
	require.NoError(t, os.WriteFile(workflowPath, []byte(yaml), 0o600))

	workflow, err := policy.ParseWorkflow(workflowPath)
	require.NoError(t, err)
	require.NotNil(t, workflow)

	assert.Equal(t, "Test Workflow", workflow.Name)
	assert.Equal(t, workflowPath, workflow.Path)
}

// TestParseWorkflow_FileNotFound tests error handling for missing files.
func TestParseWorkflow_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := policy.ParseWorkflow("/nonexistent/path/workflow.yml")
	require.Error(t, err)
}

// TestParseWorkflowBytes tests parsing workflow YAML content.
func TestParseWorkflowBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yaml        string
		expectName  string
		expectError bool
	}{
		{
			name: "minimal workflow",
			yaml: `
name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
			expectName: "Test",
		},
		{
			name: "workflow with permissions",
			yaml: `
name: With Permissions
on: push
permissions:
  contents: read
  actions: write
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
`,
			expectName: "With Permissions",
		},
		{
			name: "workflow with write-all permissions",
			yaml: `
name: Write All
on: push
permissions: write-all
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
`,
			expectName: "Write All",
		},
		{
			name: "workflow with concurrency",
			yaml: `
name: With Concurrency
on: push
concurrency:
  group: ${{ github.workflow }}
  cancel-in-progress: true
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
`,
			expectName: "With Concurrency",
		},
		{
			name: "workflow with simple concurrency",
			yaml: `
name: Simple Concurrency
on: push
concurrency: my-group
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
`,
			expectName: "Simple Concurrency",
		},
		{
			name: "workflow with array trigger",
			yaml: `
name: Array Trigger
on: [push, pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
`,
			expectName: "Array Trigger",
		},
		{
			name: "workflow with detailed triggers",
			yaml: `
name: Detailed Triggers
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
`,
			expectName: "Detailed Triggers",
		},
		{
			name: "workflow with pull_request_target",
			yaml: `
name: PR Target
on: pull_request_target
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
`,
			expectName: "PR Target",
		},
		{
			name: "workflow with workflow_dispatch",
			yaml: `
name: Manual
on: workflow_dispatch
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
`,
			expectName: "Manual",
		},
		{
			name: "workflow with job needs",
			yaml: `
name: With Needs
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: make build
  test:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - run: make test
`,
			expectName: "With Needs",
		},
		{
			name: "workflow with job needs array",
			yaml: `
name: With Needs Array
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: make build
  lint:
    runs-on: ubuntu-latest
    steps:
      - run: make lint
  test:
    runs-on: ubuntu-latest
    needs: [build, lint]
    steps:
      - run: make test
`,
			expectName: "With Needs Array",
		},
		{
			name: "workflow with job permissions",
			yaml: `
name: Job Permissions
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - run: echo hello
`,
			expectName: "Job Permissions",
		},
		{
			name: "workflow with job concurrency",
			yaml: `
name: Job Concurrency
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    concurrency: job-group
    steps:
      - run: echo hello
`,
			expectName: "Job Concurrency",
		},
		{
			name: "workflow with runs-on array",
			yaml: `
name: Runs On Array
on: push
jobs:
  build:
    runs-on: [self-hosted, linux]
    steps:
      - run: echo hello
`,
			expectName: "Runs On Array",
		},
		{
			name:        "invalid yaml",
			yaml:        `{{{invalid`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workflow, err := policy.ParseWorkflowBytes([]byte(tt.yaml), "test.yml")

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, workflow)
			assert.Equal(t, tt.expectName, workflow.Name)
			assert.Equal(t, "test.yml", workflow.Path)
		})
	}
}

// TestWorkflow_HasWritePermissions tests checking for write permissions.
func TestWorkflow_HasWritePermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		workflow  *policy.Workflow
		hasWrites bool
	}{
		{
			name: "no permissions",
			workflow: &policy.Workflow{
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: false,
		},
		{
			name: "read-only permissions",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					Contents: "read",
					Actions:  "read",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: false,
		},
		{
			name: "write-all permissions",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					All: "write-all",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: true,
		},
		{
			name: "contents write permission",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					Contents: "write",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: true,
		},
		{
			name: "actions write permission",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					Actions: "write",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: true,
		},
		{
			name: "checks write permission",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					Checks: "write",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: true,
		},
		{
			name: "deployments write permission",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					Deployments: "write",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: true,
		},
		{
			name: "issues write permission",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					Issues: "write",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: true,
		},
		{
			name: "packages write permission",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					Packages: "write",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: true,
		},
		{
			name: "pull-requests write permission",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					PullRequests: "write",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: true,
		},
		{
			name: "security-events write permission",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					SecurityEvents: "write",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: true,
		},
		{
			name: "statuses write permission",
			workflow: &policy.Workflow{
				Permissions: &policy.Permissions{
					Statuses: "write",
				},
				Jobs: map[string]*policy.Job{"build": {}},
			},
			hasWrites: true,
		},
		{
			name: "job-level write permission",
			workflow: &policy.Workflow{
				Jobs: map[string]*policy.Job{
					"build": {
						Permissions: &policy.Permissions{
							Contents: "write",
						},
					},
				},
			},
			hasWrites: true,
		},
		{
			name: "job-level read permission only",
			workflow: &policy.Workflow{
				Jobs: map[string]*policy.Job{
					"build": {
						Permissions: &policy.Permissions{
							Contents: "read",
						},
					},
				},
			},
			hasWrites: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.workflow.HasWritePermissions()
			assert.Equal(t, tt.hasWrites, result)
		})
	}
}

// TestParseWorkflowBytes_Triggers tests trigger parsing specifically.
func TestParseWorkflowBytes_Triggers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		yaml                 string
		hasPush              bool
		hasPullRequest       bool
		hasPullRequestTarget bool
		hasWorkflowDispatch  bool
	}{
		{
			name: "single push trigger",
			yaml: `
name: Test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo
`,
			hasPush: true,
		},
		{
			name: "array triggers",
			yaml: `
name: Test
on: [push, pull_request, workflow_dispatch]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo
`,
			hasPush:             true,
			hasPullRequest:      true,
			hasWorkflowDispatch: true,
		},
		{
			name: "map triggers",
			yaml: `
name: Test
on:
  push:
    branches: [main]
  pull_request_target:
    types: [opened]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo
`,
			hasPush:              true,
			hasPullRequestTarget: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workflow, err := policy.ParseWorkflowBytes([]byte(tt.yaml), "test.yml")
			require.NoError(t, err)
			require.NotNil(t, workflow.On)

			assert.Equal(t, tt.hasPush, workflow.On.Push != nil, "Push trigger mismatch")
			assert.Equal(t, tt.hasPullRequest, workflow.On.PullRequest != nil, "PullRequest trigger mismatch")
			assert.Equal(t, tt.hasPullRequestTarget, workflow.On.PullRequestTarget != nil, "PullRequestTarget trigger mismatch")
			assert.Equal(t, tt.hasWorkflowDispatch, workflow.On.WorkflowDispatch != nil, "WorkflowDispatch trigger mismatch")
		})
	}
}
