package tools

import (
	"context"
	"encoding/json"
	"testing"
)

// T024 (round 008) — tool registry dispatch unit tests (research Decision 1).

// stubTool is a minimal read-only Tool for registry tests.
type stubTool struct {
	name string
}

func (s stubTool) Name() string                { return s.name }
func (s stubTool) Description() string         { return "stub tool" }
func (s stubTool) Parameters() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }
func (s stubTool) Contract() ToolContract      { return ToolContract{} }
func (s stubTool) Execute(context.Context, string, ByteBudget) (string, error) {
	return "", nil
}

func TestRegistryLookupAndOrder(t *testing.T) {
	r := NewRegistry(stubTool{name: "list_files"}, stubTool{name: "read_files"})
	if got, ok := r.Lookup("read_files"); !ok || got.Name() != "read_files" {
		t.Fatalf("Lookup(read_files) = (%v, %v), want (read_files, true)", got, ok)
	}
	if _, ok := r.Lookup("nope"); ok {
		t.Error("Lookup(nope) ok = true, want false")
	}
	ts := r.Tools()
	if len(ts) != 2 || ts[0].Name() != "list_files" || ts[1].Name() != "read_files" {
		t.Fatalf("Tools() = %+v, want offer order [list_files read_files]", ts)
	}
}

func TestRegistryEmpty(t *testing.T) {
	r := NewRegistry()
	if len(r.Tools()) != 0 {
		t.Errorf("Tools() = %+v, want empty", r.Tools())
	}
	if _, ok := r.Lookup("x"); ok {
		t.Error("Lookup on an empty registry ok = true, want false")
	}
}
