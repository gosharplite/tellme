package config

import (
	"os"
	"path/filepath"
	"testing"
)

// T024 [UNIT] — MCP_SERVERS validation rules (FR-002/FR-013/TD4).
func TestValidateMCPServers_RemoteEntryRules(t *testing.T) {
	tests := []struct {
		name    string
		servers map[string]MCPServerConfig
		wantErr bool
	}{
		{"valid bearer", map[string]MCPServerConfig{"gh": {URL: "https://x", Auth: "bearer", Token: "t"}}, false},
		{"valid auto without token", map[string]MCPServerConfig{"gh": {URL: "https://x"}}, false},
		{"valid none", map[string]MCPServerConfig{"gh": {URL: "https://x", Auth: "none"}}, false},
		{"valid basic", map[string]MCPServerConfig{"gh": {URL: "https://x", Auth: "basic", Token: "t", Username: "u"}}, false},
		{"bad key", map[string]MCPServerConfig{"Bad_Key": {URL: "https://x"}}, true},
		{"missing url", map[string]MCPServerConfig{"gh": {Auth: "none"}}, true},
		{"unknown auth", map[string]MCPServerConfig{"gh": {URL: "https://x", Auth: "weird"}}, true},
		{"bearer without token", map[string]MCPServerConfig{"gh": {URL: "https://x", Auth: "bearer"}}, true},
		{"negative timeout", map[string]MCPServerConfig{"gh": {URL: "https://x", Timeout: -1}}, true},
		{"command skipped (non-fatal)", map[string]MCPServerConfig{"fs": {Command: "npx"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{MCPServers: tt.servers}
			_, err := cfg.ValidateMCPServers()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateMCPServers() err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

// TestMCPServers_EnabledDefaultsTrue pins FR-012: an absent ENABLED key is enabled.
func TestMCPServers_EnabledDefaultsTrue(t *testing.T) {
	absent := MCPServerConfig{URL: "https://x"}
	if !absent.IsEnabled() {
		t.Fatal("an absent ENABLED key must default to enabled")
	}
	no := false
	disabled := MCPServerConfig{URL: "https://x", Enabled: &no}
	if disabled.IsEnabled() {
		t.Fatal("ENABLED: false must report disabled")
	}
}

// TestMCPServers_CommandWarnedAndSkipped pins FR-013 / TD4: a COMMAND-shaped
// entry is warn+skipped (a warning names it; it is NOT a remote entry; the run
// does not fail).
func TestMCPServers_CommandWarnedAndSkipped(t *testing.T) {
	cfg := &Config{MCPServers: map[string]MCPServerConfig{"fs": {Command: "npx", Token: "t"}}}
	val, err := cfg.ValidateMCPServers()
	if err != nil {
		t.Fatalf("a COMMAND entry must not be fatal: %v", err)
	}
	if cfg.MCPServers["fs"].IsRemote() {
		t.Fatal("a COMMAND-shaped entry must not be classified remote")
	}
	found := false
	for _, w := range val.Warnings {
		if contains(w, "fs") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a warning naming the skipped stdio server; got %v", val.Warnings)
	}
}

// TestMCPServers_TolerantSubKeys pins TD4: unmodelled sub-keys carried by a real
// tell-me-go block are tolerated (ignored), so tellme still starts.
func TestMCPServers_TolerantSubKeys(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "cfg.yaml")
	yaml := "MODE: butler\n" +
		"SELECTED_PROVIDER: x\n" +
		"PROVIDERS: {}\n" +
		"MCP_SERVERS:\n" +
		"  gh:\n" +
		"    URL: https://x\n" +
		"    AUTH: bearer\n" +
		"    TOKEN: t\n" +
		"    REQUIRES_CONSENT: false\n" +
		"    ARGS: [\"a\"]\n" +
		"    DIR: /tmp\n" +
		"    ENV:\n" +
		"      X: y\n"
	if err := os.WriteFile(p, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("a config with unmodelled MCP sub-keys must load: %v", err)
	}
	if cfg.MCPServers["gh"].URL != "https://x" {
		t.Fatalf("the remote entry was not decoded: %+v", cfg.MCPServers)
	}
}

// contains is a tiny local helper (the test file is self-contained).
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
