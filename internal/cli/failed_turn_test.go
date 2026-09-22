package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
)

// Round 080 (ADR 0052; closes #161) — a FAILED turn with completed tool steps is
// persisted (the round-079 shape, a class-specific synthetic answer) while the
// failure surface is reported unchanged (phrase + exit 6/7).

func failedSteps(n int) []history.Step { return interruptedSteps(n) }

// TestFailedTurnAnswersAreLiterallyPinned fixes the two new synthetic answers
// (they are STORED in history.jsonl and re-read on resume — a literal pin forces
// a reviewed change).
func TestFailedTurnAnswersAreLiterallyPinned(t *testing.T) {
	if history.ProviderFailedTurnAnswer != "[Turn ended early: the provider request failed]" {
		t.Fatalf("ProviderFailedTurnAnswer = %q, want the pinned literal", history.ProviderFailedTurnAnswer)
	}
	if history.ToolFailedTurnAnswer != "[Turn ended early: the tool loop did not complete]" {
		t.Fatalf("ToolFailedTurnAnswer = %q, want the pinned literal", history.ToolFailedTurnAnswer)
	}
}

// TestRunTurn_FailedProviderTurnWithStepsIsPersisted pins the round's primary
// behaviour: a provider failure (a real error, NOT a cancellation) with a
// completed step is PERSISTED (one entry, the provider-failure synthetic answer),
// the failure surface is unchanged (phrase + exit 6), `stdout` is empty, and the
// informational keep line is emitted.
func TestRunTurn_FailedProviderTurnWithStepsIsPersisted(t *testing.T) {
	var out, errOut bytes.Buffer
	lp := &fakeLoop{
		runErr:    &llm.ProviderError{Provider: "p", Err: errors.New("connection reset by peer"), Transport: true},
		errResult: agentport.Result{Steps: failedSteps(1), Calls: []llm.Usage{{Reported: true}}},
	}
	st := &fakeStore{}
	res := resolution{Selected: "p", Mode: "butler", MaxHistoryTokens: 1000000, Provider: config.Provider{Model: "deepseek-v4-flash"}}
	code := runTurn(res, st, "explore", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))

	if code != ProviderError {
		t.Fatalf("code = %d, want %d (provider error) — the failure surface is unchanged", code, ProviderError)
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want empty (a failed turn writes no answer)", out.String())
	}
	if len(st.appended) != 1 {
		t.Fatalf("appended %d entries, want exactly 1", len(st.appended))
	}
	got := st.appended[0]
	if got.Answer != history.ProviderFailedTurnAnswer {
		t.Errorf("answer = %q, want the provider failure answer %q", got.Answer, history.ProviderFailedTurnAnswer)
	}
	if got.Calls != 1 || len(got.Steps) != 1 {
		t.Errorf("entry = {calls:%d, steps:%d}, want {1, 1}", got.Calls, len(got.Steps))
	}
	if !strings.Contains(errOut.String(), "tellme: the provider request failed") {
		t.Errorf("stderr = %q, want the frozen provider phrase", errOut.String())
	}
	if !strings.Contains(errOut.String(), "kept 1 completed tool step(s) in the session history") {
		t.Errorf("stderr = %q, want the informational keep line", errOut.String())
	}
}

// TestRunTurn_ToolLoopFailureWithStepsIsPersisted pins the exit-7 (tool-loop)
// class: ErrIncomplete with a completed step is persisted with the tool-failure
// answer, and the tool phrase + exit 7 are unchanged.
func TestRunTurn_ToolLoopFailureWithStepsIsPersisted(t *testing.T) {
	var out, errOut bytes.Buffer
	lp := &fakeLoop{
		runErr:    &agentport.ErrIncomplete{Reason: "the tool-loop bound was reached"},
		errResult: agentport.Result{Steps: failedSteps(2), Calls: []llm.Usage{{Reported: true}, {Reported: true}}},
	}
	st := &fakeStore{}
	code := runTurn(resolution{Selected: "p"}, st, "explore", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))

	if code != ToolError {
		t.Fatalf("code = %d, want %d (tool error)", code, ToolError)
	}
	if len(st.appended) != 1 || st.appended[0].Answer != history.ToolFailedTurnAnswer {
		t.Fatalf("appended = %+v, want one entry with the tool failure answer", st.appended)
	}
	if !strings.Contains(errOut.String(), "tellme: the tool request failed") {
		t.Errorf("stderr = %q, want the frozen tool phrase", errOut.String())
	}
}

// TestRunTurn_FailedTurnWithoutStepsWritesNothing pins the non-negotiable I-3: a
// failure before any completed step keeps today's clean abort (no write, phrase,
// exit 6).
func TestRunTurn_FailedTurnWithoutStepsWritesNothing(t *testing.T) {
	var out, errOut bytes.Buffer
	lp := &fakeLoop{runErr: &llm.ProviderError{Provider: "p", Err: errors.New("boom")}}
	st := &fakeStore{}
	code := runTurn(resolution{Selected: "p"}, st, "explore", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))

	if code != ProviderError {
		t.Fatalf("code = %d, want %d", code, ProviderError)
	}
	if len(st.appended) != 0 {
		t.Errorf("appended %d entries, want 0 (nothing to keep)", len(st.appended))
	}
	if strings.Contains(errOut.String(), "kept ") {
		t.Errorf("stderr = %q, must NOT carry the keep line when nothing was kept", errOut.String())
	}
}

// TestRunTurn_FailedTurnAppendFailureDoesNotMaskFailure pins D6: a failed-turn
// append failure is best-effort — the failure surface is still reported (exit 6 +
// the phrase) and the keep line is NOT emitted (no lie).
func TestRunTurn_FailedTurnAppendFailureDoesNotMaskFailure(t *testing.T) {
	var out, errOut bytes.Buffer
	lp := &fakeLoop{
		runErr:    &llm.ProviderError{Provider: "p", Err: errors.New("boom")},
		errResult: agentport.Result{Steps: failedSteps(1), Calls: []llm.Usage{{Reported: true}}},
	}
	st := &fakeStore{appendErr: errors.New("disk full")}
	code := runTurn(resolution{Selected: "p"}, st, "explore", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))

	if code != ProviderError {
		t.Fatalf("code = %d, want %d (the append failure must not mask the provider failure)", code, ProviderError)
	}
	if !strings.Contains(errOut.String(), "tellme: the provider request failed") {
		t.Errorf("stderr = %q, want the frozen provider phrase", errOut.String())
	}
	if strings.Contains(errOut.String(), "kept ") {
		t.Errorf("stderr = %q, must NOT claim a keep that failed", errOut.String())
	}
}
