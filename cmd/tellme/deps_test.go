package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/app/deps"
)

// Round 044 (relocated from internal/cli by ADR 0013): the assembler gate and the
// registry-set assertion now live next to the (relocated) production assembler
// agentTools(). No stream assertions (cli.Run hard-binds os.Stdin/Stdout/Stderr).

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
		t.Fatalf("registry tools = %v; want exactly the seven agent tools", got)
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

// TestNewTUIRegistryIsTheReaderTriplet pins round-044 fix-1: the `-i` suggestion
// source consumes the three-reader registry, NOT the seven-tool agent registry.
func TestNewTUIRegistryIsTheReaderTriplet(t *testing.T) {
	got := newTUIRegistry()
	if len(got.Tools()) != 3 {
		t.Fatalf("TUI registry tools = %d, want 3", len(got.Tools()))
	}
	for _, tl := range got.Tools() {
		switch tl.Name() {
		case "list_files", "read_files", "get_tree":
		default:
			t.Errorf("unexpected TUI registry tool %q", tl.Name())
		}
	}
}

// schemaWellFormed reports whether params is a JSON-object argument schema whose
// every `required` name is declared under `properties` (`required ⊆ properties`).
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

// TestAgentToolSchemasAreWellFormed is the round-031 recurrence gate (issue #64),
// relocated with the assembler: it iterates the NON-overridable assembler
// agentTools() and asserts `required ⊆ properties` for every tool (PR #65 ARCH-1).
func TestAgentToolSchemasAreWellFormed(t *testing.T) {
	if len(agentTools()) == 0 {
		t.Fatal("agentTools() is empty — the well-formedness gate would pass vacuously")
	}
	for _, tl := range agentTools() {
		if err := schemaWellFormed(tl.Parameters()); err != nil {
			t.Errorf("tool %q: %v", tl.Name(), err)
		}
	}
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

// TestSchemaWellFormedEdgeCases pins the gate's boundary behaviour (FR-007).
func TestSchemaWellFormedEdgeCases(t *testing.T) {
	if err := schemaWellFormed(json.RawMessage(`{"type":"object","properties":{"a":{"type":"string"}},"required":[]}`)); err != nil {
		t.Errorf("a zero-required schema must pass vacuously: %v", err)
	}
	for _, bad := range []string{
		`[]`,
		`"nope"`,
		`{invalid`,
		`{"type":"object","required":["x"]}`,
		`{"type":"object","properties":[],"required":[]}`,
		`{"type":"object","properties":{"a":{}},"required":["a","b"]}`,
	} {
		if err := schemaWellFormed(json.RawMessage(bad)); err == nil {
			t.Errorf("malformed schema %s must fail required ⊆ properties", bad)
		}
	}
}

// TestBuildDepsIsFullyWired is the composition-root smoke (round-044 fix-8): every
// injected seam is non-nil (deps.Dependencies.Validate), and UserHomeDir is wired
// to a non-nil resolver (RF-2).
func TestBuildDepsIsFullyWired(t *testing.T) {
	d := buildDeps()
	if err := d.Validate(); err != nil {
		t.Fatalf("buildDeps() left an unbound seam: %v", err)
	}
	if _, err := d.UserHomeDir(); err != nil {
		// A resolvable home is not required, but the resolver must be callable and
		// non-nil (RF-2: GlobalPromptTracker.destPath calls it unguarded).
		t.Logf("UserHomeDir() error (acceptable in a sandbox): %v", err)
	}
	opts := buildOptions()
	if opts.Deps.NewGateway == nil {
		t.Fatal("buildOptions() did not carry the deps")
	}
}

// TestDepsValidateCatchesMissingSeam pins the F-5 validator: a zero
// Dependencies (or one missing a seam) fails with a message naming the seam,
// rather than a nil-func deref deep inside internal/cli.
func TestDepsValidateCatchesMissingSeam(t *testing.T) {
	var zero deps.Dependencies
	err := zero.Validate()
	if err == nil {
		t.Fatal("Validate() = nil for a zero Dependencies, want an unbound-seam error")
	}
	if !strings.Contains(err.Error(), "is not wired") {
		t.Errorf("Validate() error = %q, want it to name the unwired seam", err)
	}
}
