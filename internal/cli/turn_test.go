package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
)

// fakeGateway is an in-memory llm.Gateway for runTurn tests (review finding #1:
// the turn must be unit-testable without the concrete adapter).
type fakeGateway struct {
	text string
	err  error
	got  llm.Request
}

func (f *fakeGateway) Complete(_ context.Context, req llm.Request) (llm.Response, error) {
	f.got = req
	return llm.Response{Text: f.text}, f.err
}

func factoryReturning(gw llm.Gateway, err error) gatewayFactory {
	return func(config.Provider, string) (llm.Gateway, error) { return gw, err }
}

// fakeStore is an in-memory history.Store for runTurn tests (round-007 TD-1).
type fakeStore struct {
	entries    []history.Entry
	loadErr    error
	appendErr  error
	archiveErr error
	appended   []history.Entry
	archived   bool
}

func (f *fakeStore) Load() ([]history.Entry, error) { return f.entries, f.loadErr }
func (f *fakeStore) Append(e history.Entry) error {
	f.appended = append(f.appended, e)
	return f.appendErr
}
func (f *fakeStore) Archive() error { f.archived = true; return f.archiveErr }

// stubRenderer is an answerRenderer whose behaviour the caller scripts.
type stubRenderer struct {
	out      string
	degraded bool
	warned   bool
}

func (s *stubRenderer) Render(string, int) (string, bool) { return s.out, s.degraded }
func (s *stubRenderer) WarnDegraded(io.Writer)            { s.warned = true }

// env builds a runtimeEnv over the given buffers + renderer for a unit test.
func env(out, errOut io.Writer, r answerRenderer) runtimeEnv {
	return runtimeEnv{stdout: out, stderr: errOut, renderer: r}
}

func TestRunTurn_PrintsRawAnswerAndPersists(t *testing.T) {
	var out, errOut bytes.Buffer
	fg := &fakeGateway{text: "the answer"}
	st := &fakeStore{}
	code := runTurn(resolution{Selected: "p"}, st, "ping", true, env(&out, &errOut, &stubRenderer{}), factoryReturning(fg, nil))
	if code != Success {
		t.Fatalf("code = %d, want %d (success)", code, Success)
	}
	if fg.got.Prompt != "ping" {
		t.Errorf("gateway got prompt %q, want ping", fg.got.Prompt)
	}
	if len(fg.got.Messages) != 0 {
		t.Errorf("gateway got messages %+v, want none on a fresh conversation", fg.got.Messages)
	}
	if out.String() != "the answer\n" {
		t.Errorf("stdout = %q, want %q", out.String(), "the answer\n")
	}
	if errOut.Len() != 0 {
		t.Errorf("stderr = %q, want empty", errOut.String())
	}
	if len(st.appended) != 1 || st.appended[0].Prompt != "ping" || st.appended[0].Answer != "the answer" {
		t.Errorf("persisted = %+v, want the completed exchange", st.appended)
	}
}

func TestRunTurn_CarriesPriorMessages(t *testing.T) {
	var out, errOut bytes.Buffer
	fg := &fakeGateway{text: "b2"}
	st := &fakeStore{entries: []history.Entry{{Prompt: "q1", Answer: "a1"}}}
	code := runTurn(resolution{Selected: "p"}, st, "q2", true, env(&out, &errOut, &stubRenderer{}), factoryReturning(fg, nil))
	if code != Success {
		t.Fatalf("code = %d, want success", code)
	}
	want := []llm.Message{{Role: "user", Content: "q1"}, {Role: "assistant", Content: "a1"}}
	if len(fg.got.Messages) != 2 || fg.got.Messages[0] != want[0] || fg.got.Messages[1] != want[1] {
		t.Errorf("messages = %+v, want %+v", fg.got.Messages, want)
	}
}

func TestRunTurn_LoadErrorIsEnvironmentError(t *testing.T) {
	var out, errOut bytes.Buffer
	st := &fakeStore{loadErr: errors.New("boom")}
	code := runTurn(resolution{Selected: "p"}, st, "ping", true, env(&out, &errOut, &stubRenderer{}), factoryReturning(&fakeGateway{text: "x"}, nil))
	if code != EnvironmentError {
		t.Fatalf("code = %d, want %d (environment error)", code, EnvironmentError)
	}
	if !strings.HasPrefix(errOut.String(), "tellme: the runtime home is not usable") {
		t.Errorf("stderr = %q, want the environment class phrase", errOut.String())
	}
}

func TestRunTurn_ProviderFailure(t *testing.T) {
	var out, errOut bytes.Buffer
	fg := &fakeGateway{err: &llm.ProviderError{Provider: "p", Err: errors.New("boom")}}
	code := runTurn(resolution{Selected: "p"}, &fakeStore{}, "ping", true, env(&out, &errOut, &stubRenderer{}), factoryReturning(fg, nil))
	if code != ProviderError {
		t.Fatalf("code = %d, want %d (provider error)", code, ProviderError)
	}
	if !strings.HasPrefix(errOut.String(), "tellme: the provider request failed") {
		t.Errorf("stderr = %q, want the frozen class phrase", errOut.String())
	}
	if out.Len() != 0 {
		t.Errorf("stdout = %q, want empty", out.String())
	}
}

func TestRunTurn_UnsupportedFamilyIsProviderError(t *testing.T) {
	var out, errOut bytes.Buffer
	buildErr := &llm.ProviderError{Provider: "p", Err: errors.New(`unsupported provider family "gemini"`)}
	code := runTurn(resolution{Selected: "p"}, &fakeStore{}, "ping", true, env(&out, &errOut, &stubRenderer{}), factoryReturning(nil, buildErr))
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
