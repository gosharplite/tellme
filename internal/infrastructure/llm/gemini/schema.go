// Projection of a tool declaration's parameter schema onto the provider's
// SUPPORTED schema surface (round 061, issue #127; ADR 0031).
//
// Why this exists: tellme relays a remote MCP server's advertised input schema
// into the declaration it offers the model. Vertex/Gemini parses
// `functionDeclarations[].parameters` as the CLOSED proto message
// `google.ai.generativelanguage.Schema`, so ANY keyword it does not define — or
// any value whose JSON kind does not match the field — makes the API reject the
// WHOLE request (HTTP 400, `the provider request failed`) and no turn reaches the
// model. A third-party server's annotations (the GitHub server's `x-mcp-header`)
// therefore used to break every Gemini turn.
//
// The rule is DEFAULT-DENY on BOTH axes: a keyword reaches the wire only if it is
// in the empirically-verified set below, and only with a value of the kind its
// field requires. The set was measured against the live Vertex endpoint
// (declaration-only `generateContent` calls, 2026-09-19, `gemini-3.8-flash`,
// project `websc-dev-433809`; ADR 0031 D2 records the table).
//
// The provider-side GUARANTEE is this file; the family-agnostic floor that removes
// vendor extensions for every provider lives in the MCP normalizer
// (`mcp.NormalizeMCPSchema`, S-6) — the two seams have one concern each (S-1).
package gemini

import (
	"encoding/json"
	"fmt"
	"math"
)

// schemaValueKind is the JSON value kind a supported keyword requires — the
// second axis of the provider-supported surface (round 061 / ADR 0031 D8; the
// fold of review V-061-1). A key whose value has the wrong kind is a decode error
// of the whole request on a closed proto, so a wrong-kind value is DROPPED (an
// absent keyword is always accepted) — never emitted, never coerced into an
// unmeasured shape.
type schemaValueKind int

const (
	kindAny              schemaValueKind = iota // data — passed through (`default`)
	kindString                                  // string
	kindBool                                    // bool
	kindNumber                                  // number
	kindInt                                     // int32 / int64
	kindStringList                              // repeated string
	kindEnum                                    // repeated string (a single-valued enum)
	kindType                                    // a single-valued type enum (an array is coerced)
	kindSchemaNode                              // a Schema message (a non-object degrades to {})
	kindBoolOrSchemaNode                        // a Schema message or a bool
	kindSchemaList                              // repeated Schema
	kindObjectOfSchema                          // map<string, Schema>
)

// supportedSchemaValueKinds is the ONE owner of the surface: the key set AND each
// key's required value kind. `supportedSchemaKeys` is derived from it, so a key
// cannot exist without a kind (round 061 / ADR 0031 D8; the fold of V-061-1).
var supportedSchemaValueKinds = map[string]schemaValueKind{
	"type":                 kindType,
	"description":          kindString,
	"title":                kindString,
	"format":               kindString,
	"pattern":              kindString,
	"default":              kindAny,
	"enum":                 kindEnum,
	"properties":           kindObjectOfSchema,
	"required":             kindStringList,
	"propertyOrdering":     kindStringList,
	"items":                kindSchemaNode,
	"additionalProperties": kindBoolOrSchemaNode,
	"oneOf":                kindSchemaList,
	"allOf":                kindSchemaList,
	"minimum":              kindNumber,
	"maximum":              kindNumber,
	"minLength":            kindInt,
	"maxLength":            kindInt,
	"minItems":             kindInt,
	"maxItems":             kindInt,
	"nullable":             kindBool,
}

// supportedSchemaKeys is derived from the kind table (one literal source).
func supportedSchemaKeys() map[string]bool {
	out := make(map[string]bool, len(supportedSchemaValueKinds))
	for k := range supportedSchemaValueKinds {
		out[k] = true
	}
	return out
}

// freeformParameters is the fail-closed declaration used when a schema cannot be
// parsed (a declaration the wire always accepts).
const freeformParameters = `{"type":"object","properties":{}}`

