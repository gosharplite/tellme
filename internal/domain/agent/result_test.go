package agent_test

// Round 049 (R5.3 of #92; ADR 0018) behaviour-preservation pins for the
// relocated crossing contracts (`Result`, `ErrIncomplete`, `ToolDefs`). These
// pin the contracts' observable behaviour so the pure relocation is proven
// non-behavioural: the field set, the error classification, and the wire-def
// projection (incl. registration order) are unchanged.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// pinTool is a minimal canned tool for the ToolDefs projection pin.
type pinTool struct {
	name string
	desc string
}

func (p pinTool) Name() string                 { return p.name }
func (p pinTool) Description() string          { return p.desc }
func (p pinTool) Parameters() json.RawMessage  { return json.RawMessage(`{"type":"object"}`) }
func (p pinTool) Contract() tools.ToolContract { return tools.ToolContract{} }
func (p pinTool) Execute(context.Context, string, tools.ByteBudget) (string, error) {
	return "", nil
}

// TestResultFieldSet pins the relocated `Result` shape (compile-checked via the
// literal) so a field rename/removal/drift is a build failure for every
// consumer.
func TestResultFieldSet(t *testing.T) {
	r := agentport.Result{
		Answer: "a",
		Steps:  []history.Step{{Tool: "read_files"}},
		Usage:  llm.Usage{PromptTokens: 1, CompletionTokens: 2},
		Calls:  []llm.Usage{{PromptTokens: 1}},
	}
	if r.Answer != "a" || len(r.Steps) != 1 || r.Usage.PromptTokens != 1 || len(r.Calls) != 1 {
		t.Fatalf("Result field set drifted: %+v", r)
	}
}

// TestErrIncomplete pins the error contract: the reason alone, the reason plus
// the wrapped cause, and the errors.As round-trip the CLI relies on.
func TestErrIncomplete(t *testing.T) {
	plain := &agentport.ErrIncomplete{Reason: "the tool-loop bound was reached"}
	if got := plain.Error(); got != "the tool-loop bound was reached" {
		t.Errorf("Error() = %q, want the reason alone", got)
	}
	if plain.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", plain.Unwrap())
	}

	cause := errors.New("boom")
	wrapped := &agentport.ErrIncomplete{Reason: "no tools are registered", Err: cause}
	if got := wrapped.Error(); got != "no tools are registered: boom" {
		t.Errorf("Error() = %q, want reason + cause", got)
	}
	if !errors.Is(wrapped, cause) {
		t.Errorf("errors.Is(wrapped, cause) = false, want true")
	}

	var target *agentport.ErrIncomplete
	if !errors.As(error(wrapped), &target) || target != wrapped {
		t.Errorf("errors.As round-trip failed")
	}
}

// TestToolDefs pins the projection: a nil registry projects to nil, and a
// registry projects one llm.ToolDef per tool in REGISTRATION ORDER, mapping
// Name/Description/Parameters verbatim (the round-011 RF-1 invariant: the
// pre-flight estimate counts exactly what the loop sends).
func TestToolDefs(t *testing.T) {
	if defs := agentport.ToolDefs(nil); defs != nil {
		t.Fatalf("ToolDefs(nil) = %v, want nil", defs)
	}

	reg := tools.NewRegistry(
		pinTool{name: "read_files", desc: "read"},
		pinTool{name: "write_file", desc: "write"},
	)
	defs := agentport.ToolDefs(reg)
	if len(defs) != 2 {
		t.Fatalf("ToolDefs len = %d, want 2", len(defs))
	}
	if defs[0].Name != "read_files" || defs[1].Name != "write_file" {
		t.Errorf("ToolDefs order = [%q, %q], want registration order", defs[0].Name, defs[1].Name)
	}
	if defs[0].Description != "read" {
		t.Errorf("ToolDefs[0].Description = %q, want read", defs[0].Description)
	}
	if string(defs[0].Parameters) != `{"type":"object"}` {
		t.Errorf("ToolDefs[0].Parameters = %s", defs[0].Parameters)
	}
}
