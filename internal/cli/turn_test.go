package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/render"
)

// fakeGateway is an in-memory llm.Gateway for runTurn tests (review finding #1:
// the turn must be unit-testable without the concrete adapter).
type fakeGateway struct {
	text   string
	usage  llm.Usage
	err    error
	got    llm.Request
	script []llm.Response // when set, responses are served in order (round-010 tool-loop test)
	served int
}

func (f *fakeGateway) Complete(_ context.Context, req llm.Request) (llm.Response, error) {
	f.got = req
	if len(f.script) > 0 {
		i := f.served
		if i >= len(f.script) {
			i = len(f.script) - 1
		}
		f.served++
		return f.script[i], f.err
	}
	return llm.Response{Text: f.text, Usage: f.usage}, f.err
}

// fakeStore is an in-memory history.Store for runTurn tests (round-007 TD-1).
type fakeStore struct {
	entries     []history.Entry
	loadErr     error
	appendErr   error
	archiveErr  error
	rollbackErr error
	appended    []history.Entry
	archived    bool
	rolled      []int // the counts passed to Rollback
}

func (f *fakeStore) Load() ([]history.Entry, error) { return f.entries, f.loadErr }
func (f *fakeStore) Append(e history.Entry) error {
	f.appended = append(f.appended, e)
	return f.appendErr
}
func (f *fakeStore) Archive() error { f.archived = true; return f.archiveErr }

// Rollback is the round-081 store capability: an in-memory clamp emulation for
// runTurn/rollback unit tests.
func (f *fakeStore) Rollback(n int) (int, error) {
	f.rolled = append(f.rolled, n)
	if f.rollbackErr != nil {
		return 0, f.rollbackErr
	}
	if n <= 0 || len(f.entries) == 0 {
		return 0, nil
	}
	removed := n
	if removed > len(f.entries) {
		removed = len(f.entries)
	}
	f.entries = f.entries[:len(f.entries)-removed]
	return removed, nil
}

// stubRenderer is a render.Answer whose behaviour the caller scripts.
type stubRenderer struct {
	out      string
	degraded bool
	warned   bool
}

func (s *stubRenderer) Render(string, int) (string, bool) { return s.out, s.degraded }
func (s *stubRenderer) WarnDegraded(io.Writer)            { s.warned = true }

// env builds a runtimeEnv over the given buffers + renderer for a unit test.
// The clock seam is fixed so the payload status line is deterministic.
func env(out, errOut io.Writer, r render.Answer) runtimeEnv {
	return runtimeEnv{stdout: out, stderr: errOut, renderer: r,
		clock: func() time.Time { return time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC) }}
}

func TestRunTurn_PrintsRawAnswerAndPersists(t *testing.T) {
	var out, errOut bytes.Buffer
	// Round 050 (Q4 → A): the loop is the in-package fake; the CLI's orchestration
	// (prompt/prior handoff, result persistence, answer rendering) is what is
	// pinned here.
	lp := &fakeLoop{result: agentport.Result{Answer: "the answer", Usage: llm.Usage{Reported: false}, Calls: []llm.Usage{{Reported: false}}}}
	st := &fakeStore{}
	code := runTurn(resolution{Selected: "p", Mode: "butler", MaxHistoryTokens: 1000000, Provider: config.Provider{Model: "deepseek-v4-flash"}}, st, "ping", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))
	if code != Success {
		t.Fatalf("code = %d, want %d (success)", code, Success)
	}
	if lp.gotPrompt != "ping" {
		t.Errorf("loop got prompt %q, want ping", lp.gotPrompt)
	}
	if len(lp.gotPrior) != 0 {
		t.Errorf("loop got prior %+v, want none on a fresh conversation", lp.gotPrior)
	}
	if out.String() != "the answer\n" {
		t.Errorf("stdout = %q, want %q", out.String(), "the answer\n")
	}
	// Round-050 fold TD-2(i) positional pin, restored in sentinel form (round-051
	// fold F-1): with an UNREPORTED usage the deferred post-turn status writes
	// nothing, so the PRE-FLIGHT payload line is the LAST stderr write before the
	// answer. Asserting the last line (not a bare Contains) keeps the position
	// falsifiable — dropping the round-018 `if !usage.Reported` gate reds it.
	lines := strings.Split(strings.TrimRight(errOut.String(), "\n"), "\n")
	last := lines[len(lines)-1]
	if !strings.HasPrefix(last, "<estimate ") || !strings.HasSuffix(last, "butler deepseek-v4-flash>") {
		t.Errorf("last stderr line = %q, want the PRE-FLIGHT payload estimate line as the LAST write (unreported usage defers nothing)", last)
	}
	if len(st.appended) != 1 || st.appended[0].Prompt != "ping" || st.appended[0].Answer != "the answer" {
		t.Errorf("persisted = %+v, want the completed exchange", st.appended)
	}
	if len(st.appended) == 1 && st.appended[0].Calls != 1 {
		t.Errorf("persisted Calls = %d, want 1 (a tool-less turn made one call)", st.appended[0].Calls)
	}
}

