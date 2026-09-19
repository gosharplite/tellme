package mcp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// isVendorExtension reports whether a schema KEYWORD is a vendor extension or a
// non-standard annotation that tellme strips for EVERY provider family (round
// 061 / ADR 0031 D-floor): an `x-…` extension (e.g. the GitHub MCP server's
// `x-mcp-header`, which only tells the SERVER to route the argument as an HTTP
// header and means nothing to the model) or the `$schema` dialect declaration.
//
// It is a KEYWORD predicate: a property NAME may itself begin with `x-`, so the
// caller must apply it at keyword positions only (see stripVendorExtensions).
func isVendorExtension(key string) bool {
	return strings.HasPrefix(key, "x-") || key == "$schema"
}

// schemaNodeChildKeys are the keywords whose VALUE is a nested schema node, so
// the floor descends into them.
var schemaNodeChildKeys = []string{
	"items", "additionalProperties", "not", "contains", "if", "then", "else",
	"propertyNames", "unevaluatedItems", "unevaluatedProperties",
}

// schemaNodeChildLists are the keywords whose value is a LIST of schema nodes.
var schemaNodeChildLists = []string{"anyOf", "allOf", "oneOf", "prefixItems"}

// schemaNodeChildMaps are the keywords whose value is a MAP FROM NAME TO a schema
// node — the map's keys are names (opaque), never keywords.
var schemaNodeChildMaps = []string{"properties", "patternProperties", "dependentSchemas"}

// stripVendorExtensions removes vendor-extension KEYWORDS from a decoded schema
// node in place, recursively and STRUCTURE-AWARELY (round 061 / ADR 0031 D6b;
// the fold of review B-061-1).
//
// It descends only through schema-node positions and never walks a data value:
// a declared argument whose NAME begins with `x-` is preserved (and
// `required ⊆ properties` still holds), and an `x-…` member inside a
// `default`/`enum`/`const`/`examples` VALUE is data, not a keyword.
func stripVendorExtensions(node map[string]any) {
	for k := range node {
		if isVendorExtension(k) {
			delete(node, k)
		}
	}
	for _, k := range schemaNodeChildKeys {
		if m, ok := node[k].(map[string]any); ok {
			stripVendorExtensions(m)
		}
	}
	for _, k := range schemaNodeChildLists {
		if list, ok := node[k].([]any); ok {
			for _, e := range list {
				if m, ok := e.(map[string]any); ok {
					stripVendorExtensions(m)
				}
			}
		}
	}
	for _, k := range schemaNodeChildMaps {
		children, ok := node[k].(map[string]any)
		if !ok {
			continue
		}
		for _, sub := range children { // names are opaque — never filtered
			if m, ok := sub.(map[string]any); ok {
				stripVendorExtensions(m)
			}
		}
	}
}

// freeformSchema is the well-formed object schema used when a server advertises
// no usable input schema: an object accepting free-form arguments. It satisfies
// the round-031 / issue #64 invariant `required ⊆ properties` vacuously.
func freeformSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}

// NormalizeMCPSchema verifies and normalizes a remote server's advertised tool
// input schema before the tool is offered to the model (round-032 FR-019, review
// fold B1). A dynamic MCP tool reopens the issue #64 vector from the OUTSIDE: a
// remote, non-reviewable author advertising `required` without a matching
// `properties` entry (or a non-object schema) would make a strict provider 400
// the WHOLE request.
//
// The normalizer guarantees the offered schema is a JSON object declaring its
// `properties` with `required ⊆ properties`:
//
//   - an ABSENT / empty / `null` schema degrades to the freeform object (safe);
//   - a non-object / unparseable schema, or one whose `properties`/`required`
//     shapes are invalid, or whose `required` names an undeclared property, is
//     UNSAFE and returns an error — the caller skips that tool with a warning
//     (the tool is not offered; the server's other tools still are).
//
// (Note: round-032 tasks.md T028 phrases the non-object case as "→ freeform";
// FR-019 — the authoritative spec — says a non-object/unparseable schema MUST be
// skipped. This implementation follows the spec: non-object ⇒ error ⇒ skip.)
func NormalizeMCPSchema(raw json.RawMessage) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return freeformSchema(), nil
	}

	var obj map[string]any
	if err := json.Unmarshal(trimmed, &obj); err != nil {
		return nil, fmt.Errorf("input schema is not a JSON object: %w", err)
	}
	if obj == nil {
		return nil, errors.New("input schema is not a JSON object")
	}
	if err := normalizeSchemaObject(obj); err != nil {
		return nil, err
	}
	// Round 061 (ADR 0031 D-floor): drop vendor-extension keywords for EVERY
	// provider family. The floor is family-agnostic — the OpenAI-compatible
	// transport tolerates an annotation, but the byte-identity of a declaration
	// is not worth a per-family rule, and the annotation is server-side plumbing.
	// Round 061 (ADR 0031 D-floor): drop vendor-extension KEYWORDS for every
	// family, then RE-ASSERT the round-031/#64 postcondition on the post-floor
	// object — the floor deletes keywords only and a property NAME is opaque, but
	// the invariant must hold by construction, not by argument (review B-061-1).
	stripVendorExtensions(obj)
	postProps, _ := obj["properties"].(map[string]any)
	if postProps == nil {
		postProps = map[string]any{}
	}
	if err := checkRequiredDeclared(obj["required"], postProps); err != nil {
		return nil, err
	}
	out, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("marshal normalized schema: %w", err)
	}
	return out, nil
}

// normalizeSchemaObject enforces the well-formedness invariant on a parsed
// schema object in place: an object root type, an object `properties`
// declaration, and `required ⊆ properties`.
func normalizeSchemaObject(obj map[string]any) error {
	if t, ok := obj["type"]; ok {
		if ts, isStr := t.(string); !isStr || ts != "object" {
			return fmt.Errorf("input schema root type must be object, got %v", t)
		}
	}
	props, _ := obj["properties"].(map[string]any)
	if obj["properties"] != nil && props == nil {
		return errors.New("input schema properties must be an object")
	}
	if props == nil {
		props = map[string]any{}
	}
	if err := checkRequiredDeclared(obj["required"], props); err != nil {
		return err
	}
	obj["type"] = "object"
	obj["properties"] = props
	return nil
}

// checkRequiredDeclared verifies that every `required` entry names a declared
// property (the round-031 / issue #64 invariant).
func checkRequiredDeclared(required any, props map[string]any) error {
	if required == nil {
		return nil
	}
	arr, ok := required.([]any)
	if !ok {
		return errors.New("input schema required must be an array")
	}
	for _, r := range arr {
		name, ok := r.(string)
		if !ok {
			return errors.New("input schema required entries must be strings")
		}
		if _, declared := props[name]; !declared {
			return fmt.Errorf("input schema requires %q which is not declared in properties", name)
		}
	}
	return nil
}
