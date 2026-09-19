package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// MCPDefaultTimeout is the default per-call timeout for a networked MCP tool
// (the execute_command class — NOT the 30 s local-reader default), overridable
// by the server's TIMEOUT (round-032 FR-021). It aliases the shared contract
// default (round-032 implementation-review F4).
const MCPDefaultTimeout = domaintools.DefaultToolTimeout

// ResolveMCPTimeout resolves a server's effective tool-call timeout: the server
// TIMEOUT (seconds) when positive, else MCPDefaultTimeout; then clamped to the
// contract's fixed TimeoutCeiling. It reuses the shared domain resolver so the
// ceiling and the default→clamp policy have ONE home (round-032 F4) — a server
// value cannot defeat the ceiling.
func ResolveMCPTimeout(seconds int) time.Duration {
	def := MCPDefaultTimeout
	if seconds > 0 {
		def = time.Duration(seconds) * time.Second
	}
	return domaintools.ResolveTimeout(0, def)
}

// Tool adapts a discovered MCP tool to the domaintools.Tool port (round-032
// FR-005/FR-014). It holds only the domain MCPClient port — never the SDK — so it
// lives on the product side of the confinement boundary while remaining
// network-agnostic.
type Tool struct {
	server      string
	tool        string
	name        string
	description string
	parameters  json.RawMessage
	client      domaintools.MCPClient
	timeout     time.Duration
}

// NewTool builds the adapter for one discovered tool. def.InputSchema is assumed
// already normalized (see NormalizeMCPSchema); a nil schema degrades to the
// freeform object.
func NewTool(server string, def domaintools.MCPToolDefinition, client domaintools.MCPClient, timeout time.Duration) *Tool {
	schema := def.InputSchema
	if len(schema) == 0 {
		schema = freeformSchema()
	}
	desc := def.Description
	if strings.TrimSpace(desc) == "" {
		desc = "MCP tool " + def.Name + " from server " + server
	}
	return &Tool{
		server:      server,
		tool:        def.Name,
		name:        NamespacedName(server, def.Name),
		description: desc,
		parameters:  schema,
		client:      client,
		timeout:     timeout,
	}
}

// Name is the deterministic namespaced wire name.
func (t *Tool) Name() string { return t.name }

// Description is the server-advertised description (or a generated fallback).
func (t *Tool) Description() string { return t.description }

// MCPPayloadKey is the envelope property carrying the remote server's own
// arguments. tellme owns this envelope; the server never sees it (round 056 /
// ADR 0025 D1/D2). It aliases the shared domain constant so the wire key has one
// authoritative spelling (round-056 review R-056-1).
const MCPPayloadKey = domaintools.PayloadArgKey

// ReasonKey is the envelope's tellme-owned reason property (required). It is
// rendered by tellme and never forwarded to the server (round 056 / ADR 0025).
// It aliases the shared domain constant (round-056 review R-056-1).
const ReasonKey = domaintools.ReasonArgKey

// reasonDescription is the tellme-authored description of the envelope's
// required `reason` property — the ask the model reads in the offered
// declaration. It lives only in tellme's declaration; the server's definition is
// never mutated.
const reasonDescription = "Why you are calling this tool (required). Put the tool's own arguments in MCP_PAYLOAD."

// Parameters is the model-visible declaration tellme OFFERS for the MCP tool
// (round 056 / ADR 0025 D1): tellme's OWN envelope — a required `reason` plus
// `MCP_PAYLOAD`, whose subschema is the remote server's advertised input schema
// carried VERBATIM (never mutated; only positioned inside the envelope). The
// server's definition is untouched and the system prompt is unchanged — the
// declared `reason` is the same elicitation mechanism the native tools use.
func (t *Tool) Parameters() json.RawMessage { return mcpEnvelope(t.parameters) }

// mcpEnvelope builds the offered declaration around the server's advertised
// schema (already normalized — see NormalizeMCPSchema). It composes the envelope
// STRUCTURALLY (json.Marshal of typed values; the server schema is carried as a
// json.RawMessage, so it is relayed verbatim) rather than by string formatting —
// Go's %q is Go-escaping, not JSON-escaping, and string interpolation of a
// third-party schema is the round-031/#64 catastrophic class. A non-JSON schema
// degrades to the freeform object (round-056 review R-056-2 hardening).
func mcpEnvelope(serverSchema json.RawMessage) json.RawMessage {
	if len(serverSchema) == 0 || !json.Valid(serverSchema) {
		return freeformEnvelope()
	}
	reasonProp, err := json.Marshal(map[string]string{"type": "string", "description": reasonDescription})
	if err != nil {
		return freeformEnvelope()
	}
	env := struct {
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}{
		Type: "object",
		Properties: map[string]json.RawMessage{
			ReasonKey:     reasonProp,
			MCPPayloadKey: serverSchema,
		},
		Required: []string{ReasonKey},
	}
	out, err := json.Marshal(env)
	if err != nil {
		return freeformEnvelope()
	}
	return out
}

