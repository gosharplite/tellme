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
// by the server's TIMEOUT (round-032 FR-021).
const MCPDefaultTimeout = 300 * time.Second

// MCPTimeoutCeiling is the contract's FIXED timeout ceiling (mirrors
// agent.TimeoutCeiling): a server-set TIMEOUT must not defeat it (round-032
// FR-021 / TD5 — distinct from the contract's token-bound ceiling).
const MCPTimeoutCeiling = 7200 * time.Second

// ResolveMCPTimeout resolves a server's effective tool-call timeout: the server
// TIMEOUT (seconds) when positive, else MCPDefaultTimeout; clamped to
// MCPTimeoutCeiling. The loop applies the same ceiling as a backstop.
func ResolveMCPTimeout(seconds int) time.Duration {
	t := MCPDefaultTimeout
	if seconds > 0 {
		t = time.Duration(seconds) * time.Second
	}
	if t > MCPTimeoutCeiling {
		t = MCPTimeoutCeiling
	}
	return t
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

// Parameters is the normalized input schema offered to the model.
func (t *Tool) Parameters() json.RawMessage { return t.parameters }

// Contract exposes the MCP tool's default timeout so the loop resolves the same
// effective bound as for a native tool.
func (t *Tool) Contract() domaintools.ToolContract {
	return domaintools.ToolContract{DefaultTimeout: t.timeout}
}

// Execute calls the tool on its server and returns the result text, bounded at
// the source to the byte budget (the loop keeps its own backstop). A nil args
// object is normalised to `{}`; ANY call-time failure is a recoverable nil-error
// result (round-032 FR-018 / TD1).
func (t *Tool) Execute(ctx context.Context, arguments string, budget domaintools.ByteBudget) (string, error) {
	args := map[string]interface{}{}
	if s := strings.TrimSpace(arguments); s != "" {
		_ = json.Unmarshal([]byte(s), &args) // a non-object / unparseable arg degrades to {}
		if args == nil {
			args = map[string]interface{}{}
		}
	}
	res, err := t.client.CallTool(ctx, t.tool, args)
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
