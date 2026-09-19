package gemini

import (
	"encoding/json"
	"fmt"
	"reflect"
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

// assertContainment walks a projected declaration STRUCTURE-AWARELY and asserts
// every schema KEYWORD it carries is in the provider's supported surface — the
// named owner (round-061 FR-007 / review F-061-1). Property NAMES and data
// values (default/enum/const/examples) are not keywords and are not judged.
func assertContainment(t *testing.T, node any, path string) {
	t.Helper()
	m, ok := node.(map[string]any)
	if !ok {
		return
	}
	for k, v := range m {
		if !SupportedSchemaKeys()[k] {
			t.Errorf("unsupported keyword %q at %s reached the wire", k, path)
			continue
		}
		// SHAPE half (V-061-1): the value must have the kind the key requires.
		if !SchemaValueKindOK(k, v) {
			t.Errorf("keyword %q at %s has a value of the wrong kind: %v", k, path, v)
		}
		switch k {
		case "properties":
			props, _ := v.(map[string]any)
			for name, sub := range props {
				assertContainment(t, sub, path+".properties."+name)
			}
		case "items", "additionalProperties":
			assertContainment(t, v, path+"."+k)
		case "oneOf", "allOf":
			if list, ok := v.([]any); ok {
				for i, e := range list {
					assertContainment(t, e, fmt.Sprintf("%s.%s[%d]", path, k, i))
				}
			}
		}
	}
}

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
	// The reject surface is the OWNER, read by the containment walk below — no
	// second transcription of the deny-list (review nit 4).
	// The supported surface must still be there (the projection is a carve-out,
	// not a lobotomy).
	for _, want := range []string{`"owner"`, `"repo"`, `"body"`, `"Repository owner"`, `"Repository name"`,
		`"enum"`, `"minItems"`, `"maxItems"`, `"default"`, `"title"`, `"nested"`, `"deep"`, `"required"`} {
		if !strings.Contains(wire, want) {
			t.Errorf("the projected declaration lost %s; parameters=%s", want, wire)
		}
	}
	// CONTAINMENT (review F-061-1): every keyword the declaration carries must be
	// in the provider's OWNER set — the gate fails on a keyword nobody listed.
	var tree any
	if err := json.Unmarshal([]byte(wire), &tree); err != nil {
		t.Fatalf("decode projected declaration: %v", err)
	}
	assertContainment(t, tree, "parameters")
}

// measuredSupportedKeys is the GOLDEN set: the keywords the round-061 probe
// measured as accepted (ADR 0031 D2). The owner must equal it exactly — so
// DELETING a key from the owner (which would silently strip it from every
// declaration) fails here, and ADDING an unmeasured key fails here too (review
// F-061-1: "deleting `items` turns nothing red").
var measuredSupportedKeys = []string{
	"type", "description", "properties", "required", "items", "enum", "format",
	"title", "default", "nullable", "pattern", "minimum", "maximum", "minLength",
	"maxLength", "minItems", "maxItems", "oneOf", "allOf", "additionalProperties",
	"propertyOrdering",
}

// TestSupportedSchemaKeys_MatchTheMeasuredProbe pins the owner set to the probe.
func TestSupportedSchemaKeys_MatchTheMeasuredProbe(t *testing.T) {
	golden := map[string]bool{}
	for _, k := range measuredSupportedKeys {
		golden[k] = true
	}
	for k := range SupportedSchemaKeys() {
		if !golden[k] {
			t.Errorf("the owner declares %q, which the probe did not verify as accepted — add it only with a fresh probe (ADR 0031 D2)", k)
		}
	}
	for _, k := range measuredSupportedKeys {
		if !SupportedSchemaKeys()[k] {
			t.Errorf("the owner is MISSING the measured-accepted keyword %q — deleting it would silently strip it from every declaration", k)
		}
	}
}

