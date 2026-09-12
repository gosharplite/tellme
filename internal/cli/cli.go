package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/home"
)

// The pinned unresolved reason categories
// (specs/truth/features/cli/diagnostics/dsl.md).
const (
	reasonHomeUnset        = "home-unset"
	reasonHomeUnusable     = "home-unusable"
	reasonConfigMissing    = "config-missing"
	reasonConfigInvalid    = "config-invalid"
	reasonProviderMismatch = "provider-mismatch"
	reasonProviderInvalid  = "provider-invalid"
)

// options are the parsed CLI flags.
type options struct {
	configPath string
	diagnostic bool
	version    bool
}

// resolution is the outcome of resolving home → configuration → workspace. On a
// resolve failure the partially populated fields (Home, Path, Selected,
// Provider, Workspace) are still returned so a renderer can produce an
// actionable message.
type resolution struct {
	Home     string
	Path     string // the configuration path (explicit or defaulted)
	Explicit bool   // whether -c was given
	Selected string // the effective selected provider (when reached)
	// Provider is the resolved (variable-expanded) selected provider entry. It
	// is carried here so Slice 004 can construct the provider transport without
	// re-loading or re-parsing the configuration (review finding #3).
	Provider  config.Provider
	Mode      string // the effective mode (when reached)
	Workspace string // the resolved workspace path (when reached)
}

// resolveError carries the pinned reason category plus the underlying cause.
type resolveError struct {
	Reason string
	Err    error
}

func (e *resolveError) Error() string {
	if e.Err != nil {
		return e.Reason + ": " + e.Err.Error()
	}
	return e.Reason
}

func (e *resolveError) Unwrap() error { return e.Err }

// Run is the CLI entrypoint: main passes argv and the injected build version,
// and Run returns the process exit code. It parses flags, then dispatches to the
// version path, the diagnostic reporting path, or the boot path.
func Run(args []string, version string) int {
	opts, ok := parseFlags(args)
	if !ok {
		return emitUsageError()
	}
	if opts.version {
		fmt.Printf("tellme %s\n", version)
		return Success
	}

	homeDir := os.Getenv("TELL_ME_HOME")

	// -d is the reporting path: it always produces a report (Decision 2).
	if opts.diagnostic {
		return renderDiagnostic(homeDir, opts.configPath)
	}
	return renderBoot(homeDir, opts.configPath)
}

// parseFlags parses argv. ok is false on an unrecognized or invalid flag.
func parseFlags(args []string) (opts *options, ok bool) {
	fs := pflag.NewFlagSet("tellme", pflag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	o := &options{}
	fs.StringVarP(&o.configPath, "config", "c", "", "Path to the YAML configuration file.")
	fs.BoolVarP(&o.diagnostic, "diagnostics", "d", false, "Report configuration and home resolution, then exit.")
	fs.BoolVar(&o.version, "version", false, "Print the build version and exit.")
	if err := fs.Parse(args); err != nil {
		return nil, false
	}
	return o, true
}

// resolve is the single resolution algorithm shared by the boot path and the
// diagnostic path: home → config path → load/validate → effective selected
// provider → effective mode → workspace. On failure it returns a *resolveError
// carrying the pinned reason category; the two callers differ only in how they
// render it (boot message + code vs. diagnostic report).
func resolve(homeDir, configPath string) (resolution, *resolveError) {
	res := resolution{Home: homeDir, Path: configPath, Explicit: configPath != ""}

	// Step 1 — resolve TELL_ME_HOME first, always (FR-006).
	if homeDir == "" {
		return res, &resolveError{Reason: reasonHomeUnset}
	}

	// Step 3 — the config path: -c when given, else the default for the mode seed.
	if !res.Explicit {
		res.Path = defaultConfigPath(homeDir)
	}

	// Step 4 — load + validate the file (FR-002, FR-005).
	cfg, err := config.Load(res.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return res, &resolveError{Reason: reasonConfigMissing, Err: err}
		}
		return res, &resolveError{Reason: reasonConfigInvalid, Err: err}
	}

	// Step 5 — the effective selected provider must be in the registry (FR-003).
	res.Selected = cfg.EffectiveSelectedProvider(os.Getenv("TELL_ME_SELECTED_PROVIDER"))
	if !cfg.ProviderInRegistry(res.Selected) {
		return res, &resolveError{Reason: reasonProviderMismatch}
	}

	// Step 5b — resolve the selected provider entry: expand ${VAR} placeholders
	// FIRST, then validate the RESOLVED state (FR-001..FR-009).
	//
	// Ordering matters (review finding #2): validating before expansion would let
	// a mandatory field whose value is a placeholder that resolves to empty
	// (e.g. `URL: "${UNSET_ENDPOINT:-}"`) pass the non-empty invariant and then
	// degrade to "" — evading validation. Expanding first makes Validate() see
	// the final, post-substitution value, and guarantees the Provider carried for
	// Slice 004 is the fully-expanded one.
	prov := cfg.Providers[res.Selected]
	if err := prov.Expand(); err != nil {
		return res, &resolveError{Reason: reasonProviderInvalid, Err: err}
	}
	if err := prov.Validate(); err != nil {
		return res, &resolveError{Reason: reasonProviderInvalid, Err: err}
	}
	cfg.Providers[res.Selected] = prov
	res.Provider = prov

	// Step 6 — effective mode + prepare the session workspace (FR-007/008/009).
	res.Mode = cfg.EffectiveMode(os.Getenv("TELL_ME_MODE"))
	workspace, err := home.EnsureWorkspace(homeDir, res.Mode)
	res.Workspace = workspace.Path
	if err != nil {
		return res, &resolveError{Reason: reasonHomeUnusable, Err: err}
	}
	return res, nil
}

