package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

// Round 087 (ADR 0058) — the cross-invocation MCP tool cache steps.

func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the tools of the MCP server "([^"]*)" have already been discovered$`, givenMCPToolsAlreadyDiscovered)
		ctx.Given(`^the tools of the MCP server "([^"]*)" were discovered more than a day ago$`, givenMCPToolsDiscoveredLongAgo)
		ctx.Given(`^the MCP server "([^"]*)" has since stopped answering$`, givenMCPServerStoppedAnswering)
		ctx.Then(`^tellme remembered the tools of the MCP server "([^"]*)"$`, thenMCPToolsRemembered)
		ctx.Then(`^tellme contacted the MCP server "([^"]*)"$`, thenContactedMCPServer)
	})
}

// givenMCPToolsAlreadyDiscovered arranges a WARM, FRESH cache entry for the
// server: the prelude must reuse it and dial nothing.
func givenMCPToolsAlreadyDiscovered(ctx context.Context, server string) error {
	return scenarioFrom(ctx).writeMCPToolCache(server, 0)
}

// givenMCPToolsDiscoveredLongAgo arranges a WARM but STALE cache entry (older
// than the 24 h freshness window): the prelude must serve it (no pre-request
// dial) and refresh after the answer.
func givenMCPToolsDiscoveredLongAgo(ctx context.Context, server string) error {
	return scenarioFrom(ctx).writeMCPToolCache(server, 48*time.Hour)
}

// givenMCPServerStoppedAnswering closes the fake so a pre-request revalidation
// would fail — making a stale entry's "served from cache" behaviour falsifiable.
func givenMCPServerStoppedAnswering(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	fake := sc.mcpFake(server)
	if fake == nil {
		return fmt.Errorf("the MCP server %q must be started before it can stop", server)
	}
	fake.Close()
	return nil
}

// thenMCPToolsRemembered asserts the cache file holds an entry for the server
// (the cold-discovery write happened).
func thenMCPToolsRemembered(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	p := filepath.Join(sc.home, mcpCacheFileName)
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("tellme must remember the server's tools, but the cache file is unreadable: %w", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("the cache file must be valid JSON: %w", err)
	}
	// FR-009 (fold F-087-5): the cache must never persist a credential.
	for _, banned := range []string{"token", "Token", "TOKEN", "Authorization", "Bearer", "secret"} {
		if strings.Contains(string(data), banned) {
			return fmt.Errorf("the MCP tool cache must not persist a credential; found %q in %s", banned, p)
		}
	}
	if _, ok := m[server]; !ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return fmt.Errorf("the cache must hold an entry for %q; got keys %v", server, keys)
	}
	return nil
}

// thenContactedMCPServer asserts the prelude/run dialed the server at least once
// (the cold-discovery witness).
func thenContactedMCPServer(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	fake := sc.mcpFake(server)
	if fake == nil {
		return fmt.Errorf("no fake MCP server %q was started", server)
	}
	if fake.ConnectionCount() == 0 {
		return fmt.Errorf("tellme was expected to contact the MCP server %q, but it received no request", server)
	}
	return nil
}
