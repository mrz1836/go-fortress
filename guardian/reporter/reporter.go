package reporter

import (
	"context"
	"io"
	"os"
)

// Reporter generates output in a specific format.
type Reporter interface {
	// Name returns the reporter identifier (e.g., "jsonl", "sarif").
	Name() string

	// Write outputs the report to the given writer.
	Write(ctx context.Context, report *Report, w io.Writer) error

	// WriteFile outputs the report to a file.
	WriteFile(ctx context.Context, report *Report, path string) error
}

// Registry manages available reporters.
type Registry struct {
	reporters map[string]Reporter
	order     []string
}

// NewRegistry creates a new reporter registry.
func NewRegistry() *Registry {
	return &Registry{
		reporters: make(map[string]Reporter),
		order:     []string{},
	}
}

// Register adds a reporter.
func (r *Registry) Register(rep Reporter) {
	name := rep.Name()
	if _, exists := r.reporters[name]; !exists {
		r.order = append(r.order, name)
	}
	r.reporters[name] = rep
}

// Get returns a reporter by name.
func (r *Registry) Get(name string) (Reporter, bool) {
	rep, ok := r.reporters[name]
	return rep, ok
}

// All returns all registered reporters in registration order.
func (r *Registry) All() []Reporter {
	result := make([]Reporter, 0, len(r.order))
	for _, name := range r.order {
		result = append(result, r.reporters[name])
	}
	return result
}

// IsCI detects if running in a CI environment.
func IsCI() bool {
	// Check for common CI environment variables
	ciVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"CIRCLECI",
		"TRAVIS",
		"JENKINS_URL",
		"BUILDKITE",
	}

	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}

// IsGitHubActions detects if running in GitHub Actions.
func IsGitHubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}
