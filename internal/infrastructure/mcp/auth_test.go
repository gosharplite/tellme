package mcp

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/config"
)

// T025 [UNIT] — auth resolution by mode + token-not-logged (FR-006/FR-017).
func TestResolveAuthorization_Modes(t *testing.T) {
	ghToken := func(context.Context) (string, error) { return "gh-tok", nil }
	ghFail := func(context.Context) (string, error) { return "", context.DeadlineExceeded }

	tests := []struct {
		name     string
		server   config.MCPServerConfig
		gh       TokenSource
		want     string
		wantWarn bool
	}{
		{"none is anonymous", config.MCPServerConfig{URL: "https://x", Auth: "none"}, nil, "", false},
		{"bearer uses explicit token", config.MCPServerConfig{URL: "https://x", Auth: "bearer", Token: "t"}, nil, "Bearer t", false},
		{"basic base64s user:token", config.MCPServerConfig{URL: "https://x", Auth: "basic", Token: "t", Username: "u"}, nil,
			"Basic " + base64.StdEncoding.EncodeToString([]byte("u:t")), false},
		{"auto with explicit token wins", config.MCPServerConfig{URL: "https://x", Token: "t"}, nil, "Bearer t", false},
		{"auto github-host falls back to source", config.MCPServerConfig{URL: "https://api.github.com/mcp"}, ghToken, "Bearer gh-tok", false},
		{"auto non-github anonymous", config.MCPServerConfig{URL: "https://example.com/mcp"}, ghToken, "", false},
		{"gh uses the source", config.MCPServerConfig{URL: "https://x", Auth: "gh"}, ghToken, "Bearer gh-tok", false},
		{"gh source failure warns + anonymous", config.MCPServerConfig{URL: "https://x", Auth: "gh"}, ghFail, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warn := ResolveAuthorization(context.Background(), tt.server, tt.gh)
			if got != tt.want {
				t.Fatalf("header=%q, want %q", got, tt.want)
			}
			if (warn != "") != tt.wantWarn {
				t.Fatalf("warning=%q, wantWarn=%v", warn, tt.wantWarn)
			}
			if strings.Contains(warn, "t") && tt.server.Token != "" && strings.Contains(warn, tt.server.Token) {
				t.Fatalf("the warning leaked the token: %q", warn)
			}
		})
	}
}
