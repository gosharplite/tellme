// Package agent holds the tool-loop orchestrator: the bounded think→act→observe
// cycle that lets one prompt run call the declared tools and iterate to a final
// answer (round-008 research Decision 3). Keeping it here — not in the CLI
// presentation layer — is the RF-1 god-object guard.
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
	"github.com/gosharplite/tellme/internal/ui"
)

// DefaultToolTimeout bounds each individual tool execution when the loop is not
// given an explicit timeout (round-008 FR-009). It aliases the shared domain
// constant (round-032 implementation-review F4).
const DefaultToolTimeout = tools.DefaultToolTimeout

// ErrIncomplete reports that a tool-using run could not reach a final answer —
// the iteration bound was reached, or the model requested a tool that tellme
// does not provide. The CLI maps it to the frozen class phrase
// `the tool request failed` and exit code 7 (round-008 FR-010).
type ErrIncomplete struct {
	Reason string
	Err    error
}

// Error renders the reason plus the underlying cause, if any.
func (e *ErrIncomplete) Error() string {
	if e.Err != nil {
		return e.Reason + ": " + e.Err.Error()
	}
	return e.Reason
}

// Unwrap exposes the underlying cause for errors.Is / errors.As.
func (e *ErrIncomplete) Unwrap() error { return e.Err }

// AgentResult is the outcome of one prompt run: the final answer text, the
// ordered tool steps performed (for persistence), and the provider's reported
// usage of the final completion. Surfacing Usage here — rather than discarding
// it inside the loop — is what lets the CLI render the post-turn payload status
// line (round-009 BLOCKER-2). On a multi-step tool run it is the usage of the
// final completion (the response that produced the answer).
type AgentResult struct {
	Answer string
	Steps  []history.Step
	Usage  llm.Usage
	// Calls holds EVERY provider call's usage for the turn, in call order
	// (round 018), so the CLI can compute the turn cost (`$#2`) and persist each
	// call to the usage log. `Usage` remains the just-returned (final) call.
	Calls []llm.Usage
}

// AgentLoop drives the bounded think→act→observe cycle for one prompt run.
type AgentLoop struct {
	Gateway  llm.Gateway
	Registry tools.Registry
	MaxLoops int
	Stderr   io.Writer
	// Observer, when set, is notified of each waiting phase (round 019) so a
	// presenter (the CLI-injected spinner) can label / clear / restore the
	// indicator per phase (round-019 research Decision 7).
	Observer agentport.LoopObserver
	// Now, when set, supplies the clock reading for a tool-loop log line (round
	// 022). Nil falls back to time.Now, so construction and unit tests stay
	// simple; the injected seam (the CLI's env.now) keeps the line deterministic
	// and shares one clock with the chrome / payload lines (round-009/017
	// precedent).
	Now func() time.Time
	// EffectiveBudget is the resolved run-static token budget (round-024
	// FR-015): min(MAX_HISTORY_TOKENS, the active model's configured context
	// window). The loop resolves each call's bound from it (default = /4, ceiling
	// = /2); <= 0 falls back to defaultEffectiveBudget.
	EffectiveBudget int
	// ToolUsage, when set, records each EXECUTED tool invocation's outcome
	// (round 026) through the injected sink. Nil = a no-op; a sink error is
	// swallowed (best-effort), so accounting never breaks a turn.
	ToolUsage history.ToolUsageSink
}

