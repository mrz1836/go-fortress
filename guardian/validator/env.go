package validator

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// EnvValidator validates .env.base files for schema compliance.
// Implements FR-013: .env.base schema validation.
type EnvValidator struct{}

// NewEnvValidator creates a new env validator.
func NewEnvValidator() *EnvValidator {
	return &EnvValidator{}
}

// Name returns the validator identifier.
func (v *EnvValidator) Name() string {
	return "env"
}

// Validate analyzes .env.base files for schema compliance.
func (v *EnvValidator) Validate(_ context.Context, path string) ([]Finding, error) {
	// Only validate .env.base files
	if !isEnvBaseFile(path) {
		return nil, nil
	}

	file, err := os.Open(path) //nolint:gosec // path is validated via isEnvBaseFile
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	defer func() { _ = file.Close() }()

	var findings []Finding
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var currentSection string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Track section headers (comments with special formatting)
		if strings.HasPrefix(trimmed, "#") {
			if strings.Contains(trimmed, "===") || strings.Contains(trimmed, "---") {
				// This looks like a section header
				currentSection = extractSectionName(trimmed)
			}
			continue
		}

		// Parse variable definition
		findings = append(findings, v.validateEnvLine(path, lineNum, trimmed, currentSection)...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return findings, nil
}

// envVarPattern matches environment variable definitions.
var envVarPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)

// validateEnvLine checks a single line for compliance.
func (v *EnvValidator) validateEnvLine(path string, lineNum int, line, section string) []Finding {
	var findings []Finding

	matches := envVarPattern.FindStringSubmatch(line)
	if len(matches) != 3 {
		// Not a valid variable definition
		findings = append(findings, Finding{
			RuleID:     "env/invalid-syntax",
			Severity:   SeverityError,
			Message:    "invalid environment variable syntax",
			File:       path,
			Line:       lineNum,
			Source:     SourceEnv,
			Suggestion: "Use format: VARIABLE_NAME=value",
		})
		return findings
	}

	varName := matches[1]
	varValue := matches[2]

	// Check naming conventions
	findings = append(findings, v.checkNamingConvention(path, lineNum, varName, section)...)

	// Check value types
	findings = append(findings, v.checkValueType(path, lineNum, varName, varValue)...)

	return findings
}

// checkNamingConvention validates variable naming follows conventions.
func (v *EnvValidator) checkNamingConvention(path string, lineNum int, varName, _ string) []Finding {
	var findings []Finding

	// Must be UPPER_SNAKE_CASE
	if !isUpperSnakeCase(varName) {
		findings = append(findings, Finding{
			RuleID:     "env/naming-convention",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("variable '%s' should be UPPER_SNAKE_CASE", varName),
			File:       path,
			Line:       lineNum,
			Source:     SourceEnv,
			Suggestion: fmt.Sprintf("Rename to %s", toUpperSnakeCase(varName)),
		})
	}

	// Version variables should end with _VERSION
	if isLikelyVersion(varName) && !strings.HasSuffix(varName, "_VERSION") {
		findings = append(findings, Finding{
			RuleID:     "env/version-naming",
			Severity:   SeverityNote,
			Message:    fmt.Sprintf("variable '%s' looks like a version; consider naming it with _VERSION suffix", varName),
			File:       path,
			Line:       lineNum,
			Source:     SourceEnv,
			Suggestion: "Use *_VERSION suffix for version variables",
		})
	}

	return findings
}

// checkValueType validates the value type is appropriate.
func (v *EnvValidator) checkValueType(path string, lineNum int, varName, varValue string) []Finding {
	var findings []Finding

	// Check for empty values on required-looking variables
	if varValue == "" && looksRequired(varName) {
		findings = append(findings, Finding{
			RuleID:     "env/empty-required",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("variable '%s' appears to be required but has no value", varName),
			File:       path,
			Line:       lineNum,
			Source:     SourceEnv,
			Suggestion: "Set a default value or mark as optional",
		})
	}

	// Check for unquoted values with spaces
	if strings.Contains(varValue, " ") && !isQuoted(varValue) {
		findings = append(findings, Finding{
			RuleID:     "env/unquoted-spaces",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("variable '%s' contains spaces but is not quoted", varName),
			File:       path,
			Line:       lineNum,
			Source:     SourceEnv,
			Suggestion: "Wrap value in quotes",
		})
	}

	// Check boolean values are consistent
	if isBooleanVar(varName) && !isValidBooleanValue(varValue) {
		findings = append(findings, Finding{
			RuleID:     "env/boolean-value",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("variable '%s' appears to be boolean but has value '%s'", varName, varValue),
			File:       path,
			Line:       lineNum,
			Source:     SourceEnv,
			Suggestion: "Use 'true' or 'false'",
		})
	}

	return findings
}

// Helper functions

func isEnvBaseFile(path string) bool {
	base := filepath.Base(path)
	return base == ".env.base"
}

func extractSectionName(comment string) string {
	// Remove comment markers and decorators
	clean := strings.TrimLeft(comment, "# ")
	clean = strings.Trim(clean, "=- ")
	return clean
}

func isUpperSnakeCase(s string) bool {
	for _, c := range s {
		isUpper := c >= 'A' && c <= 'Z'
		isDigit := c >= '0' && c <= '9'
		isUnderscore := c == '_'

		if !isUpper && !isDigit && !isUnderscore {
			return false
		}
	}

	return true
}

func toUpperSnakeCase(s string) string {
	var result strings.Builder
	for i, c := range s {
		if c >= 'a' && c <= 'z' {
			result.WriteRune(c - 32) // Convert to uppercase
		} else if c >= 'A' && c <= 'Z' {
			if i > 0 && s[i-1] >= 'a' && s[i-1] <= 'z' {
				result.WriteRune('_')
			}
			result.WriteRune(c)
		} else if c == '-' {
			result.WriteRune('_')
		} else {
			result.WriteRune(c)
		}
	}
	return result.String()
}

func isLikelyVersion(varName string) bool {
	// Check if variable name suggests it holds a version
	return strings.Contains(strings.ToLower(varName), "ver") &&
		!strings.HasSuffix(varName, "_VERSION")
}

func looksRequired(varName string) bool {
	// Certain prefixes suggest required values
	required := []string{"API_", "DB_", "SECRET_", "KEY_", "TOKEN_", "URL_", "HOST_", "PORT_"}
	for _, prefix := range required {
		if strings.HasPrefix(varName, prefix) {
			return true
		}
	}
	return false
}

func isQuoted(value string) bool {
	return (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'"))
}

func isBooleanVar(varName string) bool {
	return strings.HasPrefix(varName, "ENABLE_") ||
		strings.HasPrefix(varName, "DISABLE_") ||
		strings.HasSuffix(varName, "_ENABLED") ||
		strings.HasSuffix(varName, "_DISABLED")
}

func isValidBooleanValue(value string) bool {
	lower := strings.ToLower(value)
	validBools := []string{"true", "false", "1", "0", "yes", "no", "on", "off"}
	for _, v := range validBools {
		if lower == v {
			return true
		}
	}
	return false
}
