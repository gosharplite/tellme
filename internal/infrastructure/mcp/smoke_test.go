// Package mcp is the confined adapter layer for the Model Context Protocol: the
// ONLY production package allowed to import github.com/modelcontextprotocol/go-sdk
// (enforced by the verify-mcp-sdk-confinement Makefile gate, round-032
// research Decision 2). Every other layer consumes the tools.MCPClient domain
// port.
//
// This file is the round-032 T001 Setup smoke test: it proves the SDK is
// loadable and constructible so the dependency lands compilable before any
// product behaviour is written.
package mcp

import (
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestSDKLoads is the Setup smoke test (T001): the SDK package imports and a
// client can be constructed. It exercises no MCP behaviour.
func TestSDKLoads(t *testing.T) {
	c := sdk.NewClient(&sdk.Implementation{Name: "tellme", Version: "smoke"}, nil)
	if c == nil {
		t.Fatal("mcp.NewClient returned nil")
	}
}
