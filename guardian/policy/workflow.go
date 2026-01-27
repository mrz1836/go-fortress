package policy

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

// Workflow represents a parsed GitHub Actions workflow.
type Workflow struct {
	// Name is the workflow name.
	Name string

	// Path is the file path relative to repository root.
	Path string

	// Permissions declared at workflow level.
	Permissions *Permissions

	// Jobs in this workflow.
	Jobs map[string]*Job

	// On defines the workflow triggers.
	On *WorkflowTriggers

	// Concurrency settings if defined.
	Concurrency *Concurrency

	// Raw is the original YAML content.
	Raw []byte
}

// Job represents a workflow job.
type Job struct {
	Name        string
	Permissions *Permissions
	Steps       []*Step
	RunsOn      []string
	If          string
	Needs       []string
	Concurrency *Concurrency
}

// Step represents a job step.
type Step struct {
	ID   string
	Name string
	Uses string // Action reference (e.g., "actions/checkout@v4")
	Run  string
	With map[string]interface{}
	Env  map[string]string
	If   string
	Line int // Line number in workflow file
}

// Permissions defines GitHub token permissions.
type Permissions struct {
	Actions        string
	Checks         string
	Contents       string
	Deployments    string
	Issues         string
	Packages       string
	PullRequests   string
	SecurityEvents string
	Statuses       string
	All            string // For "read-all", "write-all", or "{}" format
}

// Concurrency defines concurrency settings.
type Concurrency struct {
	Group            string
	CancelInProgress bool
}

// WorkflowTriggers defines when the workflow runs.
type WorkflowTriggers struct {
	Push              *PushTrigger
	PullRequest       *PullRequestTrigger
	PullRequestTarget *PullRequestTrigger
	WorkflowDispatch  *WorkflowDispatchTrigger
	Schedule          []ScheduleTrigger
}

// PushTrigger defines push event configuration.
type PushTrigger struct {
	Branches []string
	Tags     []string
	Paths    []string
}

// PullRequestTrigger defines pull request event configuration.
type PullRequestTrigger struct {
	Branches []string
	Types    []string
	Paths    []string
}

// WorkflowDispatchTrigger defines manual trigger configuration.
type WorkflowDispatchTrigger struct {
	Inputs map[string]interface{}
}

// ScheduleTrigger defines cron schedule.
type ScheduleTrigger struct {
	Cron string
}

// ParseWorkflow reads and parses a workflow file.
func ParseWorkflow(path string) (*Workflow, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path from trusted validator input
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return ParseWorkflowBytes(data, path)
}

// ParseWorkflowBytes parses workflow content from bytes.
func ParseWorkflowBytes(data []byte, path string) (*Workflow, error) {
	var raw rawWorkflow
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	workflow := &Workflow{
		Name: raw.Name,
		Path: path,
		Raw:  data,
		Jobs: make(map[string]*Job),
	}

	// Parse permissions
	if raw.Permissions != nil {
		workflow.Permissions = parsePermissions(raw.Permissions)
	}

	// Parse triggers
	workflow.On = parseTriggers(raw.On)

	// Parse concurrency
	if raw.Concurrency != nil {
		workflow.Concurrency = parseConcurrency(raw.Concurrency)
	}

	// Parse jobs
	for name, job := range raw.Jobs {
		workflow.Jobs[name] = parseJob(job)
	}

	return workflow, nil
}

// rawWorkflow is the intermediate YAML structure.
type rawWorkflow struct {
	Name        string            `yaml:"name"`
	On          interface{}       `yaml:"on"`
	Permissions interface{}       `yaml:"permissions"`
	Concurrency interface{}       `yaml:"concurrency"`
	Jobs        map[string]rawJob `yaml:"jobs"`
	Env         map[string]string `yaml:"env"`
}

type rawJob struct {
	Name        string      `yaml:"name"`
	RunsOn      interface{} `yaml:"runs-on"`
	Permissions interface{} `yaml:"permissions"`
	Steps       []rawStep   `yaml:"steps"`
	If          string      `yaml:"if"`
	Needs       interface{} `yaml:"needs"`
	Concurrency interface{} `yaml:"concurrency"`
}

type rawStep struct {
	ID   string                 `yaml:"id"`
	Name string                 `yaml:"name"`
	Uses string                 `yaml:"uses"`
	Run  string                 `yaml:"run"`
	With map[string]interface{} `yaml:"with"`
	Env  map[string]string      `yaml:"env"`
	If   string                 `yaml:"if"`
}

