package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/render"
)

// dispatchDeps is the shared double set the dispatch-precedence tests use.
type dispatchDeps struct {
	historyStore func(string) history.Store
	turnsStore   func(string) history.TurnsLogStore
	listing      func() render.Listing
	usageReport  func() int
}

func newDispatchDeps() dispatchDeps {
	return dispatchDeps{
		historyStore: func(string) history.Store {
			return &fakeStore{entries: []history.Entry{{Prompt: "q", Answer: "a"}}}
		},
		turnsStore: func(string) history.TurnsLogStore {
			ts := &fakeTurnsLogStore{}
			ts.buf.WriteString("TURNS\n")
			return ts
		},
		listing:     func() render.Listing { return &fakeListing{} },
		usageReport: func() int { return Success },
	}
}

// TestDispatchReportingPrecedence pins the round-053 (ADR 0022, fold F-53-4)
// offline-reporting order: `-d` → `-l` → `-t` → `--tool-usage`. With both `-l`
// and `-t` set, the listing wins; `-t` alone prints the turn log.
func TestDispatchReportingPrecedence(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	d := newDispatchDeps()

	t.Run("-l wins over -t", func(t *testing.T) {
		var out, errOut bytes.Buffer
		f := &flags{list: 1, listSet: true, turns: true}
		code, handled := dispatchReporting(f, home, runtimeEnv{stdout: &out, stderr: &errOut}, false, d.historyStore, d.turnsStore, d.listing, d.usageReport)
		if !handled || code != Success {
			t.Fatalf("dispatched=%v code=%d, want handled success", handled, code)
		}
		if out.String() != "[MODEL]\na\n\n" {
			t.Fatalf("stdout = %q, want the -l listing (‑l precedes ‑t)", out.String())
		}
		if strings.Contains(out.String(), "TURNS") {
			t.Fatalf("stdout = %q, the turns log must NOT print when -l is also set", out.String())
		}
	})

	t.Run("-t alone prints the turn log", func(t *testing.T) {
		var out, errOut bytes.Buffer
		f := &flags{turns: true}
		code, handled := dispatchReporting(f, home, runtimeEnv{stdout: &out, stderr: &errOut}, false, d.historyStore, d.turnsStore, d.listing, d.usageReport)
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
		code, handled := dispatchReporting(f, home, runtimeEnv{stdout: &out, stderr: &errOut}, false, d.historyStore, d.turnsStore, d.listing, func() int {
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

// TestDispatchReportingBack pins the round-081 (ADR 0053) `-b` tiers: the
// `-l`×`-b` compose (list then roll back) and the `-b 0` refusal.
func TestDispatchReportingBack(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	d := newDispatchDeps()

	t.Run("-l composes with -b (list then roll back)", func(t *testing.T) {
		var out, errOut bytes.Buffer
		st := &fakeStore{entries: []history.Entry{{Prompt: "q1", Answer: "a1"}, {Prompt: "q2", Answer: "a2"}}}
		shared := func(string) history.Store { return st }
		f := &flags{list: 1, listSet: true, back: 1, backSet: true}
		code, handled := dispatchReporting(f, home, runtimeEnv{stdout: &out, stderr: &errOut}, false, shared, d.turnsStore, d.listing, d.usageReport)
		if !handled || code != Success {
			t.Fatalf("dispatched=%v code=%d, want handled success", handled, code)
		}
		if !strings.Contains(out.String(), "[MODEL]\na2\n") {
			t.Fatalf("stdout = %q, want the -l listing printed before the rollback", out.String())
		}
		if !strings.Contains(out.String(), "Rolled back 1 turn") {
			t.Fatalf("stdout = %q, want the rollback confirmation after the listing", out.String())
		}
		if len(st.rolled) != 1 || st.rolled[0] != 1 {
			t.Fatalf("rollback calls = %v, want [1] (the -l×-b compose must roll back)", st.rolled)
		}
	})

	t.Run("-b 0 is a usage error", func(t *testing.T) {
		var out, errOut bytes.Buffer
		f := &flags{back: 0, backSet: true}
		code, handled := dispatchReporting(f, home, runtimeEnv{stdout: &out, stderr: &errOut}, false, d.historyStore, d.turnsStore, d.listing, d.usageReport)
		if !handled || code != UsageError {
			t.Fatalf("dispatched=%v code=%d, want handled usage error", handled, code)
		}
		if !strings.Contains(errOut.String(), "the command-line usage is invalid") {
			t.Fatalf("stderr = %q, want the usage class phrase", errOut.String())
		}
	})
}
