package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/history"
)

// TestDispatchReportingPrecedence pins the round-053 (ADR 0022, fold F-53-4)
// offline-reporting order: `-d` → `-l` → `-t` → `--tool-usage`. With both `-l`
// and `-t` set, the listing wins; `-t` alone prints the turn log.
func TestDispatchReportingPrecedence(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	historyStore := func(string) history.Store {
		return &fakeStore{entries: []history.Entry{{Prompt: "q", Answer: "a"}}}
	}
	turnsStore := func(string) history.TurnsLogStore {
		ts := &fakeTurnsLogStore{}
		ts.buf.WriteString("TURNS\n")
		return ts
	}

	t.Run("-l wins over -t", func(t *testing.T) {
		var out, errOut bytes.Buffer
		f := &flags{list: 1, listSet: true, turns: true}
		code, handled := dispatchReporting(f, home, runtimeEnv{stdout: &out, stderr: &errOut}, historyStore, turnsStore, func() int { return Success })
		if !handled || code != Success {
			t.Fatalf("dispatched=%v code=%d, want handled success", handled, code)
		}
		if out.String() != "assistant: a\n" {
			t.Fatalf("stdout = %q, want the -l listing (‑l precedes ‑t)", out.String())
		}
		if strings.Contains(out.String(), "TURNS") {
			t.Fatalf("stdout = %q, the turns log must NOT print when -l is also set", out.String())
		}
	})

	t.Run("-t alone prints the turn log", func(t *testing.T) {
		var out, errOut bytes.Buffer
		f := &flags{turns: true}
		code, handled := dispatchReporting(f, home, runtimeEnv{stdout: &out, stderr: &errOut}, historyStore, turnsStore, func() int { return Success })
		if !handled || code != Success {
			t.Fatalf("dispatched=%v code=%d, want handled success", handled, code)
		}
		if out.String() != "TURNS\n" {
			t.Fatalf("stdout = %q, want the turn log", out.String())
		}
	})

	t.Run("--tool-usage still last", func(t *testing.T) {
		var out, errOut bytes.Buffer
		f := &flags{turns: true, toolUsage: true}
		code, handled := dispatchReporting(f, home, runtimeEnv{stdout: &out, stderr: &errOut}, historyStore, turnsStore, func() int {
			_, _ = out.WriteString("TOOLUSAGE\n")
			return Success
		})
		if !handled || code != Success {
			t.Fatalf("dispatched=%v code=%d, want handled success", handled, code)
		}
		if out.String() != "TURNS\n" {
			t.Fatalf("stdout = %q, want the turn log (‑t precedes --tool-usage)", out.String())
		}
	})
}
