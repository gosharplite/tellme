package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"github.com/gosharplite/tellme/internal/app/deps"
	appsuggestions "github.com/gosharplite/tellme/internal/app/suggestions"
	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/render"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	domaintui "github.com/gosharplite/tellme/internal/domain/tui"
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

// flags are the parsed CLI flags (renamed from `options` in round 044 / G2 to
// avoid the options / turnOptions / Options tri-collision).
type flags struct {
	configPath  string
	diagnostic  bool
	version     bool
	raw         bool
	newSession  bool
	list        int
	listSet     bool
	interactive bool
	toolUsage   bool
}

// Options is the single injected value Run consumes (round 044 / ADR 0013): the
// domain-typed Dependencies (built by the composition root) plus the ONE
// presentation seam — the interactive TUI prompt port. Round 048 (ADR 0017)
// replaced the unexported-typed RunTUIPrompt field with the exported domain
// interface domaintui.Prompter (closing review-deferral F-4): the composition
// root injects the internal/ui adapter, so internal/cli no longer imports
// internal/ui/tui/prompt.
type Options struct {
	Deps     deps.Dependencies
	Prompter domaintui.Prompter
}

// Validate reports whether every injected seam is wired, so a composition-root
// mistake fails loudly at the boundary instead of surfacing later as a confusing
// per-path error (round-044 F-5 precedent, extended by round 048 / ADR 0017:
// deps.Dependencies.Validate reflects over func fields only, so it cannot cover
// the Prompter interface seam).
func (o Options) Validate() error {
	if o.Prompter == nil {
		return errors.New("cli: the interactive prompt port (Options.Prompter) is not wired")
	}
	return o.Deps.Validate()
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
	Provider config.Provider
	// WrapWidth is the resolved rendered width (round 006): the effective
	// WRAP_WIDTH / TELL_ME_WRAP_WIDTH, or 0 for the renderer default.
	WrapWidth int
	// MaxToolLoop is the resolved tool-loop bound (round 008): MAX_TOOL_LOOP
	// (env/config, default 1000).
	MaxToolLoop int
	// MaxHistoryTokens is the resolved payload budget (round 009): the value the
	// payload status line measures against (MAX_HISTORY_TOKENS, default 1000000).
	MaxHistoryTokens int
	// EffectiveBudget is the run-static tool resource budget (round 024 D5,
	// FR-015): min(MaxHistoryTokens, the active model's configured
	// MODELS.<model>.CONTEXT_WINDOW) — the value the tool-result bound derives
	// from, and the value the payload status line now renders.
	EffectiveBudget int
	// Person is the resolved PERSON — the persona sent to the provider as the
	// leading `system` message of every request (round 011).
	Person string
	// MCPServers is the validated MCP_SERVERS registry (round 032); the
	// prompt-path discovery reads it. MCPWarnings carries non-fatal validation
	// warnings (a skipped COMMAND/stdio entry) to surface on the diagnostic
	// stream when a prompt run begins.
	MCPServers  map[string]config.MCPServerConfig
	MCPWarnings []string
	// Pricing is the active model's config-only `MODELS` rates (round 018);
	// Priced is false when the model has no entry, so the post-turn cost renders
	// `$0.0000` (research D2).
	Pricing   config.PricingRates
	Priced    bool
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

// runtimeEnv bundles the process I/O streams, the terminal detector, and the
// answer renderer for one CLI invocation (round-006 review **Obs 2**). Threading
// them as a single value keeps the call signatures stable as the flag/env
// surface grows (additional flags, multi-turn sessions), instead of widening
// `run`/`renderTurn`/`runTurn`/`writeAnswer` with more stream primitives. The
// renderer is built once per invocation (see `Run`) and reused across the turn.
type runtimeEnv struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	isTTY  func(any) bool
	// stderrTTY is the diagnostic-stream (stderr) terminal probe (round 019).
	// Nil disables the turn spinner — the gate is the stderr stream, so there is
	// no fallback to the shared (stdin) probe.
	stderrTTY func(any) bool
	renderer  render.Answer
	// clock is the injected time seam for the payload status line (round 009);
	// nil falls back to time.Now.
	clock func() time.Time
}

// round 044: the history/usage/toolUsage-store, gateway and tool-registry factory
// vars moved to cmd/tellme + deps.Dependencies; the renderer var was deleted
// (built inline via ui.NewRenderer()) (ADR 0013). round 048 (ADR 0017): the TUI
// runner var is gone too — the interactive prompt is reached through the
// injected domaintui.Prompter port, with no in-package default (a nil port is a
// loud EnvironmentError; see runTUIPrompt).

// runInteractiveTUI runs the interactive TUI prompt (round 015) for one
// invocation: it announces on the diagnostic stream, builds the multi-source
// suggestion engine over the injected Dependencies, and drives the prompt bound
// to stderr/stdin through the injected domain port. It returns the composed
// prompt (ok) on submit and never writes to stdout. Round 048 (ADR 0017): the
// terminal runtime is reached only through p — internal/cli no longer imports
// internal/ui/tui/prompt.
func runInteractiveTUI(ctx context.Context, res resolution, env runtimeEnv, newPromptTracker func(home string, userHome func() (string, error)) history.PromptTracker, newTUIRegistry func() domaintools.Registry, userHome func() (string, error), p domaintui.Prompter) (string, bool, error) {
	_, _ = fmt.Fprintln(env.stderr, TUIHint)

	tracker := newPromptTracker(res.Home, userHome)
	defer func() { _ = tracker.Close(context.Background()) }()
	// Round 028: the first-use seed runs once, here, before the first suggestion
	// read. It is the segregated Seeder capability (NOT part of PromptTracker —
	// interface segregation, matching the round-026 port split); construction stays
	// pure (the tracker is built twice per `-i` run, so the seed is not a
	// constructor side effect — TD-2); the seed is best-effort.
	if seeder, ok := tracker.(history.Seeder); ok {
		_ = seeder.Seed(ctx)
	}
	reg := newTUIRegistry()
	engine := appsuggestions.New(
		appsuggestions.TrackerPrompts{Tracker: tracker},
		appsuggestions.OSSWorkspace{},
		appsuggestions.RegistryTools{Registry: reg},
	)
	src := tuiSource{svc: engine}
	// No startup disk I/O: the dashboard header was retired (round-016 FR-004 /
	// architect D3), so the history store is no longer read to build the prompt.
	return p.Run(ctx, env.stdin, env.stderr, src, tuiDebounceDuration(p))
}

