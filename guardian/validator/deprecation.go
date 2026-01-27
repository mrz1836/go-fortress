package validator

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v4"
)

// DeprecationValidator detects deprecated actions and runner labels.
type DeprecationValidator struct{}

// NewDeprecationValidator creates a new deprecation validator.
func NewDeprecationValidator() *DeprecationValidator {
	return &DeprecationValidator{}
}

// Name returns the validator identifier.
func (v *DeprecationValidator) Name() string {
	return "deprecation"
}

// Validate checks for deprecated actions and runners in workflow files.
func (v *DeprecationValidator) Validate(_ context.Context, workflowPath string) ([]Finding, error) {
	data, err := os.ReadFile(workflowPath) //nolint:gosec // path from trusted validator input
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var workflow workflowYAML
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		// Skip files that can't be parsed (intentional: not all yaml files are workflows)
		return nil, nil //nolint:nilerr // intentionally skip unparseable files
	}

	var findings []Finding

	for jobName, job := range workflow.Jobs {
		// Check runner labels
		findings = append(findings, v.checkRunnerLabels(workflowPath, jobName, job.RunsOn)...)

		// Check action references in steps
		for i, step := range job.Steps {
			if step.Uses != "" {
				findings = append(findings, v.checkActionDeprecation(workflowPath, jobName, i+1, step.Uses)...)
			}
		}
	}

	return findings, nil
}

// workflowYAML is a minimal workflow structure for deprecation checking.
type workflowYAML struct {
	Jobs map[string]jobYAML `yaml:"jobs"`
}

type jobYAML struct {
	RunsOn interface{} `yaml:"runs-on"` // Can be string or array
	Steps  []stepYAML  `yaml:"steps"`
}

type stepYAML struct {
	Uses string `yaml:"uses"`
}

// Deprecated runners and their replacements.
//
//nolint:gochecknoglobals // intentional lookup table
var deprecatedRunners = map[string]string{
	"ubuntu-18.04": "ubuntu-22.04 or ubuntu-24.04",
	"ubuntu-16.04": "ubuntu-22.04 or ubuntu-24.04",
	"macos-10.15":  "macos-14 or macos-15",
	"macos-11":     "macos-14 or macos-15",
	"windows-2016": "windows-2022 or windows-2025",
	"windows-2019": "windows-2022 or windows-2025",
}

// Deprecated actions and their replacements.
//
//nolint:gochecknoglobals // intentional lookup table
var deprecatedActions = map[string]deprecatedAction{
	"actions/checkout@v1": {
		replacement: "actions/checkout@v4",
		reason:      "v1 uses deprecated Node.js 12",
	},
	"actions/checkout@v2": {
		replacement: "actions/checkout@v4",
		reason:      "v2 uses deprecated Node.js 12",
	},
	"actions/checkout@v3": {
		replacement: "actions/checkout@v4",
		reason:      "v3 uses deprecated Node.js 16",
	},
	"actions/setup-node@v1": {
		replacement: "actions/setup-node@v4",
		reason:      "v1 uses deprecated Node.js 12",
	},
	"actions/setup-node@v2": {
		replacement: "actions/setup-node@v4",
		reason:      "v2 uses deprecated Node.js 12",
	},
	"actions/setup-node@v3": {
		replacement: "actions/setup-node@v4",
		reason:      "v3 uses deprecated Node.js 16",
	},
	"actions/setup-go@v1": {
		replacement: "actions/setup-go@v5",
		reason:      "v1 uses deprecated Node.js 12",
	},
	"actions/setup-go@v2": {
		replacement: "actions/setup-go@v5",
		reason:      "v2 uses deprecated Node.js 12",
	},
	"actions/setup-go@v3": {
		replacement: "actions/setup-go@v5",
		reason:      "v3 uses deprecated Node.js 16",
	},
	"actions/setup-go@v4": {
		replacement: "actions/setup-go@v5",
		reason:      "v4 uses deprecated Node.js 16",
	},
	"actions/setup-python@v1": {
		replacement: "actions/setup-python@v5",
		reason:      "v1 uses deprecated Node.js 12",
	},
	"actions/setup-python@v2": {
		replacement: "actions/setup-python@v5",
		reason:      "v2 uses deprecated Node.js 12",
	},
	"actions/setup-python@v3": {
		replacement: "actions/setup-python@v5",
		reason:      "v3 uses deprecated Node.js 16",
	},
	"actions/setup-python@v4": {
		replacement: "actions/setup-python@v5",
		reason:      "v4 uses deprecated Node.js 16",
	},
	"actions/cache@v1": {
		replacement: "actions/cache@v4",
		reason:      "v1 uses deprecated Node.js 12",
	},
	"actions/cache@v2": {
		replacement: "actions/cache@v4",
		reason:      "v2 uses deprecated Node.js 12",
	},
	"actions/cache@v3": {
		replacement: "actions/cache@v4",
		reason:      "v3 uses deprecated Node.js 16",
	},
	"actions/upload-artifact@v1": {
		replacement: "actions/upload-artifact@v4",
		reason:      "v1 uses deprecated Node.js 12 and artifact format",
	},
	"actions/upload-artifact@v2": {
		replacement: "actions/upload-artifact@v4",
		reason:      "v2 uses deprecated Node.js 12 and artifact format",
	},
	"actions/upload-artifact@v3": {
		replacement: "actions/upload-artifact@v4",
		reason:      "v3 uses deprecated artifact format",
	},
	"actions/download-artifact@v1": {
		replacement: "actions/download-artifact@v4",
		reason:      "v1 uses deprecated Node.js 12 and artifact format",
	},
	"actions/download-artifact@v2": {
		replacement: "actions/download-artifact@v4",
		reason:      "v2 uses deprecated Node.js 12 and artifact format",
	},
	"actions/download-artifact@v3": {
		replacement: "actions/download-artifact@v4",
		reason:      "v3 uses deprecated artifact format",
	},
	"set-output": {
		replacement: "$GITHUB_OUTPUT",
		reason:      "set-output workflow command is deprecated",
	},
	"save-state": {
		replacement: "$GITHUB_STATE",
		reason:      "save-state workflow command is deprecated",
	},
}