// SupportedSchemaKeys returns a COPY of the provider's supported KEY set — the
// named owner both the projection and its regression gate read (round-061
// FR-007; folded per reviews F-061-1 and V-061-1). A copy keeps the owner single:
// a caller cannot alias, add to, or delete from it.
func SupportedSchemaKeys() map[string]bool { return supportedSchemaKeys() }

// SchemaValueKindOK reports whether a value has the JSON kind the key requires
// (the shape half of the gate; round 061 / ADR 0031 D8). The gate reads this so
// it checks SHAPE as well as key membership — a wrong-typed value is as invisible
// to a key-only walk as an unlisted key was before F-061-1.
func SchemaValueKindOK(key string, value any) bool {
	kind, ok := supportedSchemaValueKinds[key]
	if !ok {
		return false
	}
	_, keep := applyValueKind(kind, value)
	return keep
}

// ProjectSchema projects a tool declaration's parameter schema onto the
// provider's supported surface, recursively (round 061 / ADR 0031). It is pure
// and never fails: an empty schema passes through, and an unparseable or
// non-object schema degrades to the freeform object.
func ProjectSchema(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return json.RawMessage(freeformParameters)
	}
	obj, ok := v.(map[string]any)
	if !ok {
		// A declaration's parameters must be an object; anything else (a scalar,
		// a list) is unrepresentable — fail closed.
		return json.RawMessage(freeformParameters)
	}
	out, err := json.Marshal(projectValue(obj))
	if err != nil {
		return json.RawMessage(freeformParameters)
	}
	return out
}

// applyValueKind returns the value to keep for a key (and whether to keep it).
// A kind-matched value passes through; the measured coercions apply (`type`
// array, `enum` scalars, a non-object schema node); anything else is dropped.
func applyValueKind(kind schemaValueKind, val any) (any, bool) {
	switch kind {
	case kindAny:
		return val, true
	case kindString:
		s, ok := val.(string)
		return s, ok
	case kindBool:
		b, ok := val.(bool)
		return b, ok
	case kindNumber:
		f, ok := val.(float64)
		return f, ok
	case kindInt:
		f, ok := val.(float64)
		return f, ok && f == math.Trunc(f)
	case kindStringList:
		return normalizeStringList(val)
	case kindEnum:
		return normalizeEnum(val)
	case kindType:
		return normalizeType(val)
	case kindSchemaNode:
		return projectSubschema(val)
	case kindBoolOrSchemaNode:
		if b, ok := val.(bool); ok {
			return b, true
		}
		return projectSubschema(val)
	case kindSchemaList:
		return projectList(val)
	case kindObjectOfSchema:
		return projectProperties(val)
	}
	return nil, false
}

// normalizeStringList keeps a `repeated string` value only when EVERY member is a
// string (one bad member drops the keyword).
func normalizeStringList(val any) (any, bool) {
	list, ok := val.([]any)
	if !ok {
		return nil, false
	}
	out := make([]any, 0, len(list))
	for _, e := range list {
		s, ok := e.(string)
		if !ok {
			return nil, false
		}
		out = append(out, s)
	}
	return out, true
}

// normalizeType rewrites a supported `type` value to the single-valued form the
// wire accepts (ADR 0031 D2): a string passes through; an ARRAY is reduced to its
// lone non-`null` member (the `"null"` member is recorded by projectValue as
// `nullable: true`; an ambiguous list is dropped); any other kind is dropped.
func normalizeType(val any) (any, bool) {
	if s, ok := val.(string); ok {
		return s, true
	}
	list, ok := val.([]any)
	if !ok {
		return nil, false
	}
	kept := make([]any, 0, len(list))
	for _, t := range list {
		if s, isStr := t.(string); isStr && s == "null" {
			continue
		}
		kept = append(kept, t)
	}
	if len(kept) != 1 {
		return nil, false // zero or ambiguous members — drop rather than guess
	}
	return kept[0], true
}

// scalarTypeNames are the schema types a Gemini `enum` may sit beside — the probe
// rejected an enum on an OBJECT or ARRAY type and an enum with no type at all
// (ADR 0031 D2, R-2).
var scalarTypeNames = map[string]bool{"string": true, "integer": true, "number": true, "boolean": true}