func TestRunTurn_CarriesPriorMessages(t *testing.T) {
	var out, errOut bytes.Buffer
	// Round 050 (Q4 → A): runTurn's job is to hand the loaded prior turns + the
	// prompt to the loop. The wire-message construction (BuildMessages) is a loop
	// concern, covered by internal/agent's own tests.
	lp := &fakeLoop{result: agentport.Result{Answer: "b2", Usage: llm.Usage{Reported: true}}}
	st := &fakeStore{entries: []history.Entry{{Prompt: "q1", Answer: "a1"}}}
	code := runTurn(resolution{Selected: "p"}, st, "q2", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))
	if code != Success {
		t.Fatalf("code = %d, want success", code)
	}
	if lp.gotPrompt != "q2" {
		t.Errorf("loop prompt = %q, want q2", lp.gotPrompt)
	}
	want := []history.Entry{{Prompt: "q1", Answer: "a1"}}
	if !reflect.DeepEqual(lp.gotPrior, want) {
		t.Errorf("prior = %+v, want %+v", lp.gotPrior, want)
	}
}

func TestRunTurn_LoadErrorIsEnvironmentError(t *testing.T) {
	var out, errOut bytes.Buffer
	st := &fakeStore{loadErr: errors.New("boom")}
	code := runTurn(resolution{Selected: "p"}, st, "ping", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(&fakeLoop{}))
	if code != EnvironmentError {
		t.Fatalf("code = %d, want %d (environment error)", code, EnvironmentError)
	}
	if !strings.HasPrefix(errOut.String(), "tellme: the runtime home is not usable") {
		t.Errorf("stderr = %q, want the environment class phrase", errOut.String())
	}
}

func TestRunTurn_ProviderFailure(t *testing.T) {
	var out, errOut bytes.Buffer
	// Round 050 (Q4 → A): the provider/transport failure now surfaces from the
	// loop seam (the real loop returns the unwrapped error); the CLI maps it to
	// the frozen class phrase + code 6.
	lp := &fakeLoop{runErr: &llm.ProviderError{Provider: "p", Err: errors.New("boom")}}
	code := runTurn(resolution{Selected: "p"}, &fakeStore{}, "ping", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))
	if code != ProviderError {
		t.Fatalf("code = %d, want %d (provider error)", code, ProviderError)
	}
	if !strings.Contains(errOut.String(), "tellme: the provider request failed") {
		t.Errorf("stderr = %q, want the frozen class phrase", errOut.String())
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want empty", out.String())
	}
}

func TestRunTurn_UnsupportedFamilyIsProviderError(t *testing.T) {
	var out, errOut bytes.Buffer
	buildErr := &llm.ProviderError{Provider: "p", Err: errors.New(`unsupported provider family "gemini"`)}
	code := runTurn(resolution{Selected: "p"}, &fakeStore{}, "ping", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), depsWithGateway(nil, buildErr))
	if code != ProviderError {
		t.Fatalf("code = %d, want %d (provider error)", code, ProviderError)
	}
	if !strings.Contains(errOut.String(), "unsupported provider family") {
		t.Errorf("stderr = %q, want the actionable reason", errOut.String())
	}
}