// freeformEnvelope is the well-formed envelope whose MCP_PAYLOAD is the empty
// object schema — the degrade target when the server's schema is absent or not
// valid JSON.
func freeformEnvelope() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"reason":{"type":"string"},"MCP_PAYLOAD":{"type":"object","properties":{}}},"required":["reason"]}`)
}

// envelopeViolation is the recoverable result tellme returns, WITHOUT contacting
// the server, when an MCP call is not a valid envelope (round 056 / ADR 0025 D2).
const envelopeViolation = `error: an MCP tool call must be {"reason":"...","MCP_PAYLOAD":{...}}; nothing was sent to the server`

// Contract exposes the MCP tool's default timeout so the loop resolves the same
// effective bound as for a native tool.
func (t *Tool) Contract() domaintools.ToolContract {
	return domaintools.ToolContract{DefaultTimeout: t.timeout}
}

// Execute calls the tool on its server and returns the result text, bounded at
// the source to the byte budget (the loop keeps its own backstop). A nil args
// object is normalised to `{}`; ANY call-time failure is a recoverable nil-error
// result (round-032 FR-018 / TD1).
//
// NOTE (F8, confirmed intended): because every call-time failure is a nil-error
// result, round-026 classifies an MCP failure as `ok` and the round-022 log line
// shows no error — consistent with round-026's recorded forward item
// ("recoverable inline failures count as ok"). A future distinction would need
// an explicit error marker on the fed-back result text.
func (t *Tool) Execute(ctx context.Context, arguments string, budget domaintools.ByteBudget) (string, error) {
	payload, ok := unwrapEnvelope(arguments)
	if !ok {
		// Round 056 (ADR 0025 D2): a shape violation is REFUSED — the server is
		// never contacted — and the model receives a recoverable result (nil error,
		// the round-032 TD1 convention) asking it to retry with the envelope.
		return envelopeViolation, nil
	}
	res, err := t.client.CallTool(ctx, t.tool, payload)
	if err != nil {
		// CallTool never returns a call-time error today (TD1/R3); if a future
		// implementation did, surface it as a recoverable result text.
		return "error: " + err.Error(), nil
	}
	text := res.Text
	if b := int(budget); b > 0 && len(text) > b {
		text = strings.ToValidUTF8(text[:b], "") + domaintools.TruncationMarker
	}
	return text, nil
}

// unwrapEnvelope validates the round-056 MCP call envelope and returns the
// object to forward to the server (the `MCP_PAYLOAD` contents). It returns
// ok=false for any shape violation — a top-level key other than `reason` /
// `MCP_PAYLOAD`, or a `MCP_PAYLOAD` present but not a JSON object. An ABSENT
// `MCP_PAYLOAD` (with a reason) is a legitimate empty payload ({}).
//
// It deliberately does NOT re-implement the reason-presence half of the rule —
// the loop's universal gate owns that (ADR 0025 D3/D4: one owner per rule).
func unwrapEnvelope(arguments string) (map[string]interface{}, bool) {
	raw := map[string]interface{}{}
	if s := strings.TrimSpace(arguments); s != "" {
		// UseNumber preserves integer literals beyond 2^53 through the decode
		// (round-056 review TD-056-5): a plain Unmarshal would turn an int64-shaped
		// argument into float64 and silently alter it on the way to the server.
		dec := json.NewDecoder(strings.NewReader(s))
		dec.UseNumber()
		if err := dec.Decode(&raw); err != nil {
			return nil, false
		}
		if raw == nil {
			raw = map[string]interface{}{}
		}
	}
	payload := map[string]interface{}{}
	for k, v := range raw {
		switch k {
		case ReasonKey:
			// tellme's own field: rendered by the loop, never forwarded.
		case MCPPayloadKey:
			if v == nil {
				return nil, false // present-but-null is not an object
			}
			obj, isObj := v.(map[string]interface{})
			if !isObj {
				return nil, false
			}
			payload = obj
		default:
			return nil, false // a stray top-level key is a shape violation
		}
	}
	return payload, true
}