// Run performs one prompt run. It sends the conversation (the replayed prior
// turns followed by the current prompt), executes any tools the model requests,
// feeds their results back, and repeats until the model returns a final answer
// or MaxLoops tool rounds have been made. It returns the final answer text, the
// ordered tool steps performed, and the final completion's reported usage. An
// incomplete run returns *ErrIncomplete; a provider/transport failure is
// returned unwrapped (so the caller maps it to the provider class phrase + code
// 6).
//
// Wire chronology (review PR #25 BLOCKER-1): the active turn's user prompt is
// ALWAYS the first message of the active turn, and each tool exchange is
// appended AFTER it — so every request keeps the OpenAI-mandated order
// `[prior…, user(prompt), assistant(tool_calls), tool(result), …]`. The first
// request carries the prompt via Request.Prompt (the round-004/007 shape); later
// tool rounds fold the whole active turn into Request.Messages with an empty
// Prompt, so the adapter never moves the prompt behind the tool activity.
func (a *AgentLoop) Run(ctx context.Context, prompt string, prior []history.Entry) (AgentResult, error) {
	maxLoops := a.MaxLoops
	if maxLoops <= 0 {
		maxLoops = 1
	}

	base := BuildMessages(prior)
	// turn is the active turn's conversation: it starts with the user prompt and
	// appends each tool exchange after it (chronological order).
	turn := []llm.Message{{Role: "user", Content: prompt}}
	var steps []history.Step
	var calls []llm.Usage

	for i := 0; ; i++ {
		req := llm.Request{Tools: a.toolDefs()}
		// The fused base+turn wire slice the call-begin hook carries, so the CLI
		// computes the per-call estimate WITHOUT the loop owning a persona or an
		// estimator field (ADR 0005 D2). Built ONCE per call (round 034 review
		// REFACTOR-2): for i == 0 the request carries the prompt via Request.Prompt
		// (its Messages is just `base`), so the two legitimately differ; for i >= 1
		// the request messages ARE the fused slice, so share it.
		wire := append(append(make([]llm.Message, 0, len(base)+len(turn)), base...), turn...)
		if i == 0 {
			req.Prompt = prompt
			req.Messages = base
		} else {
			req.Messages = wire
		}
		a.notifyCallBegin(i, wire)
		a.notifyInferenceStart()
		resp, err := a.Gateway.Complete(ctx, req)
		a.notifyInferenceEnd()
		if err != nil {
			return AgentResult{Steps: steps, Calls: calls}, err
		}
		calls = append(calls, resp.Usage)
		final := len(resp.ToolCalls) == 0
		if final {
			// The final call produced the answer: fire the call-end hook with no
			// round reasons (ADR 0005 D1). The CLI defers its tail past the answer.
			a.notifyCallEnd(i, resp.Usage, nil, true)
			return AgentResult{Answer: resp.Text, Steps: steps, Usage: resp.Usage, Calls: calls}, nil
		}
		if i >= maxLoops {
			a.notifyCallEnd(i, resp.Usage, reasonsOf(resp.ToolCalls), false)
			return AgentResult{Steps: steps, Calls: calls}, &ErrIncomplete{Reason: "the tool-loop bound was reached"}
		}

		// The model requested tools: echo the assistant tool-call message, run
		// each tool, and feed the results back — appended after the user prompt.
		a.notifyToolsStart(toolNames(resp.ToolCalls))
		a.logEngine(i+1, maxLoops)
		turn = append(turn, llm.Message{Role: "assistant", ToolCalls: resp.ToolCalls})
		for _, tc := range resp.ToolCalls {
			if a.Registry == nil {
				return AgentResult{Steps: steps, Calls: calls}, &ErrIncomplete{Reason: "no tools are registered"}
			}
			tool, ok := a.Registry.Lookup(tc.Name)
			if !ok {
				return AgentResult{Steps: steps, Calls: calls}, &ErrIncomplete{Reason: fmt.Sprintf("tool %q is not available", tc.Name)}
			}
			a.logAction(tc)
			tctx, cancel := context.WithTimeout(ctx, a.callTimeout(tool, tc.Arguments))
			byteBudget := a.callByteBudget(tc.Arguments)
			result, terr := tool.Execute(tctx, tc.Arguments, tools.ByteBudget(byteBudget))
			// Read the per-call deadline signal BEFORE cancel() (round 026): it is
			// the structural `timeout` signal, independent of the tool result text.
			toolTimedOut := errors.Is(tctx.Err(), context.DeadlineExceeded)
			cancel()
			if terr != nil {
				// A recoverable tool error is fed back as the tool's result (non-terminal).
				result = "error: " + terr.Error()
			} else {
				// The loop's raw-byte-length backstop (round-024 Q1/D4): inert for a
				// compliant tool that bounded at the source to the same byte budget.
				result = clampBytes(result, byteBudget)
			}
			a.recordToolUsage(tc.Name, terr, toolTimedOut)
			a.logResult(tc, result)
			turn = append(turn, llm.Message{Role: "tool", Content: result, ToolCallID: tc.ID})
			steps = append(steps, history.Step{Tool: tc.Name, Arguments: tc.Arguments, Result: result, Signature: tc.Signature})
		}
		a.notifyToolsEnd()
		// Round 034 (ADR 0005 D1): the call-end hook fires at the END of the
		// call's phase (inference + its tool round), carrying the round's reasons
		// so the CLI emits the grouped post-call tail after the results.
		a.notifyCallEnd(i, resp.Usage, reasonsOf(resp.ToolCalls), false)
	}
}