// tuiDebounceDuration resolves the suggestion-refresh debounce. The hermetic E2E
// sets TELL_ME_TUI_DEBOUNCE=0 so the scripted keys observe suggestions
// synchronously (round-016); otherwise the port's default (~100 ms) applies.
func tuiDebounceDuration(p domaintui.Prompter) time.Duration {
	if v := strings.TrimSpace(os.Getenv("TELL_ME_TUI_DEBOUNCE")); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms >= 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return p.DefaultDebounceDuration()
}

// tuiSource adapts the suggestion engine to the prompt's Source seam, carrying
// ctx so a superseded fetch can be cancelled (round-016 architect D1).
type tuiSource struct {
	svc *appsuggestions.Service
}

// Suggest returns the engine's suggestion texts for the query.
func (s tuiSource) Suggest(ctx context.Context, query string) []string {
	sugg := s.svc.Suggest(ctx, query)
	out := make([]string, 0, len(sugg))
	for _, x := range sugg {
		out = append(out, x.Text)
	}
	return out
}

// tuiRequested reports whether the opt-in interactive prompt is enabled: the
// `-i`/`--interactive` flag, or the config `USE_TUI_PROMPT` key (round-015
// FR-001). The terminal-stdin requirement is enforced separately by the caller.
func tuiRequested(homeDir string, f *flags) bool {
	if f.interactive {
		return true
	}
	if cfg, err := config.Load(defaultConfigPath(homeDir)); err == nil {
		return cfg.UseTUIPrompt
	}
	return false
}

// runTUIPrompt resolves the setup and runs the interactive TUI prompt (round
// 015). A submitted prompt is recorded in the shared log — only for the
// interactive prompt — and then runs exactly one reasoning turn; an
// aborted/empty submission sends no request and exits success.
func runTUIPrompt(homeDir string, f *flags, env runtimeEnv, opts Options) int {
	dp := opts.Deps
	// Best-effort resolution: the TUI engages even when the setup is unresolved
	// (an aborted/empty submission owes no request — the round-012 ordering
	// rationale). The resolution feeds the dashboard and the submit path.
	res, _ := resolve(homeDir, f.configPath)
	p := opts.Prompter
	if p == nil {
		// Unreachable in production: the composition root always injects the port
		// (asserted by Options.Validate; see cmd/tellme's smoke test).
		// [round-048 divergence — ADR 0017] Reuse the environment class phrase so the
		// closed phrase vocabulary is not widened; the trailing detail is
		// contract-free. The CAUSE here is a composition-root wiring fault, not an
		// unusable home — a deliberate recorded mismatch for a future vocabulary
		// round.
		_, _ = fmt.Fprintf(env.stderr, "tellme: the runtime home is not usable (interactive prompt: no runner injected)\n")
		return EnvironmentError
	}
	text, ok, err := runInteractiveTUI(context.Background(), res, env, dp.NewPromptTracker, dp.NewTUIRegistry, dp.UserHomeDir, p)
	if err != nil {
		return emitProviderError(env.stderr, err)
	}
	if !ok || text == "" {
		return Success
	}
	// Record in the shared log (round-015 FR-009) — only the interactive prompt
	// writes it.
	tracker := dp.NewPromptTracker(res.Home, dp.UserHomeDir)
	_ = tracker.Append(context.Background(), text)
	_ = tracker.Close(context.Background())
	return renderTurn(homeDir, f.configPath, text, turnOptions{raw: f.raw, chrome: true, echo: true}, env, dp)
}

// Run is the CLI entrypoint: main passes argv and the injected build version,
// and Run returns the process exit code. It binds the real process streams, the
// default terminal detector, and the production renderer into a runtimeEnv, then
// delegates to run (round-005 research Decision 4 — the seam keeps the
// input/output-mode selection unit-testable; round-006 review Obs 2 — one value
// instead of many stream primitives).
func Run(args []string, version string, opts Options) int {
	return run(args, version, opts, runtimeEnv{
		stdin:     os.Stdin,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		isTTY:     terminalDetector(),
		stderrTTY: stderrTerminalDetector(),
		renderer:  opts.Deps.NewAnswer(),
		clock:     time.Now,
	})
}

// stderrTerminalDetector returns the process's diagnostic-stream (stderr)
// terminal probe. The diagnostic environment seam TELL_ME_FORCE_STDERR_TTY forces
// the stderr probe to report a terminal so the round-019 spinner is E2E-drivable
// without a pty (mirroring the round-012 TELL_ME_FORCE_STDIN_TTY seam). Otherwise
// the real isatty probe (defaultIsTerminal) is used. It does NOT wire a
// standard-output probe — round-006 / PR #16 Obs 1 stays OPEN.
func stderrTerminalDetector() func(any) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TELL_ME_FORCE_STDERR_TTY"))) {
	case "1", "true", "yes":
		return func(any) bool { return true }
	default:
		return defaultIsTerminal
	}
}

// stderrIsTerminal reports whether the diagnostic stream (stderr) is a terminal
// via the dedicated stderr probe. It deliberately does NOT fall back to the
// shared (stdin) probe: the round-019 gate is the *stderr* stream (FR-006), and a
// fallback would make every `runtimeEnv{isTTY: true}` construction report a
// terminal stderr. With no stderr probe set the spinner is off (the safe default).
func (e runtimeEnv) stderrIsTerminal() bool {
	if e.stderrTTY == nil {
		return false
	}
	return e.stderrTTY(e.stderr)
}

// terminalDetector returns the process's terminal probe. When the diagnostic
// environment seam TELL_ME_FORCE_STDIN_TTY is truthy every stream is reported as
// a terminal, so the interactive multi-line read can be exercised end-to-end
// against a pipe without a pty (round-012 review RF1). Otherwise the real isatty
// probe (defaultIsTerminal) is used.
func terminalDetector() func(any) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TELL_ME_FORCE_STDIN_TTY"))) {
	case "1", "true", "yes":
		return func(any) bool { return true }
	default:
		return defaultIsTerminal
	}
}

