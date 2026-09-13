// Package agent holds the tool-loop orchestrator: the bounded think→act→observe
// cycle that lets one prompt run call the declared tools and iterate to a final
// answer (round-008 research Decision 3). Keeping it here — not in the CLI
// presentation layer — is the RF-1 god-object guard.
package agent

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
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

// AgentLoop drives the bounded think→act→observe cycle for one prompt run.
type AgentLoop struct {
	Gateway     llm.Gateway
	Registry    tools.Registry
	MaxLoops    int
	ToolTimeout time.Duration
	Stderr      io.Writer
}

// Run performs one prompt run. It sends the conversation (the replayed prior
// turns followed by the current prompt), executes any tools the model requests,
// feeds their results back, and repeats until the model returns a final answer
// or MaxLoops tool rounds have been made. It returns the final answer text and
// the ordered tool steps performed (for persistence). An incomplete run returns
// *ErrIncomplete; a provider/transport failure is returned unwrapped (so the
// caller maps it to the provider class phrase + code 6).
func (a *AgentLoop) Run(ctx context.Context, prompt string, prior []history.Entry) (string, []history.Step, error) {
	maxLoops := a.MaxLoops
	if maxLoops <= 0 {
		maxLoops = 1
	}
	toolTimeout := a.ToolTimeout
	if toolTimeout <= 0 {
		toolTimeout = DefaultToolTimeout
	}

	messages := buildMessages(prior)
	var steps []history.Step

	for i := 0; ; i++ {
		resp, err := a.Gateway.Complete(ctx, llm.Request{
			Prompt:   prompt,
			Messages: messages,
			Tools:    a.toolDefs(),
		})
		if err != nil {
			return "", steps, err
		}
		if len(resp.ToolCalls) == 0 {
			return resp.Text, steps, nil
		}
		if i >= maxLoops {
			return "", steps, &ErrIncomplete{Reason: "the tool-loop bound was reached"}
		}

		// The model requested tools: echo the assistant tool-call message, run
		// each tool, and feed the results back.
		messages = append(messages, llm.Message{Role: "assistant", ToolCalls: resp.ToolCalls})
		for _, tc := range resp.ToolCalls {
			if a.Registry == nil {
				return "", steps, &ErrIncomplete{Reason: "no tools are registered"}
			}
			tool, ok := a.Registry.Lookup(tc.Name)
			if !ok {
				return "", steps, &ErrIncomplete{Reason: fmt.Sprintf("tool %q is not available", tc.Name)}
			}
			tctx, cancel := context.WithTimeout(ctx, toolTimeout)
			result, terr := tool.Execute(tctx, tc.Arguments)
			cancel()
			if terr != nil {
				// A recoverable tool error is fed back as the tool's result (non-terminal).
				result = "error: " + terr.Error()
			}
			a.logStep(tc, result)
			messages = append(messages, llm.Message{Role: "tool", Content: result, ToolCallID: tc.ID})
			steps = append(steps, history.Step{Tool: tc.Name, Arguments: tc.Arguments, Result: result})
		}
	}
}

// toolDefs projects the registry's tools into the wire definitions offered to
// the model.
func (a *AgentLoop) toolDefs() []llm.ToolDef {
	if a.Registry == nil {
		return nil
	}
	ts := a.Registry.Tools()
	defs := make([]llm.ToolDef, 0, len(ts))
	for _, t := range ts {
		defs = append(defs, llm.ToolDef{Name: t.Name(), Description: t.Description(), Parameters: t.Parameters()})
	}
	return defs
}

// logStep emits one discrete tool-loop log line to the diagnostic stream
// (round-008 Decision 7): the tool name, its arguments, and its result. This is
// NOT token streaming.
func (a *AgentLoop) logStep(tc llm.ToolCall, result string) {
	if a.Stderr == nil {
		return
	}
	_, _ = fmt.Fprintf(a.Stderr, "[tool] %s arguments=%s result=%s\n",
		tc.Name, oneLine(tc.Arguments), oneLine(truncate(result, 200)))
}

// buildMessages replays the persisted prior turns into the conversation sent to
// the provider: each turn is the user prompt, then the turn's tool steps (an
// assistant tool-call + a tool result per step), then the assistant answer. Tool
// steps carry no stored id, so a deterministic `call_step_<n>` id is synthesised
// on replay (round-008 Decision 5 / TD-2).
func buildMessages(prior []history.Entry) []llm.Message {
	var msgs []llm.Message
	for _, e := range prior {
		msgs = append(msgs, llm.Message{Role: "user", Content: e.Prompt})
		for i, s := range e.Steps {
			id := fmt.Sprintf("call_step_%d", i+1)
			msgs = append(msgs, llm.Message{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: id, Name: s.Tool, Arguments: s.Arguments}}})
			msgs = append(msgs, llm.Message{Role: "tool", Content: s.Result, ToolCallID: id})
		}
		msgs = append(msgs, llm.Message{Role: "assistant", Content: e.Answer})
	}
	return msgs
}

// oneLine folds newlines so a log line stays single-line.
func oneLine(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "\r", " ")
}

// truncate bounds a log field to n characters.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
