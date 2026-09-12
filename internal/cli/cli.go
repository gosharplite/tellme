package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/pflag"

	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/home"
	infrallm "github.com/gosharplite/tellme/internal/infrastructure/llm"
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
	// is carried here so the reasoning turn can construct the provider transport
	// without re-loading or re-parsing the configuration (review finding #3).
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
// and Run returns the process exit code. It binds the real process streams and
// the default terminal detector, then delegates to run (round-005 research
// Decision 4 — the seam keeps the input/output-mode selection unit-testable).
func Run(args []string, version string) int {
	return run(args, version, os.Stdin, os.Stdout, os.Stderr, defaultIsTerminal)
}

// run parses flags, then dispatches to the version path, the diagnostic
// reporting path, the prompt-bearing reasoning turn, or the boot path. stdin is
// read on the non-explicit-mode dispatch path (round-005 FR-010) — the reasoning
// turn and the empty→boot fall-through, when stdin is not a terminal; the version
// and diagnostic paths never read it.
func run(args []string, version string, stdin io.Reader, stdout, stderr io.Writer, isTTY func(any) bool) int {
	opts, flagArgs, ok := parseFlags(args, stderr)
	if !ok {
		return emitUsageError(stderr)
	}
	if opts.version {
		_, _ = fmt.Fprintf(stdout, "tellme %s\n", version)
		return Success
	}

	homeDir := os.Getenv("TELL_ME_HOME")

	// -d is the reporting path: it always produces a report (Decision 2), and it
	// takes precedence over a prompt or piped input (round-004 Decision 7 /
	// round-005 FR-010).
	if opts.diagnostic {
		return renderDiagnostic(homeDir, opts.configPath, stdout)
	}
	// The prompt turn reads piped input only when stdin is not a terminal and
	// combines it with the positional argument(s); an empty result falls through
	// to boot (round-005 FR-001..FR-005).
	prompt, err := resolvePrompt(flagArgs, stdin, isTTY)
	if err != nil {
		// Review finding F1: reuse the existing environment class phrase so the
		// closed nine-phrase vocabulary (specs/truth/features/cli/dsl.md) is not
		// widened; the trailing stdin detail is contract-free.
		_, _ = fmt.Fprintf(stderr, "tellme: the runtime home is not usable (standard input: %v)\n", err)
		return EnvironmentError
	}
	if prompt != "" {
		return renderTurn(homeDir, opts.configPath, prompt, stdout, stderr)
	}
	return renderBoot(homeDir, opts.configPath, stdout, stderr)
}

// parseFlags parses argv, returning the parsed flags and the positional
// arguments (the prompt parts, if any). ok is false on an unrecognized or
// invalid flag. Flag errors are written to the injected stderr (review finding
// F3: no direct os.Stderr coupling).
func parseFlags(args []string, stderr io.Writer) (opts *options, flagArgs []string, ok bool) {
	fs := pflag.NewFlagSet("tellme", pflag.ContinueOnError)
	fs.SetOutput(stderr)
	o := &options{}
	fs.StringVarP(&o.configPath, "config", "c", "", "Path to the YAML configuration file.")
	fs.BoolVarP(&o.diagnostic, "diagnostics", "d", false, "Report configuration and home resolution, then exit.")
	fs.BoolVar(&o.version, "version", false, "Print the build version and exit.")
	if err := fs.Parse(args); err != nil {
		return nil, nil, false
	}
	return o, fs.Args(), true
}

// resolve is the single resolution algorithm shared by the boot, diagnostic, and
// turn paths: home → config path → load/validate → effective selected provider →
// effective mode → workspace. On failure it returns a *resolveError carrying the
// pinned reason category; the callers differ only in how they render it.
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
	// the reasoning turn is the fully-expanded one.
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
func renderBoot(homeDir, configPath string, stdout, stderr io.Writer) int {
	res, rerr := resolve(homeDir, configPath)
	if rerr != nil {
		return emitBootError(stderr, res, rerr)
	}
	_, _ = fmt.Fprintln(stdout, "configuration: ready")
	_, _ = fmt.Fprintln(stdout, "session workspace: "+res.Workspace)
	return Success
}

// gatewayFactory builds the provider gateway for a resolved provider entry. It
// is the composition seam (review finding #1): the presentation layer never
// couples to a concrete adapter constructor, tests can inject a fake
// llm.Gateway, and an un-adapted family surfaces as an actionable error.
type gatewayFactory func(prov config.Provider, name string) (llm.Gateway, error)

// newGateway is the production gateway factory (a var so tests may override it).
var newGateway gatewayFactory = infrallm.NewGateway