// run parses flags, then dispatches to the version path, the diagnostic
// reporting path, the history-listing path, the prompt-bearing reasoning turn,
// the fresh-session path, or the boot path (round-007 Decision 5 precedence:
// `--version` → `-d` → `-l` → (`--new`) prompt turn → boot). stdin is read on
// the prompt path only (round-005 FR-010); the version, diagnostic, and `-l`
// paths never read it.
func run(args []string, version string, scoped Options, env runtimeEnv) int {
	dp := scoped.Deps
	f, flagArgs, ok := parseFlags(args, env.stderr)
	if !ok {
		return emitUsageError(env.stderr)
	}
	if f.version {
		_, _ = fmt.Fprintf(env.stdout, "tellme %s\n", version)
		return Success
	}

	homeDir := os.Getenv("TELL_ME_HOME")

	// The offline reporting commands run in precedence order — -d → -l →
	// --tool-usage — before any prompt or stdin access.
	if code, handled := dispatchReporting(f, homeDir, env, dp.NewHistoryStore, func() int {
		return renderToolUsage(env, dp.NewToolRegistry, dp.NewToolUsageStore, dp.UserHomeDir, dp.NewLines())
	}); handled {
		return code
	}
	// The prompt turn reads piped input only when stdin is not a terminal and
	// combines it with the positional argument(s); an empty result falls through
	// to the fresh-session or boot path (round-005 FR-001..FR-005).
	prompt, err := resolvePrompt(flagArgs, env.stdin, env.isTTY)
	if err != nil {
		// Reuse the existing environment class phrase so the closed phrase
		// vocabulary (specs/truth/features/cli/dsl.md) is not widened; the
		// trailing stdin detail is contract-free.
		_, _ = fmt.Fprintf(env.stderr, "tellme: the runtime home is not usable (standard input: %v)\n", err)
		return EnvironmentError
	}
	if prompt != "" {
		return renderTurn(homeDir, f.configPath, prompt, turnOptions{raw: f.raw, newSession: f.newSession, chrome: true}, env, dp)
	}
	// Round 012 (amended, A8) — a prompt-less invocation on a terminal reads an
	// interactive multi-line prompt: print the hint to stderr and read stdin to EOF
	// (Ctrl+D). With --new the session is archived FIRST (so the fresh session is
	// used, and an empty/cancel still starts fresh), then the reader engages. An
	// empty or cancelled submission sends no request and exits success (round-012
	// research Decisions 1–5). POSIX-only; there is no Windows variant.
	//
	// Ordering note (round-012 review TD3): the reader engages BEFORE setup
	// resolution, deliberately. An empty/cancelled submission owes no request and
	// therefore requires no configuration (Decision 4 / Clarify Q3), so readiness
	// cannot gate the read without changing that contract; a misconfigured setup
	// therefore surfaces after the read, via renderTurn. The error-masking hazard
	// the review flagged (exit 0 on a broken config for `< /dev/null`) is fixed at
	// its root by the real isatty probe (B1), which routes a non-terminal stdin —
	// /dev/null included — to the boot path instead of here.
	//
	// SIGTERM note (round-012 review TD4): a SIGTERM during the read cancels the
	// context and exits success (0), matching the existing runTurn convention for
	// an operator-initiated interruption of a prompt turn.
	if env.isTTY(env.stdin) {
		// `--new` archives BEFORE the interactive read for BOTH terminal reader
		// surfaces — the `-i` TUI prompt and the plain reader — so a fresh session
		// starts regardless of the submission (round 027: the `-i` surface
		// previously dropped `--new`, so the header counted the prior history). It
		// archives before resolving the configuration: a prompt-less `--new` is an
		// archive command that works offline, so — unlike the prompt-bearing
		// `--new "<prompt>"` form, which resolves first — a broken config still
		// archives here and then fails when the turn resolves (round-012 review).
		if f.newSession {
			if code := renderNewSession(homeDir, env, dp); code != Success {
				return code
			}
		}
		// Round 015 — the opt-in interactive TUI prompt engages here (only when
		// enabled AND stdin is a terminal); the plain reader below stays the
		// default. The dispatch runs through the injected domaintui.Prompter port
		// (round 048 / ADR 0017) so the matrix is unit-testable (PR #38 review
		// directive ④).
		if tuiRequested(homeDir, f) {
			return runTUIPrompt(homeDir, f, env, scoped)
		}
		ictx, icancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		text, ok := readInteractivePrompt(ictx, env.stdin, env.stderr)
		icancel()
		if !ok || text == "" {
			return Success
		}
		return renderTurn(homeDir, f.configPath, text, turnOptions{raw: f.raw, chrome: true}, env, dp)
	}
	// A prompt-less --new on a NON-terminal keeps its round-007 behaviour: archive
	// the session and exit (the reader never engages on a non-terminal).
	if f.newSession {
		return renderNewSession(homeDir, env, dp)
	}
	return renderBoot(homeDir, f.configPath, env)
}