func parsePermissions(raw interface{}) *Permissions {
	perm := &Permissions{}

	switch v := raw.(type) {
	case string:
		// "read-all", "write-all", or "{}"
		perm.All = v
	case map[string]interface{}:
		if s, ok := v["actions"].(string); ok {
			perm.Actions = s
		}
		if s, ok := v["checks"].(string); ok {
			perm.Checks = s
		}
		if s, ok := v["contents"].(string); ok {
			perm.Contents = s
		}
		if s, ok := v["deployments"].(string); ok {
			perm.Deployments = s
		}
		if s, ok := v["issues"].(string); ok {
			perm.Issues = s
		}
		if s, ok := v["packages"].(string); ok {
			perm.Packages = s
		}
		if s, ok := v["pull-requests"].(string); ok {
			perm.PullRequests = s
		}
		if s, ok := v["security-events"].(string); ok {
			perm.SecurityEvents = s
		}
		if s, ok := v["statuses"].(string); ok {
			perm.Statuses = s
		}
	}

	return perm
}

func parseConcurrency(raw interface{}) *Concurrency {
	conc := &Concurrency{}

	switch v := raw.(type) {
	case string:
		conc.Group = v
	case map[string]interface{}:
		if g, ok := v["group"].(string); ok {
			conc.Group = g
		}
		if c, ok := v["cancel-in-progress"].(bool); ok {
			conc.CancelInProgress = c
		}
	}

	return conc
}

func parseTriggers(raw interface{}) *WorkflowTriggers {
	triggers := &WorkflowTriggers{}

	switch v := raw.(type) {
	case string:
		// Simple trigger like "push"
		switch v {
		case "push":
			triggers.Push = &PushTrigger{}
		case "pull_request":
			triggers.PullRequest = &PullRequestTrigger{}
		case "pull_request_target":
			triggers.PullRequestTarget = &PullRequestTrigger{}
		case "workflow_dispatch":
			triggers.WorkflowDispatch = &WorkflowDispatchTrigger{}
		}
	case []interface{}:
		// Array of triggers
		for _, t := range v {
			if s, ok := t.(string); ok {
				switch s {
				case "push":
					triggers.Push = &PushTrigger{}
				case "pull_request":
					triggers.PullRequest = &PullRequestTrigger{}
				case "pull_request_target":
					triggers.PullRequestTarget = &PullRequestTrigger{}
				case "workflow_dispatch":
					triggers.WorkflowDispatch = &WorkflowDispatchTrigger{}
				}
			}
		}
	case map[string]interface{}:
		// Detailed trigger configuration
		if _, ok := v["push"]; ok {
			triggers.Push = &PushTrigger{}
		}
		if _, ok := v["pull_request"]; ok {
			triggers.PullRequest = &PullRequestTrigger{}
		}
		if _, ok := v["pull_request_target"]; ok {
			triggers.PullRequestTarget = &PullRequestTrigger{}
		}
		if _, ok := v["workflow_dispatch"]; ok {
			triggers.WorkflowDispatch = &WorkflowDispatchTrigger{}
		}
	}

	return triggers
}

func parseJob(raw rawJob) *Job {
	job := &Job{
		Name:  raw.Name,
		Steps: make([]*Step, 0, len(raw.Steps)),
		If:    raw.If,
	}

	// Parse runs-on
	switch v := raw.RunsOn.(type) {
	case string:
		job.RunsOn = []string{v}
	case []interface{}:
		for _, r := range v {
			if s, ok := r.(string); ok {
				job.RunsOn = append(job.RunsOn, s)
			}
		}
	}

	// Parse permissions
	if raw.Permissions != nil {
		job.Permissions = parsePermissions(raw.Permissions)
	}

	// Parse needs
	switch v := raw.Needs.(type) {
	case string:
		job.Needs = []string{v}
	case []interface{}:
		for _, n := range v {
			if s, ok := n.(string); ok {
				job.Needs = append(job.Needs, s)
			}
		}
	}

	// Parse concurrency
	if raw.Concurrency != nil {
		job.Concurrency = parseConcurrency(raw.Concurrency)
	}

	// Parse steps
	for _, s := range raw.Steps {
		job.Steps = append(job.Steps, &Step{
			ID:   s.ID,
			Name: s.Name,
			Uses: s.Uses,
			Run:  s.Run,
			With: s.With,
			Env:  s.Env,
			If:   s.If,
		})
	}

	return job
}

// HasWritePermissions checks if the workflow has write permissions.
func (w *Workflow) HasWritePermissions() bool {
	if w.Permissions != nil {
		if w.Permissions.All == "write-all" {
			return true
		}
		if hasWritePerm(w.Permissions) {
			return true
		}
	}

	for _, job := range w.Jobs {
		if job.Permissions != nil && hasWritePerm(job.Permissions) {
			return true
		}
	}

	return false
}

func hasWritePerm(p *Permissions) bool {
	return p.Contents == "write" ||
		p.Actions == "write" ||
		p.Checks == "write" ||
		p.Deployments == "write" ||
		p.Issues == "write" ||
		p.Packages == "write" ||
		p.PullRequests == "write" ||
		p.SecurityEvents == "write" ||
		p.Statuses == "write"
}
