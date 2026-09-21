package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 077 (ADR 0049) — the offered MCP declaration's description makes the
// CALLABLE wire name positively discoverable: a tellme-authored prefix note,
// added OUTSIDE the server's own (unchanged) text, plus a fallback that names
// the callable name (never the bare upstream name).
//
// Folds (PR #157 architect review): F-1 the "server text relayed unchanged"
// claim is pinned by BYTE equality over a whitespace/CRLF-bearing fixture (a
// HasPrefix/HasSuffix pair over a whitespace-free fixture could not redden for
// it); F-2 the envelope-neutrality claim is pinned by the exact property-name
// SET; F-3 the fallback's provenance is pinned on the fallback SEGMENT (the
// prefix note's `t.name` already contains the server key).

// A description-bearing server whose text carries leading/trailing whitespace
// and a newline: the note prefixes it and the server's text follows IMMEDIATELY
// and byte-for-byte (F-1).
func TestMCPToolName_DescriptionIsNotePlusServerTextByteExact(t *testing.T) {
	const serverText = "  Get details of the authenticated GitHub user.\n\tUse when the request is about the profile.  "
	tool := NewTool("github", domaintools.MCPToolDefinition{Name: "get_me", Description: serverText}, &recordingClient{}, time.Second)

	want := fmt.Sprintf(callNameNoteFormat, "mcp_github_get_me") + serverText
	if got := tool.Description(); got != want {
		t.Fatalf("description must be the note concatenated with the server text BYTE-EXACT (no trim, no interposition):\n got %q\nwant %q", got, want)
	}
	if tool.Name() != "mcp_github_get_me" {
		t.Fatalf("name = %q, want mcp_github_get_me", tool.Name())
	}
	// The server's text must be the IMMEDIATE suffix (nothing interposed).
	if !strings.HasSuffix(tool.Description(), serverText) {
		t.Fatalf("the server's text must immediately follow the note; got %q", tool.Description())
	}
}

// An empty server description: the synthesized fallback names the CALLABLE wire
// name — never the bare upstream name — and states the source server (the
// fallback SEGMENT is asserted, not the whole string, since `t.name` already
// contains the server key — F-3).
func TestMCPToolName_EmptyDescriptionFallbackNamesTheCallableName(t *testing.T) {
	tool := NewTool("github", domaintools.MCPToolDefinition{Name: "get_me"}, &recordingClient{}, time.Second)

	got := tool.Description()
	// The whole description is the note + the fallback body (byte-exact, F-1).
	note := fmt.Sprintf(callNameNoteFormat, "mcp_github_get_me")
	want := note + "MCP tool mcp_github_get_me from server github"
	if got != want {
		t.Fatalf("fallback description:\n got %q\nwant %q", got, want)
	}
	// The fallback body names the callable name AND states the source server
	// (F-3: assert the fallback SEGMENT, which the note alone cannot satisfy).
	body := strings.TrimPrefix(got, note)
	if !strings.Contains(body, "mcp_github_get_me") {
		t.Fatalf("the fallback body must name the callable name; got %q", body)
	}
	if !strings.Contains(body, "from server github") {
		t.Fatalf("the fallback body must state the source server; got %q", body)
	}
	if strings.Contains(body, "MCP tool get_me from server") {
		t.Fatalf("the fallback body must NOT present the bare upstream name as callable; got %q", body)
	}
}

// A 64-byte-truncated (hash-suffixed) name: the note names the POST-truncation
// wire name — it can never lie about a truncated name (ADR 0049 D5).
func TestMCPToolName_NoteNamesTheTruncatedWireName(t *testing.T) {
	server := strings.Repeat("s", 24)
	def := domaintools.MCPToolDefinition{Name: strings.Repeat("t", 200), Description: "d"}
	tool := NewTool(server, def, &recordingClient{}, time.Second)

	if len(tool.Name()) > 64 {
		t.Fatalf("name %q exceeds 64 bytes", tool.Name())
	}
	want := fmt.Sprintf(callNameNoteFormat, tool.Name()) + "d"
	if got := tool.Description(); got != want {
		t.Fatalf("the note must name the post-truncation wire name:\n got %q\nwant %q", got, want)
	}
}

// The change is DESCRIPTION-ONLY: the offered envelope's SHAPE is unchanged —
// its property-name set is EXACTLY {reason, MCP_PAYLOAD} (F-2: a property-name
// set pin reddens if a stray property is added; the earlier Contains-only pin
// could not).
func TestMCPToolName_OfferedEnvelopeShapeIsUnchanged(t *testing.T) {
	schema := []byte(`{"type":"object","properties":{"q":{"type":"string"}},"required":["q"]}`)
	tool := NewTool("srv", domaintools.MCPToolDefinition{Name: "t", Description: "d", InputSchema: schema}, &recordingClient{}, time.Second)

	var env struct {
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(tool.Parameters(), &env); err != nil {
		t.Fatalf("the offered parameters are not a JSON object: %v", err)
	}
	if env.Type != "object" {
		t.Fatalf("envelope root type = %q, want object", env.Type)
	}
	names := make([]string, 0, len(env.Properties))
	for k := range env.Properties {
		names = append(names, k)
	}
	sort.Strings(names)
	want := []string{MCPPayloadKey, ReasonKey}
	sort.Strings(want)
	if !equalStrings(names, want) {
		t.Fatalf("the envelope's property set = %v, want exactly %v", names, want)
	}
	if len(env.Required) != 1 || env.Required[0] != ReasonKey {
		t.Fatalf("the envelope must require exactly %q, got %v", ReasonKey, env.Required)
	}
	// The server's schema still rides MCP_PAYLOAD verbatim, and the note never
	// enters the schema (the round-077 text change cannot touch the shape).
	if !strings.Contains(string(env.Properties[MCPPayloadKey]), string(schema)) {
		t.Fatalf("the server's schema must ride MCP_PAYLOAD verbatim; got %s", env.Properties[MCPPayloadKey])
	}
	if strings.Contains(string(tool.Parameters()), "Call this tool as") {
		t.Fatalf("the call-name note must not enter the offered schema; got %s", tool.Parameters())
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
