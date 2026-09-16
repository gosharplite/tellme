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
