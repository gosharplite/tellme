package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Round-034 T023 unit pin for the persistence invariant (ADR 0005 D4 / FR-010b):
// on a COMPLETED turn the persisted record set is the `Reported` subset of
// `result.Calls`, written by exactly ONE AppendBatch; on every error exit, and on
// a completed turn whose final call reports no usage, NOTHING is persisted. The
// per-call tail is display-only — it never adds a per-call append.

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

// withUsageStore swaps the package usage-store factory for the test's lifetime.
func withUsageStore(t *testing.T, us history.UsageStore) {
	t.Helper()
	prev := newUsageStore
	newUsageStore = func(string) history.UsageStore { return us }
	t.Cleanup(func() { newUsageStore = prev })
}

// withTestTooling swaps the tool registry and tool-usage store for hermetic runs.
func withTestTooling(t *testing.T) {
	t.Helper()
	prevReg := newToolRegistry
	newToolRegistry = func() domaintools.Registry {
		return domaintools.NewRegistry(noopTool{name: "read_files"})
	}
	prevUsage := newToolUsageStore
	newToolUsageStore = func() history.ToolUsageStore { return noopUsageStore{} }
	t.Cleanup(func() {
		newToolRegistry = prevReg
		newToolUsageStore = prevUsage
	})
}

func TestRunTurn_PersistenceInvariant(t *testing.T) {
	withTestTooling(t)
	toolCall := llm.ToolCall{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["n.txt"],"reason":"r"}`}
	res := resolution{Selected: "p", Mode: "butler", MaxHistoryTokens: 1000000, Workspace: t.TempDir(), Provider: config.Provider{Model: "m"}}

	t.Run("completed turn persists the Reported subset in one batch", func(t *testing.T) {
		var out, errOut bytes.Buffer
		us := &capturingUsageStore{}
		withUsageStore(t, us)
		fg := &fakeGateway{script: []llm.Response{
			{ToolCalls: []llm.ToolCall{toolCall}, Usage: llm.Usage{Reported: false}},
			{Text: "done", Usage: llm.Usage{Reported: true, PromptTokens: 42, CachedTokens: 4, CompletionTokens: 2}},
		}}
		if code := runTurn(res, &fakeStore{}, "ping", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), factoryReturning(fg, nil)); code != Success {
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
		withUsageStore(t, us)
		fg := &fakeGateway{err: &llm.ProviderError{Provider: "p", Err: errors.New("boom")}}
		if code := runTurn(res, &fakeStore{}, "ping", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), factoryReturning(fg, nil)); code != ProviderError {
			t.Fatalf("code = %d, want the provider error code", code)
		}
		if us.appendCalls != 0 {
			t.Errorf("AppendBatch called %d times on an error exit, want 0", us.appendCalls)
		}
	})

	t.Run("final call with no usage persists nothing", func(t *testing.T) {
		var out, errOut bytes.Buffer
		us := &capturingUsageStore{}
		withUsageStore(t, us)
		fg := &fakeGateway{text: "done", usage: llm.Usage{Reported: false}}
		if code := runTurn(res, &fakeStore{}, "ping", turnOptions{raw: true}, env(&out, &errOut, &stubRenderer{}), factoryReturning(fg, nil)); code != Success {
			t.Fatalf("code = %d, want success", code)
		}
		if us.appendCalls != 0 {
			t.Errorf("AppendBatch called %d times when the final call reports no usage, want 0", us.appendCalls)
		}
	})
}
