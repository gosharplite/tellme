package cli

import (
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// Round 035 (issue #72): the round-034 per-call tail is written by the `call`
// half of the composite observer while the round-019/025 spinner is still live,
// so the tail's first grouped `[Tool Reason]` shares a row with the spinner
// frame and the frame residue survives the finished turn. The fix is a
// phase-boundary yield on the composite: the spinner is synchronously cleared
// BEFORE the tail's write on a NON-final call, and never resumed there (the next
// waiting phase re-activates it). These tests pin that ordering contract — the
// layer a flat capture cannot witness (research Decision 5).

// recordingCallObserver records the call-half hooks into a shared sequence.
type recordingCallObserver struct{ seq *[]string }

func (o *recordingCallObserver) OnCallBegin(int, []llm.Message) {
	*o.seq = append(*o.seq, "call.OnCallBegin")
}

func (o *recordingCallObserver) OnCallEnd(int, llm.Usage, []string, bool) {
	*o.seq = append(*o.seq, "call.OnCallEnd")
}

// recordingSpinner records the spinner-half hooks into the same sequence; only
// the yield hooks matter here (a yield is YieldIndicator, a restore is
// RestoreIndicator — round 045's split vocabulary).
type recordingSpinner struct{ seq *[]string }

func (s *recordingSpinner) OnCallBegin(int, []llm.Message)           {}
func (s *recordingSpinner) OnCallEnd(int, llm.Usage, []string, bool) {}

func (s *recordingSpinner) OnInferenceStart() { *s.seq = append(*s.seq, "spinner.OnInferenceStart") }
func (s *recordingSpinner) OnInferenceEnd()   {}
func (s *recordingSpinner) OnToolsStart([]string) {
	*s.seq = append(*s.seq, "spinner.OnToolsStart")
}
func (s *recordingSpinner) OnToolsEnd()       {}
func (s *recordingSpinner) YieldIndicator()   { *s.seq = append(*s.seq, "spinner.yield") }
func (s *recordingSpinner) RestoreIndicator() { *s.seq = append(*s.seq, "spinner.restore") }

// seqString renders a recorded sequence for readable failure messages.
func seqString(seq []string) string { return "[" + strings.Join(seq, ", ") + "]" }

// TestCompositeOnCallEndYieldsSpinnerBeforeNonFinalTail pins the fix's core
// ordering: on a NON-final call the spinner is cleared BEFORE the tail is
// written, and it is NOT resumed (the phase boundary leaves the line clean for
// the next call's frame — research Decision 1). This is the RED pin: today the
// composite writes the tail with no yield.
func TestCompositeOnCallEndYieldsSpinnerBeforeNonFinalTail(t *testing.T) {
	var seq []string
	c := compositeObserver{
		call:    &recordingCallObserver{seq: &seq},
		spinner: &recordingSpinner{seq: &seq},
	}

	c.OnCallEnd(0, llm.Usage{}, []string{"read the notes"}, false)

	want := []string{"spinner.yield", "call.OnCallEnd"}
	if got := seqString(seq); got != seqString(want) {
		t.Fatalf("non-final OnCallEnd ordering = %s, want %s (the spinner must yield before the tail write and not restore)", got, seqString(want))
	}
}

// TestCompositeOnCallEndDoesNotYieldOnFinalCall pins that the FINAL call writes
// no tail at call-end time (the renderer defers it past the answer), so the
// composite must not clear or resume the spinner there — keeping the final path
// byte-identical to before the fix.
func TestCompositeOnCallEndDoesNotYieldOnFinalCall(t *testing.T) {
	var seq []string
	c := compositeObserver{
		call:    &recordingCallObserver{seq: &seq},
		spinner: &recordingSpinner{seq: &seq},
	}

	c.OnCallEnd(0, llm.Usage{}, nil, true)

	want := []string{"call.OnCallEnd"}
	if got := seqString(seq); got != seqString(want) {
		t.Fatalf("final OnCallEnd ordering = %s, want %s (the deferred final tail must not touch the spinner)", got, seqString(want))
	}
}

// TestCompositeOnCallEndNilSafe pins the nil-safe delegation: a gated-off
// spinner (nil) must not panic and must still deliver the tail to the call half.
func TestCompositeOnCallEndNilSafe(t *testing.T) {
	var seq []string
	c := compositeObserver{call: &recordingCallObserver{seq: &seq}}

	c.OnCallEnd(0, llm.Usage{}, []string{"r"}, false)

	want := []string{"call.OnCallEnd"}
	if got := seqString(seq); got != seqString(want) {
		t.Fatalf("nil-spinner OnCallEnd ordering = %s, want %s", got, seqString(want))
	}
}
