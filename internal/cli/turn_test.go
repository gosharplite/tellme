package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/config"
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

func TestRunTurn_PrintsRawAnswer(t *testing.T) {
	var out, errOut bytes.Buffer
	fg := &fakeGateway{text: "the answer"}
	code := runTurn(resolution{Selected: "p"}, "ping", true, env(&out, &errOut, &stubRenderer{}), factoryReturning(fg, nil))
	if code != Success {
		t.Fatalf("code = %d, want %d (success)", code, Success)
	}
	if fg.got.Prompt != "ping" {
		t.Errorf("gateway got prompt %q, want ping", fg.got.Prompt)
	}
	if out.String() != "the answer\n" {
		t.Errorf("stdout = %q, want %q", out.String(), "the answer\n")
	}
	if errOut.Len() != 0 {
		t.Errorf("stderr = %q, want empty", errOut.String())
	}
}

func TestRunTurn_ProviderFailure(t *testing.T) {
	var out, errOut bytes.Buffer
	fg := &fakeGateway{err: &llm.ProviderError{Provider: "p", Err: errors.New("boom")}}
	code := runTurn(resolution{Selected: "p"}, "ping", true, env(&out, &errOut, &stubRenderer{}), factoryReturning(fg, nil))
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
	code := runTurn(resolution{Selected: "p"}, "ping", true, env(&out, &errOut, &stubRenderer{}), factoryReturning(nil, buildErr))
	if code != ProviderError {
		t.Fatalf("code = %d, want %d (provider error)", code, ProviderError)
	}
	if !strings.Contains(errOut.String(), "unsupported provider family") {
		t.Errorf("stderr = %q, want the actionable reason", errOut.String())
	}
}
