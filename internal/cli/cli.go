package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"github.com/gosharplite/tellme/internal/agent"
	appsuggestions "github.com/gosharplite/tellme/internal/app/suggestions"
	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	"github.com/gosharplite/tellme/internal/home"
	infrhistory "github.com/gosharplite/tellme/internal/infrastructure/history"
	infrallm "github.com/gosharplite/tellme/internal/infrastructure/llm"
	infratools "github.com/gosharplite/tellme/internal/infrastructure/tools"
	"github.com/gosharplite/tellme/internal/ui"
	tuiprompt "github.com/gosharplite/tellme/internal/ui/tui/prompt"
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
	configPath  string
	diagnostic  bool
	version     bool
	raw         bool
	newSession  bool
	list        int
	listSet     bool
	interactive bool
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
	// Person is the resolved PERSON — the persona sent to the provider as the
	// leading `system` message of every request (round 011).
	Person string
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
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	isTTY    func(any) bool
	renderer answerRenderer
	// clock is the injected time seam for the payload status line (round 009);
	// nil falls back to time.Now.
	clock func() time.Time
}

// answerRenderer renders a Markdown answer to ANSI (round 006). It is the seam
// that keeps the render/raw mode selection unit-testable without a real
// renderer. On degradation it returns the (sanitized) text the caller should
// fall back to.
type answerRenderer interface {
	// Render returns the rendered answer, and whether rendering degraded (in
	// which case the returned string is the sanitized raw fallback text).
	Render(markdown string, width int) (text string, degraded bool)
	// WarnDegraded emits the one-time degradation warning.
	WarnDegraded(w io.Writer)
}

// newRenderer builds the production answer renderer (a var so tests may inject a
// fake).
var newRenderer = func() answerRenderer { return ui.NewRenderer() }

// historyStoreFactory builds the session-history store for a resolved workspace.
// It is the DI seam (round-007 TD-1): the presentation layer never couples to a
// concrete file adapter, and tests can inject an in-memory fake without disk
// I/O. It mirrors gatewayFactory / answerRenderer.
type historyStoreFactory func(workspace string) history.Store

// newHistoryStore is the production history-store factory (a var so tests may
// override it).
var newHistoryStore historyStoreFactory = func(workspace string) history.Store {
	return infrhistory.NewFileStore(workspace)
}

// usageStoreFactory builds the per-mode usage-log store for a resolved workspace
// (round 018). It mirrors historyStoreFactory so the presentation layer never
// couples to the concrete file adapter.
type usageStoreFactory func(workspace string) history.UsageStore

// newUsageStore is the production usage-store factory (a var so tests may
// override it).
var newUsageStore usageStoreFactory = func(workspace string) history.UsageStore {
	return infrhistory.NewUsageStore(workspace)
}

// tuiPromptRunner runs the interactive TUI prompt (round 015) for one invocation
// and returns the composed prompt text plus whether a prompt was submitted (ok).
// It is the DI seam (PR #38 review directive ④): the -i / USE_TUI_PROMPT / non-TTY
// dispatch matrix is unit-testable without a terminal event loop, mirroring
// gatewayFactory / historyStoreFactory.
type tuiPromptRunner func(ctx context.Context, res resolution, env runtimeEnv) (string, bool, error)

// newTUIPromptRunner is the production TUI runner (a var so tests may inject a
// fake).
var newTUIPromptRunner tuiPromptRunner = defaultRunTUIPrompt

// defaultRunTUIPrompt runs the interactive TUI prompt (round 015): it announces
// on the diagnostic stream, drives the prompt bound to stderr/stdin through the
// Bubble Tea runtime, and returns the composed prompt (ok) on submit. It never
// writes to stdout (PR #38 review BLOCKER).
func defaultRunTUIPrompt(ctx context.Context, res resolution, env runtimeEnv) (string, bool, error) {
	_, _ = fmt.Fprintln(env.stderr, TUIHint)

	tracker := infrhistory.NewGlobalPromptTracker(res.Home)
	defer func() { _ = tracker.Close(context.Background()) }()
	reg := domaintools.NewRegistry(infratools.NewFilesystemTools()...)
	engine := appsuggestions.New(
		appsuggestions.TrackerPrompts{Tracker: tracker},
		appsuggestions.OSSWorkspace{},
		appsuggestions.RegistryTools{Registry: reg},
	)
	src := tuiSource{svc: engine}
	// No startup disk I/O: the dashboard header was retired (round-016 FR-004 /
	// architect D3), so the history store is no longer read to build the prompt.
	return tuiprompt.Run(ctx, env.stdin, env.stderr, src, tuiDebounceDuration())
}

