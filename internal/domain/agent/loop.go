package agent

import (
	"context"
	"io"
	"time"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// This file declares the loop's CONSTRUCTION/EXECUTION contract — the injected
// domain port through which a caller obtains and runs the bounded
// think→act→observe cycle, without naming the concrete loop implementation.
//
// It was introduced in round 050 (R5.4 of #92; ADR 0019) — the sub-slice 2 of
// the `internal/cli -> internal/agent` de-coupling. Round 049 (ADR 0018) had
// already relocated the loop's crossing VALUES (Result/ErrIncomplete/ToolDefs)
// here, leaving exactly one `internal/agent` reference in production code: the
// FORMER `&agent.AgentLoop{…}` construction in `internal/cli`. This port removes
// that last reference (historical), so `internal/cli` depends only on the domain
// (plus stdlib) for the loop; the construction now lives behind `agent.NewLoop`.
//
// RULE-C purity: the contract references only stdlib (context/io/time) and
// domain (history/llm/tools) types — no `internal/agent` type crosses, so the
// concrete loop struct stays private to its package and the inversion is a
// domain-only seam. LoopSpec's fields are exactly the loop's construction
// inputs, every one already domain/stdlib-typed.

// Loop is the domain port for one bounded prompt run. It is the construction/
// execution seam a caller (internal/cli) consumes instead of building the
// concrete loop directly. The concrete adapter is `agent.NewLoop` (round 050;
// ADR 0019).
type Loop interface {
	// Run performs one prompt run over the replayed prior turns plus the current
	// prompt, executing any tools the model requests until a final answer or the
	// iteration bound. It returns the run outcome (answer/steps/usage); an
	// incomplete run returns *ErrIncomplete; a provider/transport failure is
	// returned unwrapped.
	Run(ctx context.Context, prompt string, prior []history.Entry) (Result, error)
}

// LoopSpec carries the loop's construction inputs. Every field is domain- or
// stdlib-typed, so a LoopFactory implementation (the adapter) can build the
// concrete loop without the composition site naming any `internal/agent` or
// `internal/ui` type.
type LoopSpec struct {
	// Gateway is the provider gateway the loop calls for each inference round.
	Gateway llm.Gateway
	// Registry is the tool registry whose definitions are offered to the model.
	Registry tools.Registry
	// MaxLoops bounds the tool-execution iterations within one run.
	MaxLoops int
	// Stderr is the loop's diagnostic stream (its tool-line/log sink).
	Stderr io.Writer
	// Now, when set, supplies the loop's clock readings; nil = time.Now.
	Now func() time.Time
	// EffectiveBudget is the resolved run-static token budget (<= 0 = default).
	EffectiveBudget int
	// ToolUsage, when set, records each executed tool invocation (nil = no-op).
	ToolUsage history.ToolUsageSink
	// Lines, when set, renders the loop's four diagnostic tool lines (nil = none).
	Lines ToolLineRenderer
	// Observer, when set, is notified of each waiting phase (nil = none).
	Observer LoopObserver
}

// LoopFactory builds a Loop from a LoopSpec. It is the injected seam (a
// func-typed field on the composition-root dependencies value): the adapter
// (`agent.NewLoop`) is constructed once at the composition root and the
// application layer obtains the loop through it, naming no `internal/agent`
// type.
type LoopFactory func(LoopSpec) Loop
