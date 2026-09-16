package config

import "testing"

// MCP_SERVERS ${VAR} expansion (round-032 SC-002 / issue #67).
//
// The tests drive the injected lookup (expandMCPServersWithLookup) rather than
// mutating process-global environment state, so the table runs in memory and in
// parallel — mirroring the round-003 expand_test.go discipline.

func TestExpandMCPServersWithLookup_ResolvesAndKeepsRaw(t *testing.T) {
	t.Parallel()

	lookup := func(key string) (string, bool) {
		switch key {
		case "MCP_TOKEN":
			return "s3cr3t", true
		case "MCP_HOST":
			return "mcp.example.com", true
		case "MCP_USER":
			return "alice", true
		default:
			return "", false
		}
	}

	cfg := &Config{MCPServers: map[string]MCPServerConfig{
		"shop": {
			URL:      "https://${MCP_HOST}/mcp",
			Token:    "${MCP_TOKEN}",
			Username: "${MCP_USER:-fallback}",
		},
	}}

	cfg.expandMCPServersWithLookup(lookup)
	got := cfg.MCPServers["shop"]

	if got.URL != "https://mcp.example.com/mcp" {
		t.Errorf("URL = %q, want %q", got.URL, "https://mcp.example.com/mcp")
	}
	if got.Token != "s3cr3t" {
		t.Errorf("Token = %q, want %q", got.Token, "s3cr3t")
	}
	if got.Username != "alice" {
		t.Errorf("Username = %q, want %q", got.Username, "alice")
	}
}

func TestExpandMCPServersWithLookup_DefaultAndUnset(t *testing.T) {
	t.Parallel()

	// Only MCP_SET is defined; MCP_UNSET is not.
	lookup := func(key string) (string, bool) {
		if key == "MCP_SET" {
			return "resolved", true
		}
		return "", false
	}

	cfg := &Config{MCPServers: map[string]MCPServerConfig{
		"a": {Token: "${MCP_SET:-d}"},     // set wins over default
		"b": {Token: "${MCP_UNSET:-d}"},   // unset falls back to the default
		"c": {Token: "${MCP_UNSET}"},      // unset, no default → raw preserved (non-fatal)
		"d": {Token: "https://x${MCP_UN"}, // malformed unclosed → raw preserved (non-fatal)
		"e": {Token: "plain-token"},       // no ${...} → untouched
	}}

	cfg.expandMCPServersWithLookup(lookup)

	want := map[string]string{
		"a": "resolved",
		"b": "d",
		"c": "${MCP_UNSET}",
		"d": "https://x${MCP_UN",
		"e": "plain-token",
	}
	for name, w := range want {
		if got := cfg.MCPServers[name].Token; got != w {
			t.Errorf("MCPServers[%q].Token = %q, want %q", name, got, w)
		}
	}
}

// TestExpandMCPServers_UnsetTokenDoesNotFailValidation pins the issue's
// non-fatal requirement: an unset ${VAR} token keeps its literal (non-empty)
// text, so a bearer entry is NOT rejected at startup — the server reaches the
// dial path and is warn+skipped, exactly as before the expansion existed.
func TestExpandMCPServers_UnsetTokenDoesNotFailValidation(t *testing.T) {
	t.Parallel()

	lookup := func(string) (string, bool) { return "", false }
	cfg := &Config{MCPServers: map[string]MCPServerConfig{
		"github": {URL: "https://api.githubcopilot.com/mcp/", Auth: "bearer", Token: "${GITHUB_TOKEN}"},
	}}
	cfg.expandMCPServersWithLookup(lookup)

	if _, err := cfg.ValidateMCPServers(); err != nil {
		t.Fatalf("an unset ${VAR} token must not fail validation: %v", err)
	}
	if got := cfg.MCPServers["github"].Token; got != "${GITHUB_TOKEN}" {
		t.Errorf("the unresolved token must be preserved literally; got %q", got)
	}
}

// TestExpandMCPServers_ProcessEnvFallback verifies the production entry point
// resolves against the process environment. Deliberately NOT parallel:
// t.Setenv mutates process-global state.
func TestExpandMCPServers_ProcessEnvFallback(t *testing.T) {
	t.Setenv("TELLME_MCP_EXPAND_TOKEN", "from-process-env")

	cfg := &Config{MCPServers: map[string]MCPServerConfig{
		"shop": {URL: "https://x/mcp", Token: "${TELLME_MCP_EXPAND_TOKEN}"},
	}}
	cfg.ExpandMCPServers()

	if got := cfg.MCPServers["shop"].Token; got != "from-process-env" {
		t.Errorf("Token = %q, want %q", got, "from-process-env")
	}
}
