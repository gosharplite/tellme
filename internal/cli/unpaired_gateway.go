package cli

import (
	"context"
	"fmt"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/render"
)

// unpairedGateway is the round-068 (ADR 0038) diagnostic decorator: it wraps the
// resolved provider gateway and, before each Complete, surfaces any tool calls a
// Gemini/Vertex round left UNANSWERED (M < N) as a `[Tool …]` line on `stderr` —
// resolving ADR 0037 §Forward RF-067-1 (the unpaired-call account had no live
// consumer). Detection is the family-neutral single owner, llm.UnpairedToolCalls;
// the shipped `M == N` path yields an empty account, so nothing new is printed
// (I-7). The diagnostic is informational (Q2 → A): the request proceeds unchanged
// (the emitted body carries the M parts produced), and the line is never written
// to turns.log (Q1 → A). The OpenAI-compatible family carries no round-boundary
// drop, so for it the account is always empty — the decorator is inert there.
type unpairedGateway struct {
	inner llm.Gateway
	emit  func(ids []string) // nil disables the diagnostic (a no-op decorator)
}

func (g unpairedGateway) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	if g.emit != nil {
		if ids := llm.UnpairedToolCalls(req.Messages); len(ids) > 0 {
			g.emit(ids)
		}
	}
	return g.inner.Complete(ctx, req)
}

// withUnpairedDiagnostic wraps a gateway so a short round is reported on stderr
// before the request is sent. A nil emit (no stderr seam) returns the gateway
// unwrapped.
func withUnpairedDiagnostic(gw llm.Gateway, emit func(ids []string)) llm.Gateway {
	if emit == nil {
		return gw
	}
	return unpairedGateway{inner: gw, emit: emit}
}

// unpairedEmitter builds the diagnostic emit callback for a turn: it writes the
// `[Tool …]` line to the diagnostic stream (stderr) through the render.Lines port
// (so internal/cli names no internal/ui type — ADR 0020). A nil lines or stderr
// yields nil (the decorator is then a no-op passthrough).
func unpairedEmitter(env runtimeEnv, lines render.Lines) func(ids []string) {
	if lines == nil || env.stderr == nil {
		return nil
	}
	return func(ids []string) {
		if len(ids) == 0 {
			return
		}
		_, _ = fmt.Fprintln(env.stderr, lines.UnpairedCalls(env.now(), ids))
	}
}
