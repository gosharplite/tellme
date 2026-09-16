package cli

import (
	"fmt"
	"time"

	"github.com/gosharplite/tellme/internal/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	"github.com/gosharplite/tellme/internal/ui"
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
// The loop owns no estimator/persona; the CLI-computed estimate is the only path
// to the number (ADR 0005 D2). The metrics/`Ready` session field is a recorded
// DISPLAY-ONLY divergence (ADR 0005 D4): it folds each call as it arrives, even a
// call the final persistence gate will not write.
type callRenderer struct {
	env        runtimeEnv
	res        resolution
	reg        domaintools.Registry
	chrome     bool
	priorCalls int

	pricing  ui.Pricing
	session  history.UsageSummary
	turnCost float64

	finalTail func()
}

// newCallRenderer builds the per-call renderer, seeding the session roll-up from
// the persisted summary so the `Ready` session totals accumulate across the
// session (round-018 D7).
func newCallRenderer(env runtimeEnv, res resolution, reg domaintools.Registry, chrome bool, priorCalls int) *callRenderer {
	r := &callRenderer{
		env:        env,
		res:        res,
		reg:        reg,
		chrome:     chrome,
		priorCalls: priorCalls,
		pricing:    ui.Pricing{Hit: res.Pricing.HIT, Miss: res.Pricing.MISS, Comp: res.Pricing.COMP},
	}
	if res.Workspace != "" {
		r.session, _ = newUsageStore(res.Workspace).Totals()
	}
	return r
}

// OnCallBegin emits the frame (rule + header + gap, chrome only) and the
// CLI-computed per-call pre-flight estimate.
func (r *callRenderer) OnCallBegin(callIndex int, messages []llm.Message) {
	if r.chrome {
		_, _ = fmt.Fprint(r.env.stderr, ui.FormatTurnOpening(r.priorCalls+callIndex+1, r.res.Mode))
	}
	estimate := llm.EstimatePayload(r.res.Person, agent.ToolDefs(r.reg), messages)
	_, _ = fmt.Fprintln(r.env.stderr, ui.FormatPayloadStatus(r.env.now(), estimate, r.res.effectiveBudget(), r.res.Mode, r.res.Provider.Model, true))
	if r.chrome {
		_, _ = fmt.Fprint(r.env.stderr, ui.FormatTurnGap())
	}
}

// OnCallEnd emits the call's tail, deferring the FINAL call's tail past the
// answer (G5).
func (r *callRenderer) OnCallEnd(callIndex int, usage llm.Usage, roundReasons []string, final bool) {
	emit := func() {
		for _, reason := range roundReasons {
			_, _ = fmt.Fprintln(r.env.stderr, ui.FormatToolReason(r.env.now(), reason))
		}
		if !usage.Reported {
			return
		}
		_, _ = fmt.Fprintln(r.env.stderr, ui.FormatPayloadStatus(r.env.now(), usage.PromptTokens, r.res.effectiveBudget(), r.res.Mode, r.res.Provider.Model, false))
		r.emitMetrics(usage)
	}
	if final {
		r.finalTail = emit
		return
	}
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
	miss := usage.PromptTokens - usage.CachedTokens
	cost := ui.ComputeCost(r.pricing, miss, usage.CachedTokens, usage.CompletionTokens, usage.ThinkingTokens)
	r.turnCost += cost
	r.session.Add(history.UsageRecord{
		CachedTokens:   usage.CachedTokens,
		PromptTokens:   usage.PromptTokens,
		ResponseTokens: usage.CompletionTokens,
		ThinkingTokens: usage.ThinkingTokens,
		Cost:           cost,
	})
	_, _ = fmt.Fprintln(r.env.stderr, ui.FormatMetrics(r.env.now(), r.res.Selected, ui.UsageCounts{
		Miss:       miss,
		Hit:        usage.CachedTokens,
		Completion: usage.CompletionTokens,
		Thinking:   usage.ThinkingTokens,
	}))
	_, _ = fmt.Fprintln(r.env.stderr, ui.FormatReady(cost, r.turnCost, r.session.Cost, r.session.Miss, r.session.Hit, r.session.Out, ui.HitRate(r.session.Hit, r.session.Miss)))
}

// persistTurnUsage writes the turn's usage ONCE (ADR 0005 D4/FR-010b): the
// Reported subset of result.Calls in a single AppendBatch — never per call — and
// ONLY when the FINAL call reports usage (the round-018 gate). A best-effort
// write; an empty workspace writes nothing.
func persistTurnUsage(env runtimeEnv, res resolution, result agent.AgentResult) {
	if res.Workspace == "" || !result.Usage.Reported {
		return
	}
	pricing := ui.Pricing{Hit: res.Pricing.HIT, Miss: res.Pricing.MISS, Comp: res.Pricing.COMP}
	now := env.now()
	records := make([]history.UsageRecord, 0, len(result.Calls))
	for _, c := range result.Calls {
		if !c.Reported {
			continue
		}
		miss := c.PromptTokens - c.CachedTokens
		cost := ui.ComputeCost(pricing, miss, c.CachedTokens, c.CompletionTokens, c.ThinkingTokens)
		records = append(records, history.UsageRecord{
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
	_ = newUsageStore(res.Workspace).AppendBatch(records)
}
