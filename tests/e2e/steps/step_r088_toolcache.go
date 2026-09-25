package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 088 (ADR 0059) — the header-routing + cache-location steps.

func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a remote MCP server "([^"]*)" that offers a tool "([^"]*)" whose "([^"]*)" argument is routed as a header$`, givenMCPServerHeaderRoutedTool)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to use the MCP tool "([^"]*)" from the server "([^"]*)" with the reason "([^"]*)" and the argument "([^"]*)" set to "([^"]*)" and then answers with "([^"]*)"$`, givenProviderAsksMCPToolWithArgument)
		ctx.Then(`^tellme remembered the tools of the MCP server "([^"]*)" in the session workspace$`, thenMCPToolsRememberedInWorkspace)
	})
}

// givenMCPServerHeaderRoutedTool arranges a fake MCP server advertising {tool}
// with the {arg} property annotated `x-mcp-header` (the GitHub MCP server's
// shape). The SDK server then REQUIRES the client to send `Mcp-Param-{arg}` and
// rejects a call without it — so a cached tool that skips `tools/list` fails.
func givenMCPServerHeaderRoutedTool(ctx context.Context, server, tool, arg string) error {
	sc := scenarioFrom(ctx)
	fake := sc.startMCPFake(server, mcptest.Options{Tool: tool, Result: "ok", HeaderRouted: arg})
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL()})
	return nil
}

// givenProviderAsksMCPToolWithArgument scripts the fake provider to return a tool
// call naming the namespaced MCP tool `mcp_{server}_{tool}` carrying the round-056
// envelope `{"reason":"{reason}","MCP_PAYLOAD":{"{arg}":"{val}"}}`, then a final
// answer {answer}; and writes the resolvable default configuration. The
// `MCP_PAYLOAD` carries the header-routed argument so the SDK can lift it onto
// the `Mcp-Param-*` header (round 088).
func givenProviderAsksMCPToolWithArgument(ctx context.Context, provider, tool, server, reason, arg, val, answer string) error {
	sc := scenarioFrom(ctx)
	payload, err := json.Marshal(map[string]string{arg: val})
	if err != nil {
		return err
	}
	env := mcpEnvelopeJSON(unescapeText(reason), string(payload))
	name := "mcp_" + server + "_" + tool
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: name, Arguments: env},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedTool = name
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

// thenMCPToolsRememberedInWorkspace asserts the cache file lives in the per-mode
// SESSION WORKSPACE (output/<mode>/mcp-toolcache.json) and NOT at the home root
// (round 088 / ADR 0059).
func thenMCPToolsRememberedInWorkspace(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	data, err := os.ReadFile(sc.mcpCacheWorkspacePath())
	if err != nil {
		return fmt.Errorf("the cache must live in the session workspace %s: %w", sc.mcpCacheWorkspacePath(), err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("the cache file must be valid JSON: %w", err)
	}
	if _, ok := m[server]; !ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return fmt.Errorf("the workspace cache must hold an entry for %q; got %v", server, keys)
	}
	if _, err := os.Stat(sc.mcpCacheHomeRootPath()); err == nil {
		return fmt.Errorf("the cache must NOT live at the home root (%s)", sc.mcpCacheHomeRootPath())
	}
	return nil
}
