package mcp

import (
	"strings"
	"testing"
	"time"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 077 (ADR 0049) — the offered MCP declaration's description makes the
// CALLABLE wire name positively discoverable: a tellme-authored prefix note,
// added OUTSIDE the server's own (unchanged) text, plus a fallback that names
// the callable name (never the bare upstream name).

// A description-bearing server: the note prefixes the server text, and the
// server's own words are preserved byte-identically as its suffix.
func TestMCPToolName_DescriptionPrefixesTheServerTextVerbatim(t *testing.T) {
	const serverText = "Get details of the authenticated GitHub user."
	tool := NewTool("github", domaintools.MCPToolDefinition{Name: "get_me", Description: serverText}, &recordingClient{}, time.Second)

	got := tool.Description()
	wantPrefix := `Call this tool as "mcp_github_get_me". `
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("description %q does not start with the call-name note %q", got, wantPrefix)
	}
	if !strings.HasSuffix(got, serverText) {
		t.Fatalf("the server's own text must be relayed unchanged as the suffix; got %q", got)
	}
	if tool.Name() != "mcp_github_get_me" {
		t.Fatalf("name = %q, want mcp_github_get_me", tool.Name())
	}
}

// An empty server description: the synthesized fallback names the CALLABLE wire
// name — never the bare upstream name (the round-077 correctness fix).
func TestMCPToolName_EmptyDescriptionFallbackNamesTheCallableName(t *testing.T) {
	tool := NewTool("github", domaintools.MCPToolDefinition{Name: "get_me"}, &recordingClient{}, time.Second)

	got := tool.Description()
	if !strings.Contains(got, "mcp_github_get_me") {
		t.Fatalf("the fallback must name the callable name; got %q", got)
	}
	if strings.Contains(got, "MCP tool get_me from server") {
		t.Fatalf("the fallback must NOT name the bare upstream name as callable; got %q", got)
	}
	if !strings.Contains(got, "github") {
		t.Fatalf("the fallback must still state the source server; got %q", got)
	}
	// The prefix note is present too, so the callable name leads.
	if !strings.HasPrefix(got, `Call this tool as "mcp_github_get_me". `) {
		t.Fatalf("the fallback must still carry the call-name note; got %q", got)
	}
}

// A 64-byte-truncated (hash-suffixed) name: the note names the POST-truncation
// wire name — it can never lie about a truncated name.
func TestMCPToolName_NoteNamesTheTruncatedWireName(t *testing.T) {
	server := strings.Repeat("s", 24)
	def := domaintools.MCPToolDefinition{Name: strings.Repeat("t", 200), Description: "d"}
	tool := NewTool(server, def, &recordingClient{}, time.Second)

	if len(tool.Name()) > 64 {
		t.Fatalf("name %q exceeds 64 bytes", tool.Name())
	}
	note := `Call this tool as "` + tool.Name() + `". `
	if !strings.HasPrefix(tool.Description(), note) {
		t.Fatalf("the note must name the post-truncation wire name; desc=%q note=%q", tool.Description(), note)
	}
}

// The change is DESCRIPTION-ONLY: the offered envelope (Parameters) is
// unaffected — the MCP_PAYLOAD still carries the server's schema verbatim.
func TestMCPToolName_OfferedSchemaIsUnchanged(t *testing.T) {
	schema := []byte(`{"type":"object","properties":{"q":{"type":"string"}},"required":["q"]}`)
	tool := NewTool("srv", domaintools.MCPToolDefinition{Name: "t", Description: "d", InputSchema: schema}, &recordingClient{}, time.Second)

	got := string(tool.Parameters())
	if !strings.Contains(got, string(schema)) {
		t.Fatalf("the server's schema must ride MCP_PAYLOAD verbatim; got %s", got)
	}
	// The description text must not leak into the schema.
	if strings.Contains(got, "Call this tool as") {
		t.Fatalf("the call-name note must not enter the offered schema; got %s", got)
	}
}
