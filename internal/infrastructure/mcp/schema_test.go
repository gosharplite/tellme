package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

// T028(a) [UNIT] — schema normalization / well-formedness (FR-019 / B1).
func TestNormalizeMCPSchema_AbsentIsFreeform(t *testing.T) {
	for _, in := range []json.RawMessage{nil, []byte(""), []byte("null"), []byte("  ")} {
		out, err := NormalizeMCPSchema(in)
		if err != nil {
			t.Fatalf("an absent schema must degrade to freeform, got err=%v", err)
		}
		if !strings.Contains(string(out), `"type":"object"`) {
			t.Fatalf("freeform schema malformed: %s", out)
		}
	}
}

func TestNormalizeMCPSchema_ValidObjectPreserved(t *testing.T) {
	in := json.RawMessage(`{"type":"object","properties":{"a":{"type":"string"}},"required":["a"]}`)
	out, err := NormalizeMCPSchema(in)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err := json.Unmarshal(out, &obj); err != nil {
		t.Fatal(err)
	}
	props, _ := obj["properties"].(map[string]any)
	req, _ := obj["required"].([]any)
	for _, r := range req {
		if _, ok := props[r.(string)]; !ok {
			t.Fatalf("required %q not in properties — required ⊄ properties", r)
		}
	}
}

func TestNormalizeMCPSchema_RequiredWithoutPropertyRejected(t *testing.T) {
	// the #64-mirror malformed shape: required names an undeclared property.
	in := json.RawMessage(`{"type":"object","required":["x"]}`)
	if _, err := NormalizeMCPSchema(in); err == nil {
		t.Fatal("a required entry with no matching property must be rejected")
	}
}

func TestNormalizeMCPSchema_NonObjectRejected(t *testing.T) {
	for _, s := range []string{`"string"`, `[1,2]`, `123`, `true`} {
		if _, err := NormalizeMCPSchema(json.RawMessage(s)); err == nil {
			t.Fatalf("%s must be rejected as unsafe (FR-019)", s)
		}
	}
}

func TestNormalizeMCPSchema_BadShapesRejected(t *testing.T) {
	for _, s := range []string{
		`{"type":"object","properties":"nope"}`,
		`{"type":"object","properties":{},"required":"nope"}`,
		`{"type":"string"}`,
	} {
		if _, err := NormalizeMCPSchema(json.RawMessage(s)); err == nil {
			t.Fatalf("%s must be rejected", s)
		}
	}
}

// T-061 [UNIT] — the vendor-extension floor (round 061 / ADR 0031 D-floor): an
// `x-…` annotation (the GitHub server's `x-mcp-header`) and `$schema` are dropped
// for EVERY family, recursively (root, nested property, inside `items`/`oneOf`),
// while the declared arguments and the supported surface survive.
func TestNormalizeMCPSchema_StripsVendorExtensions(t *testing.T) {
	in := json.RawMessage(`{
	  "type":"object",
	  "$schema":"https://json-schema.org/draft/2020-12/schema",
	  "x-root":"gone",
	  "properties":{
	    "owner":{"type":"string","description":"Repository owner","x-mcp-header":"owner"},
	    "repo":{"type":"string","description":"Repository name","x-mcp-header":"repo"},
	    "files":{"type":"array","description":"d","items":{"type":"string","x-nested":"gone"}},
	    "alt":{"anyOf":[{"type":"string","x-branch":"gone"},{"type":"number"}]}
	  },
	  "required":["owner","repo"]
	}`)
	out, err := NormalizeMCPSchema(in)
	if err != nil {
		t.Fatalf("a well-formed schema with vendor extensions must normalize, got err=%v", err)
	}
	got := string(out)
	for _, gone := range []string{"x-mcp-header", "$schema", "x-root", "x-nested", "x-branch"} {
		if strings.Contains(got, gone) {
			t.Errorf("the floor kept the vendor keyword %q; got %s", gone, got)
		}
	}
	for _, kept := range []string{"owner", "repo", "Repository owner", "Repository name", "files", "items", "anyOf", "required"} {
		if !strings.Contains(got, kept) {
			t.Errorf("the floor dropped %q, which is not a vendor extension; got %s", kept, got)
		}
	}
}
