package cli

import (
	"fmt"
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

	pricing  llm.Pricing
	session  history.UsageSummary
	turnCost float64

	// renderedToolRound is set once the turn has produced a non-final call — i.e.
	// the model requested tools (and it also covers the bound-reached path). It
	// gates the round-039 post-status blank so a TOOL-LESS turn gains no blank
	// (round-039 review B1: FR-009 / the edge-case list / ADR 0008 D5's closing
	// sentence all say a non-tool turn is unchanged).
	renderedToolRound bool

	finalTail func()
}

// newCallRenderer builds the per-call renderer, seeding the session roll-up from
// the persisted summary so the `Ready` session totals accumulate across the
// session (round-018 D7).
func newCallRenderer(env runtimeEnv, res resolution, reg domaintools.Registry, chrome bool, priorCalls int, dp deps.Dependencies) *callRenderer {
	r := &callRenderer{
		env:        env,
		res:        res,
		reg:        reg,
		lines:      dp.NewLines(),
		chrome:     chrome,
		priorCalls: priorCalls,
		pricing:    llm.Pricing{Hit: res.Pricing.HIT, Miss: res.Pricing.MISS, Comp: res.Pricing.COMP},
	}
	if res.Workspace != "" {
		r.session, _ = dp.NewUsageStore(res.Workspace).Totals()
	}
	return r
}

// OnCallBegin emits the frame (rule + header + gap, chrome only) and the
// CLI-computed per-call pre-flight estimate.
func (r *callRenderer) OnCallBegin(callIndex int, messages []llm.Message) {
	if r.chrome {
		_, _ = fmt.Fprint(r.env.stderr, r.lines.TurnOpening(r.priorCalls+callIndex+1, r.res.Mode))
	}
	estimate := llm.EstimatePayload(r.res.Person, agentport.ToolDefs(r.reg), messages)
	_, _ = fmt.Fprintln(r.env.stderr, r.lines.PayloadStatus(r.env.now(), estimate, r.res.effectiveBudget(), r.res.Mode, r.res.Provider.Model, true))
	if r.chrome {
		_, _ = fmt.Fprint(r.env.stderr, r.lines.TurnGap())
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
			_, _ = fmt.Fprintln(r.env.stderr)
			for _, reason := range roundReasons {
				_, _ = fmt.Fprintln(r.env.stderr, r.lines.ToolReason(r.env.now(), reason))
			}
		}
		if !usage.Reported {
			return
		}
		if r.renderedToolRound {
			_, _ = fmt.Fprintln(r.env.stderr)
		}
		_, _ = fmt.Fprintln(r.env.stderr, r.lines.PayloadStatus(r.env.now(), usage.PromptTokens, r.res.effectiveBudget(), r.res.Mode, r.res.Provider.Model, false))
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
	_, _ = fmt.Fprintln(r.env.stderr, r.lines.Metrics(now, r.res.Selected, metrics.UsageCounts{
		Miss:       rec.PromptTokens - rec.CachedTokens,
		Hit:        rec.CachedTokens,
		Completion: rec.ResponseTokens,
		Thinking:   rec.ThinkingTokens,
	}))
	_, _ = fmt.Fprintln(r.env.stderr, r.lines.Ready(cost, r.turnCost, r.session.Cost, r.session.Miss, r.session.Hit, r.session.Out, llm.HitRate(r.session.Hit, r.session.Miss)))
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
