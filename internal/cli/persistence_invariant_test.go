package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Round-034 T023 unit pin for the persistence invariant (ADR 0005 D4 / FR-010b):
// on a COMPLETED turn the persisted record set is the `Reported` subset of
// `result.Calls`, written by exactly ONE AppendBatch; on every error exit, and on
// a completed turn whose final call reports no usage, NOTHING is persisted. The
// per-call tail is display-only — it never adds a per-call append.
//
// Round 044: the package-level factory vars are gone; each subtest builds an
// injected deps.Dependencies (the RF-1 fixture) instead of swapping globals.

// capturingUsageStore is a history.UsageStore double capturing AppendBatch.
type capturingUsageStore struct {
	batches     [][]history.UsageRecord
	appendCalls int
}

func (s *capturingUsageStore) Append(history.UsageRecord) error { return nil }
func (s *capturingUsageStore) AppendBatch(recs []history.UsageRecord) error {
	s.appendCalls++
	s.batches = append(s.batches, recs)
	return nil
}
func (s *capturingUsageStore) Load() ([]history.UsageRecord, error) { return nil, nil }
func (s *capturingUsageStore) Totals() (history.UsageSummary, error) {
	return history.UsageSummary{}, nil
}
func (s *capturingUsageStore) Archive() error { return nil }

// noopUsageStore is a history.ToolUsageStore double that records nothing (so a
// unit test never writes the operator's real ~/.tellme log).
type noopUsageStore struct{}

func (noopUsageStore) Record(string, history.ToolOutcome) error { return nil }
func (noopUsageStore) Aggregate() (map[string]history.ToolUsageCounts, error) {
	return map[string]history.ToolUsageCounts{}, nil
}

// noopTool is a harmless domaintools.Tool so a unit test tool round never touches
// the real filesystem.
type noopTool struct{ name string }

func (t noopTool) Name() string        { return t.name }
func (t noopTool) Description() string { return "noop test tool" }
func (t noopTool) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (t noopTool) Contract() domaintools.ToolContract { return domaintools.ToolContract{} }
func (t noopTool) Execute(context.Context, string, domaintools.ByteBudget) (string, error) {
	return "ok", nil
}

// persistenceDeps builds the deps for the persistence test: the noop tool-usage
// store, the given usage store, and the given loop double (round 050 / Q4 → A:
// the loop is the in-package fake; the CLI's persistence of its Result is what is
// pinned).
func persistenceDeps(us history.UsageStore, lp *fakeLoop) deps.Dependencies {
	return depsWithLoop(lp, func(d *deps.Dependencies) {
		d.NewUsageStore = func(string) history.UsageStore { return us }
		d.NewToolUsageStore = func(func() (string, error)) history.ToolUsageStore { return noopUsageStore{} }
	})
}

func TestRunTurn_PersistenceInvariant(t *testing.T) {
	res := resolution{Selected: "p", Mode: "butler", MaxHistoryTokens: 1000000, Workspace: t.TempDir(), Provider: config.Provider{Model: "m"}}

	t.Run("completed turn persists the Reported subset in one batch", func(t *testing.T) {
		var out, errOut bytes.Buffer
		us := &capturingUsageStore{}
		lp := &fakeLoop{result: agentport.Result{
			Answer: "done",
			Usage:  llm.Usage{Reported: true, PromptTokens: 42, CachedTokens: 4, CompletionTokens: 2},
			Calls: []llm.Usage{
				{Reported: false},
				{Reported: true, PromptTokens: 42, CachedTokens: 4, CompletionTokens: 2},
			},
		}}
		if code := runTurn(res, &fakeStore{}, "ping", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), persistenceDeps(us, lp)); code != Success {
			t.Fatalf("code = %d, want success", code)
		}
		if us.appendCalls != 1 {
			t.Fatalf("AppendBatch calls = %d, want exactly 1 (one batch per turn)", us.appendCalls)
		}
		if got := len(us.batches[0]); got != 1 {
			t.Errorf("persisted %d records, want 1 (only the Reported call)", got)
		}
	})

	t.Run("error exit persists nothing", func(t *testing.T) {
		var out, errOut bytes.Buffer
		us := &capturingUsageStore{}
		lp := &fakeLoop{runErr: &llm.ProviderError{Provider: "p", Err: errors.New("boom")}}
		if code := runTurn(res, &fakeStore{}, "ping", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), persistenceDeps(us, lp)); code != ProviderError {
			t.Fatalf("code = %d, want the provider error code", code)
		}
		if us.appendCalls != 0 {
			t.Errorf("AppendBatch called %d times on an error exit, want 0", us.appendCalls)
		}
	})

	t.Run("final call with no usage persists nothing", func(t *testing.T) {
		var out, errOut bytes.Buffer
		us := &capturingUsageStore{}
		lp := &fakeLoop{result: agentport.Result{Answer: "done", Usage: llm.Usage{Reported: false}, Calls: []llm.Usage{{Reported: false}}}}
		if code := runTurn(res, &fakeStore{}, "ping", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), persistenceDeps(us, lp)); code != Success {
			t.Fatalf("code = %d, want success", code)
		}
		if us.appendCalls != 0 {
			t.Errorf("AppendBatch called %d times when the final call reports no usage, want 0", us.appendCalls)
		}
	})
}
