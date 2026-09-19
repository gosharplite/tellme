// Projection of a tool declaration's parameter schema onto the provider's
// SUPPORTED schema surface (round 061, issue #127; ADR 0031).
//
// Why this exists: tellme relays a remote MCP server's advertised input schema
// into the declaration it offers the model. Vertex/Gemini parses
// `functionDeclarations[].parameters` as the CLOSED proto message
// `google.ai.generativelanguage.Schema`, so ANY keyword it does not define makes
// the API reject the WHOLE request (HTTP 400, `the provider request failed`) and
// no turn reaches the model. A third-party server's annotations (the GitHub
// server's `x-mcp-header`) therefore used to break every Gemini turn.
//
// The rule is DEFAULT-DENY: a keyword reaches the wire only if it is in the
// empirically-verified set below. The set was measured against the live Vertex
// endpoint (declaration-only `generateContent` calls, 2026-09-19,
// `gemini-3.8-flash`, project `websc-dev-433809`; ADR 0031 records the table):
//
//	accepted  — type, description, properties, required, items, enum, format,
//	            title, default, nullable, pattern, minimum, maximum, minLength,
//	            maxLength, minItems, maxItems, oneOf, allOf,
//	            additionalProperties, propertyOrdering, `type: "null"`
//	rejected  — x-* (any vendor extension), $schema, const, examples,
//	            deprecated, readOnly, writeOnly, multipleOf, uniqueItems,
//	            $ref, $defs, definitions, and `anyOf` whenever any other schema
//	            keyword sits beside it ("when using any_of, it must be the only
//	            field set")
//
// `anyOf` is deliberately NOT allowlisted: it is accepted only when it is the
// sole schema keyword on its node, a shape the projection cannot guarantee, so
// default-deny drops it (ADR 0031 §Forward records the enrichment idea).
//
// This is the provider-side GUARANTEE. The family-agnostic floor that removes
// vendor extensions for every provider lives in the MCP normalizer
// (`mcp.NormalizeMCPSchema`, S-6) — the two seams have one concern each (S-1).
package gemini

import (
	"encoding/json"
	"fmt"
)

// supportedSchemaKeys is the named owner of the provider's supported schema
// surface — the single home the projection and its regression pin both read
// (round-061 S-2/FR-007).
var supportedSchemaKeys = map[string]bool{
	"type":                 true,
	"description":          true,
	"properties":           true,
	"required":             true,
	"items":                true,
	"enum":                 true,
	"format":               true,
	"title":                true,
	"default":              true,
	"nullable":             true,
	"pattern":              true,
	"minimum":              true,
	"maximum":              true,
	"minLength":            true,
	"maxLength":            true,
	"minItems":             true,
	"maxItems":             true,
	"oneOf":                true,
	"allOf":                true,
	"additionalProperties": true,
	"propertyOrdering":     true,
}

// freeformParameters is the fail-closed declaration used when a schema cannot be
// parsed (a declaration the wire always accepts).
const freeformParameters = `{"type":"object","properties":{}}`

// SupportedSchemaKeys returns the provider's supported schema surface — the
// NAMED OWNER both the projection and its regression gate read (round-061
// FR-007; the fold of review F-061-1). Callers must treat the result as
// read-only.
func SupportedSchemaKeys() map[string]bool { return supportedSchemaKeys }

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

// stringEnumMember renders a non-string `enum` member as the string Gemini's
// `Schema.enum` (repeated string) requires — measured rejected otherwise
// (ADR 0031 D2, value-shape probe: `enum: [1,2,3]` → TYPE_STRING).
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

// normalizeTypeKeyword rewrites an ARRAY `type` into the single-valued form
// Gemini's `Schema.type` accepts (measured: `type: ["string"]` is rejected as an
// unknown name — ADR 0031 D2, value-shape probe). Per JSON-Schema `type`
// semantics: drop the `"null"` member and record it as `nullable: true`; keep the
// lone remaining member; drop the keyword entirely when the remainder is empty or
// ambiguous (more than one member — no faithful single value exists).
func normalizeTypeKeyword(node map[string]any) {
	list, ok := node["type"].([]any)
	if !ok {
		return
	}
	nullable := false
	kept := make([]any, 0, len(list))
	for _, t := range list {
		s, isStr := t.(string)
		if isStr && s == "null" {
			nullable = true
			continue
		}
		kept = append(kept, t)
	}
	switch {
	case len(kept) == 1:
		node["type"] = kept[0]
	case len(kept) == 0:
		delete(node, "type")
	default:
		delete(node, "type") // ambiguous — drop rather than guess a member type
	}
	if nullable {
		if _, set := node["nullable"]; !set {
			node["nullable"] = true
		}
	}
}

// normalizeEnumKeyword coerces every `enum` member to a string (Gemini's
// `Schema.enum` is `repeated string`; a numeric member is rejected — ADR 0031 D2).
func normalizeEnumKeyword(node map[string]any) {
	if raw, ok := node["enum"].([]any); ok {
		out := make([]any, len(raw))
		for i, v := range raw {
			out[i] = stringEnumMember(v)
		}
		node["enum"] = out
	}
}

// projectValue projects one schema node: a map keeps only allowlisted keys (and
// recurses where structure lives), a list projects its elements, a scalar is
// returned as-is.
func projectValue(v any) any {
	switch node := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(node))
		for k, val := range node {
			if !supportedSchemaKeys[k] {
				continue
			}
			switch k {
			case "properties":
				out[k] = projectProperties(val)
			case "items", "additionalProperties":
				out[k] = projectValue(val)
			case "oneOf", "allOf":
				out[k] = projectList(val)
			default:
				out[k] = val
			}
		}
		normalizeTypeKeyword(out)
		normalizeEnumKeyword(out)
		return out
	case []any:
		return projectList(node)
	}
	return v
}

// projectProperties projects each declared property's subschema.
func projectProperties(v any) any {
	props, ok := v.(map[string]any)
	if !ok {
		return v
	}
	out := make(map[string]any, len(props))
	for name, sub := range props {
		out[name] = projectValue(sub)
	}
	return out
}

// projectList projects each element of a keyword list (oneOf/allOf).
func projectList(v any) any {
	list, ok := v.([]any)
	if !ok {
		return v
	}
	out := make([]any, len(list))
	for i, e := range list {
		out[i] = projectValue(e)
	}
	return out
}