// renderTurn resolves the setup, then runs exactly one reasoning turn against
// the resolved provider and prints the answer (round-004 FR-001..FR-005). A
// resolve failure is rendered as a boot error; an unsupported family or a
// provider/transport failure is rendered with the frozen provider class phrase
// and exit code 6.
func renderTurn(homeDir, configPath, prompt string, stdout, stderr io.Writer) int {
	res, rerr := resolve(homeDir, configPath)
	if rerr != nil {
		return emitBootError(stderr, res, rerr)
	}
	return runTurn(res, prompt, stdout, stderr, newGateway)
}

// runTurn performs one reasoning turn through an injected gateway factory. The
// context is cancelled on SIGINT/SIGTERM so a stalled provider can be
// interrupted (review finding #2).
func runTurn(res resolution, prompt string, out, errOut io.Writer, factory gatewayFactory) int {
	gw, err := factory(res.Provider, res.Selected)
	if err != nil {
		return emitProviderError(errOut, err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	resp, err := gw.Complete(ctx, llm.Request{Prompt: prompt})
	if err != nil {
		return emitProviderError(errOut, err)
	}
	_, _ = fmt.Fprintln(out, resp.Text)
	return Success
}

// emitBootError maps a resolve failure to its actionable stderr message + code.
func emitBootError(stderr io.Writer, res resolution, rerr *resolveError) int {
	switch rerr.Reason {
	case reasonConfigMissing:
		if res.Explicit {
			_, _ = fmt.Fprintf(stderr, "tellme: the configuration could not be found at %s\n", res.Path)
		} else {
			_, _ = fmt.Fprintf(stderr, "tellme: no configuration could be found at %s\n", res.Path)
		}
		return ConfigError
	case reasonConfigInvalid:
		_, _ = fmt.Fprintf(stderr, "tellme: the configuration could not be parsed at %s\n", res.Path)
		return ConfigError
	case reasonProviderMismatch:
		_, _ = fmt.Fprintf(stderr, "tellme: the selected provider is not in the registry (%q)\n", res.Selected)
		return ConfigError
	case reasonProviderInvalid:
		_, _ = fmt.Fprintf(stderr, "tellme: the provider configuration is invalid: provider %q: %v\n", res.Selected, rerr.Err)
		return ConfigError
	case reasonHomeUnusable:
		if errors.Is(rerr.Err, home.ErrNotDirectory) {
			_, _ = fmt.Fprintf(stderr, "tellme: the workspace path is not a directory (%s)\n", res.Workspace)
		} else {
			_, _ = fmt.Fprintf(stderr, "tellme: the runtime home is not usable (%v)\n", rerr.Err)
		}
		return EnvironmentError
	default: // reasonHomeUnset
		_, _ = fmt.Fprintln(stderr, "tellme: the runtime home is not usable")
		return EnvironmentError
	}
}

// emitProviderError maps a provider/transport failure to its frozen class phrase
// and dedicated exit code (round-004 FR-006..FR-008, Clarify Q3). The trailing
// detail is contract-free; newlines are folded so exactly one line carries the
// phrase.
func emitProviderError(w io.Writer, err error) int {
	detail := strings.ReplaceAll(err.Error(), "\n", " ")
	_, _ = fmt.Fprintf(w, "tellme: the provider request failed: %s\n", detail)
	return ProviderError
}

// renderDiagnostic runs the -d path. It always emits a plain report and returns 0
// when resolution succeeded, else the dedicated diagnostic "unresolved" code.
// Round 002: the machine-readable `--json` form was removed; `--json` is no
// longer a flag, so any use of it is an unrecognized-flag usage error.
func renderDiagnostic(homeDir, configPath string, stdout io.Writer) int {
	res, rerr := resolve(homeDir, configPath)
	emitDiagnosticText(stdout, res, rerr)
	if rerr != nil {
		return DiagnosticUnresolvedError
	}
	return Success
}

// emitDiagnosticText writes the plain-text report.
func emitDiagnosticText(w io.Writer, res resolution, rerr *resolveError) {
	_, _ = fmt.Fprintln(w, "tellme diagnostic")
	if rerr == nil {
		_, _ = fmt.Fprintln(w, "configuration: resolved")
		_, _ = fmt.Fprintln(w, "runtime_home: "+res.Home)
		_, _ = fmt.Fprintln(w, "session_workspace: "+res.Workspace)
		return
	}
	_, _ = fmt.Fprintln(w, "configuration: unresolved")
	_, _ = fmt.Fprintln(w, "reason: "+rerr.Reason)
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
func emitUsageError(w io.Writer) int {
	_, _ = fmt.Fprintln(w, "tellme: the command-line usage is invalid")
	return UsageError
}
