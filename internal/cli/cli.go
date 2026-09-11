package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/home"
)

// Run is the CLI entrypoint: main passes argv and the injected build version,
// and Run returns the process exit code.
//
// Round 001 (narrow foundation): parse flags, then either report the build
// version, run the setup diagnostic (reporting path), or run the boot path
// (resolve the runtime home, configuration, and per-mode workspace, then report
// readiness).
func Run(args []string, version string) int {
	fs := pflag.NewFlagSet("tellme", pflag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	configPath := fs.StringP("config", "c", "", "Path to the YAML configuration file.")
	diagnostic := fs.BoolP("diagnostics", "d", false, "Report configuration and home resolution, then exit.")
	asJSON := fs.Bool("json", false, "Emit machine-readable output for the diagnostic.")
	showVersion := fs.Bool("version", false, "Print the build version and exit.")
	if err := fs.Parse(args); err != nil {
		return emitUsageError()
	}

	// Version is independent of setup: report and exit (FR-010).
	if *showVersion {
		fmt.Printf("tellme %s\n", version)
		return Success
	}

	// Step 1 — resolve TELL_ME_HOME first, always (FR-006).
	homeDir := os.Getenv("TELL_ME_HOME")

	// -d is the reporting path: it always produces a report (Decision 2).
	if *diagnostic {
		return runDiagnostic(homeDir, *configPath, *asJSON)
	}

	if homeDir == "" {
		fmt.Fprintln(os.Stderr, "tellme: the runtime home is not usable")
		return EnvironmentError
	}

	// Step 3 — the config path: -c when given, else the default for the mode seed.
	explicit := *configPath != ""
	path := *configPath
	if !explicit {
		path = defaultConfigPath(homeDir)
	}

	// Step 4 — load + validate the file (FR-002, FR-005).
	cfg, err := config.Load(path)
	if err != nil {
		return emitConfigLoadError(path, explicit, err)
	}

	// Step 5 — validate the effective selected provider against the registry
	// (FR-003).
	selected := cfg.EffectiveSelectedProvider(os.Getenv("TELL_ME_SELECTED_PROVIDER"))
	if !cfg.ProviderInRegistry(selected) {
		fmt.Fprintf(os.Stderr, "tellme: the selected provider is not in the registry (%q)\n", selected)
		return ConfigError
	}

	// Step 6 — resolve the effective mode and prepare the session workspace
	// (FR-007, FR-008, FR-009).
	mode := cfg.EffectiveMode(os.Getenv("TELL_ME_MODE"))
	workspace, err := home.EnsureWorkspace(homeDir, mode)
	if err != nil {
		return emitWorkspaceError(workspace, err)
	}

	fmt.Println("configuration: ready")
	fmt.Println("session workspace: " + workspace.Path)
	return Success
}

// diagnostic is the resolved-or-unresolved report produced by -d (Decision 2).
type diagnostic struct {
	resolved  bool
	reason    string // one of the pinned categories when not resolved
	home      string
	workspace string
}

// The pinned unresolved reason categories (specs/truth/features/cli/diagnostics/dsl.md).
const (
	reasonHomeUnset        = "home-unset"
	reasonHomeUnusable     = "home-unusable"
	reasonConfigMissing    = "config-missing"
	reasonConfigInvalid    = "config-invalid"
	reasonProviderMismatch = "provider-mismatch"
)

// runDiagnostic produces the setup report and returns its exit code: 0 when
// resolution succeeded, else the dedicated diagnostic "unresolved" code.
func runDiagnostic(homeDir, configPath string, asJSON bool) int {
	result := resolveDiagnostic(homeDir, configPath)

	if asJSON {
		emitDiagnosticJSON(result)
	} else {
		emitDiagnosticText(result)
	}

	if !result.resolved {
		return DiagnosticUnresolvedError
	}
	return Success
}

// resolveDiagnostic resolves home → configuration → selected provider →
// workspace, capturing the first failure as a pinned reason category.
func resolveDiagnostic(homeDir, configPath string) diagnostic {
	if homeDir == "" {
		return diagnostic{reason: reasonHomeUnset}
	}
	result := diagnostic{home: homeDir}

	explicit := configPath != ""
	path := configPath
	if !explicit {
		path = defaultConfigPath(homeDir)
	}
	cfg, err := config.Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			result.reason = reasonConfigMissing
		} else {
			result.reason = reasonConfigInvalid
		}
		return result
	}

	selected := cfg.EffectiveSelectedProvider(os.Getenv("TELL_ME_SELECTED_PROVIDER"))
	if !cfg.ProviderInRegistry(selected) {
		result.reason = reasonProviderMismatch
		return result
	}

	mode := cfg.EffectiveMode(os.Getenv("TELL_ME_MODE"))
	workspace, err := home.EnsureWorkspace(homeDir, mode)
	if err != nil {
		result.reason = reasonHomeUnusable
		return result
	}

	result.resolved = true
	result.workspace = workspace.Path
	return result
}