// ownerFixtures is one declaration per measured-supported keyword, each carrying
// that keyword at a schema position.
var ownerFixtures = map[string]string{
	"type":                 `{"type":"object","properties":{"a":{"type":"string"}}}`,
	"description":          `{"type":"object","properties":{"a":{"type":"string","description":"v"}}}`,
	"properties":           `{"type":"object","properties":{"a":{"type":"string"}}}`,
	"required":             `{"type":"object","properties":{"a":{"type":"string"}},"required":["a"]}`,
	"items":                `{"type":"object","properties":{"a":{"type":"array","items":{"type":"string"}}}}`,
	"enum":                 `{"type":"object","properties":{"a":{"type":"string","enum":["x"]}}}`,
	"format":               `{"type":"object","properties":{"a":{"type":"string","format":"date-time"}}}`,
	"title":                `{"type":"object","properties":{"a":{"type":"string","title":"t"}}}`,
	"default":              `{"type":"object","properties":{"a":{"type":"string","default":"v"}}}`,
	"nullable":             `{"type":"object","properties":{"a":{"type":"string","nullable":true}}}`,
	"pattern":              `{"type":"object","properties":{"a":{"type":"string","pattern":"^a$"}}}`,
	"minimum":              `{"type":"object","properties":{"a":{"type":"integer","minimum":1}}}`,
	"maximum":              `{"type":"object","properties":{"a":{"type":"integer","maximum":9}}}`,
	"minLength":            `{"type":"object","properties":{"a":{"type":"string","minLength":1}}}`,
	"maxLength":            `{"type":"object","properties":{"a":{"type":"string","maxLength":9}}}`,
	"minItems":             `{"type":"object","properties":{"a":{"type":"array","items":{"type":"string"},"minItems":1}}}`,
	"maxItems":             `{"type":"object","properties":{"a":{"type":"array","items":{"type":"string"},"maxItems":1}}}`,
	"oneOf":                `{"type":"object","properties":{"a":{"oneOf":[{"type":"string"}]}}}`,
	"allOf":                `{"type":"object","properties":{"a":{"allOf":[{"type":"string"}]}}}`,
	"additionalProperties": `{"type":"object","properties":{"a":{"type":"string"}},"additionalProperties":false}`,
	"propertyOrdering":     `{"type":"object","properties":{"a":{"type":"string"}},"propertyOrdering":["a"]}`,
}

// TestProjectedDeclaration_CoversTheOwnerSet is the COVERAGE half of the gate
// (review F-061-1): every keyword in the owner set survives projection, so
// deleting one from the set cannot silently strip it from every declaration.
func TestProjectedDeclaration_CoversTheOwnerSet(t *testing.T) {
	for _, k := range measuredSupportedKeys {
		fixture, ok := ownerFixtures[k]
		if !ok {
			t.Fatalf("the coverage pin has no fixture for measured key %q — add one", k)
		}
		if got := string(ProjectSchema(json.RawMessage(fixture))); !strings.Contains(got, `"`+k+`"`) {
			t.Errorf("measured key %q did not survive projection: %s", k, got)
		}
	}
}

// TestProjectSchema_CoercedShapesAreTheMeasuredOnes pins the SHAPES THE COERCION
// EMITS against the probe (ADR 0031 D2, extended by review R-2): `nullable`
// without a `type` is REJECTED, and an `enum` beside a non-scalar `type` (or with
// no type) is REJECTED — so the coercion must not produce either. A non-object
// subschema is coerced to the accepted empty `{}` (review R-3).
func TestProjectSchema_CoercedShapesAreTheMeasuredOnes(t *testing.T) {
	got := string(ProjectSchema(json.RawMessage(`{"type":"object","properties":{
	  "onlynull":{"type":["null"]},
	  "arr":{"type":"array","items":{"type":"string"},"enum":["a","b"]},
	  "notype":{"enum":["a","b"]},
	  "obj":{"type":"object","enum":["{}"]},
	  "int":{"type":"integer","enum":[1,2,3]},
	  "boolprop":{"type":"boolean","enum":[true,false]},
	  "weird":true,
	  "itemsweird":{"type":"array","items":false},
	  "union":{"oneOf":[true,{"type":"string"}]},
	  "unionall":{"allOf":[false]}
	}}`)))
	// A `["null"]`-only type must leave NEITHER `type` NOR `nullable` (a nullable
	// with no type is rejected).
	var tree map[string]any
	if err := json.Unmarshal([]byte(got), &tree); err != nil {
		t.Fatalf("decode: %v", err)
	}
	props := tree["properties"].(map[string]any)
	onlyNull := props["onlynull"].(map[string]any)
	if _, has := onlyNull["type"]; has {
		t.Errorf("a [\"null\"]-only type must be dropped; got %v", onlyNull)
	}
	if _, has := onlyNull["nullable"]; has {
		t.Errorf("`nullable` without a `type` is rejected by the wire; got %v", onlyNull)
	}
	for _, k := range []string{"arr", "notype", "obj"} {
		if _, has := props[k].(map[string]any)["enum"]; has {
			t.Errorf("%s: an enum beside a non-scalar or absent type must be dropped; got %v", k, props[k])
		}
	}
	for k, want := range map[string]string{"int": `["1","2","3"]`, "boolprop": `["true","false"]`} {
		if !strings.Contains(got, `"enum":`+want) {
			t.Errorf("%s: want the coerced string enum %s; got %s", k, want, got)
		}
	}
	// R-5: the ELEMENT position of oneOf/allOf is a schema node too.
	for k, want := range map[string]any{"union": []any{map[string]any{}, map[string]any{"type": "string"}}, "unionall": []any{map[string]any{}}} {
		if !reflect.DeepEqual(props[k].(map[string]any)["oneOf"], want) && !reflect.DeepEqual(props[k].(map[string]any)["allOf"], want) {
			t.Errorf("%s: a non-object union element must degrade to {}; got %v", k, props[k])
		}
	}
	for _, k := range []string{"weird", "itemsweird"} {
		sub := props[k]
		if k == "itemsweird" {
			sub = props[k].(map[string]any)["items"]
		}
		if !reflect.DeepEqual(sub, map[string]any{}) {
			t.Errorf("%s: a non-object subschema must degrade to {}; got %v", k, sub)
		}
	}
}