// typeCarriesNull reports whether a `type` value carried the `"null"` member.
func typeCarriesNull(val any) bool {
	list, ok := val.([]any)
	if !ok {
		return false
	}
	for _, t := range list {
		if s, isStr := t.(string); isStr && s == "null" {
			return true
		}
	}
	return false
}

// normalizeEnum coerces scalar `enum` members to their string form (Gemini's
// `Schema.enum` is `repeated string`) and DROPS non-scalar / `null` members; an
// empty result drops the keyword (ADR 0031 D2/D6a′; the object-member nit of the
// V-061-1 review — a stringified object would hand the model a fake value).
func normalizeEnum(val any) (any, bool) {
	list, ok := val.([]any)
	if !ok {
		return nil, false
	}
	out := make([]any, 0, len(list))
	for _, m := range list {
		switch m.(type) {
		case string:
			out = append(out, m)
		case float64, bool:
			out = append(out, stringEnumMember(m))
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

// stringEnumMember renders a non-string SCALAR `enum` member as the string
// Gemini's `Schema.enum` (repeated string) requires — measured rejected
// otherwise (ADR 0031 D2: `enum: [1,2,3]` → TYPE_STRING).
func stringEnumMember(v any) any {
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}

// projectValue projects one schema node: it keeps only the keys in the value-kind
// table (the single owner) and applies each key's required kind, recursing where
// the kind says the value is (or contains) a schema node.
func projectValue(v any) any {
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}
	out := make(map[string]any, len(m))
	for k, val := range m {
		kind, ok := supportedSchemaValueKinds[k]
		if !ok {
			continue // not on the supported surface
		}
		if got, keep := applyValueKind(kind, val); keep {
			out[k] = got
		}
	}
	// An `enum` is only accepted beside a SCALAR type (ADR D6a′ / R-2): the probe
	// rejects an enum on an object/array type or with no type at all, so the
	// keyword drops whenever its node cannot legally carry it.
	if _, hasEnum := out["enum"]; hasEnum {
		if t, _ := out["type"].(string); !scalarTypeNames[t] {
			delete(out, "enum")
		}
	}
	// A `["null"]`-only type yields no `type`, and `nullable` is only accepted
	// BESIDE a type (ADR D6a′) — drop a `nullable` derived from such a type.
	if _, typed := out["type"]; !typed && typeCarriesNull(m["type"]) {
		delete(out, "nullable")
	}
	// Record the `"null"` member of an array type as `nullable: true` (only when a
	// type remains).
	if _, typed := out["type"]; typed && typeCarriesNull(m["type"]) {
		if _, set := out["nullable"]; !set {
			out["nullable"] = true
		}
	}
	return out
}

// projectProperties projects each declared property's subschema. A subschema that
// is not an object (a boolean schema, a bare scalar, …) is unrepresentable as a
// Schema message, so it degrades to the accepted empty `{}` (ADR D6c / R-3), and
// a `properties` value that is not a map drops the keyword.
func projectProperties(v any) (any, bool) {
	props, ok := v.(map[string]any)
	if !ok {
		return nil, false
	}
	out := make(map[string]any, len(props))
	for name, sub := range props {
		out[name], _ = projectSubschema(sub)
	}
	return out, true
}

// projectSubschema projects a value that must BE a schema node, coercing a
// non-object to the empty `{}` (ADR D6c / R-3).
func projectSubschema(v any) (any, bool) {
	if _, ok := v.(map[string]any); !ok {
		return map[string]any{}, true
	}
	return projectValue(v), true
}

// projectList projects each element of a keyword list (oneOf/allOf). The element
// position IS a schema node, so a non-object element degrades to the accepted
// empty `{}` (ADR D6c / R-5); a non-list value drops the keyword.
func projectList(v any) (any, bool) {
	list, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]any, len(list))
	for i, e := range list {
		out[i], _ = projectSubschema(e)
	}
	return out, true
}