// parseFlags parses argv, returning the parsed flags and the positional
// arguments (the prompt parts, if any). ok is false on an unrecognized or
// invalid flag. Flag errors are written to the injected stderr (review finding
// F3: no direct os.Stderr coupling).
func parseFlags(args []string, stderr io.Writer) (f *flags, flagArgs []string, ok bool) {
	fs := pflag.NewFlagSet("tellme", pflag.ContinueOnError)
	fs.SetOutput(stderr)
	o := &flags{}
	fs.StringVarP(&o.configPath, "config", "c", "", "Path to the YAML configuration file.")
	fs.BoolVarP(&o.diagnostic, "diagnostics", "d", false, "Report configuration and home resolution, then exit.")
	fs.BoolVar(&o.version, "version", false, "Print the build version and exit.")
	fs.BoolVarP(&o.raw, "raw", "r", false, "Print the answer as raw text (no Markdown rendering).")
	fs.BoolVar(&o.newSession, "new", false, "Start a fresh session, archiving the current session history.")
	fs.IntVarP(&o.list, "list", "l", 0, "List the last N messages of the session history and exit.")
	fs.BoolVarP(&o.interactive, "interactive", "i", false, "Open the interactive TUI prompt (requires a terminal).")
	fs.BoolVar(&o.toolUsage, "tool-usage", false, "Report per-tool invocation counts across all sessions, then exit.")
	if err := fs.Parse(args); err != nil {
		return nil, nil, false
	}
	o.listSet = fs.Changed("list")
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

	// Step 4b — resolve and validate the rendered width (round-006 FR-006): the
	// helper owns resolve+validate, so a non-integer override or a negative value
	// from either source is a configuration error. This step runs before provider
	// resolution, so a config that is both width-invalid and provider-invalid
	// reports the width error (recorded precedence).
	width, werr := cfg.EffectiveWrapWidth(os.Getenv("TELL_ME_WRAP_WIDTH"))
	if werr != nil {
		return res, &resolveError{Reason: reasonConfigInvalid, Err: werr}
	}
	res.WrapWidth = width

	// Step 4c — resolve the tool-loop bound (round-008 FR-006): MAX_TOOL_LOOP
	// (env/config, default 1000). A non-integer or negative value is a
	// configuration error.
	maxLoops, lerr := cfg.EffectiveMaxToolLoop(os.Getenv("MAX_TOOL_LOOP"))
	if lerr != nil {
		return res, &resolveError{Reason: reasonConfigInvalid, Err: lerr}
	}
	res.MaxToolLoop = maxLoops

	// Step 4d — resolve the payload budget (round-009 FR-010): MAX_HISTORY_TOKENS
	// (env/config, default 1000000). A non-integer or negative value is a
	// configuration error.
	budget, bErr := cfg.EffectiveMaxHistoryTokens(os.Getenv("MAX_HISTORY_TOKENS"))
	if bErr != nil {
		return res, &resolveError{Reason: reasonConfigInvalid, Err: bErr}
	}
	res.MaxHistoryTokens = budget

	// Step 4d.1 — expand ${VAR} / ${VAR:-default} in the MCP_SERVERS string
	// fields, best-effort (round-032 SC-002 / issue #67): a `TOKEN:
	// "${GITHUB_TOKEN}"` entry authenticates instead of being sent literally (and
	// then warn+skipped). An unresolved ${VAR} keeps its literal text AND emits a
	// non-fatal diagnostic warning naming the field/variable, so the cause is
	// visible instead of an opaque "could not be reached"; the load never fails.
	// Expansion runs BEFORE validation so the validator sees the resolved values.
	res.MCPWarnings = append(res.MCPWarnings, cfg.ExpandMCPServers()...)

	// Step 4e — validate the MCP_SERVERS registry (round-032 FR-002/FR-013): a
	// malformed REMOTE entry reuses the configuration-invalid class phrase (a
	// stable, classed failure); a COMMAND (stdio) entry is warn+skipped, surfaced
	// on the prompt path. The registry is carried on the resolution so the
	// prompt-path discovery reads the exact validated set.
	if mcpVal, mcpErr := cfg.ValidateMCPServers(); mcpErr != nil {
		return res, &resolveError{Reason: reasonConfigInvalid, Err: mcpErr}
	} else {
		res.MCPWarnings = append(res.MCPWarnings, mcpVal.Warnings...)
	}
	res.MCPServers = cfg.MCPServers

	// Step 5 — the effective selected provider must be in the registry (FR-003).
	res.Selected = cfg.EffectiveSelectedProvider(os.Getenv("TELL_ME_SELECTED_PROVIDER"))
	if !cfg.ProviderInRegistry(res.Selected) {
		return res, &resolveError{Reason: reasonProviderMismatch}
	}

	// Step 5b — resolve the selected provider entry: expand ${VAR} placeholders
	// FIRST, then validate the RESOLVED state (FR-001..FR-009).
	prov := cfg.Providers[res.Selected]
	if err := prov.Expand(); err != nil {
		return res, &resolveError{Reason: reasonProviderInvalid, Err: err}
	}
	if err := prov.Validate(); err != nil {
		return res, &resolveError{Reason: reasonProviderInvalid, Err: err}
	}
	cfg.Providers[res.Selected] = prov
	res.Provider = prov
	res.Pricing, res.Priced = cfg.PricingFor(prov.Model)

	// Step 5c — the run-static effective budget (round 024 FR-015): the tool
	// resource bound cap, min(MAX_HISTORY_TOKENS, the model's configured context
	// window). With no window configured it is MAX_HISTORY_TOKENS (so the payload
	// line and the bound are byte-identical to the model-blind default). The
	// window is the optional per-model MODELS.<model>.CONTEXT_WINDOW.
	res.EffectiveBudget = res.MaxHistoryTokens
	if window, ok := cfg.ContextWindowFor(prov.Model); ok {
		if window < res.EffectiveBudget {
			res.EffectiveBudget = window
		}
	} else {
		logNoWindowOnce(prov.Model)
	}

	// Step 6 — effective mode + prepare the session workspace (FR-007/008/009).
	res.Person = cfg.Person
	res.Mode = cfg.EffectiveMode(os.Getenv("TELL_ME_MODE"))
	workspace, err := home.EnsureWorkspace(homeDir, res.Mode)
	res.Workspace = workspace.Path
	if err != nil {
		return res, &resolveError{Reason: reasonHomeUnusable, Err: err}
	}
	return res, nil
}

// renderBoot runs the boot path and reports readiness or an actionable error.
func renderBoot(homeDir, configPath string, env runtimeEnv) int {
	res, rerr := resolve(homeDir, configPath)
	if rerr != nil {
		return emitBootError(env.stderr, res, rerr)
	}
	_, _ = fmt.Fprintln(env.stdout, "configuration: ready")
	_, _ = fmt.Fprintln(env.stdout, "session workspace: "+res.Workspace)
	return Success
}

