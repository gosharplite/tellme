package cli

import (
	"encoding/json"
	"fmt"
	"testing"
)

// Round 024 T042 / round 029 / round 033: the production registry factory offers
// exactly the seven agent tools — the three filesystem readers, the write pair
// (write_file, replace_text), the command tool, and the read-only skills listing
// tool (list_skills) — and never summarize_history or pipe_commands.

func TestNewToolRegistryOffersAgentTools(t *testing.T) {
	reg := newToolRegistry()
	got := map[string]bool{}
	for _, tl := range reg.Tools() {
		got[tl.Name()] = true
	}
	want := map[string]bool{
		"list_files":      true,
		"read_files":      true,
		"get_tree":        true,
		"write_file":      true,
		"replace_text":    true,
		"execute_command": true,
		"list_skills":     true,
	}
	if len(got) != len(want) {
		t.Fatalf("registry tools = %v; want exactly list_files, read_files, get_tree, write_file, replace_text, execute_command, list_skills", got)
	}
	for name := range want {
		if !got[name] {
			t.Errorf("registry is missing %q", name)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("registry offers unexpected tool %q", name)
		}
	}
}

// schemaWellFormed reports whether params is a JSON-object argument schema whose
// every `required` name is declared under `properties` (`required ⊆ properties`) —
// the well-formedness a strict provider (Vertex/Gemini) enforces (round 031, issue
// #64). An unparseable schema, a non-object schema, or a missing/non-object
// `properties` section is an error (FR-007); an empty `required` passes vacuously.
func schemaWellFormed(params json.RawMessage) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(params, &root); err != nil {
		return fmt.Errorf("required ⊆ properties: schema is not a JSON object: %w", err)
	}
	rawProps, ok := root["properties"]
	if !ok {
		return fmt.Errorf("required ⊆ properties: schema declares no properties section")
	}
	var props map[string]json.RawMessage
	if err := json.Unmarshal(rawProps, &props); err != nil {
		return fmt.Errorf("required ⊆ properties: properties is not an object: %w", err)
	}
	var required []string
	if rawReq, ok := root["required"]; ok {
		if err := json.Unmarshal(rawReq, &required); err != nil {
			return fmt.Errorf("required ⊆ properties: required is not an array: %w", err)
		}
	}
	for _, name := range required {
		if _, ok := props[name]; !ok {
			return fmt.Errorf("required ⊆ properties: required %q is not declared under properties", name)
		}
	}
	return nil
}

// TestAgentToolSchemasAreWellFormed is the round-031 recurrence gate (issue #64):
// it reads the NON-overridable production assembler agentTools() — NOT the
// newToolRegistry var, a DI seam a test could override, which would let the gate
// read a fake registry and pass vacuously (PR #65 ARCH-1) — and asserts every
// advertised tool satisfies `required ⊆ properties`. It must fail non-vacuously
// against the 5 violating tools until the shared schema builder is fixed.
func TestAgentToolSchemasAreWellFormed(t *testing.T) {
	if len(agentTools()) == 0 {
		t.Fatal("agentTools() is empty — the well-formedness gate would pass vacuously")
	}
	for _, tl := range agentTools() {
		if err := schemaWellFormed(tl.Parameters()); err != nil {
			t.Errorf("tool %q: %v", tl.Name(), err)
		}
	}
	// Belt-and-braces (PR #65 fold-review): the assembler the gate validates and the
	// registry the transport sends must expose the same tool set, so the two cannot
	// drift apart. Names only — the invariant above stays asserted on the assembler.
	assembler := map[string]bool{}
	for _, tl := range agentTools() {
		assembler[tl.Name()] = true
	}
	fromRegistry := map[string]bool{}
	for _, tl := range newToolRegistry().Tools() {
		fromRegistry[tl.Name()] = true
	}
	if len(assembler) != len(fromRegistry) {
		t.Errorf("assembler tools = %v; registry tools = %v; want the same set", assembler, fromRegistry)
	}
	for name := range assembler {
		if !fromRegistry[name] {
			t.Errorf("assembler tool %q is missing from the registry", name)
		}
	}
}

// TestSchemaWellFormedEdgeCases pins the gate's boundary behaviour the spec names
// (FR-007): a zero-`required` schema passes vacuously; a non-object/unparseable or
// property-less schema fails; a required name missing from properties fails.
func TestSchemaWellFormedEdgeCases(t *testing.T) {
	if err := schemaWellFormed(json.RawMessage(`{"type":"object","properties":{"a":{"type":"string"}},"required":[]}`)); err != nil {
		t.Errorf("a zero-required schema must pass vacuously: %v", err)
	}
	for _, bad := range []string{
		`[]`,                                 // not an object
		`"nope"`,                             // not an object
		`{invalid`,                           // unparseable
		`{"type":"object","required":["x"]}`, // no properties section
		`{"type":"object","properties":[],"required":[]}`,              // properties not an object
		`{"type":"object","properties":{"a":{}},"required":["a","b"]}`, // b not declared
	} {
		if err := schemaWellFormed(json.RawMessage(bad)); err == nil {
			t.Errorf("malformed schema %s must fail required ⊆ properties", bad)
		}
	}
}