// TestProjectSchema_NormalizesValueShapes pins the value-shape normalization the
// live probe forced (ADR 0031 D2, value-shape rows; review F-061-2):
// `type` array → single member (+ nullable), ambiguous → dropped; `enum` members
// → strings.
func TestProjectSchema_NormalizesValueShapes(t *testing.T) {
	got := string(ProjectSchema(json.RawMessage(
		`{"type":"object","properties":{"a":{"type":["string","null"]},"b":{"type":["string","number"]},"c":{"type":"integer","enum":[1,2,3]}}}`)))
	for _, want := range []string{`"type":"string"`, `"nullable":true`, `"enum":["1","2","3"]`} {
		if !strings.Contains(got, want) {
			t.Errorf("value-shape normalization missing %s; got %s", want, got)
		}
	}
	// The ambiguous member list is dropped rather than guessed.
	if strings.Contains(got, `["string","number"]`) {
		t.Errorf("an ambiguous `type` array must be dropped; got %s", got)
	}
	var tree map[string]any
	if err := json.Unmarshal([]byte(got), &tree); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, present := tree["properties"].(map[string]any)["b"].(map[string]any)["type"]; present {
		t.Errorf("the ambiguous property kept a `type`; got %s", got)
	}
}

// TestProjectSchema_KeepsOnlySupportedKeys pins the projection's default-deny
// contract on a flat fixture (the named allowlist is the single owner, S-2).
func TestProjectSchema_KeepsOnlySupportedKeys(t *testing.T) {
	in := `{"type":"object","title":"t","$schema":"x","x-any":1,"properties":{"a":{"type":"string","description":"d","readOnly":true,"pattern":"^a$"}},"required":["a"],"additionalProperties":false}`
	got := string(ProjectSchema(json.RawMessage(in)))
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
		got := string(ProjectSchema(json.RawMessage(in)))
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

// TestProjectSchema_ValueKindsAreEnforced pins the SHAPE half of the surface
// (ADR 0031 D8; the fold of review V-061-1): a value whose JSON kind does not
// match its field is dropped, and a kind-correct value survives.
func TestProjectSchema_ValueKindsAreEnforced(t *testing.T) {
	bad := string(ProjectSchema(json.RawMessage(`{"type":"object","properties":{"a":{
	  "description":123,"title":true,"pattern":7,"format":{"x":1},"type":5,
	  "nullable":"yes","minItems":"3","minimum":"1","items":"nope","oneOf":{"type":"string"},
	  "required":[5],"enum":[{"x":1}],"properties":"nope"
	}},"required":"x"}`)))
	var tree map[string]any
	if err := json.Unmarshal([]byte(bad), &tree); err != nil {
		t.Fatalf("decode: %v", err)
	}
	a := tree["properties"].(map[string]any)["a"].(map[string]any)
	for _, k := range []string{"description", "title", "pattern", "format", "type", "nullable",
		"minItems", "minimum", "oneOf", "required", "enum", "properties"} {
		if _, has := a[k]; has {
			t.Errorf("wrong-kind %q survived the projection; got %v", k, a[k])
		}
	}
	if _, has := tree["required"]; has {
		t.Errorf("a wrong-kind root `required` survived: %v", tree["required"])
	}
	if got, ok := a["items"].(map[string]any); !ok || len(got) != 0 {
		t.Errorf("a non-object `items` must degrade to {}; got %v", a["items"])
	}
	// W-061-1: a `type` ARRAY whose lone member is not a string must emit NO type.
	for _, in := range []string{`{"type":[5]}`, `{"type":[true]}`, `{"type":[{}]}`, `{"type":[["string"]]}`} {
		got := string(ProjectSchema(json.RawMessage(`{"type":"object","properties":{"a":` + in + `}}`)))
		var t2 map[string]any
		if err := json.Unmarshal([]byte(got), &t2); err != nil {
			t.Fatalf("decode: %v", err)
		}
		sub := t2["properties"].(map[string]any)["a"].(map[string]any)
		if _, has := sub["type"]; has {
			t.Errorf("a non-string type-array member survived as `type`; input=%s got=%v", in, sub)
		}
	}
	good := string(ProjectSchema(json.RawMessage(`{"type":"object","properties":{"a":{
	  "type":"string","description":"d","title":"t","format":"date-time","pattern":"^a$",
	  "nullable":true,"minLength":1,"maxLength":9,"enum":["x","y"]
	}},"required":["a"],"propertyOrdering":["a"]}`)))
	for _, want := range []string{`"description":"d"`, `"title":"t"`, `"format":"date-time"`,
		`"pattern":"^a$"`, `"nullable":true`, `"minLength":1`, `"maxLength":9`, `"propertyOrdering":["a"]`} {
		if !strings.Contains(good, want) {
			t.Errorf("a kind-correct value was dropped: missing %s; got %s", want, good)
		}
	}
}

// TestSchemaValueKindOK_MirrorsTheTable pins the gate helper to the table: every
// key accepts its own kind and rejects a foreign one.
func TestSchemaValueKindOK_MirrorsTheTable(t *testing.T) {
	okValues := map[string]any{
		"type": "string", "description": "d", "title": "t", "format": "f", "pattern": "p",
		"default": []any{1}, "enum": []any{"x"}, "properties": map[string]any{"a": map[string]any{}},
		"required": []any{"a"}, "propertyOrdering": []any{"a"}, "items": map[string]any{},
		"additionalProperties": false, "oneOf": []any{map[string]any{}}, "allOf": []any{map[string]any{}},
		"minimum": 1.0, "maximum": 9.0, "minLength": 1.0, "maxLength": 9.0,
		"minItems": 1.0, "maxItems": 9.0, "nullable": true,
	}
	for k, v := range okValues {
		if !SchemaValueKindOK(k, v) {
			t.Errorf("SchemaValueKindOK(%q, %v) = false; want true", k, v)
		}
	}
	for _, wrong := range []struct {
		k string
		v any
	}{{"description", 1}, {"nullable", "yes"}, {"required", "x"}, {"minItems", "3"}, {"type", 5}, {"properties", "nope"}} {
		if SchemaValueKindOK(wrong.k, wrong.v) {
			t.Errorf("SchemaValueKindOK(%q, %v) accepted a wrong kind", wrong.k, wrong.v)
		}
	}
	if SchemaValueKindOK("not-a-key", "x") {
		t.Error("an unknown key must not be whitelisted by the kind helper")
	}
	if SupportedSchemaKeys()["not-a-key"] {
		t.Error("an unknown key must not be in the owner set")
	}
}

// TestProjectSchema_OutputSatisfiesTheGate turns "the projection and its gate
// agree" into a checked invariant (fold of review W-061-1): for a set of
// adversarial inputs, the projected output must pass the SAME predicate the gate
// uses — every keyword in the owner set, with the required value kind.
func TestProjectSchema_OutputSatisfiesTheGate(t *testing.T) {
	inputs := []string{
		recordedEnvelope,
		`{"type":"object","properties":{"a":{"type":[5]}}}`,
		`{"type":"object","properties":{"a":{"type":[true]}}}`,
		`{"type":"object","properties":{"a":{"type":[{}]}}}`,
		`{"type":"object","properties":{"a":{"type":[["string"]]}}}`,
		`{"type":"object","properties":{"a":{"type":["string","null"],"enum":[1,2]}}}`,
		`{"type":"object","properties":{"a":true,"b":"nope","c":7}}`,
		`{"type":"object","properties":{"a":{"oneOf":[true,{"type":"string"}]}}}`,
		`{"type":"object","properties":{"a":{"items":false,"oneOf":{"x":1},"required":[5]}}}`,
		`{"type":"object","properties":{"a":{"description":123,"nullable":"yes","minItems":"3"}}}`,
	}
	for _, in := range inputs {
		out := ProjectSchema(json.RawMessage(in))
		var tree any
		if err := json.Unmarshal(out, &tree); err != nil {
			t.Fatalf("the projected output is not JSON for %s: %v", in, err)
		}
		assertContainment(t, tree, "projected")
	}
}
