package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/gosharplite/tellme/internal/app/deps"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/metrics"
	"github.com/gosharplite/tellme/internal/domain/render"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// callRenderer is the round-034 per-AI-endpoint-call renderer (ADR 0005 D1/D3):
// the sole owner of the turn's status frames and tails. It is bound as the
// `call` half of the CLI's compositeObserver, so the loop drives it through the
// call-begin / call-end hooks.
//
//   - OnCallBegin emits the status FRAME (rule + `╭─⠿ Turn N - <mode>` +
//     pre-flight payload) once per call. N = priorCalls + callIndex + 1 — the
//     number ADVANCES within a prompt (G6); the pre-flight estimate is computed
//     by the CLI from the hook's fused base+turn wire messages (G1/G10).
//   - OnCallEnd emits the call's TAIL (grouped `[Tool Reason]` lines + measured
//     payload + metrics + `Ready`), EXCEPT for the final call whose tail is
//     DEFERRED (stored and emitted after the answer — G5).
//
// Round 051 (R5.5 of #92; ADR 0020): the BYTES come from the injected domain
// render.Lines port (internal/ui owns the formatting); the cost/metrics types
// are domain types (llm.Pricing / metrics.UsageCounts / history.ToolUsageRow), so
// internal/cli names no internal/ui type.
type callRenderer struct {
	env        runtimeEnv
	res        resolution
	reg        domaintools.Registry
	lines      render.Lines
	chrome     bool
	priorCalls int

	// fileLines is the PLAIN renderer for the turn-log file leg (round 054 fold
	// B-54-1): the artifact is written with colour off, never stripped afterwards.
	// turnsLog is the file writer (nil when no turn log is open).
	fileLines render.Lines
	turnsLog  io.Writer

	pricing  llm.Pricing
	session  history.UsageSummary
	turnCost float64

	// renderedToolRound is set once the turn has produced a non-final call — i.e.
	// the model requested tools (and it also covers the bound-reached path). It
	// gates the round-039 post-status blank so a TOOL-LESS turn gains no blank
	// (round-039 review B1: FR-009 / the edge-case list / ADR 0008 D5's closing
	// sentence all say a non-tool turn is unchanged).
	renderedToolRound bool

	// prevEstimate / haveEstimate carry the round-057 payload-increment baseline
	// (ADR 0027): the last ESTIMATED payload emitted this process, so the next
	// estimate can show its growth (`+<delta>`). It is in-memory and
	// session-scoped — a fresh process has no predecessor (the first estimate
	// renders `+0`). No persistence.
	prevEstimate int
	haveEstimate bool

	finalTail func()
}

// newCallRenderer builds the per-call renderer, seeding the session roll-up from
// the persisted summary so the `Ready` session totals accumulate across the
// session (round-018 D7).
func newCallRenderer(env runtimeEnv, res resolution, reg domaintools.Registry, chrome bool, priorCalls int, colour bool, dp deps.Dependencies) *callRenderer {
	r := &callRenderer{
		env:        env,
		res:        res,
		reg:        reg,
		lines:      dp.NewLines(colour),
		fileLines:  dp.NewLines(false),
		turnsLog:   env.turnsLog,
		chrome:     chrome,
		priorCalls: priorCalls,
		pricing:    llm.Pricing{Hit: res.Pricing.HIT, Miss: res.Pricing.MISS, Comp: res.Pricing.COMP},
	}
	if res.Workspace != "" {
		r.session, _ = dp.NewUsageStore(res.Workspace).Totals()
	}
	return r
}

// emit renders ONE chrome line through build for each sink (round 054 fold
// B-54-1): the diagnostic stream gets the render from r.lines (coloured when the
// gate is on), and — when a turn log is open — the FILE leg gets the render from
// r.fileLines (PLAIN, colour off). build receives the renderer for its sink, so
// the file leg is never produced by stripping colour after the fact.
func (r *callRenderer) emit(newline bool, build func(l render.Lines) string) {
	writeChromeLine(r.env.stderr, build(r.lines), newline)
	if r.turnsLog != nil {
		writeChromeLine(r.turnsLog, build(r.fileLines), newline)
	}
}

// writeChromeLine writes one chrome line, with or without a trailing newline.
func writeChromeLine(w io.Writer, s string, newline bool) {
	if newline {
		_, _ = fmt.Fprintln(w, s)
		return
	}
	_, _ = fmt.Fprint(w, s)
}

// OnCallBegin emits the frame (rule + header + gap, chrome only) and the
// CLI-computed per-call pre-flight estimate.
func (r *callRenderer) OnCallBegin(callIndex int, messages []llm.Message) {
	turn := r.priorCalls + callIndex + 1
	if r.chrome {
		r.emit(false, func(l render.Lines) string { return l.TurnOpening(turn, r.res.Mode) })
	}
	estimate := llm.EstimatePayload(r.res.Person, agentport.ToolDefs(r.reg), messages)
	// Round 057 (ADR 0027): the estimated line shows the increment over the
	// previous estimate emitted this process (0 when there is none). The tracker
	// is in-memory and session-scoped; no persistence.
	delta := 0
	if r.haveEstimate {
		delta = estimate - r.prevEstimate
	}
	r.prevEstimate = estimate
	r.haveEstimate = true
	r.emit(true, func(l render.Lines) string {
		return l.PayloadEstimate(r.env.now(), estimate, delta, r.res.Mode, r.res.Provider.Model)
	})
	if r.chrome {
		r.emit(false, func(l render.Lines) string { return l.TurnGap() })
	}
}

