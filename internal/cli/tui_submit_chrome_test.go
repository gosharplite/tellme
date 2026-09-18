package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
)

// TestRunTurn_EchoToggle (round-023 T008): with opts.echo the submitted prompt is
// written to stderr verbatim as its own block BEFORE the input-capture
// acknowledgement; with echo=false it is not echoed.
func TestRunTurn_EchoToggle(t *testing.T) {
	clk := func() time.Time { return time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC) }
	for _, echo := range []bool{true, false} {
		var buf bytes.Buffer
		lp := &fakeLoop{result: agentport.Result{Answer: "ANSWER", Usage: llm.Usage{Reported: true}}}
		e := runtimeEnv{stdout: &buf, stderr: &buf, renderer: &stubRenderer{out: "ANSWER"}, clock: clk}
		res := resolution{Selected: "p", Mode: "butler", MaxHistoryTokens: 1000000, Provider: config.Provider{Model: "deepseek-v4-flash"}}
		if code := runTurn(res, &fakeStore{}, "hello world", turnOptions{raw: true, chrome: true, echo: echo}, e, depsWithLoop(lp)); code != Success {
			t.Fatalf("code = %d, want success", code)
		}
		out := buf.String()
		echoIdx := strings.Index(out, "hello world")
		ackIdx := strings.Index(out, "Input captured")
		if echo {
			if echoIdx < 0 || ackIdx < 0 || echoIdx > ackIdx {
				t.Fatalf("echo=true: want the prompt echoed before the ack; out=%q", out)
			}
		} else if echoIdx >= 0 {
			t.Fatalf("echo=false: the prompt was echoed; out=%q", out)
		}
	}
}

// TestTUISubmitResumesChromeAndEchoes (round-023 T008/T010): an -i submission
// resumes the standard turn chrome on stderr AND echoes the submitted prompt
// before the acknowledgement — the wiring runTUIPrompt -> renderTurn{chrome:true,
// echo:true} -> runTurn.
func TestTUISubmitResumesChromeAndEchoes(t *testing.T) {
	clearAmbientOverrides(t)
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "configs"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "MODE: butler\nPERSON: p\nSELECTED_PROVIDER: m\nPROVIDERS:\n  m:\n    TYPE: deepseek\n    MODEL: deepseek-v4-flash\n    URL: https://example.invalid\n"
	if err := os.WriteFile(filepath.Join(home, "configs", "butler.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TELL_ME_HOME", home)

	dp := defaultTestDeps(func(d *deps.Dependencies) {
		d.NewGateway = func(config.Provider, string, string) (llm.Gateway, error) { return &fakeGateway{text: "ok"}, nil }
		d.NewHistoryStore = func(string) history.Store { return &fakeStore{} }
	})
	opts := Options{Deps: dp, Prompter: fakePrompter{result: "hi there", ok: true}}

	var out, errOut strings.Builder
	e := runtimeEnv{
		stdin:    strings.NewReader(""),
		stdout:   &out,
		stderr:   &errOut,
		isTTY:    func(any) bool { return true },
		renderer: &stubRenderer{out: "ok"},
		clock:    func() time.Time { return time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC) },
	}
	if code := run([]string{"-i"}, "dev", opts, e); code != Success {
		t.Fatalf("code = %d, want success; stderr=%q", code, errOut.String())
	}
	errs := errOut.String()
	echoIdx := strings.Index(errs, "hi there")
	ackIdx := strings.Index(errs, "Input captured")
	if echoIdx < 0 || ackIdx < 0 || echoIdx > ackIdx {
		t.Fatalf("the -i submit did not echo the prompt before the ack; stderr=%q", errs)
	}
	if !strings.Contains(errs, strings.Repeat("─", 80)) {
		t.Fatalf("the -i submit did not render the turn rule; stderr=%q", errs)
	}
	if !strings.Contains(errs, "╭─⠿ Turn 1 - butler") {
		t.Fatalf("the -i submit did not render the turn header; stderr=%q", errs)
	}
	if strings.Contains(out.String(), "hi there") {
		t.Fatalf("the echoed prompt leaked onto stdout; stdout=%q", out.String())
	}
}
