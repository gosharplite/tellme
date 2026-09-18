package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
)

// TestRunTurnTeesChromeIntoTurnsLog pins the round-053 (ADR 0022) writer seam —
// fold F-53-1. The per-call renderer's chrome (the status frame + tail) must
// reach the injected turns-log store on the prompt path (positive), and must NOT
// on a run with no workspace (negative) — so deleting the tee reds this test
// (the review's witness (b) made permanent).
func TestRunTurnTeesChromeIntoTurnsLog(t *testing.T) {
	t.Parallel()

	captureStore := func(got **fakeTurnsLogStore) func(*deps.Dependencies) {
		return func(d *deps.Dependencies) {
			d.NewTurnsLogStore = func(string) history.TurnsLogStore {
				*got = &fakeTurnsLogStore{}
				return *got
			}
		}
	}

	t.Run("a prompt turn tees the chrome", func(t *testing.T) {
		t.Parallel()
		var got *fakeTurnsLogStore
		var out, errOut bytes.Buffer
		lp := &fakeLoop{result: agentport.Result{Answer: "ok", Usage: llm.Usage{Reported: true}}}
		res := resolution{Selected: "p", Mode: "butler", Workspace: t.TempDir(), MaxHistoryTokens: 1000000, Provider: config.Provider{Model: "deepseek-v4-flash"}}
		code := runTurn(res, &fakeStore{}, "ping", turnOptions{raw: true, chrome: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp, captureStore(&got)))
		if code != Success {
			t.Fatalf("code = %d, want success", code)
		}
		if got == nil {
			t.Fatal("the turns-log store was never constructed (the tee is missing)")
		}
		logged := got.buf.String()
		if !strings.Contains(logged, "<turn-opening") || !strings.Contains(logged, "<payload") {
			t.Fatalf("turns.log %q must carry the chrome (the status frame + payload line)", logged)
		}
		// The chrome still reaches stderr (stderr is the multiwriter's first sink).
		if !strings.Contains(errOut.String(), "<turn-opening") {
			t.Fatalf("the chrome must still reach stderr; stderr=%q", errOut.String())
		}
	})

	t.Run("a run with no workspace tees nothing", func(t *testing.T) {
		t.Parallel()
		var got *fakeTurnsLogStore
		var out, errOut bytes.Buffer
		lp := &fakeLoop{result: agentport.Result{Answer: "ok", Usage: llm.Usage{Reported: true}}}
		res := resolution{Selected: "p", Mode: "butler", Workspace: "", MaxHistoryTokens: 1000000, Provider: config.Provider{Model: "deepseek-v4-flash"}}
		code := runTurn(res, &fakeStore{}, "ping", turnOptions{raw: true, chrome: true}, env(&out, &errOut, &stubRenderer{}), depsWithLoop(lp, captureStore(&got)))
		if code != Success {
			t.Fatalf("code = %d, want success", code)
		}
		if got != nil && got.buf.Len() != 0 {
			t.Fatalf("an empty workspace must tee nothing; got %q", got.buf.String())
		}
	})
}