// OnCallEnd emits the call's tail, deferring the FINAL call's tail past the
// answer (G5).
func (r *callRenderer) OnCallEnd(callIndex int, usage llm.Usage, roundReasons []string, final bool) {
	emit := func() {
		// Round 039: the trailing grouped `[Tool Reason]` block is preceded by
		// exactly ONE blank line and has NO blank between its lines.
		//
		// Round 046 (R4 of #92, ADR 0015): the blank-reason predicate is
		// single-owned upstream — the loop filters the round's reasons through
		// ui.ToolLineRenderer.ReasonLine, so roundReasons never carries a reason
		// that renders no line.
		if len(roundReasons) > 0 {
			r.emit(true, func(render.Lines) string { return "" })
			for _, reason := range roundReasons {
				r.emit(true, func(l render.Lines) string { return l.ToolReason(r.env.now(), reason) })
			}
		}
		if !usage.Reported {
			return
		}
		if r.renderedToolRound {
			r.emit(true, func(render.Lines) string { return "" })
		}
		r.emit(true, func(l render.Lines) string {
			return l.PayloadStatus(r.env.now(), usage.PromptTokens, r.res.effectiveBudget(), r.res.Mode, r.res.Provider.Model, false)
		})
		r.emitMetrics(usage)
	}
	if final {
		r.finalTail = emit
		return
	}
	r.renderedToolRound = true
	emit()
}

// EmitFinalTail emits the deferred final call's tail (after the answer).
func (r *callRenderer) EmitFinalTail() {
	if r.finalTail != nil {
		r.finalTail()
		r.finalTail = nil
	}
}

// emitMetrics renders the call's metrics line and the `╰─⠿ Ready` session summary,
// folding the call into the display-only session roll-up.
func (r *callRenderer) emitMetrics(usage llm.Usage) {
	now := r.env.now()
	rec, cost := usageRecordOf(r.pricing, r.res.Selected, r.res.Provider.Model, now.Format(time.RFC3339), usage)
	r.turnCost += cost
	r.session.Add(rec)
	counts := metrics.UsageCounts{
		Miss:       rec.PromptTokens - rec.CachedTokens,
		Hit:        rec.CachedTokens,
		Completion: rec.ResponseTokens,
		Thinking:   rec.ThinkingTokens,
	}
	r.emit(true, func(l render.Lines) string { return l.Metrics(now, r.res.Selected, counts) })
	r.emit(true, func(l render.Lines) string {
		return l.Ready(cost, r.turnCost, r.session.Cost, r.session.Miss, r.session.Hit, r.session.Out, llm.HitRate(r.session.Hit, r.session.Miss))
	})
}

// usageRecordOf builds one call's persisted usage record and its cost from the
// SINGLE-SOURCED formula (round 034 review REFACTOR-1): the miss is
// `prompt − cached`, the cost is derived from the config `MODELS` pricing, and
// the total is `prompt + response + thinking`.
func usageRecordOf(pricing llm.Pricing, selected, model, ts string, c llm.Usage) (history.UsageRecord, float64) {
	miss := c.PromptTokens - c.CachedTokens
	cost := llm.ComputeCost(pricing, miss, c.CachedTokens, c.CompletionTokens, c.ThinkingTokens)
	return history.UsageRecord{
		Timestamp:      ts,
		Provider:       selected,
		Model:          model,
		CachedTokens:   c.CachedTokens,
		PromptTokens:   c.PromptTokens,
		ResponseTokens: c.CompletionTokens,
		TotalTokens:    c.PromptTokens + c.CompletionTokens + c.ThinkingTokens,
		ThinkingTokens: c.ThinkingTokens,
		Cost:           cost,
	}, cost
}

// persistTurnUsage writes the turn's usage ONCE (ADR 0005 D4/FR-010b): the
// Reported subset of result.Calls in a single AppendBatch — never per call — and
// ONLY when the FINAL call reports usage (the round-018 gate).
func persistTurnUsage(env runtimeEnv, res resolution, result agentport.Result, dp deps.Dependencies) {
	if res.Workspace == "" || !result.Usage.Reported {
		return
	}
	pricing := llm.Pricing{Hit: res.Pricing.HIT, Miss: res.Pricing.MISS, Comp: res.Pricing.COMP}
	ts := env.now().Format(time.RFC3339)
	records := make([]history.UsageRecord, 0, len(result.Calls))
	for _, c := range result.Calls {
		if !c.Reported {
			continue
		}
		rec, _ := usageRecordOf(pricing, res.Selected, res.Provider.Model, ts, c)
		records = append(records, rec)
	}
	_ = dp.NewUsageStore(res.Workspace).AppendBatch(records)
}