// toolDefs projects the registry's tools into the wire definitions offered to
// the model.
func (a *AgentLoop) toolDefs() []llm.ToolDef { return ToolDefs(a.Registry) }

// ToolDefs projects a tool registry into the wire definitions offered to the
// model. Exported so the CLI's pre-flight estimate counts exactly what the loop
// sends (round-011 RF-1), mirroring the BuildMessages reuse (round 009).
func ToolDefs(reg tools.Registry) []llm.ToolDef {
	if reg == nil {
		return nil
	}
	ts := reg.Tools()
	defs := make([]llm.ToolDef, 0, len(ts))
	for _, t := range ts {
		defs = append(defs, llm.ToolDef{Name: t.Name(), Description: t.Description(), Parameters: t.Parameters()})
	}
	return defs
}

// notifyInferenceStart / notifyInferenceEnd / notifyToolsStart / notifyToolsEnd
// forward the loop's waiting-phase transitions to the observer when one is set
// (round 019). Keeping the nil guard here holds Run's cyclomatic complexity below
// the cyclop gate (max 15).
func (a *AgentLoop) notifyInferenceStart() {
	if a.Observer != nil {
		a.Observer.OnInferenceStart()
	}
}

func (a *AgentLoop) notifyInferenceEnd() {
	if a.Observer != nil {
		a.Observer.OnInferenceEnd()
	}
}

func (a *AgentLoop) notifyToolsStart(names []string) {
	if a.Observer != nil {
		a.Observer.OnToolsStart(names)
	}
}

func (a *AgentLoop) notifyToolsEnd() {
	if a.Observer != nil {
		a.Observer.OnToolsEnd()
	}
}

// notifyCallBegin fires the round-034 call-begin hook (ADR 0005 D1/D2): the
// CLI observer computes the per-call pre-flight estimate from the fused base+turn
// wire messages. A nil observer is a no-op (the seam's default).
func (a *AgentLoop) notifyCallBegin(callIndex int, messages []llm.Message) {
	if a.Observer != nil {
		a.Observer.OnCallBegin(callIndex, messages)
	}
}

// notifyCallEnd fires the round-034 call-end hook (ADR 0005 D1): the CLI
// observer renders the call's tail from its usage, the round's reasons, and
// whether the call produced the final answer.
func (a *AgentLoop) notifyCallEnd(callIndex int, usage llm.Usage, roundReasons []string, final bool) {
	if a.Observer != nil {
		a.Observer.OnCallEnd(callIndex, usage, roundReasons, final)
	}
}

// reasonsOf returns the non-empty top-level `reason` of each requested call, in
// call order (the grouped post-call tail; round 034 FR-005).
func reasonsOf(calls []llm.ToolCall) []string {
	var out []string
	for _, tc := range calls {
		// Round 036 (issue #74): a blank reason contributes no grouped tail line
		// either. The guard tests the SANITIZED value but APPENDS THE RAW value —
		// the pure formatter trims later, so the tail always shows the trimmed
		// reason (round-036 review N-4). This is the production filter the tail
		// relies on; the tail's own guard is defensive only (see call_renderer.go).
		if r := toolReason(tc.Arguments); strings.TrimSpace(r) != "" {
			out = append(out, r)
		}
	}
	return out
}

// Round-034 decomposed tool-call rendering (ADR 0005). Each piece is rendered by
// a pure `internal/ui` formatter and stamped from the injected clock seam; the
// Observer hooks still wrap every write (round-019). logEngine fires once per
// EXECUTED round (FR-001); logAction fires per call at its begin — the reason,
// then the action (FR-002/FR-003); logResult fires per call when it completes
// (FR-004).
func (a *AgentLoop) logEngine(step, total int) {
	a.withToolLog(func() {
		_, _ = fmt.Fprintln(a.Stderr, ui.FormatToolEngine(a.now(), step, total))
	})
}

