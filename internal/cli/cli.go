package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/domain/render"
	domaintui "github.com/gosharplite/tellme/internal/domain/tui"
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
	configPath string
	diagnostic bool
	// help is the round-074 `-h`/`--help` flag; helpText is the rendered flag
	// list (captured at parse time so run can print it to stdout and exit 0
	// without rebuilding the FlagSet).
	help        bool
	helpText    string
	version     bool
	raw         bool
	newSession  bool
	list        int
	listSet     bool
	back        int
	backSet     bool
	interactive bool
	toolUsage   bool
	turns       bool
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

// cliDeps is the CLI's injected seam bundle (round 098; ADR 0068): the
// composition root's deps.Dependencies plus the interactive-prompt port. The
// offline reporters and the interactive path used to take up to FOUR adjacent
// func-typed parameters — an injected-closure list, which the parameter-shape
// gate reds (arm (c)) and whose house idiom is a named spec struct (ToolSetSpec,
// LoopSpec, turnOptions). ONE bundle replaces all five lists.
type cliDeps struct {
	dp       deps.Dependencies
	prompter domaintui.Prompter
}

// cliDepsFrom adapts the public Options at the CLI boundary.
func cliDepsFrom(o Options) cliDeps { return cliDeps{dp: o.Deps, prompter: o.Prompter} }

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
	// stdoutTTY is the standard-output (stdout) terminal probe (round 073;
	// ADR 0045). It is tellme's first stdout probe and is used ONLY by the `-l`
	// listing's header-accent gate — the prompt path wires no stdout chrome, so
	// round-006 / PR #16 Obs 1 stays OPEN. Nil reports "not a terminal" (the
	// safe, byte-plain default).
	stdoutTTY func(any) bool
	renderer  render.Answer
	// clock is the injected time seam for the payload status line (round 009);
	// nil falls back to time.Now.
	clock func() time.Time
	// turnsLog, when non-nil, is the session turn-log writer (round 053; ADR
	// 0022): the per-call renderer writes the PLAIN render of each chrome line
	// here (round 054 fold B-54-1 — the artifact stays control-free even when the
	// stderr leg is coloured). nil on the offline / non-prompt paths. The
	// prompt-surface ack and the `-i` echo never write here (round-053 F-53-3(i)).
	turnsLog io.Writer
}

// round 044: the history/usage/toolUsage-store, gateway and tool-registry factory
// vars moved to cmd/tellme + deps.Dependencies; the renderer var was deleted
// (built inline via ui.NewRenderer()) (ADR 0013). round 048 (ADR 0017): the TUI
// runner var is gone too — the interactive prompt is reached through the
// injected domaintui.Prompter port, with no in-package default (a nil port is a
// loud EnvironmentError; see runTUIPrompt).

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
		stdoutTTY: stdoutTerminalDetector(),
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

// stdoutTerminalDetector returns the process's standard-output (stdout) terminal
// probe. The diagnostic environment seam TELL_ME_FORCE_STDOUT_TTY forces the
// stdout probe to report a terminal so the round-073 listing's header accents are
// E2E-drivable without a pty (mirroring the TELL_ME_FORCE_STDERR_TTY seam). It is
// tellme's first stdout probe; it is wired for the `-l` listing colour ONLY, so
// round-006 / PR #16 Obs 1 (stdout *chrome* on a prompt turn) stays OPEN.
func stdoutTerminalDetector() func(any) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TELL_ME_FORCE_STDOUT_TTY"))) {
	case "1", "true", "yes":
		return func(any) bool { return true }
	default:
		return defaultIsTerminal
	}
}