// renderTurn resolves the setup, optionally archives the current session
// (`--new`), then runs exactly one reasoning turn against the resolved provider
// and prints the answer (round-004 FR-001..FR-005; round-007 FR-005). A resolve
// failure is rendered as a boot error; an unsupported family or a
// provider/transport failure is rendered with the frozen provider class phrase
// and exit code 6; a history failure reuses the environment class phrase.
// turnOptions carries a turn's execution flags as named fields rather than
// positional booleans (round-017 implementation-review finding 1 — "boolean
// blindness": renderTurn/runTurn had three consecutive positional bools).
type turnOptions struct {
	raw        bool
	newSession bool
	chrome     bool
	// echo is true exclusively for the -i interactive-prompt submission: the
	// submitted prompt is echoed on stderr before the input-capture
	// acknowledgement, because the editor box that held it was cleared
	// (round 023; FR-007/FR-009). It is false for the positional / Ctrl+D
	// surfaces (they already show the typed text).
	echo bool
}

func renderTurn(homeDir, configPath, prompt string, opts turnOptions, env runtimeEnv, dp deps.Dependencies) int {
	res, rerr := resolve(homeDir, configPath)
	if rerr != nil {
		return emitBootError(env.stderr, res, rerr)
	}
	store := dp.NewHistoryStore(res.Workspace)
	if opts.newSession {
		if err := store.Archive(); err != nil {
			return emitHistoryError(env.stderr, err)
		}
		if err := dp.NewUsageStore(res.Workspace).Archive(); err != nil {
			return emitHistoryError(env.stderr, err)
		}
	}
	return runTurn(res, store, prompt, opts, env, dp)
}

