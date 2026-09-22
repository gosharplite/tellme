package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
)

// Round 079 (ADR 0051) — the operator-interruption persistence contract.
//
// The E2E carries the live signal path (a command that SIGINTs tellme mid-turn);
// these unit pins carry the branches a Gherkin scenario cannot reach cheaply: the
// TYPED detection (through the adapter's ProviderError wrap), the zero-step
// no-write, the append-failure degradation, the non-cancellation guard, and the
// synthetic-answer LITERAL (the round-064/076 literal-pin precedent).

// interruptedSteps is one completed tool step, as the loop records it.
func interruptedSteps(n int) []history.Step {
	steps := make([]history.Step, 0, n)
	for i := 0; i < n; i++ {
		steps = append(steps, history.Step{Tool: "execute_command", Arguments: `{"command":"x"}`, Result: "Exit Code: -1\n"})
	}
	return steps
}

// TestInterruptedTurnAnswerIsLiterallyPinned fixes the synthetic answer's exact
// text (a literal pin: a reviewed change is required to alter it, since it is
// STORED in history.jsonl and re-read on resume).
func TestInterruptedTurnAnswerIsLiterallyPinned(t *testing.T) {
	if history.InterruptedTurnAnswer != "[Turn interrupted by operator via Ctrl+C]" {
		t.Fatalf("InterruptedTurnAnswer = %q, want the pinned literal", history.InterruptedTurnAnswer)
	}
}

// TestRunTurn_InterruptedWithStepsPersistsPartialTurn pins the round's primary
// behaviour: an operator interruption (a cancelled turn context) that completed a
// tool step is PERSISTED as one entry closed with the synthetic answer, returns
// Success (0), writes nothing to stdout, and reports the interruption on stderr.
// The cancellation is delivered WRAPPED in the adapter's ProviderError, so the
// TYPED detection (errors.Is) is exercised end-to-end rather than a bare sentinel.
func TestRunTurn_InterruptedWithStepsPersistsPartialTurn(t *testing.T) {
	var out, errOut bytes.Buffer
	cancelled := &llm.ProviderError{Provider: "p", Err: context.Canceled}
	lp := &fakeLoop{
		runErr:    cancelled,
		errResult: agentport.Result{Steps: interruptedSteps(1), Calls: []llm.Usage{{Reported: true}}},
	}
	st := &fakeStore{}
	res := resolution{Selected: "p", Mode: "butler", MaxHistoryTokens: 1000000, Provider: config.Provider{Model: "deepseek-v4-flash"}}
	code := runTurn(res, st, "explore", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))

	if code != Success {
		t.Fatalf("code = %d, want %d (success) — an operator interruption that kept its work is not a failure", code, Success)
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want empty (the interrupted path writes no answer)", out.String())
	}
	if len(st.appended) != 1 {
		t.Fatalf("appended %d entries, want exactly 1", len(st.appended))
	}
	got := st.appended[0]
	if got.Prompt != "explore" {
		t.Errorf("prompt = %q, want %q", got.Prompt, "explore")
	}
	if got.Answer != history.InterruptedTurnAnswer {
		t.Errorf("answer = %q, want the synthetic %q", got.Answer, history.InterruptedTurnAnswer)
	}
	if got.Calls != 1 {
		t.Errorf("calls = %d, want 1 (the completed inference rounds)", got.Calls)
	}
	if len(got.Steps) != 1 {
		t.Errorf("steps = %d, want 1 (the completed tool step)", len(got.Steps))
	}
	if !strings.Contains(errOut.String(), "interrupted by operator") {
		t.Errorf("stderr = %q, want the informational interruption line", errOut.String())
	}
	if strings.Contains(errOut.String(), "tellme: ") {
		t.Errorf("stderr = %q, must NOT carry a `tellme: ` class phrase (no failure occurred)", errOut.String())
	}
}

// TestRunTurn_InterruptedWithoutStepsWritesNothing pins the non-negotiable I-2:
// a turn interrupted BEFORE any tool step completed keeps today's clean abort —
// no history write, the frozen provider phrase, exit 6.
func TestRunTurn_InterruptedWithoutStepsWritesNothing(t *testing.T) {
	var out, errOut bytes.Buffer
	lp := &fakeLoop{
		runErr:    &llm.ProviderError{Provider: "p", Err: context.Canceled},
		errResult: agentport.Result{},
	}
	st := &fakeStore{}
	code := runTurn(resolution{Selected: "p"}, st, "explore", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))

	if code != ProviderError {
		t.Fatalf("code = %d, want %d (provider error) — zero steps keeps today's clean abort", code, ProviderError)
	}
	if len(st.appended) != 0 {
		t.Errorf("appended %d entries, want 0 (nothing to keep)", len(st.appended))
	}
	if !strings.Contains(errOut.String(), "tellme: the provider request failed") {
		t.Errorf("stderr = %q, want the frozen provider phrase", errOut.String())
	}
}

// TestRunTurn_InterruptedAppendFailureIsEnvironmentError pins the degradation: a
// failed partial-save append surfaces the existing environment-class phrase (exit
// 4) rather than a crash or a silent loss.
func TestRunTurn_InterruptedAppendFailureIsEnvironmentError(t *testing.T) {
	var out, errOut bytes.Buffer
	lp := &fakeLoop{
		runErr:    &llm.ProviderError{Provider: "p", Err: context.Canceled},
		errResult: agentport.Result{Steps: interruptedSteps(1), Calls: []llm.Usage{{Reported: true}}},
	}
	st := &fakeStore{appendErr: errors.New("disk full")}
	code := runTurn(resolution{Selected: "p"}, st, "explore", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))

	if code != EnvironmentError {
		t.Fatalf("code = %d, want %d (environment error) — a failed partial save degrades to the history phrase", code, EnvironmentError)
	}
	if !strings.Contains(errOut.String(), "tellme: the runtime home is not usable") {
		t.Errorf("stderr = %q, want the environment class phrase", errOut.String())
	}
}

// TestRunTurn_NonCancellationErrorWithStepsPersistsAsFailure pins the typed guard
// AND the round-080 change: a NON-cancellation failure that carries completed
// steps is NOT an interruption — it stays a failure (exit 6) — but round 080 now
// PERSISTS the completed steps with the FAILURE synthetic answer (never the
// operator-interruption answer).
func TestRunTurn_NonCancellationErrorWithStepsPersistsAsFailure(t *testing.T) {
	var out, errOut bytes.Buffer
	lp := &fakeLoop{
		runErr:    &llm.ProviderError{Provider: "p", Err: errors.New("connection reset")},
		errResult: agentport.Result{Steps: interruptedSteps(2), Calls: []llm.Usage{{Reported: true}, {Reported: true}}},
	}
	st := &fakeStore{}
	code := runTurn(resolution{Selected: "p"}, st, "explore", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))

	if code != ProviderError {
		t.Fatalf("code = %d, want %d (provider error) — only a cancellation is an interruption", code, ProviderError)
	}
	if len(st.appended) != 1 {
		t.Fatalf("appended %d entries, want 1 (round 080 keeps the completed steps)", len(st.appended))
	}
	if st.appended[0].Answer != history.ProviderFailedTurnAnswer {
		t.Errorf("answer = %q, want the failure answer %q (not the operator-interruption one)", st.appended[0].Answer, history.ProviderFailedTurnAnswer)
	}
}
