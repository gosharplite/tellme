// Package di is the composition root for the MCP client (round-032): it wires the
// confined SDK adapter and the bounded external token source behind seams the CLI
// injects, so the presentation layer never couples to the protocol library or to
// a raw process spawn.
package di

import (
	"context"
	"os/exec"
	"strings"
	"time"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	mcp "github.com/gosharplite/tellme/internal/infrastructure/mcp"
)

// TokenResolver resolves a token from an external source (the `gh` CLI) for a
// server's credential. It is the injectable seam (round-032 FR-020, review fold
// B3): tests inject a fake so NO `gh` process is ever spawned. It aliases
// mcp.TokenSource so the CLI seam and the adapter agree on one type.
type TokenResolver = mcp.TokenSource

// NewRemoteClient builds the remote (Streamable HTTP) MCP client adapter for a
// server, given the already-resolved Authorization header (empty = anonymous)
// and the two bounding deadlines: discoveryTimeout caps the connect/list phase
// (the fixed fast-fail bound) and callTimeout caps the tool calls. Credential
// RESOLUTION is the caller's job; this is the transport construction.
func NewRemoteClient(ctx context.Context, url, authorization string, discoveryTimeout, callTimeout time.Duration) (domaintools.MCPClient, error) {
	return mcp.NewRemoteClient(ctx, url, authorization, discoveryTimeout, callTimeout)
}

// NewGhTokenResolver returns the production token resolver: it spawns `gh auth
// token` bounded by bound (the SAME fixed fast-fail deadline as discovery —
// round-032 FR-020), trimming the output. On timeout/failure it returns an error
// the caller turns into a warn + anonymous fallback (never a run failure). It is
// bounded and ctx-carrying — a deliberate improvement over the reference's
// unbounded `gh` shell-out (round-032 research Decision 4, R6).
func NewGhTokenResolver(bound time.Duration) mcp.TokenSource {
	return func(ctx context.Context) (string, error) {
		cctx, cancel := context.WithTimeout(ctx, bound)
		defer cancel()
		cmd := exec.CommandContext(cctx, "gh", "auth", "token")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	}
}