// runTurn performs one reasoning turn through the injected dependencies and a
// history store. The resumed conversation is loaded and carried ahead of the
// current prompt; the completed turn is appended after the provider answers
// (append-after-complete, round-007 FR-001/FR-003). The answer is written by the
// runtimeEnv's renderer. The context is cancelled on SIGINT/SIGTERM so a stalled
// provider can be interrupted (review finding #2).
func runTurn(res resolution, store history.Store, prompt string, opts turnOptions, env runtimeEnv, dp deps.Dependencies) int {
	gw, err := dp.NewGateway(res.Provider, res.Selected, res.Person)
	if err != nil {
		return emitProviderError(env.stderr, err)
	}
	prior, err := store.Load()
	if err != nil {
		return emitHistoryError(env.stderr, err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Round-019 elapsed epoch: the spinner's turn-scoped timer starts at prompt
	// capture — the moment the input-capture acknowledgement fires (research D4).
	turnStart := env.now()

	// Round 019 — the live progress spinner: a diagnostic-stream-only indicator
	// that labels / clears / restores per waiting phase. Round 051 (R5.5 of #92;
	// ADR 0020): the spinner + the `[Tool Output]` coordinator are built behind
	// the injected domain seam (deps.NewProgress) — internal/cli names no
	// internal/ui type. A nil indicator means the spinner is gated off.
	// Round 052 (closes #115 R-2; ADR 0021): the progress object is built BEFORE
	// the registry so its `[Tool Output]` sink can be injected at the command
	// tool's construction (no post-construction rebind).
	prog := dp.NewProgress(env.stderr, env.now, res.Provider.Model, turnStart, stderrColumns(env), toolOutputIdleGap(dp.NewLines()), spinnerGate(opts, env.stderrIsTerminal()))
	ind := prog.Indicator
	if ind != nil {
		defer ind.Stop() // panic-safe residue guard (idempotent)
	}

	// Pre-flight payload status (round-009 FR-001): the estimated size of the
	// assembled conversation — the resumed turns (via the shared projection,
	// including tool steps — TD-1) plus the current prompt — measured against the
	// payload budget. Diagnostic only, on stderr.
	// Round 052 (closes #115 R-2; ADR 0021): the registry is built with the
	// `[Tool Output]` sink injected at construction (`prog.ToolOutput`) — the
	// round-034 `BindToolOutput` rebind no longer exists.
	reg := dp.NewToolRegistry(prog.ToolOutput)
	// Round 033 (FR-009): bind the `list_skills` catalog source on the
	// prompt-bearing turn path ONLY — the runtime home is resolved here. The load
	// stays lazy (inside the tool's Execute), so no registration reads docs/skills
	// and the offline paths never touch it.
	dp.BindSkillsCatalog(reg, home.SkillsDir(res.Home))
	// Round 034: the per-prompt pre-flight line is retired in favour of the
	// per-call estimate computed by the call renderer from the loop's messages.
	// Round-017 turn chrome: a prompt-bearing turn opens with the input-capture
	// acknowledgement and the rule/header frame, wrapping the pre-flight payload
	// line. It is true for the positional / piped / Ctrl+D reader surfaces and —
	// round 023 — the `-i` submit path; it is false for the non-prompt paths
	// (FR-007).
	if opts.chrome {
		// Round 023: the `-i` submit echoes the submitted prompt as its own
		// diagnostic block before the acknowledgement (the editor box that held it
		// was cleared), so the operator still sees what they sent. Written verbatim
		// — embedded newlines preserved (FR-007/FR-009).
		if opts.echo {
			_, _ = fmt.Fprintln(env.stderr, prompt)
		}
		emitInputCaptured(env, dp.NewLines())
	}
	// Round 032 (F9) — discover MCP tools BEFORE the turn frames, so a
	// slow/unreachable server's bounded wait is never silent and the per-call
	// estimate counts the offered tools (round 034: the frame is per call).
	reg, closeMCP := augmentRegistryWithMCP(ctx, res, reg, env.stderr, dp)
	defer closeMCP()

	// Round 034 (ADR 0005 D1/D3): the CLI's call renderer is the sole per-call
	// renderer — the status frame (rule + `╭─⠿ Turn N - <mode>` + pre-flight
	// estimate) at each call's begin, and the tail (grouped reasons + measured
	// payload + metrics + `Ready`) at each call's end, with the FINAL call's tail
	// deferred past the answer (G5). The loop fires the call hooks.
	renderer := newCallRenderer(env, res, reg, opts.chrome, turnNumber(prior)-1, dp)

	// Round 034 (ADR 0005 D1): the loop keeps a single observer — the composite
	// composes the per-call block renderer with the round-019 spinner.
	observer := compositeObserver{call: renderer, spinner: ind}

	// Round 050 (R5.4 of #92; ADR 0019): the loop is obtained through the injected
	// domain port (deps.LoopFactory) — internal/cli names no internal/agent type.
	// The port carries the loop's construction inputs; the CLI keeps its ui wiring
	// (the tool-line renderer, the spinner/observer composite) by design (Q1 → A).
	loop := dp.LoopFactory(agentport.LoopSpec{
		Gateway:         gw,
		Registry:        reg,
		MaxLoops:        res.MaxToolLoop,
		EffectiveBudget: res.EffectiveBudget,
		Stderr:          env.stderr,
		// Round 022: share the CLI clock seam so the tool-log line and the
		// chrome/payload lines use one clock (the loop falls back to time.Now when
		// unset).
		Now: env.now,
		// Round 026: the user-global tool-usage sink. Best-effort; a turn that uses
		// no tool leaves ~/.tellme untouched (the adapter creates the file lazily).
		ToolUsage: dp.NewToolUsageStore(dp.UserHomeDir),
		// Round 046 (R4 of #92, ADR 0015): the loop renders its four tool lines
		// through the injected presentation port — internal/ui owns the bytes and
		// the single-owned blank-reason predicate, so internal/agent imports no
		// internal/ui (the final layer-discipline baseline entry is gone).
		Lines: dp.NewToolLines(),
		// Round 034 (ADR 0005 D1) + round 050: the single observer (composite) is
		// supplied through the spec.
		Observer: observer,
	})
	// Round 052 (closes #115 R-2; ADR 0021): the `[Tool Output]` sink is now
	// injected into the command tool at CONSTRUCTION (via dp.NewToolRegistry
	// above) — the round-034 post-construction `dp.BindToolOutput(reg, …)` rebind
	// is gone. The block renders unconditionally; the coordinator owns the writer
	// + the spinner and applies the WS-A idle-gap liveness (mutual exclusion +
	// join; ADR 0009 D3/D4, superseding the round-034 whole-block pause).
	result, err := loop.Run(ctx, prompt, prior)
	if ind != nil {
		// Synchronous clear before any interleaved write (the answer, the
		// post-turn lines) so no frame survives into the completed turn. The
		// clear leaves the cursor at column 0 of the erased line, so the answer
		// on `stdout` (and a class phrase on `stderr`) continues on that line.
		ind.Stop()
	}
	if err != nil {
		var inc *agentport.ErrIncomplete
		if errors.As(err, &inc) {
			return emitToolError(env.stderr, inc)
		}
		return emitProviderError(env.stderr, err)
	}
	// Persist the turn with its AI-endpoint-call count (round 027): the number of
	// inference rounds this turn made (`len(result.Calls)`), summed across the
	// session to number the next turn's header.
	if err := store.Append(history.Entry{Prompt: prompt, Answer: result.Answer, Calls: len(result.Calls), Steps: result.Steps}); err != nil {
		return emitHistoryError(env.stderr, err)
	}
	// Round 034 (ADR 0005 D4/FR-010b): persist the turn's usage ONCE — the
	// Reported subset of result.Calls in one AppendBatch (never per call).
	persistTurnUsage(env, res, result, dp)
	env.writeAnswer(result.Answer, opts.raw, res.WrapWidth)
	// The final AI-endpoint call's tail was deferred; emit it now so the closing
	// status trails the answer (G5).
	renderer.EmitFinalTail()
	return Success
}

// emitPayloadStatus writes one payload status line to the diagnostic stream
// (stderr) using the runtime's injected clock seam (round-009 FR-001/FR-006).
// `estimated` selects the pre-flight `~` form; the measured form omits it. The
// line carries no `tellme: ` prefix (FR-014) and names the effective mode and the
// provider's configured MODEL (TD-2).
// effectiveBudget returns the budget the payload status line renders: the
// run-static EffectiveBudget when set (the resolve() path), else MaxHistoryTokens
// (directly-constructed resolutions in unit tests).
func (r resolution) effectiveBudget() int {
	if r.EffectiveBudget > 0 {
		return r.EffectiveBudget
	}
	return r.MaxHistoryTokens
}

// (round 034, 4B: emitPayloadStatus is retired — the per-call renderer renders
// the pre-flight and measured payload lines from the call hooks.)

// logNoWindowOnce emits a one-time debug log when the active model has no
// configured CONTEXT_WINDOW, so the operator knows the tool bound tracks
// MAX_HISTORY_TOKENS (round-024 research D5; the recorded implementation note).
// It goes to the debug log — never to the operator streams — so byte-exact
// stderr assertions are unaffected.
var noWindowOnce sync.Once

func logNoWindowOnce(model string) {
	noWindowOnce.Do(func() {
		slog.Debug("tool resource contract: active model has no configured context window; the bound tracks MAX_HISTORY_TOKENS", "model", model)
	})
}

// now returns the current time from the injected clock seam (falling back to
// time.Now) — the shared round-009/017 clock seam.
func (e runtimeEnv) now() time.Time {
	if e.clock != nil {
		return e.clock()
	}
	return time.Now()
}

// (round 034, 4B: emitPostTurnStatus is retired — the per-call renderer emits
// each call's metrics/`Ready` tail and persistTurnUsage writes the one AppendBatch.)

// emitInputCaptured writes the round-017 input-capture acknowledgement to the
// diagnostic stream (the reference's `[HH:MM:SS] Input captured. Processing...`).
func emitInputCaptured(env runtimeEnv, lines render.Lines) {
	_, _ = fmt.Fprintln(env.stderr, lines.InputCaptured(env.now()))
}

// (round 034, 4B: emitTurnOpening / emitTurnGap are retired — the per-call
// renderer emits the frame via ui.FormatTurnOpening/FormatTurnGap directly.)

// turnNumber is the round-017/027 turn-header number: the session's running
// AI-endpoint-call index — one more than the total number of provider calls
// (inference rounds) the prior completed turns made (`history.TotalCalls`, which
// treats an entry without a count as one — round-027 Decision 5). The header is
// emitted before the turn, so this is the index of this prompt's FIRST call —
// the value tell-me-go prints at that prompt's first call.
func turnNumber(prior []history.Entry) int {
	return history.TotalCalls(prior) + 1
}

// spinnerGate reports whether the turn spinner should be drawn: only on a
// non-TUI prompt-bearing surface (chrome), when -r/--raw is off, and when the
// diagnostic stream (stderr) is a terminal (round-019 FR-006/FR-008).
func spinnerGate(opts turnOptions, stderrIsTerminal bool) bool {
	return opts.chrome && !opts.raw && stderrIsTerminal
}

// stderrColumns reports the terminal width for the spinner's row-aware clear
// (round 025). The diagnostic environment seam TELL_ME_FORCE_STDERR_COLS overrides
// the real stderr-width probe so the row-aware clear is drivable in E2E/unit
// without a pty (mirroring TELL_ME_FORCE_STDERR_TTY); 0 means the width is unknown,
// so the presenter degrades to a single-row best-effort clear.
func stderrColumns(env runtimeEnv) func() int {
	// Resolve the override ONCE at construction (mirroring the stderrTTY / clock
	// seams), so the seam does not re-read the environment on every frame; the
	// real probe is still re-issued per call so a mid-turn resize is observed.
	if n, err := strconv.Atoi(strings.TrimSpace(os.Getenv("TELL_ME_FORCE_STDERR_COLS"))); err == nil && n > 0 {
		return func() int { return n }
	}
	return func() int { return terminalColumns(env.stderr) }
}

// toolOutputIdleGap resolves the round-040 WS-A idle-gap threshold (ADR 0009 D3):
// the hermetic seam TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS (milliseconds; 0 = admit
// immediately; unset/invalid = the 3 s default). Resolved once at construction,
// mirroring the stderrTTY / stderrColumns / tuiDebounceDuration seams.
func toolOutputIdleGap(lines render.Lines) time.Duration {
	if v := strings.TrimSpace(os.Getenv("TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS")); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms >= 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return lines.DefaultToolOutputIdleGap()
}

// dispatchReporting handles the offline reporting commands in precedence order
// — `-d` → `-l` → `--tool-usage` — before any prompt or stdin access. It returns
// the exit code and whether a reporting command handled the run. (`--version` is
// handled by the caller, ahead of the reporting batch.) Extracted from `run` so
// its cyclomatic complexity stays under the cyclop gate (round 026).
func dispatchReporting(f *flags, homeDir string, env runtimeEnv, newHistoryStore func(workspace string) history.Store, toolUsageReport func() int) (int, bool) {
	// -d is the reporting path: it always produces a report, and it takes
	// precedence over a prompt or piped input (round-004 Decision 7).
	if f.diagnostic {
		return renderDiagnostic(homeDir, f.configPath, env.stdout), true
	}
	// -l lists the last N messages and exits, strictly offline (round-007). A
	// non-positive N is a usage error, evaluated before any network or stdin.
	if f.listSet {
		if f.list <= 0 {
			return emitUsageError(env.stderr), true
		}
		return renderHistoryList(homeDir, f.list, env, newHistoryStore), true
	}
	// --tool-usage is the offline tool-usage report: like --version it needs no
	// configuration, no TELL_ME_HOME, and no workspace (round-026 FR-006).
	if f.toolUsage {
		return toolUsageReport(), true
	}
	return 0, false
}

// renderToolUsage prints the offline per-tool roll-up (round-026 FR-006..FR-009):
// it enumerates the LIVE registry, streams the user-global tool-usage log into
// per-tool counts, and writes plain text to stdout. It is `--version`-class — it
// requires neither a configuration, nor TELL_ME_HOME, nor a session workspace.
//
// A GENUINE read failure of the log is surfaced as a one-line diagnostic on the
// diagnostic stream (the report path is offline, so stderr is free), so an
// unreadable log is distinguishable from "no tool ever used"; the all-zero report
// still prints and the command succeeds.
func renderToolUsage(env runtimeEnv, newToolRegistry func(domaintools.OutputSink) domaintools.Registry, newToolUsageStore func(func() (string, error)) history.ToolUsageStore, userHome func() (string, error), lines render.Lines) int {
	// The offline report never executes a tool, so the registry is built with a
	// nil `[Tool Output]` sink (round 052; ADR 0021).
	reg := newToolRegistry(nil)
	tools := reg.Tools()
	counts, err := newToolUsageStore(userHome).Aggregate()
	if err != nil {
		_, _ = fmt.Fprintf(env.stderr, "[tool-usage] could not read the usage log: %v\n", err)
		counts = nil
	}
	rows := make([]history.ToolUsageRow, 0, len(tools))
	for _, t := range tools {
		c := counts[t.Name()] // zero value when the tool has no records
		rows = append(rows, history.ToolUsageRow{Tool: t.Name(), Counts: c})
	}
	_, _ = fmt.Fprint(env.stdout, lines.ToolUsage(rows))
	return Success
}

// renderHistoryList lists the last N persisted messages (round-007 FR-007..FR-009)
// and exits — strictly offline, no provider request. It resolves only the
// workspace (no configuration/provider requirement), so listing works even when
// the configuration is absent.
func renderHistoryList(homeDir string, n int, env runtimeEnv, newHistoryStore func(workspace string) history.Store) int {
	ws, rerr := resolveWorkspace(homeDir)
	if rerr != nil {
		return emitBootError(env.stderr, resolution{Home: homeDir, Workspace: ws}, rerr)
	}
	entries, err := newHistoryStore(ws).Load()
	if err != nil {
		return emitHistoryError(env.stderr, err)
	}
	msgs := toMessages(entries)
	if len(msgs) > n {
		msgs = msgs[len(msgs)-n:]
	}
	for _, m := range msgs {
		_, _ = fmt.Fprintf(env.stdout, "%s: %s\n", m.Role, m.Content)
	}
	return Success
}

// renderNewSession starts a fresh session without a prompt: it archives the
// active history (retaining it) and returns success (round-007 FR-005/FR-006).
func renderNewSession(homeDir string, env runtimeEnv, dp deps.Dependencies) int {
	ws, rerr := resolveWorkspace(homeDir)
	if rerr != nil {
		return emitBootError(env.stderr, resolution{Home: homeDir, Workspace: ws}, rerr)
	}
	if err := dp.NewHistoryStore(ws).Archive(); err != nil {
		return emitHistoryError(env.stderr, err)
	}
	if err := dp.NewUsageStore(ws).Archive(); err != nil {
		return emitHistoryError(env.stderr, err)
	}
	return Success
}

// resolveWorkspace resolves only the runtime home + effective mode + session
// workspace (no configuration/provider), for the session commands `-l` and
// `--new` that must work offline.
func resolveWorkspace(homeDir string) (string, *resolveError) {
	if homeDir == "" {
		return "", &resolveError{Reason: reasonHomeUnset}
	}
	ws, err := home.EnsureWorkspace(homeDir, historyMode(homeDir))
	if err != nil {
		return ws.Path, &resolveError{Reason: reasonHomeUnusable, Err: err}
	}
	return ws.Path, nil
}

// historyMode resolves the effective mode for a session command: the
// TELL_ME_MODE override when set, else the configuration's MODE when the default
// configuration is loadable, else "butler".
func historyMode(homeDir string) string {
	if m := os.Getenv("TELL_ME_MODE"); m != "" {
		return m
	}
	if cfg, err := config.Load(defaultConfigPath(homeDir)); err == nil {
		return cfg.EffectiveMode("")
	}
	return "butler"
}

// toMessages flattens persisted entries into the ordered conversation messages
// (user prompt, assistant answer, …).
func toMessages(entries []history.Entry) []llm.Message {
	msgs := make([]llm.Message, 0, len(entries)*2)
	for _, e := range entries {
		msgs = append(msgs, llm.Message{Role: "user", Content: e.Prompt})
		msgs = append(msgs, llm.Message{Role: "assistant", Content: e.Answer})
	}
	return msgs
}

// writeAnswer writes the provider's answer to the environment's stdout
// (round-006 FR-001/FR-004): the answer bytes verbatim under -r/--raw, or the
// Markdown-rendered form by default. Rendering is gated by -r ALONE — never by
// whether stdout is a terminal. On renderer degradation the renderer's sanitized
// fallback text is written and a one-time non-class warning goes to stderr (the
// frozen `tellme: {phrase}` vocabulary is untouched).
func (e runtimeEnv) writeAnswer(answer string, raw bool, width int) {
	if raw {
		e.writeRawAnswer(answer)
		return
	}
	rendered, degraded := e.renderer.Render(answer, width)
	if degraded {
		e.renderer.WarnDegraded(e.stderr)
		e.writeRawAnswer(rendered) // research D5: the degraded fallback is the sanitized text
		return
	}
	if trimmed := strings.Trim(rendered, "\n"); trimmed != "" {
		_, _ = fmt.Fprint(e.stdout, trimmed+"\n\n")
	}
}

// writeRawAnswer prints text verbatim followed by exactly one CLI-appended
// terminating newline (round-005 FR-006). It serves both the raw (-r) path and
// the sanitized degraded fallback (round-006 research D5).
func (e runtimeEnv) writeRawAnswer(answer string) {
	_, _ = fmt.Fprintln(e.stdout, answer)
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
		// A present-but-invalid value (e.g. a negative rendered width) uses the
		// general configuration-invalid class phrase (round-006 FR-006); a YAML
		// parse failure keeps the parse phrase. The sentinel prefix is stripped so
		// the message does not double the wording.
		if errors.Is(rerr.Err, config.ErrInvalidValue) {
			detail := strings.TrimPrefix(rerr.Err.Error(), config.ErrInvalidValue.Error()+": ")
			_, _ = fmt.Fprintf(stderr, "tellme: the configuration is invalid: %s\n", detail)
		} else {
			_, _ = fmt.Fprintf(stderr, "tellme: the configuration could not be parsed at %s\n", res.Path)
		}
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

// augmentRegistryWithMCP performs the round-032 prompt-path MCP discovery: it
// discovers each enabled remote MCP server's tools (bounded, non-stall), offers
// them ALONGSIDE the native tools, and surfaces any warn+skip messages on the
// diagnostic stream. Discovery runs ONLY on the prompt path (this function is
// called from runTurn), so an offline run makes no MCP network contact. The
// returned close hook tears down the discovered clients when the turn ends.
func augmentRegistryWithMCP(ctx context.Context, res resolution, reg domaintools.Registry, stderr io.Writer, dp deps.Dependencies) (domaintools.Registry, func()) {
	d := dp.MCPDiscoverer(ctx, res.MCPServers)
	for _, w := range res.MCPWarnings {
		_, _ = fmt.Fprintln(stderr, w)
	}
	for _, w := range d.Warnings {
		_, _ = fmt.Fprintln(stderr, w)
	}
	if len(d.Tools) > 0 {
		reg = domaintools.NewRegistry(append(reg.Tools(), d.Tools...)...)
	}
	closeFn := func() {}
	if d.Closer != nil {
		closeFn = func() { _ = d.Closer.Close() }
	}
	return reg, closeFn
}

// emitToolError maps an incomplete tool loop to the frozen tool class phrase and
// dedicated exit code (round-008 FR-010 / Clarify R2 Q1). The trailing detail is
// contract-free; newlines are folded so exactly one line carries the phrase.
func emitToolError(w io.Writer, err error) int {
	detail := strings.ReplaceAll(err.Error(), "\n", " ")
	_, _ = fmt.Fprintf(w, "tellme: the tool request failed: %s\n", detail)
	return ToolError
}

// emitHistoryError maps a session-history read/write failure to the environment
// class phrase and exit code (round-007 Decision 6 / Clarify Q3). The trailing
// detail is contract-free; newlines are folded so exactly one line carries the
// phrase, keeping the frozen vocabulary at ten.
func emitHistoryError(w io.Writer, err error) int {
	detail := strings.ReplaceAll(err.Error(), "\n", " ")
	_, _ = fmt.Fprintf(w, "tellme: the runtime home is not usable (session history: %s)\n", detail)
	return EnvironmentError
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
