package gemini

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// recordedEnvelope is the declaration an MCP tool offers for a tool whose
// advertised schema carries the GitHub server's shape: two annotated arguments
// (`owner`/`repo` with `x-mcp-header`), a `title`, a `default`, a `minItems`/
// `maxItems` array, and an `anyOf` union — i.e. both an unsupported vendor
// extension AND unsupported standard keywords.
const recordedEnvelope = `{
  "type": "object",
  "properties": {
    "reason": {"type": "string", "description": "why"},
    "MCP_PAYLOAD": {
      "type": "object",
      "title": "params",
      "properties": {
        "owner": {"type": "string", "description": "Repository owner", "x-mcp-header": "owner"},
        "repo": {"type": "string", "description": "Repository name", "x-mcp-header": "repo"},
        "body": {"type": "string", "description": "The text", "default": "n/a"},
        "files": {"type": "array", "description": "A string or an array", "minItems": 1, "maxItems": 100,
                  "items": {"type": "string"}},
        "files2": {"description": "union", "anyOf": [{"type": "string"}, {"type": "array", "items": {"type": "string"}}]},
        "kind": {"type": "string", "enum": ["a", "b"], "uniques": 1, "const": "a"},
        "nested": {"type": "object", "description": "d", "properties": {"deep": {"type": "string", "$ref": "#/x", "readOnly": true}}}
      },
      "required": ["owner", "repo"]
    }
  },
  "required": ["reason"]
}`

// forbiddenKeys are the keywords the live probe measured as REJECTED (ADR 0031):
// any of them on the wire fails the whole request.
var forbiddenKeys = []string{
	"x-mcp-header", "$schema", "$ref", "$defs", "definitions", "const", "examples",
	"deprecated", "readOnly", "writeOnly", "multipleOf", "uniqueItems", "anyOf",
	"uniques",
}

// TestProjectToolDeclarations_NoUnsupportedKeywordReachesTheWire is the round-061
// regression pin (S-5 / SC-002). It drives the PRODUCTION declaration path
// (`buildToolDeclarations`) with the envelope an MCP tool offers and asserts that
// no keyword the provider rejects survives — at any depth. Deleting the
// projection turns this test RED (witness reproduced during the round).
func TestProjectToolDeclarations_NoUnsupportedKeywordReachesTheWire(t *testing.T) {
	decls := buildToolDeclarations([]llm.ToolDef{{
		Name:        "mcp_github_add_issue_comment",
		Description: "MCP tool",
		Parameters:  json.RawMessage(recordedEnvelope),
	}})
	if len(decls) != 1 {
		t.Fatalf("want 1 declaration, got %d", len(decls))
	}
	raw, err := json.Marshal(decls[0]["parameters"])
	if err != nil {
		t.Fatalf("marshal projected parameters: %v", err)
	}
	wire := string(raw)
	for _, k := range forbiddenKeys {
		if strings.Contains(wire, `"`+k+`"`) {
			t.Errorf("unsupported keyword %q reached the Gemini wire; parameters=%s", k, wire)
		}
	}
	// The supported surface must still be there (the projection is a carve-out,
	// not a lobotomy): the arguments, their types/descriptions, the enum, the
	// array bounds and the nested structure all survive.
	for _, want := range []string{`"owner"`, `"repo"`, `"body"`, `"Repository owner"`, `"Repository name"`,
		`"enum"`, `"minItems"`, `"maxItems"`, `"default"`, `"title"`, `"nested"`, `"deep"`, `"required"`} {
		if !strings.Contains(wire, want) {
			t.Errorf("the projected declaration lost %s; parameters=%s", want, wire)
		}
	}
}

// TestProjectSchema_KeepsOnlySupportedKeys pins the projection's default-deny
// contract on a flat fixture (the named allowlist is the single owner, S-2).
func TestProjectSchema_KeepsOnlySupportedKeys(t *testing.T) {
	in := `{"type":"object","title":"t","$schema":"x","x-any":1,"properties":{"a":{"type":"string","description":"d","readOnly":true,"pattern":"^a$"}},"required":["a"],"additionalProperties":false}`
	got := string(projectSchema(json.RawMessage(in)))
	for _, gone := range []string{"$schema", "x-any", "readOnly"} {
		if strings.Contains(got, gone) {
			t.Errorf("projection kept %q; got %s", gone, got)
		}
	}
	for _, kept := range []string{"title", "type", "description", "pattern", "required", "additionalProperties", "properties"} {
		if !strings.Contains(got, kept) {
			t.Errorf("projection dropped %q; got %s", kept, got)
		}
	}
}

// TestProjectSchema_FailsClosed covers the "unrepresentable" edge case: a
// non-JSON schema degrades to the freeform object (never a new failure mode).
func TestProjectSchema_FailsClosed(t *testing.T) {
	for _, in := range []string{``, `not json`, `"a string"`} {
		got := string(projectSchema(json.RawMessage(in)))
		if in == "" {
			if got != "" {
				t.Errorf("empty schema: want empty passthrough, got %q", got)
			}
			continue
		}
		if got != freeformParameters {
			t.Errorf("schema %q: want the freeform object, got %s", in, got)
		}
	}
}
