// Package agent holds the tool-loop orchestrator: the bounded think→act→observe
// cycle that lets one prompt run call the declared tools and iterate to a final
// answer (round-008 research Decision 3). Keeping it here — not in the CLI
// presentation layer — is the RF-1 god-object guard.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
	"github.com/gosharplite/tellme/internal/ui"
)

// DefaultToolTimeout bounds each individual tool execution when the loop is not
// given an explicit timeout (round-008 FR-009).
const DefaultToolTimeout = 300 * time.Second

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
	Gateway     llm.Gateway
	Registry    tools.Registry
	MaxLoops    int
	ToolTimeout time.Duration
	Stderr      io.Writer
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
		if i == 0 {
			req.Prompt = prompt
			req.Messages = base
		} else {
			req.Messages = append(append(make([]llm.Message, 0, len(base)+len(turn)), base...), turn...)
		}
		a.notifyInferenceStart()
		resp, err := a.Gateway.Complete(ctx, req)
		a.notifyInferenceEnd()
		if err != nil {
			return AgentResult{Steps: steps, Calls: calls}, err
		}
		calls = append(calls, resp.Usage)
		if len(resp.ToolCalls) == 0 {
			return AgentResult{Answer: resp.Text, Steps: steps, Usage: resp.Usage, Calls: calls}, nil
		}
		if i >= maxLoops {
			return AgentResult{Steps: steps, Calls: calls}, &ErrIncomplete{Reason: "the tool-loop bound was reached"}
		}

		// The model requested tools: echo the assistant tool-call message, run
		// each tool, and feed the results back — appended after the user prompt.
		a.notifyToolsStart(toolNames(resp.ToolCalls))
		turn = append(turn, llm.Message{Role: "assistant", ToolCalls: resp.ToolCalls})
		for _, tc := range resp.ToolCalls {
			if a.Registry == nil {
				return AgentResult{Steps: steps, Calls: calls}, &ErrIncomplete{Reason: "no tools are registered"}
			}
			tool, ok := a.Registry.Lookup(tc.Name)
			if !ok {
				return AgentResult{Steps: steps, Calls: calls}, &ErrIncomplete{Reason: fmt.Sprintf("tool %q is not available", tc.Name)}
			}
			tctx, cancel := context.WithTimeout(ctx, a.callTimeout(tool, tc.Arguments))
			byteBudget := a.callByteBudget(tc.Arguments)
			result, terr := tool.Execute(tctx, tc.Arguments, tools.ByteBudget(byteBudget))
			cancel()
			if terr != nil {
				// A recoverable tool error is fed back as the tool's result (non-terminal).
				result = "error: " + terr.Error()
			} else {
				// The loop's raw-byte-length backstop (round-024 Q1/D4): inert for a
				// compliant tool that bounded at the source to the same byte budget.
				result = clampBytes(result, byteBudget)
			}
			a.logStep(tc)
			turn = append(turn, llm.Message{Role: "tool", Content: result, ToolCallID: tc.ID})
			steps = append(steps, history.Step{Tool: tc.Name, Arguments: tc.Arguments, Result: result, Signature: tc.Signature})
		}
		a.notifyToolsEnd()
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

// logStep emits one discrete tool-loop log line to the diagnostic stream
// (round-008 Decision 7; reshaped round 022): a single timestamped line naming
// the tool and, when the call states one, its `reason` —
// `[HH:MM:SS] [Tool] <name> - <reason>`. The raw call arguments and result are
// deliberately NOT echoed (the operator-chosen shape; round-022 research
// Decision 1/3). This is NOT token streaming. The line is rendered by the pure
// `internal/ui` formatter and stamped from the injected clock seam; the Observer
// hooks still wrap the write so the round-019 spinner can clear/restore around it.
func (a *AgentLoop) logStep(tc llm.ToolCall) {
	if a.Stderr == nil {
		return
	}
	if a.Observer != nil {
		a.Observer.BeforeToolLog()
	}
	_, _ = fmt.Fprintln(a.Stderr, ui.FormatToolLog(a.now(), tc.Name, toolReason(tc.Arguments)))
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
