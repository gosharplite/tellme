package ui

// Round 045 (R3 of #92) — the single named owner of the progress-indicator YIELD
// POLICY.
//
// THE POLICY (authoritative; every yield in the turn path routes through here):
//
//   - Yield clears the indicator because a non-indicator line is about to be
//     written to the diagnostic stream, so that line starts on its own cleared
//     row. A yield is ALWAYS clear-only: the caller never resumes here.
//   - Restore resumes the indicator (a synchronous first frame). Restoration is a
//     PHASE-BOUNDARY act — the next waiting phase, the end of a streaming block,
//     or (via Admit) the idle-gap watcher; never mid-write.
//   - Admit resumes with the first frame drawn on the redraw goroutine — the one
//     resume that must NOT block inside another component's critical section (the
//     `[Tool Output]` idle-gap resume, admitted under the block writer's mutex).
//
// WHY ONE OWNER: the MECHANISM was never the problem — Spinner.deactivate(),
// resume() and AdmitResume() already give a synchronous, goroutine-joined clear,
// an epoch-preserving resume, and (round 025) a row-aware clear; that machinery
// stays in spinner.go (round-019 research D10; ADR 0009 D4). What rounds 035/040
// left unnamed was the POLICY: three homes (the loop's withToolLog, the CLI
// compositeObserver's per-call tail, and this package's ToolOutputCoordinator)
// each decided locally when to clear and whether to resume, and each carried its
// own copy of the rationale. This type is that policy's ONE home: the loop and
// the composite reach it through the agent-facing port adapter on *Spinner, and
// the coordinator holds it directly.
//
// A yield is NOT the turn teardown: Spinner.Stop() remains the idempotent
// end-of-turn clear (it is called after the loop returns and by the panic-safe
// residue guard); a mid-turn clear is Yield. Policy recorded in ADR 0014
// (docs/decisions/0014-yield-policy-owner.md).
type YieldController struct{ sp *Spinner }

// NewYieldController makes the turn spinner the yield policy's owner. A nil
// spinner is allowed and makes every method a no-op — the gated-off case
// (round-019 FR-006), so callers need no nil check of their own.
func NewYieldController(sp *Spinner) YieldController { return YieldController{sp: sp} }

// Enabled reports whether the controller owns a live spinner (false when gated
// off), so a caller can skip work that would be a no-op.
func (y YieldController) Enabled() bool { return y.sp != nil }

// Yield clears the indicator (clear-only; see the type doc). It is nil-safe.
func (y YieldController) Yield() {
	if y.sp != nil {
		y.sp.deactivate()
	}
}

// Restore resumes the indicator with a synchronous first frame (a phase-boundary
// act; see the type doc). It is nil-safe.
func (y YieldController) Restore() {
	if y.sp != nil {
		y.sp.resume()
	}
}

// Admit resumes the indicator with the first frame drawn on the redraw goroutine,
// for a caller that must not block inside another component's critical section
// (the `[Tool Output]` idle-gap resume). It is nil-safe.
func (y YieldController) Admit() {
	if y.sp != nil {
		y.sp.AdmitResume()
	}
}
