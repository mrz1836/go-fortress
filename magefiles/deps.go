//go:build mage

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mg"
)

// errToolsMissing is returned when required tools are missing or have wrong versions.
var errToolsMissing = errors.New("some tools are missing or have wrong versions; run 'magex deps:install'")

// Deps is the namespace for dependency management commands.
type Deps mg.Namespace

// Tool versions - these should match .env.base
const (
	ActVersion        = "v0.2.84"
	ActionlintVersion = "v1.6.27"
)

// Install installs all required tools for the project.
func (Deps) Install(ctx context.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, "Installing project dependencies...")

	// Install act
	if err := installGoTool(ctx, "github.com/nektos/act", ActVersion); err != nil {
		return fmt.Errorf("installing act: %w", err)
	}

	// Install actionlint
	if err := installGoTool(ctx, "github.com/rhysd/actionlint/cmd/actionlint", ActionlintVersion); err != nil {
		return fmt.Errorf("installing actionlint: %w", err)
	}

	_, _ = fmt.Fprintln(os.Stdout, "All dependencies installed successfully!")

	return nil
}

// Check verifies all required tools are installed with correct versions.
func (Deps) Check(ctx context.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, "Checking tool versions...")

	checks := []struct {
		name    string
		cmd     string
		args    []string
		version string
	}{
		{"act", "act", []string{"--version"}, ActVersion},
		{"actionlint", "actionlint", []string{"--version"}, ActionlintVersion},
	}

	allOK := true

	for _, check := range checks {
		installed, err := getToolVersion(ctx, check.cmd, check.args)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "  %s: NOT INSTALLED\n", check.name)
			allOK = false

			continue
		}

		if strings.Contains(installed, strings.TrimPrefix(check.version, "v")) {
			_, _ = fmt.Fprintf(os.Stdout, "  %s: %s ✓\n", check.name, check.version)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "  %s: %s (expected %s) ✗\n", check.name, installed, check.version)
			allOK = false
		}
	}

	if !allOK {
		return errToolsMissing
	}

	return nil
}

// Update updates all tools to their pinned versions.
func (Deps) Update(ctx context.Context) error {
	_, _ = fmt.Fprintln(os.Stdout, "Updating tools to pinned versions...")

	return Deps{}.Install(ctx)
}

// installGoTool installs a Go tool at a specific version.
func installGoTool(ctx context.Context, pkg, version string) error {
	pkgWithVersion := pkg + "@" + version
	_, _ = fmt.Fprintf(os.Stdout, "  Installing %s...\n", pkgWithVersion)

	cmd := exec.CommandContext(ctx, "go", "install", pkgWithVersion)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// getToolVersion returns the installed version of a tool.
func getToolVersion(ctx context.Context, cmd string, args []string) (string, error) {
	c := exec.CommandContext(ctx, cmd, args...)
	output, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