// tuiDebounceDuration resolves the suggestion-refresh debounce. The hermetic E2E
// sets TELL_ME_TUI_DEBOUNCE=0 so the scripted keys observe suggestions
// synchronously (round-016); otherwise the reference ~100 ms applies.
func tuiDebounceDuration() time.Duration {
	if v := strings.TrimSpace(os.Getenv("TELL_ME_TUI_DEBOUNCE")); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms >= 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return tuiprompt.DefaultDebounceDuration
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
func tuiRequested(homeDir string, opts *options) bool {
	if opts.interactive {
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
func runTUIPrompt(homeDir string, opts *options, env runtimeEnv) int {
	// Best-effort resolution: the TUI engages even when the setup is unresolved
	// (an aborted/empty submission owes no request — the round-012 ordering
	// rationale). The resolution feeds the dashboard and the submit path.
	res, _ := resolve(homeDir, opts.configPath)
	text, ok, err := newTUIPromptRunner(context.Background(), res, env)
	if err != nil {
		return emitProviderError(env.stderr, err)
	}
	if !ok || text == "" {
		return Success
	}
	// Record in the shared log (round-015 FR-009) — only the interactive prompt
	// writes it.
	tracker := infrhistory.NewGlobalPromptTracker(res.Home)
	_ = tracker.Append(context.Background(), text)
	_ = tracker.Close(context.Background())
	return renderTurn(homeDir, opts.configPath, text, turnOptions{raw: opts.raw}, env)
}

// Run is the CLI entrypoint: main passes argv and the injected build version,
// and Run returns the process exit code. It binds the real process streams, the
// default terminal detector, and the production renderer into a runtimeEnv, then
// delegates to run (round-005 research Decision 4 — the seam keeps the
// input/output-mode selection unit-testable; round-006 review Obs 2 — one value
// instead of many stream primitives).
func Run(args []string, version string) int {
	return run(args, version, runtimeEnv{
		stdin:    os.Stdin,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		isTTY:    terminalDetector(),
		renderer: newRenderer(),
		clock:    time.Now,
	})
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
func run(args []string, version string, env runtimeEnv) int {
	opts, flagArgs, ok := parseFlags(args, env.stderr)
	if !ok {
		return emitUsageError(env.stderr)
	}
	if opts.version {
		_, _ = fmt.Fprintf(env.stdout, "tellme %s\n", version)
		return Success
	}

	homeDir := os.Getenv("TELL_ME_HOME")

	// -d is the reporting path: it always produces a report (Decision 2), and it
	// takes precedence over a prompt or piped input (round-004 Decision 7 /
	// round-005 FR-010).
	if opts.diagnostic {
		return renderDiagnostic(homeDir, opts.configPath, env.stdout)
	}
	// -l is a terminal reporting command: it lists the last N messages and exits,
	// strictly offline (round-007 FR-007/FR-008). A non-positive N is a usage
	// error, evaluated before any network or stdin access (RF-2).
	if opts.listSet {
		if opts.list <= 0 {
			return emitUsageError(env.stderr)
		}
		return renderHistoryList(homeDir, opts.list, env)
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
		return renderTurn(homeDir, opts.configPath, prompt, turnOptions{raw: opts.raw, newSession: opts.newSession, chrome: true}, env)
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
		// Round 015 — the opt-in interactive TUI prompt engages here (only when
		// enabled AND stdin is a terminal); the plain reader below stays the
		// default. The dispatch delegates to the tuiPromptRunner seam so the
		// matrix is unit-testable (PR #38 review directive ④).
		if tuiRequested(homeDir, opts) {
			return runTUIPrompt(homeDir, opts, env)
		}
		if opts.newSession {
			// Archive BEFORE resolving the configuration: a prompt-less --new is an
			// archive command that works offline, so — unlike the prompt-bearing
			// `--new "<prompt>"` form, which resolves first — a broken config still
			// archives here and then fails when the turn resolves (round-012 review).
			if code := renderNewSession(homeDir, env); code != Success {
				return code
			}
		}
		ictx, icancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		text, ok := readInteractivePrompt(ictx, env.stdin, env.stderr)
		icancel()
		if !ok || text == "" {
			return Success
		}
		return renderTurn(homeDir, opts.configPath, text, turnOptions{raw: opts.raw, chrome: true}, env)
	}
	// A prompt-less --new on a NON-terminal keeps its round-007 behaviour: archive
	// the session and exit (the reader never engages on a non-terminal).
	if opts.newSession {
		return renderNewSession(homeDir, env)
	}
	return renderBoot(homeDir, opts.configPath, env)
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
	fs.BoolVarP(&o.raw, "raw", "r", false, "Print the answer as raw text (no Markdown rendering).")
	fs.BoolVar(&o.newSession, "new", false, "Start a fresh session, archiving the current session history.")
	fs.IntVarP(&o.list, "list", "l", 0, "List the last N messages of the session history and exit.")
	fs.BoolVarP(&o.interactive, "interactive", "i", false, "Open the interactive TUI prompt (requires a terminal).")
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

// gatewayFactory builds the provider gateway for a resolved provider entry. It
// is the composition seam (review finding #1): the presentation layer never
// couples to a concrete adapter constructor, tests can inject a fake
// llm.Gateway, and an un-adapted family surfaces as an actionable error.
type gatewayFactory func(prov config.Provider, name, persona string) (llm.Gateway, error)

// newGateway is the production gateway factory (a var so tests may override it).
var newGateway gatewayFactory = infrallm.NewGateway

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
}

func renderTurn(homeDir, configPath, prompt string, opts turnOptions, env runtimeEnv) int {
	res, rerr := resolve(homeDir, configPath)
	if rerr != nil {
		return emitBootError(env.stderr, res, rerr)
	}
	store := newHistoryStore(res.Workspace)
	if opts.newSession {
		if err := store.Archive(); err != nil {
			return emitHistoryError(env.stderr, err)
		}
		if err := newUsageStore(res.Workspace).Archive(); err != nil {
			return emitHistoryError(env.stderr, err)
		}
	}
	return runTurn(res, store, prompt, opts, env, newGateway)
}

// runTurn performs one reasoning turn through an injected gateway factory and
// history store. The resumed conversation is loaded and carried ahead of the
// current prompt; the completed turn is appended after the provider answers
// (append-after-complete, round-007 FR-001/FR-003). The answer is written by the
// runtimeEnv's renderer. The context is cancelled on SIGINT/SIGTERM so a stalled
// provider can be interrupted (review finding #2).
func runTurn(res resolution, store history.Store, prompt string, opts turnOptions, env runtimeEnv, factory gatewayFactory) int {
	gw, err := factory(res.Provider, res.Selected, res.Person)
	if err != nil {
		return emitProviderError(env.stderr, err)
	}
	prior, err := store.Load()
	if err != nil {
		return emitHistoryError(env.stderr, err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Pre-flight payload status (round-009 FR-001): the estimated size of the
	// assembled conversation — the resumed turns (via the shared projection,
	// including tool steps — TD-1) plus the current prompt — measured against the
	// payload budget. Diagnostic only, on stderr.
	reg := newToolRegistry(store, gw)
	assembled := append(append(make([]llm.Message, 0, len(prior)+1), agent.BuildMessages(prior)...), llm.Message{Role: "user", Content: prompt})
	// Round-017 turn chrome: on surfaces (A)/(B) the turn opens with the
	// input-capture acknowledgement and the rule/header frame, wrapping the
	// pre-flight payload line. It is false for the `-i` submit path and the
	// non-prompt paths (FR-007).
	if opts.chrome {
		emitInputCaptured(env)
		// Turn <N> = the session's completed-turn count + 1, derived from the
		// loaded active history. Forward item (round-017 review finding 3): when
		// sliding-window summarisation/archival lands, `len(prior)` will
		// undercount and this must use a total-lifetime count, e.g. a
		// history.Store.Count() method.
		emitTurnOpening(env, len(prior)+1, res.Mode)
	}
	emitPayloadStatus(env, res, llm.EstimatePayload(res.Person, agent.ToolDefs(reg), assembled), true)
	if opts.chrome {
		emitTurnGap(env)
	}

	loop := &agent.AgentLoop{
		Gateway:  gw,
		Registry: reg,
		MaxLoops: res.MaxToolLoop,
		Stderr:   env.stderr,
	}
	result, err := loop.Run(ctx, prompt, prior)
	if err != nil {
		var inc *agent.ErrIncomplete
		if errors.As(err, &inc) {
			return emitToolError(env.stderr, inc)
		}
		return emitProviderError(env.stderr, err)
	}
	if err := store.Append(history.Entry{Prompt: prompt, Answer: result.Answer, Steps: result.Steps}); err != nil {
		return emitHistoryError(env.stderr, err)
	}
	env.writeAnswer(result.Answer, opts.raw, res.WrapWidth)
	// Post-turn payload status (round-009 FR-006): the provider's measured prompt
	// tokens, written AFTER the answer so it trails the response (the pre-flight
	// line led it). Omitted when the provider reported no usage.
	if result.Usage.Reported {
		emitPayloadStatus(env, res, result.Usage.PromptTokens, false)
	}
	// Round 018 — the post-turn status lines (the metrics line + the `╰─⠿ Ready`
	// summary), written AFTER the measured payload line; suppressed when the
	// just-returned call reports no usage (FR-012).
	emitPostTurnStatus(env, res, result)
	return Success
}

// emitPayloadStatus writes one payload status line to the diagnostic stream
// (stderr) using the runtime's injected clock seam (round-009 FR-001/FR-006).
// `estimated` selects the pre-flight `~` form; the measured form omits it. The
// line carries no `tellme: ` prefix (FR-014) and names the effective mode and the
// provider's configured MODEL (TD-2).
func emitPayloadStatus(env runtimeEnv, res resolution, tokens int, estimated bool) {
	_, _ = fmt.Fprintln(env.stderr, ui.FormatPayloadStatus(env.now(), tokens, res.MaxHistoryTokens, res.Mode, res.Provider.Model, estimated))
}

// now returns the current time from the injected clock seam (falling back to
// time.Now) — the shared round-009/017 clock seam.
func (e runtimeEnv) now() time.Time {
	if e.clock != nil {
		return e.clock()
	}
	return time.Now()
}

// emitPostTurnStatus writes the round-018 post-turn status to the diagnostic
// stream and persists the turn's per-call usage to the per-mode usage log. It is
// a no-op when the just-returned call reports no usage (FR-012); otherwise it:
//   - loads the session's prior usage, appends one record per reported call, and
//     emits the metrics line (`M/H/C/Th` of the just-returned call) and the
//     `╰─⠿ Ready` summary (three costs + the session token totals + hit-rate).
//
// The usage-log write is best-effort: a log failure never breaks a completed
// turn (the answer is already on stdout).
func emitPostTurnStatus(env runtimeEnv, res resolution, result agent.AgentResult) {
	if !result.Usage.Reported {
		return
	}
	us := newUsageStore(res.Workspace)
	prior, _ := us.Load()

	pricing := ui.Pricing{Hit: res.Pricing.HIT, Miss: res.Pricing.MISS, Comp: res.Pricing.COMP}
	now := env.now()

	var turnCost, lastCost float64
	turnRecords := make([]history.UsageRecord, 0, len(result.Calls))
	for _, c := range result.Calls {
		if !c.Reported {
			continue
		}
		miss := c.PromptTokens - c.CachedTokens
		cost := ui.ComputeCost(pricing, miss, c.CachedTokens, c.CompletionTokens, c.ThinkingTokens)
		turnCost += cost
		turnRecords = append(turnRecords, history.UsageRecord{
			Timestamp:      now.Format(time.RFC3339),
			Provider:       res.Selected,
			Model:          res.Provider.Model,
			CachedTokens:   c.CachedTokens,
			PromptTokens:   c.PromptTokens,
			ResponseTokens: c.CompletionTokens,
			TotalTokens:    c.PromptTokens + c.CompletionTokens + c.ThinkingTokens,
			ThinkingTokens: c.ThinkingTokens,
			Cost:           cost,
		})
	}
	if len(turnRecords) > 0 {
		lastCost = turnRecords[len(turnRecords)-1].Cost
	}
	for _, rec := range turnRecords {
		if err := us.Append(rec); err != nil {
			return // best-effort: never break the turn over the usage log
		}
	}

	sMiss, sHit, sOut := 0, 0, 0
	sessionCost := 0.0
	add := func(r history.UsageRecord) {
		sMiss += r.PromptTokens - r.CachedTokens
		sHit += r.CachedTokens
		sOut += r.ResponseTokens + r.ThinkingTokens
		sessionCost += r.Cost
	}
	for _, r := range prior {
		add(r)
	}
	for _, r := range turnRecords {
		add(r)
	}

	last := result.Usage
	_, _ = fmt.Fprintln(env.stderr, ui.FormatMetrics(env.now(), res.Selected, ui.UsageCounts{
		Miss:       last.PromptTokens - last.CachedTokens,
		Hit:        last.CachedTokens,
		Completion: last.CompletionTokens,
		Thinking:   last.ThinkingTokens,
	}))
	_, _ = fmt.Fprintln(env.stderr, ui.FormatReady(lastCost, turnCost, sessionCost, sMiss, sHit, sOut, ui.HitRate(sHit, sMiss)))
}

// emitInputCaptured writes the round-017 input-capture acknowledgement to the
// diagnostic stream (the reference's `[HH:MM:SS] Input captured. Processing...`).
func emitInputCaptured(env runtimeEnv) {
	_, _ = fmt.Fprintln(env.stderr, ui.FormatInputCaptured(env.now()))
}

// emitTurnOpening writes the round-017 frame opening (a leading blank line, the
// 80-column rule, and the `╭─⠿ Turn <N> - <mode>` header) to the diagnostic stream.
func emitTurnOpening(env runtimeEnv, turn int, mode string) {
	_, _ = fmt.Fprint(env.stderr, ui.FormatTurnOpening(turn, mode))
}

// emitTurnGap writes the blank line that separates the frame from the answer.
func emitTurnGap(env runtimeEnv) {
	_, _ = fmt.Fprint(env.stderr, ui.FormatTurnGap())
}

// renderHistoryList lists the last N persisted messages (round-007 FR-007..FR-009)
// and exits — strictly offline, no provider request. It resolves only the
// workspace (no configuration/provider requirement), so listing works even when
// the configuration is absent.
func renderHistoryList(homeDir string, n int, env runtimeEnv) int {
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
func renderNewSession(homeDir string, env runtimeEnv) int {
	ws, rerr := resolveWorkspace(homeDir)
	if rerr != nil {
		return emitBootError(env.stderr, resolution{Home: homeDir, Workspace: ws}, rerr)
	}
	if err := newHistoryStore(ws).Archive(); err != nil {
		return emitHistoryError(env.stderr, err)
	}
	if err := newUsageStore(ws).Archive(); err != nil {
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

// toolRegistryFactory builds the tool registry offered to the model for one
// prompt run. It is the DI seam for the tool layer (round-008 TD-1, review
// PR #25): the presentation layer never hard-wires the concrete tool adapters
// (mirroring gatewayFactory / historyStoreFactory), so tests can inject a fake
// registry.
type toolRegistryFactory func(store history.Store, gw llm.Gateway) domaintools.Registry

// newToolRegistry is the production registry factory (a var so tests may
// override it). It assembles the two read-only filesystem tools plus the
// LLM-backed session-summarisation tool (round-008 research Decisions 4 & 10).
var newToolRegistry toolRegistryFactory = func(store history.Store, gw llm.Gateway) domaintools.Registry {
	ts := infratools.NewFilesystemTools()
	ts = append(ts, infratools.NewSummarizeHistoryTool(store, gw))
	return domaintools.NewRegistry(ts...)
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
