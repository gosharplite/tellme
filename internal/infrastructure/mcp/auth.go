package mcp

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/gosharplite/tellme/internal/config"
)

// TokenSource resolves a token from an external source (the `gh` CLI). It is the
// injectable seam (round-032 FR-020) so no test spawns `gh`.
type TokenSource func(ctx context.Context) (string, error)

// CredentialWarning is the non-fatal warning emitted when the external token
// source cannot be reached (the auth falls back to anonymous).
const CredentialWarning = "[mcp] the credential source could not be reached; continuing anonymously"

// ResolveAuthorization resolves the Authorization header value for a server by
// its effective auth mode (round-032 FR-006, research Decision 4):
//
//   - none          → anonymous ("")
//   - bearer        → "Bearer <TOKEN>" (the explicit token always wins)
//   - basic         → "Basic base64(<USERNAME>:<TOKEN>)"
//   - gh            → the token source, as Bearer; a source failure warns and
//     falls back to anonymous (never fails the run)
//   - auto (default)→ the explicit TOKEN when present; else a GitHub-hosted host
//     falls back to the token source; else anonymous
//
// The resolved header (and the token it carries) is NEVER logged (FR-017). A
// non-empty warning signals a non-fatal fallback (the token source could not be
// reached).
func ResolveAuthorization(ctx context.Context, server config.MCPServerConfig, gh TokenSource) (header string, warning string) {
	switch server.EffectiveAuth() {
	case "none":
		return "", ""
	case "bearer":
		if strings.TrimSpace(server.Token) == "" {
			return "", ""
		}
		return "Bearer " + server.Token, ""
	case "basic":
		cred := base64.StdEncoding.EncodeToString([]byte(server.Username + ":" + server.Token))
		return "Basic " + cred, ""
	case "gh":
		return fromTokenSource(ctx, gh)
	case "auto":
		if strings.TrimSpace(server.Token) != "" {
			return "Bearer " + server.Token, ""
		}
		if isGitHubHost(server.URL) {
			return fromTokenSource(ctx, gh)
		}
		return "", ""
	default:
		return "", ""
	}
}

// fromTokenSource resolves a token via the injected source, falling back to
// anonymous (with a warning) on any failure.
func fromTokenSource(ctx context.Context, gh TokenSource) (string, string) {
	if gh == nil {
		return "", ""
	}
	tok, err := gh(ctx)
	if err != nil || strings.TrimSpace(tok) == "" {
		return "", CredentialWarning
	}
	return "Bearer " + strings.TrimSpace(tok), ""
}

// isGitHubHost reports whether a server URL's host is github-hosted (used by the
// `auto` mode's token-source fallback detection).
func isGitHubHost(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == "github.com" || strings.HasSuffix(host, ".github.com") ||
		host == "githubcopilot.com" || strings.HasSuffix(host, ".githubcopilot.com")
}