// renderBoot runs the boot path and reports readiness or an actionable error.
func renderBoot(homeDir, configPath string) int {
	res, rerr := resolve(homeDir, configPath)
	if rerr != nil {
		return emitBootError(res, rerr)
	}
	fmt.Println("configuration: ready")
	fmt.Println("session workspace: " + res.Workspace)
	return Success
}

// emitBootError maps a resolve failure to its actionable stderr message + code.
func emitBootError(res resolution, rerr *resolveError) int {
	switch rerr.Reason {
	case reasonConfigMissing:
		if res.Explicit {
			fmt.Fprintf(os.Stderr, "tellme: the configuration could not be found at %s\n", res.Path)
		} else {
			fmt.Fprintf(os.Stderr, "tellme: no configuration could be found at %s\n", res.Path)
		}
		return ConfigError
	case reasonConfigInvalid:
		fmt.Fprintf(os.Stderr, "tellme: the configuration could not be parsed at %s\n", res.Path)
		return ConfigError
	case reasonProviderMismatch:
		fmt.Fprintf(os.Stderr, "tellme: the selected provider is not in the registry (%q)\n", res.Selected)
		return ConfigError
	case reasonProviderInvalid:
		fmt.Fprintf(os.Stderr, "tellme: the provider configuration is invalid: provider %q: %v\n", res.Selected, rerr.Err)
		return ConfigError
	case reasonHomeUnusable:
		if errors.Is(rerr.Err, home.ErrNotDirectory) {
			fmt.Fprintf(os.Stderr, "tellme: the workspace path is not a directory (%s)\n", res.Workspace)
		} else {
			fmt.Fprintf(os.Stderr, "tellme: the runtime home is not usable (%v)\n", rerr.Err)
		}
		return EnvironmentError
	default: // reasonHomeUnset
		fmt.Fprintln(os.Stderr, "tellme: the runtime home is not usable")
		return EnvironmentError
	}
}

// renderDiagnostic runs the -d path. It always emits a plain report and returns 0
// when resolution succeeded, else the dedicated diagnostic "unresolved" code.
// Round 002: the machine-readable `--json` form was removed; `--json` is no
// longer a flag, so any use of it is an unrecognized-flag usage error.
func renderDiagnostic(homeDir, configPath string) int {
	res, rerr := resolve(homeDir, configPath)
	emitDiagnosticText(res, rerr)
	if rerr != nil {
		return DiagnosticUnresolvedError
	}
	return Success
}

// emitDiagnosticText writes the plain-text report.
func emitDiagnosticText(res resolution, rerr *resolveError) {
	fmt.Println("tellme diagnostic")
	if rerr == nil {
		fmt.Println("configuration: resolved")
		fmt.Println("runtime_home: " + res.Home)
		fmt.Println("session_workspace: " + res.Workspace)
		return
	}
	fmt.Println("configuration: unresolved")
	fmt.Println("reason: " + rerr.Reason)
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

// emitUsageError writes the usage-error stderr message and returns the usage
// error code (FR-014).
func emitUsageError() int {
	fmt.Fprintln(os.Stderr, "tellme: the command-line usage is invalid")
	return UsageError
}
