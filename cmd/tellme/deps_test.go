package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/cli"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	infrallm "github.com/gosharplite/tellme/internal/infrastructure/llm"
	"github.com/gosharplite/tellme/internal/infrastructure/llm/gemini"
	infratools "github.com/gosharplite/tellme/internal/infrastructure/tools"
)

// Round 044 (relocated from internal/cli by ADR 0013): the assembler gate and the
// registry-set assertion now live next to the (relocated) production assembler
// agentTools(). No stream assertions (cli.Run hard-binds os.Stdin/Stdout/Stderr).

func TestNewToolRegistryOffersAgentTools(t *testing.T) {
	reg := newToolRegistry(deps.ToolSetSpec{})
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
	for _, tl := range newToolRegistry(deps.ToolSetSpec{}).Tools() {
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
	// Round 048 (F-4 / ADR 0017): the presentation port is an interface seam that
	// deps.Dependencies.Validate's func-kind predicate cannot see, so the
	// composition root asserts it explicitly — a dropped Prompter would otherwise
	// compile and surface only as the -i path failing.
	if opts.Prompter == nil {
		t.Fatal("buildOptions() left the interactive prompt port (Options.Prompter) unwired")
	}
	if err := opts.Validate(); err != nil {
		t.Fatalf("buildOptions() produced invalid Options: %v", err)
	}
}

// TestOptionsValidateCatchesUnwiredPrompter pins the round-048 F-4 validator: an
// Options with a wired Dependencies but a nil Prompter fails loudly, naming the
// port, rather than dereferencing nil on the -i path.
func TestOptionsValidateCatchesUnwiredPrompter(t *testing.T) {
	opts := cli.Options{Deps: buildDeps()} // Prompter deliberately nil
	err := opts.Validate()
	if err == nil {
		t.Fatal("Options.Validate() = nil with a nil Prompter, want a wiring error")
	}
	if !strings.Contains(err.Error(), "Prompter") {
		t.Errorf("Options.Validate() error = %q, want it to name the Prompter seam", err)
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

// TestAgentToolDeclarationsSurviveTheGeminiProjection closes round-061 RF-061-4
// durably: every NATIVE tool declaration must be *semantically identical* after
// the closed-wire projection — i.e. nothing tellme itself declares is trimmed.
// (Semantic, not byte, equality: the projection re-serialises, so key order
// changes.)
func TestAgentToolDeclarationsSurviveTheGeminiProjection(t *testing.T) {
	for _, tl := range agentTools() {
		raw := tl.Parameters()
		if len(raw) == 0 {
			continue
		}
		var before, after any
		if err := json.Unmarshal(raw, &before); err != nil {
			t.Fatalf("tool %q: decode declaration: %v", tl.Name(), err)
		}
		projected := gemini.ProjectSchema(raw)
		if err := json.Unmarshal(projected, &after); err != nil {
			t.Fatalf("tool %q: decode projected declaration: %v", tl.Name(), err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Errorf("tool %q: the Gemini projection changed a native declaration\ngot  %s\nfrom %s", tl.Name(), projected, raw)
		}
	}
}

// TestCompositionResolvesTheFamilyAwareImageCeiling pins round-063 (ADR 0033 D4)
// at the composition root: the provider family (single-owned by infrallm.Family)
// resolves to a STRICTER Gemini ceiling than the OpenAI-compatible one, and the
// vision variant of the registry offers `read_image` while the base variant does
// not (the capability gate is unchanged — Q1 → A).
func TestCompositionResolvesTheFamilyAwareImageCeiling(t *testing.T) {
	geminiCeiling := infratools.ImageCeilingForFamily(infrallm.Family("gemini"))
	openAICeiling := infratools.ImageCeilingForFamily(infrallm.Family("deepseek"))
	if geminiCeiling >= openAICeiling {
		t.Fatalf("the Gemini image ceiling (%d) must be stricter than the OpenAI-compatible one (%d)", geminiCeiling, openAICeiling)
	}
	has := func(reg domaintools.Registry, name string) bool {
		for _, tl := range reg.Tools() {
			if tl.Name() == name {
				return true
			}
		}
		return false
	}
	if !has(newToolRegistry(deps.ToolSetSpec{Vision: true, ProviderType: "gemini"}), "read_image") {
		t.Error("a vision-enabled provider must be offered read_image")
	}
	if has(newToolRegistry(deps.ToolSetSpec{ProviderType: "gemini"}), "read_image") {
		t.Error("a provider without vision must NOT be offered read_image")
	}
}

// TestResolveImageCeilingPinsTheFamilyAwareConsumption pins the CEILING
// CONSUMPTION seam (round 069 / ADR 0039; PR #141 fold F2, closing RF-069-5):
// `resolveImageCeiling` is the one place the agent registry path turns a
// provider label into the family-aware inline value the `read_image` tool
// enforces, so a regression that dropped or collapsed the family resolution
// (e.g. always the OpenAI-compatible ceiling) reds HERE even though the
// `reading-a-local-image` E2E cannot distinguish the two families (its oversize
// fixture exceeds BOTH ceilings). The owner tables stay pinned by
// TestCompositionResolvesTheFamilyAwareImageCeiling; this pins their wiring.
func TestResolveImageCeilingPinsTheFamilyAwareConsumption(t *testing.T) {
	cases := []struct {
		providerType string
		want         int
	}{
		{"gemini", 14 << 20},   // Gemini/Vertex — the stricter derived ceiling
		{"google", 14 << 20},   // the gemini family's other label
		{"deepseek", 32 << 20}, // OpenAI-compatible family
		{"", 32 << 20},         // an unclassified label defaults to the OpenAI-compatible ceiling
	}
	for _, c := range cases {
		got := resolveImageCeiling(deps.ToolSetSpec{Vision: true, ProviderType: c.providerType})
		if got != c.want {
			t.Errorf("resolveImageCeiling(%q) = %d, want %d", c.providerType, got, c.want)
		}
	}
	if resolveImageCeiling(deps.ToolSetSpec{ProviderType: "gemini"}) == resolveImageCeiling(deps.ToolSetSpec{ProviderType: "deepseek"}) {
		t.Error("the family-aware resolution must differ across families (Gemini stricter than OpenAI-compatible)")
	}
}
