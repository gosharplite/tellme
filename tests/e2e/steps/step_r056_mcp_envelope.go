package steps

import (
	"encoding/json"
)

// Round-056 MCP helpers (ADR 0025), kept in one file so the per-sentence step
// files stay independent (Zero Shared Edits).
//
// An MCP tool call the model makes is the envelope
// `{"reason":"...","MCP_PAYLOAD":{...}}`: `reason` is tellme's own field
// (rendered as `[Tool Reason]`, never forwarded); `MCP_PAYLOAD` carries the
// remote server's own arguments (forwarded verbatim).
const mcpPayloadKey = "MCP_PAYLOAD"

// mcpEnvelopeJSON builds the round-056 MCP call envelope with the given reason
// and a raw `MCP_PAYLOAD` JSON literal (e.g. "{}" or `{"sku":"A1"}`).
func mcpEnvelopeJSON(reason, payloadJSON string) string {
	m := map[string]json.RawMessage{}
	r, _ := json.Marshal(reason)
	m["reason"] = r
	m[mcpPayloadKey] = json.RawMessage(payloadJSON)
	out, _ := json.Marshal(m)
	return string(out)
}
