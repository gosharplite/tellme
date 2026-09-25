package tools

import (
	"context"
	"encoding/json"
)

// MCPToolDefinition is one tool a remote MCP server advertises (round-032
// research Decision 2). It is the network-free, protocol-free projection of the
// server's tool listing: the wire name, a human description, and the tool's
// argument JSON Schema.
//
// InputSchema is carried as json.RawMessage so the adapter is the only place
// that knows the protocol library's schema representation; the domain layer
// stays dependency-free.
type MCPToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
}

// MCPToolResult is the outcome of an MCP tool call: the text fed back into the
// conversation like a native tool result (round-032 FR-005/FR-018).
type MCPToolResult struct {
	// Text is the result text fed back to the model.
	Text string
}

// MCPClient is the domain port through which tellme lists a remote MCP server's
// tools and calls one (round-032 research Decision 2). It is implemented once by
// the SDK adapter in internal/infrastructure/mcp (the only package that imports
// the protocol library) so the rest of the codebase — the agent loop, the CLI —
// stays free of the protocol types (NFR-005).
//
// Invariants the port fixes for every implementation (round-032 review folds):
//
//   - TD2 / R2 — CallTool's `args` map is NEVER sent as JSON null: a nil map is
//     normalised to `{}` before the wire, because a strict server rejects
//     `"arguments": null`.
//
//   - TD1 / R3 — NO call-time failure returns a non-nil error: a tool-level
//     error reported by the server (isError), a transport/connection failure, and
//     a call against an already-closed client ALL return a nil-error MCPToolResult
//     carrying the error text. The loop therefore always takes the recoverable
//     path and the run never aborts on an MCP call-time failure (FR-018). The
//     `error` return is retained only for interface symmetry / a future genuinely
//     terminal condition; today's call-time failures never use it.
//
// ListTools is the discovery half: it lists the server's tools (a transport
// failure there is a discovery failure the caller warns+skips).
type MCPClient interface {
	ListTools(ctx context.Context) ([]MCPToolDefinition, error)
	CallTool(ctx context.Context, name string, args map[string]interface{}) (MCPToolResult, error)
	Close() error
}