type deprecatedAction struct {
	replacement string
	reason      string
}

// checkRunnerLabels checks for deprecated runner labels.
func (v *DeprecationValidator) checkRunnerLabels(path, jobName string, runsOn interface{}) []Finding {
	var findings []Finding
	var labels []string

	switch r := runsOn.(type) {
	case string:
		labels = []string{r}
	case []interface{}:
		for _, l := range r {
			if s, ok := l.(string); ok {
				labels = append(labels, s)
			}
		}
	}

	for _, label := range labels {
		if replacement, deprecated := deprecatedRunners[label]; deprecated {
			findings = append(findings, Finding{
				RuleID:     "deprecation/runner",
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("runner '%s' is deprecated in job '%s'", label, jobName),
				File:       path,
				Line:       1, // Line number not easily available from YAML
				Source:     SourceDeprecation,
				Suggestion: fmt.Sprintf("Use %s instead", replacement),
			})
		}
	}

	return findings
}

// checkActionDeprecation checks if an action reference is deprecated.
func (v *DeprecationValidator) checkActionDeprecation(path, jobName string, stepNum int, uses string) []Finding {
	var findings []Finding

	// Normalize the action reference (remove SHA if present)
	actionRef := normalizeActionRef(uses)

	if info, deprecated := deprecatedActions[actionRef]; deprecated {
		findings = append(findings, Finding{
			RuleID:     "deprecation/action",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("action '%s' is deprecated in job '%s' step %d: %s", uses, jobName, stepNum, info.reason),
			File:       path,
			Line:       1,
			Source:     SourceDeprecation,
			Suggestion: fmt.Sprintf("Upgrade to %s", info.replacement),
		})
	}

	return findings
}

// actionRefPattern matches action references like "owner/repo@version" or "owner/repo/path@version"
var actionRefPattern = regexp.MustCompile(`^([^@]+)@(.+)$`)

// normalizeActionRef extracts the action@version format, removing SHA pins.
func normalizeActionRef(uses string) string {
	matches := actionRefPattern.FindStringSubmatch(uses)
	if len(matches) != 3 {
		return uses
	}

	action := matches[1]
	version := matches[2]

	// If version looks like a SHA (40 hex chars), we can't determine deprecation
	if len(version) == 40 && isHex(version) {
		// Can't determine version from SHA
		return ""
	}

	// Normalize version: extract major version for comparison
	// e.g., "v4.1.0" -> "v4"
	if strings.HasPrefix(version, "v") {
		parts := strings.Split(version, ".")
		if len(parts) > 0 {
			version = parts[0]
		}
	}

	return action + "@" + version
}

// isHex checks if a string contains only hexadecimal characters.
func isHex(s string) bool {
	for _, c := range s {
		isDigit := c >= '0' && c <= '9'
		isLowerHex := c >= 'a' && c <= 'f'
		isUpperHex := c >= 'A' && c <= 'F'

		if !isDigit && !isLowerHex && !isUpperHex {
			return false
		}
	}

	return true
}