// stdoutIsTerminal reports whether standard output is a terminal via the
// dedicated stdout probe (round 073). With no probe set it reports false (the
// safe, byte-plain default) — never falling back to the shared (stdin) probe.
func (e runtimeEnv) stdoutIsTerminal() bool {
	if e.stdoutTTY == nil {
		return false
	}
	return e.stdoutTTY(e.stdout)
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
	// Round 074 (ADR 0046): help is a successful, offline, prompt-less action and
	// precedes every other terminal action; `-v` shares the version path below.
	if f.help {
		_, _ = fmt.Fprint(env.stdout, f.helpText)
		return Success
	}
	if f.version {
		_, _ = fmt.Fprintf(env.stdout, "tellme %s\n", version)
		return Success
	}

	homeDir := os.Getenv("TELL_ME_HOME")
	in := cliDepsFrom(scoped)
	ref := sessionRef{home: homeDir, config: f.configPath}

	// The offline reporting commands run in precedence order — -d → -l → -b (standalone)
	// → -t → --tool-usage — before any prompt or stdin access. hasPrompt tells the
	// dispatcher whether a positional prompt is present (a prompt-bearing `-b` is
	// handled by the prompt path below).
	if code, handled := dispatchReporting(f, ref, env, len(flagArgs) > 0, in); handled {
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
		return renderTurn(ref, prompt, turnOptions{raw: f.raw, newSession: f.newSession, chrome: true, rollbackTurns: f.back}, env, dp)
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
		// previously dropped `--new`, so the header counted the prior history).
		// Round 053 (closes #103; ADR 0022; fold F-53-2): a prompt-less `--new`
		// now selects the session from the `-c` config's MODE, so an explicit `-c`
		// that cannot be honoured REFUSES before archiving (Q2 → A) — the round-012
		// "a broken config still archives here" ordering is SUPERSEDED (failing
		// loudly beats archiving the WRONG session).
		if f.newSession {
			if code := renderNewSession(ref, env, in); code != Success {
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
		return renderTurn(ref, text, turnOptions{raw: f.raw, chrome: true}, env, dp)
	}
	// A prompt-less --new on a NON-terminal keeps its round-007 behaviour: archive
	// the session and exit (the reader never engages on a non-terminal).
	if f.newSession {
		return renderNewSession(ref, env, in)
	}
	return renderBoot(ref, env)
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
	// Round 074 (ADR 0046): `-h`/`--help` is an EXPLICIT flag, so pflag's
	// implicit help path (which writes to the SetOutput writer = stderr and
	// returns ErrHelp, i.e. the caller's usage error → exit 2) is superseded;
	// tellme prints the flag list to stdout and exits 0. `-v` is the shorthand
	// for `--version`.
	fs.BoolVarP(&o.help, "help", "h", false, "Print this help (the flag list) and exit.")
	fs.BoolVarP(&o.version, "version", "v", false, "Print the build version and exit.")
	fs.BoolVarP(&o.raw, "raw", "r", false, "Print the answer (and the -l history listing) as raw text, without Markdown rendering.")
	fs.BoolVar(&o.newSession, "new", false, "Start a fresh session, archiving the current session history.")
	fs.IntVarP(&o.list, "list", "l", 0, "List the last N messages of the session history and exit. Defaults to 1 when the value is omitted.")
	// Round 054 (ADR 0023): bare `-l`/`--list` defaults to 1 (reference parity).
	fs.Lookup("list").NoOptDefVal = "1"
	// Round 081 (ADR 0053): `-b`/`--back` rolls back the last N turns of the
	// session history (defaults to 1 when the value is omitted) and exits; with a
	// positional prompt it rolls back first, then runs the prompt.
	fs.IntVarP(&o.back, "back", "b", 0, "Roll back the last N turns of the session history and exit. Defaults to 1 when the value is omitted.")
	fs.Lookup("back").NoOptDefVal = "1"
	fs.BoolVarP(&o.turns, "turns", "t", false, "Print the session's turn log and exit.")
	fs.BoolVarP(&o.interactive, "interactive", "i", false, "Open the interactive TUI prompt (requires a terminal).")
	fs.BoolVar(&o.toolUsage, "tool-usage", false, "Report per-tool invocation counts across all sessions, then exit.")
	if err := fs.Parse(consumeOptionalIntArgs(args, "-l", "--list", "-b", "--back")); err != nil {
		return nil, nil, false
	}
	o.listSet = fs.Changed("list")
	o.backSet = fs.Changed("back")
	if o.help {
		// The flag list is pflag's own rendering — the text pflag's IMPLICIT
		// help path wrote to the SetOutput writer (stderr) before this round —
		// captured here so run prints it to stdout (the usage-ERROR path prints
		// only the class phrase, no block).
		o.helpText = "Usage of tellme:\n" + fs.FlagUsages()
	}
	return o, fs.Args(), true
}

// consumeOptionalIntArgs normalizes an optional-integer flag (round 054 `-l`/
// `--list`, generalised in round 081 / ADR 0053 to also cover `-b`/`--back`) so
// an ADJACENT value is consumed (`-l 5` → `-l=5`, `-b 3` → `-b=3`) instead of
// being left as a positional argument. pflag's NoOptDefVal makes a bare `-l`/`-b`
// mean the default, but it also means `-l 5` parses as `-l=1` plus a positional
// `5` — a prompt token that would otherwise be swallowed or misread. Mirroring
// the reference's `consumeOptionalIntFlag`: consume the next token ONLY when it
// is a valid integer; otherwise leave it (so `-b "prompt"` ⇒ 1 + the prompt).
// An `--` separator stops the scan.
func consumeOptionalIntArgs(args []string, names ...string) []string {
	if len(args) == 0 {
		return args
	}
	optional := make(map[string]bool, len(names))
	for _, n := range names {
		optional[n] = true
	}
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			out = append(out, args[i:]...)
			return out
		}
		if optional[arg] {
			if i+1 < len(args) {
				if _, err := strconv.Atoi(args[i+1]); err == nil {
					out = append(out, arg+"="+args[i+1])
					i++
					continue
				}
			}
			out = append(out, arg)
			continue
		}
		out = append(out, arg)
	}
	return out
}
