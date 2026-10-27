package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gosharplite/tellme/internal/app/deps"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/render"
	"github.com/gosharplite/tellme/internal/home"
)

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
	// rollbackTurns is the round-081 `-b`/`--back [N]` count: when > 0 the turn
	// rolls back that many turns of the session history FIRST, then runs (the
	// prompt-bearing form). Zero means no rollback (the flag was not given).
	rollbackTurns int
}

func renderTurn(target sessionRef, prompt string, opts turnOptions, env runtimeEnv, dp deps.Dependencies) int {
	res, rerr := resolve(target)
	if rerr != nil {
		return emitBootError(env.stderr, res, rerr)
	}
	// Round 081 (ADR 0053): a prompt-bearing `-b [N] "prompt"` rolls back the last
	// N turns FIRST, then runs the prompt against the trimmed history. `res` is the
	// same session the rollback resolves (mode precedence identical), so the rollback
	// targets the workspace this turn will use. The single store is reused for the
	// rollback, the confirmation count, the optional archive, and the turn.
	store := dp.NewHistoryStore(res.Workspace)
	if opts.rollbackTurns > 0 {
		removed, err := store.Rollback(opts.rollbackTurns)
		if err != nil {
			return emitHistoryError(env.stderr, err)
		}
		remaining, err := store.Load()
		if err != nil {
			return emitHistoryError(env.stderr, err)
		}
		_, _ = fmt.Fprint(env.stdout, rollbackConfirmation(removed, len(remaining)))
	}
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

	// Round 053 (ADR 0022) + round 054 fold B-54-1: the per-call renderer writes
	// each chrome line to BOTH sinks — the diagnostic stream (coloured per the
	// gate) and the resolved session's turns.log (PLAIN). Best-effort — an open
	// failure (or an empty workspace) leaves the chrome on stderr alone; the handle
	// is closed at run end.
	if res.Workspace != "" {
		if wc, err := dp.NewTurnsLogStore(res.Workspace).Writer(); err == nil {
			defer func() { _ = wc.Close() }()
			env.turnsLog = wc
		}
	}

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
	// Round 054 fold RF-54-2: resolve the shared terminal gate ONCE per turn (the
	// spinner and the chrome colour share one predicate, chromeColour) so the
	// three consumers cannot disagree.
	colourOn := chromeColour(opts, env)
	// Round 057 fold TD-057-2: named fields, so the spinner gate and the
	// chrome-colour gate cannot be silently swapped.
	prog := dp.NewProgress(render.ProgressSpec{
		Stream:  env.stderr,
		Now:     env.now,
		Model:   res.Provider.Model,
		Epoch:   turnStart,
		Columns: stderrColumns(env),
		IdleGap: toolOutputIdleGap(dp.NewLines(false)),
		Spinner: colourOn,
		Colour:  colourOn,
	})
	ind := prog.Indicator
	if ind != nil {
		defer ind.Stop() // panic-safe residue guard (idempotent)
	}

	// Pre-flight payload status (round-009 FR-001; round 057 ADR 0027): the
	// estimated size of the assembled conversation — the resumed turns (via the
	// shared projection, including tool steps — TD-1) plus the current prompt —
	// rendered with its increment over the previous estimate and WITHOUT a budget
	// (the budget is shown only on the measured post-turn line). Diagnostic only,
	// on stderr.
	// Round 052 (closes #115 R-2; ADR 0021): the registry is built with the
	// `[Tool Output]` sink injected at construction (`prog.ToolOutput`) — the
	// round-034 `BindToolOutput` rebind no longer exists. Round 069 (ADR 0039):
	// the build takes ONE named capability value (`deps.ToolSetSpec`), not
	// positional scalars.
	reg := dp.NewToolRegistry(deps.ToolSetSpec{Sink: prog.ToolOutput, Vision: res.Provider.Vision, ProviderType: res.Provider.Type})
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
		// Round 097 (ADR 0067): the acknowledgement is the FIFTH chrome accent
		// (bright blue, whole line). It rides the SAME gate as the spinner and the
		// payload/`Ready` accents — and it CONSUMES the already-resolved colourOn
		// (round 054 fold RF-54-2: resolve the shared terminal gate ONCE per turn so
		// the consumers cannot disagree), never re-running chromeColour here: a
		// second probe could disagree with line 808's verdict and colour the ack
		// while the spinner/payload/`Ready` accents stayed plain in the same turn.
		// NOTE: the ack gates on opts.chrome, NOT on `raw`, so under `-r` the LINE
		// stays and only its COLOUR drops; the echo above stays plain (issue #204
		// rev-1, withdrawn).
		emitInputCaptured(env, dp.NewLines(colourOn))
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
	renderer := newCallRenderer(env, res, reg, opts.chrome, turnNumber(prior)-1, colourOn, dp)

	// Round 034 (ADR 0005 D1): the loop keeps a single observer — the composite
	// composes the per-call block renderer with the round-019 spinner.
	observer := compositeObserver{call: renderer, spinner: ind}

	// Round 078 (ADR 0050): wrap the gateway with the bounded transport retry — a
	// retryable provider failure (transport, or HTTP 429/5xx) is retried at most
	// twice (1 s, 3 s) before the provider phrase + exit 6. It is the INNER wrapper
	// (below the round-068 unpaired diagnostic) so the retry re-sends the SAME
	// request through the one seam, and the retry line is written with the spinner
	// YIELDED (the indicator is built above, so the notify closure can clear the
	// live frame before the line and restore it after — never tearing it).
	gw = withProviderRetry(gw, retryNotifier(env, ind))
	// Round 068 (ADR 0038; resolves ADR 0037 RF-067-1): wrap the gateway so a
	// Gemini/Vertex round that leaves a tool call unanswered (M < N) is reported as
	// a `[Tool …]` diagnostic on stderr BEFORE the request is sent — informational,
	// never routed to turns.log (Q1/Q2 → A). The shipped M == N path is inert. It is
	// the OUTER wrapper (one emission per loop call, before the retry loop).
	gw = withUnpairedDiagnostic(gw, unpairedEmitter(env, dp.NewLines(chromeColour(opts, env))))

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
		Lines: dp.NewToolLines(colourOn),
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
		return failTurn(env, store, prompt, err, result, ctx)
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

// failTurn maps a loop error to its terminal surface (round 079 / ADR 0051 +
// round 080 / ADR 0052).
//
//   - An OPERATOR interruption (the turn context was cancelled by SIGINT/SIGTERM
//     — `ctx.Err() != nil` — or an error that wraps context.Canceled) that
//     completed at least one tool step keeps its work: the partial turn is
//     persisted, closed with `history.InterruptedTurnAnswer`, and reports Success.
//   - A FAILED turn (a provider failure, or the tool-loop failure) that completed
//     at least one tool step ALSO keeps its work: the partial turn is persisted,
//     closed with a class-specific synthetic answer, and then the failure surface
//     is reported UNCHANGED (the frozen phrase + exit 6/7) with one informational
//     stderr line.
//   - Any turn that failed/interrupted with ZERO completed steps keeps today's
//     clean abort (no write).
//
// Detection is STRUCTURAL/typed — never string-matched (NFR-004). `ctx.Err()`
// sees a SIGINT/SIGTERM `defer cancel()` has not yet run; the `errors.Is` term
// keeps the case where the adapter surfaces the cancellation only via the error;
// both cover the round-078 retry decorator's `lastErr`-on-abort (issue #161 hole
// #2: a Ctrl+C during a retry wait).
func failTurn(env runtimeEnv, store history.Store, prompt string, err error, result agentport.Result, ctx context.Context) int {
	interrupted := ctx.Err() != nil || errors.Is(err, context.Canceled)
	if len(result.Steps) > 0 && interrupted {
		return persistInterruptedTurn(env, store, prompt, result)
	}
	var inc *agentport.ErrIncomplete
	isTool := errors.As(err, &inc)
	kept := len(result.Steps) > 0
	// Round 080 (ADR 0052): a failed turn with completed steps keeps them. The
	// append is BEST-EFFORT — a failure never masks the failure surface (D6).
	stored := false
	if kept {
		answer := history.ProviderFailedTurnAnswer
		if isTool {
			answer = history.ToolFailedTurnAnswer
		}
		if aerr := store.Append(history.Entry{Prompt: prompt, Answer: answer, Calls: len(result.Calls), Steps: result.Steps}); aerr == nil {
			stored = true
		}
	}
	// The failure surface is UNCHANGED (frozen phrase + exit 6/7).
	var code int
	if isTool {
		code = emitToolError(env.stderr, inc)
	} else {
		code = emitProviderError(env.stderr, err)
	}
	if stored {
		emitKeptStepsLine(env, len(result.Steps))
	}
	return code
}

// emitKeptStepsLine writes the round-080 informational line (a failed turn with
// completed steps kept them), mirroring the round-079/078 precedent: chrome-styled
// (`[HH:MM:SS] …`) on the diagnostic stream only, control-free, carrying NO
// `tellme: ` class prefix (the round-017 exactly-one-`tellme:`-line contract
// holds) and never routed to turns.log.
func emitKeptStepsLine(env runtimeEnv, steps int) {
	if env.stderr == nil {
		return
	}
	_, _ = fmt.Fprintf(env.stderr, "[%s] kept %d completed tool step(s) in the session history\n",
		env.now().Format("15:04:05"), steps)
}

// effectiveBudget returns the budget the payload status line renders: the
// run-static EffectiveBudget when set (the resolve() path), else MaxHistoryTokens
// (directly-constructed resolutions in unit tests).
func (r resolution) effectiveBudget() int {
	if r.EffectiveBudget > 0 {
		return r.EffectiveBudget
	}
	return r.MaxHistoryTokens
}

// persistInterruptedTurn persists an operator-interrupted turn that completed at
// least one tool step (round 079; ADR 0051) and reports success. The partial turn
// is closed with the synthetic `history.InterruptedTurnAnswer`, so the stored
// entry replays as a valid `… assistant` sequence (both provider families). It
// emits ONE informational line on the diagnostic stream — chrome-styled,
// control-free, carrying no `tellme: ` class prefix (the round-017
// exactly-one-`tellme:`-line contract holds) and never routed to `turns.log`
// (the round-078 retry-line / round-068 unpaired-line precedent). A failed append
// degrades to the existing environment-class phrase via emitHistoryError (never a
// crash, never a half-written file). The synthetic answer is NOT written to
// `stdout` (the interrupted path produces no answer), so `stdout` stays empty.
func persistInterruptedTurn(env runtimeEnv, store history.Store, prompt string, result agentport.Result) int {
	entry := history.Entry{
		Prompt: prompt,
		Answer: history.InterruptedTurnAnswer,
		Calls:  len(result.Calls),
		Steps:  result.Steps,
	}
	if err := store.Append(entry); err != nil {
		return emitHistoryError(env.stderr, err)
	}
	if env.stderr != nil {
		_, _ = fmt.Fprintf(env.stderr, "[%s] interrupted by operator; kept %d completed tool step(s) in the session history\n",
			env.now().Format("15:04:05"), len(result.Steps))
	}
	return Success
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
// It is the PROMPT-SURFACE ack, not turn chrome, so it stays on stderr and is
// NOT recorded in the session turn log (round-053 fold F-53-3(i)).
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

// chromeColour reports whether the round-054 green chrome accents apply (ADR
// 0023): the diagnostic stream (stderr) is a terminal AND rendering is on
// (`-r` off) AND this is a prompt-bearing surface (opts.chrome). It is the SAME
// expression as the spinner gate — colour and the spinner share one gate — kept
// as its own named helper so the two call sites read by intent.
func chromeColour(opts turnOptions, env runtimeEnv) bool {
	return spinnerGate(opts, env.stderrIsTerminal())
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