// logAction emits the call's `[Tool Reason]` (when present) then its
// `[Tool Action]` line at call begin.
func (a *AgentLoop) logAction(tc llm.ToolCall) {
	a.withToolLog(func() {
		// Round 036 (issue #74): a blank reason (empty OR whitespace-only after
		// folding+trimming) emits NO reason line — the round-022 B1 intent. The
		// guard checks the SANITIZED value so a `"   "` / `"\n"` reason cannot
		// render a dangling prefix row.
		if reason := toolReason(tc.Arguments); strings.TrimSpace(reason) != "" {
			_, _ = fmt.Fprintln(a.Stderr, ui.FormatToolReason(a.now(), reason))
		}
		_, _ = fmt.Fprintln(a.Stderr, ui.FormatToolAction(a.now(), tc.Name, tc.Arguments))
	})
}

// logResult emits the call's `[Tool Result]` line once the tool has run.
func (a *AgentLoop) logResult(tc llm.ToolCall, result string) {
	a.withToolLog(func() {
		_, _ = fmt.Fprintln(a.Stderr, ui.FormatToolResult(a.now(), tc.Name, result))
	})
}

// withToolLog wraps one diagnostic write with the nil-Stderr guard and the
// observer's clear/restore hooks (round-019), so the spinner yields the line.
func (a *AgentLoop) withToolLog(fn func()) {
	if a.Stderr == nil {
		return
	}
	if a.Observer != nil {
		a.Observer.BeforeToolLog()
	}
	fn()
	if a.Observer != nil {
		a.Observer.AfterToolLog()
	}
}

// now returns the clock reading for a tool-loop log line: the injected Now seam
// when set, else time.Now (the nil fallback keeps construction and the unit tests
// simple).
func (a *AgentLoop) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

// classifyToolOutcome derives the recorded outcome from the loop's structural
// signals ONLY — the tool's returned error and the loop-owned per-call deadline
// (round 026 D1). It never sniffs the tool result text, so the generic loop stays
// decoupled from tool-specific markers. The round-024 FR-018 timeout is a
// nil-error result, so only the deadline distinguishes it from a success.
func classifyToolOutcome(terr error, timedOut bool) history.ToolOutcome {
	switch {
	case terr != nil:
		return history.ToolOutcomeError
	case timedOut:
		return history.ToolOutcomeTimeout
	default:
		return history.ToolOutcomeOK
	}
}

// recordToolUsage records one executed invocation through the injected sink
// (round 026). Best-effort: a nil sink or a sink error is ignored so accounting
// never breaks a turn. The write happens before the tool result is folded back
// into the conversation, so a record exists for every executed call.
func (a *AgentLoop) recordToolUsage(tool string, terr error, timedOut bool) {
	if a.ToolUsage == nil {
		return
	}
	_ = a.ToolUsage.Record(tool, classifyToolOutcome(terr, timedOut))
}

// toolReason extracts the top-level `reason` string from a tool call's raw
// arguments JSON, or "" when it is absent or the arguments are unparseable
// (round 021 Decision 4). Extracting it in the loop — rather than in each tool —
// keeps the tools free of presentation concerns and works uniformly.
func toolReason(arguments string) string {
	var probe struct {
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(arguments), &probe); err != nil {
		return ""
	}
	return probe.Reason
}

// toolNames extracts the requested tool names in call order, for the observer's
// tool-phase label (round 019).
func toolNames(calls []llm.ToolCall) []string {
	names := make([]string, 0, len(calls))
	for _, tc := range calls {
		names = append(names, tc.Name)
	}
	return names
}

// BuildMessages replays the persisted prior turns into the conversation sent to
// the provider: each turn is the user prompt, then the turn's tool steps (an
// assistant tool-call + a tool result per step), then the assistant answer. Tool
// steps carry no stored id, so a deterministic `call_step_<n>` id is synthesised
// on replay (round-008 Decision 5 / TD-2).
//
// It is exported so the CLI's pre-flight token estimate reuses the exact same
// projection — including tool steps — instead of the legacy prompt/answer-only
// projection, which would undercount a tool-using conversation (round-009 TD-1).
func BuildMessages(prior []history.Entry) []llm.Message {
	var msgs []llm.Message
	for _, e := range prior {
		msgs = append(msgs, llm.Message{Role: "user", Content: e.Prompt})
		for i, s := range e.Steps {
			id := fmt.Sprintf("call_step_%d", i+1)
			msgs = append(msgs, llm.Message{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: id, Name: s.Tool, Arguments: s.Arguments, Signature: s.Signature}}})
			msgs = append(msgs, llm.Message{Role: "tool", Content: s.Result, ToolCallID: id})
		}
		msgs = append(msgs, llm.Message{Role: "assistant", Content: e.Answer})
	}
	return msgs
}