// emitDiagnosticText writes the plain-text report.
func emitDiagnosticText(result diagnostic) {
	fmt.Println("tellme diagnostic")
	if result.resolved {
		fmt.Println("configuration: resolved")
		fmt.Println("runtime_home: " + result.home)
		fmt.Println("session_workspace: " + result.workspace)
		return
	}
	fmt.Println("configuration: unresolved")
	fmt.Println("reason: " + result.reason)
}

// emitDiagnosticJSON writes the pinned structured report.
func emitDiagnosticJSON(result diagnostic) {
	obj := map[string]string{}
	if result.resolved {
		obj["status"] = "resolved"
		obj["runtime_home"] = result.home
		obj["session_workspace"] = result.workspace
	} else {
		obj["status"] = "unresolved"
		obj["reason"] = result.reason
	}
	out, err := json.Marshal(obj)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tellme: the diagnostic report could not be produced")
		return
	}
	fmt.Println(string(out))
}

// defaultConfigPath is the default configuration path for the effective mode
// seed: $TELL_ME_HOME/configs/<seed>.yaml, where the seed is TELL_ME_MODE or
// "butler" (FR-007 discovery seed only).
func defaultConfigPath(homeDir string) string {
	seed := os.Getenv("TELL_ME_MODE")
	if seed == "" {
		seed = "butler"
	}
	return filepath.Join(homeDir, "configs", seed+".yaml")
}

// emitConfigLoadError writes the actionable stderr message for a configuration
// load failure and returns the configuration error code. A missing file is
// distinguished from a malformed one; a missing default (no -c) reads
// differently from a missing explicit path.
func emitConfigLoadError(path string, explicit bool, err error) int {
	switch {
	case errors.Is(err, os.ErrNotExist) && explicit:
		fmt.Fprintf(os.Stderr, "tellme: the configuration could not be found at %s\n", path)
	case errors.Is(err, os.ErrNotExist):
		fmt.Fprintf(os.Stderr, "tellme: no configuration could be found at %s\n", path)
	default:
		fmt.Fprintf(os.Stderr, "tellme: the configuration could not be parsed at %s\n", path)
	}
	return ConfigError
}

// emitWorkspaceError writes the actionable stderr message for a workspace
// preparation failure and returns the environment error code.
func emitWorkspaceError(workspace home.Workspace, err error) int {
	if errors.Is(err, home.ErrNotDirectory) {
		fmt.Fprintf(os.Stderr, "tellme: the workspace path is not a directory (%s)\n", workspace.Path)
	} else {
		fmt.Fprintf(os.Stderr, "tellme: the runtime home is not usable (%v)\n", err)
	}
	return EnvironmentError
}

// emitUsageError writes the usage-error stderr message and returns the usage
// error code (FR-014).
func emitUsageError() int {
	fmt.Fprintln(os.Stderr, "tellme: the command-line usage is invalid")
	return UsageError
}