func TestParseFlagsHistorySurfaces(t *testing.T) {
	// -l and --list are equivalent; the flag being present is recorded.
	for _, form := range []string{"-l", "--list"} {
		opts, _, ok := parseFlags([]string{form, "3"}, io.Discard)
		if !ok {
			t.Fatalf("parseFlags(%q 3) ok = false", form)
		}
		if !opts.listSet || opts.list != 3 {
			t.Errorf("parseFlags(%q 3): listSet=%v list=%d, want true/3", form, opts.listSet, opts.list)
		}
	}
	opts, _, ok := parseFlags([]string{"--new"}, io.Discard)
	if !ok || !opts.newSession {
		t.Errorf("parseFlags(--new): ok=%v newSession=%v, want true/true", ok, opts.newSession)
	}
}

// TestRunTurn_PostTurnStatusFollowsAnswer pins the round-009 write ordering: the
// pre-flight status line leads the answer, and the post-turn measured line trails
// it. A single interleaved buffer captures the write order of stdout and stderr.
func TestRunTurn_PostTurnStatusFollowsAnswer(t *testing.T) {
	var buf bytes.Buffer
	lp := &fakeLoop{result: agentport.Result{Answer: "ANSWER", Usage: llm.Usage{Reported: true, PromptTokens: 42}, Calls: []llm.Usage{{Reported: true, PromptTokens: 42}}}}
	res := resolution{Selected: "p", Mode: "butler", MaxHistoryTokens: 1000000, Workspace: t.TempDir(), Provider: config.Provider{Model: "deepseek-v4-flash"}}
	e := runtimeEnv{stdout: &buf, stderr: &buf, renderer: &stubRenderer{out: "ANSWER"},
		clock: func() time.Time { return time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC) }}
	if code := runTurn(res, &fakeStore{}, "ping", turnOptions{raw: true}, e, depsWithLoop(lp)); code != Success {
		t.Fatalf("code = %d, want success", code)
	}
	out := buf.String()
	pre := strings.Index(out, "<estimate ")
	answer := strings.Index(out, "ANSWER")
	post := strings.Index(out, "<payload 42/1000000")
	if pre < 0 || answer < 0 || post < 0 {
		t.Fatalf("missing markers in output: %q", out)
	}
	if pre >= answer || answer >= post {
		t.Fatalf("write order = pre(%d) answer(%d) post(%d), want pre < answer < post: %q", pre, answer, post, out)
	}
}

// TestRunTurn_ToolLoopLogPrecedesAnswer pins the round-010 tool-loop ordering at
// the unit layer: with stdout and stderr bound to ONE interleaved buffer, the
// live tool-loop log line (stderr, naming the tool) is written before the answer
// (stdout).
func TestRunTurn_ToolLoopLogPrecedesAnswer(t *testing.T) {
	// Round 050 (Q4 → A): the loop performs its diagnostic writes to the spec's
	// Stderr during Run; the CLI writes the answer after Run returns. This pins the
	// CLI's ordering guarantee (loop output precedes the answer) against the fake
	// loop's scripted writes.
	var buf bytes.Buffer
	// Round-050 fold N-4: no baked timestamp in the double's emitted line.
	lp := &fakeLoop{
		result: agentport.Result{Answer: "ANSWER", Usage: llm.Usage{Reported: true}, Calls: []llm.Usage{{Reported: true}, {Reported: true}}},
		writes: []string{"[Tool Engine] Step 1/1 read_files"},
	}
	res := resolution{Selected: "p", Mode: "butler", MaxHistoryTokens: 1000000, Provider: config.Provider{Model: "deepseek-v4-flash"}}
	e := runtimeEnv{stdout: &buf, stderr: &buf, renderer: &stubRenderer{out: "ANSWER"},
		clock: func() time.Time { return time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC) }}
	st := &fakeStore{}
	if code := runTurn(res, st, "ping", turnOptions{raw: true}, e, depsWithLoop(lp)); code != Success {
		t.Fatalf("code = %d, want success", code)
	}
	if len(st.appended) == 1 && st.appended[0].Calls != 2 {
		t.Errorf("persisted Calls = %d, want 2 (the tool round + the answer)", st.appended[0].Calls)
	}
	out := buf.String()
	tool := strings.Index(out, "read_files")
	answer := strings.Index(out, "ANSWER")
	if tool < 0 || answer < 0 {
		t.Fatalf("missing markers in output: %q", out)
	}
	if tool >= answer {
		t.Fatalf("write order = tool(%d) answer(%d), want tool < answer: %q", tool, answer, out)
	}
}

