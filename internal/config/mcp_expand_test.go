package config

import (
	"strings"
	"testing"
)

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

// TestExpandMCPServersWithLookup_WarnsOnUnset pins the SC-002 review improvement:
// an unset ${VAR} still keeps the literal value (non-fatal), but now ALSO returns
// a non-fatal diagnostic warning naming the server, field, and variable — so an
// unresolved credential is self-diagnosing rather than surfacing only as an
// opaque "could not be reached".
func TestExpandMCPServersWithLookup_WarnsOnUnset(t *testing.T) {
	t.Parallel()

	lookup := func(key string) (string, bool) {
		if key == "MCP_SET" {
			return "resolved", true
		}
		return "", false
	}
	cfg := &Config{MCPServers: map[string]MCPServerConfig{
		"hf": {URL: "https://x", Token: "${GITHUB_TOKEN}"}, // unset → warn, literal kept
		"ok": {URL: "https://${MCP_SET}", Token: "plain"},  // resolves → no warning
	}}

	warnings := cfg.expandMCPServersWithLookup(lookup)

	if len(warnings) != 1 {
		t.Fatalf("want exactly one warning (the unset token); got %v", warnings)
	}
	for _, want := range []string{"MCP_SERVERS.hf.TOKEN", "${GITHUB_TOKEN}", "unset"} {
		if !strings.Contains(warnings[0], want) {
			t.Errorf("warning %q must mention %q", warnings[0], want)
		}
	}
	if got := cfg.MCPServers["hf"].Token; got != "${GITHUB_TOKEN}" {
		t.Errorf("the unresolved token must be preserved literally; got %q", got)
	}
}

// TestExpandMCPServersWithLookup_WarnsOnMalformed pins the malformed-expression
// branch of the same diagnostic.
func TestExpandMCPServersWithLookup_WarnsOnMalformed(t *testing.T) {
	t.Parallel()

	lookup := func(string) (string, bool) { return "", false }
	cfg := &Config{MCPServers: map[string]MCPServerConfig{
		"bad": {URL: "https://x", Token: "${UNDONE"},
	}}

	warnings := cfg.expandMCPServersWithLookup(lookup)

	if len(warnings) != 1 || !strings.Contains(warnings[0], "malformed") {
		t.Fatalf("want one malformed-expression warning; got %v", warnings)
	}
	if got := cfg.MCPServers["bad"].Token; got != "${UNDONE" {
		t.Errorf("the malformed value must be preserved literally; got %q", got)
	}
}