// TestRunTurn_NonFinalTailPrecedesAnswer pins the round-050 fold TD-2(ii): the
// runTurn wiring for a NON-final call. With a `{false, true}` schedule the fake
// loop drives a non-final call (whose tail renders immediately — measured payload
// + metrics + Ready) followed by the final call (whose tail is DEFERRED past the
// answer by the call renderer / EmitFinalTail). With stdout and stderr bound to
// ONE interleaved buffer, the non-final tail's `Ready` precedes the answer and
// the deferred final tail's `Ready` trails it.
func TestRunTurn_NonFinalTailPrecedesAnswer(t *testing.T) {
	var buf bytes.Buffer
	lp := &fakeLoop{
		result: agentport.Result{Answer: "ANSWER", Usage: llm.Usage{Reported: true, PromptTokens: 42}, Calls: []llm.Usage{{Reported: true, PromptTokens: 42}, {Reported: true, PromptTokens: 42}}},
		finals: []bool{false, true},
	}
	res := resolution{Selected: "p", Mode: "butler", MaxHistoryTokens: 1000000, Workspace: t.TempDir(), Provider: config.Provider{Model: "deepseek-v4-flash"}}
	e := runtimeEnv{stdout: &buf, stderr: &buf, renderer: &stubRenderer{out: "ANSWER"},
		clock: func() time.Time { return time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC) }}
	if code := runTurn(res, &fakeStore{}, "ping", turnOptions{raw: true}, e, depsWithLoop(lp)); code != Success {
		t.Fatalf("code = %d, want success", code)
	}
	out := buf.String()
	firstReady := strings.Index(out, "<ready>")
	answer := strings.Index(out, "ANSWER")
	lastReady := strings.LastIndex(out, "<ready>")
	if firstReady < 0 || answer < 0 || lastReady < 0 {
		t.Fatalf("missing markers in output: %q", out)
	}
	if firstReady >= answer || answer >= lastReady {
		t.Fatalf("write order = nonFinalTail(%d) answer(%d) finalTail(%d), want nonFinalTail < answer < finalTail: %q", firstReady, answer, lastReady, out)
	}
}

// TestRunTurn_ChromeHeaderCountsCalls pins the round-027 header at the runTurn
// composition level: with chrome on and a prior turn that made two inference
// rounds, the turn opens at `Turn 3` (Σ calls + 1) on the diagnostic stream.
func TestRunTurn_ChromeHeaderCountsCalls(t *testing.T) {
	var out, errOut bytes.Buffer
	lp := &fakeLoop{result: agentport.Result{Answer: "the answer", Usage: llm.Usage{Reported: true}}}
	st := &fakeStore{entries: []history.Entry{{Prompt: "q1", Answer: "a1", Calls: 2}}}
	code := runTurn(resolution{Selected: "p", Mode: "butler", MaxHistoryTokens: 1000000, Provider: config.Provider{Model: "deepseek-v4-flash"}}, st, "ping", turnOptions{raw: true, chrome: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp))
	if code != Success {
		t.Fatalf("code = %d, want success", code)
	}
	if !strings.Contains(errOut.String(), "<turn-opening 3 butler>") {
		t.Errorf("stderr = %q, want the `<turn-opening 3 butler>` frame (Σ calls + 1)", errOut.String())
	}
}
