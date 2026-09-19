package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"

	mcptest "github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
)

// Round 061 (issue #127 / ADR 0031) — a remote server's argument marks must not
// reach a provider whose tool-definition reader is closed (Vertex/Gemini).
//
//   - Given: the server advertises its tool with an ANNOTATED schema (the GitHub
//     server's `x-mcp-header` shape) — the exact input that used to 400 every
//     Gemini turn.
//   - Then: the offered declaration carries no server-side mark (any family), and
//     — under the gemini family — no keyword the provider cannot read at all.
//
// The live-request acceptance (that Vertex now ACCEPTS the projected payload) is
// proven by the round-061 research probe (ADR 0031) + the hermetic unit pin; the
// E2E observes the wire bytes the fake provider recorded.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a remote MCP server "([^"]*)" that offers a tool "([^"]*)" answering "([^"]*)" whose arguments carry a server-side mark$`, givenMCPServerAnnotatedTool)
		ctx.Given(`^a configured gemini provider "([^"]*)" whose endpoint reports the offered tools and then answers with "([^"]*)"$`, givenGeminiProviderReportsTools)
		ctx.Then(`^the offered tool "([^"]*)" from the MCP server "([^"]*)" carries no server-side mark$`, thenOfferedToolNoServerMark)
		ctx.Then(`^the offered tool "([^"]*)" from the MCP server "([^"]*)" carries no keyword the provider cannot read$`, thenOfferedToolNoUnsupportedKeyword)
		ctx.Then(`^the offered tool "([^"]*)" from the MCP server "([^"]*)" still describes the arguments "([^"]*)"$`, thenOfferedToolStillDescribesArguments)
	})
}

// givenMCPServerAnnotatedTool arranges a reachable fake MCP server whose tool
// advertises the annotated (GitHub-shaped) schema and records it as
// MCP_SERVERS.{server}.
func givenMCPServerAnnotatedTool(ctx context.Context, server, tool, result string) error {
	sc := scenarioFrom(ctx)
	fake := sc.startMCPFake(server, mcptest.Options{
		Tool:   tool,
		Result: unescapeText(result),
		Schema: mcptest.AnnotatedSchema(),
	})
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL()})
	return nil
}

// givenGeminiProviderReportsTools arranges a Vertex-shaped gemini provider (the
// family whose declaration reader is closed) whose endpoint records the offered
// tools and then answers with {answer}.
func givenGeminiProviderReportsTools(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.VertexMode()
	f.Script(fakeprovider.Reply{Answer: unescapeText(answer)})
	sc.registerFake(provider, f)
	keyPath, err := sc.writeServiceAccountKey("secrets/key.json", f.URL()+"/token")
	if err != nil {
		return err
	}
	return sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), keyPath, 40960)
}

// declaredToolSchema returns the parameters of the offered declaration
// `mcp_{server}_{tool}` as raw JSON, shape-agnostic across the OpenAI-shaped and
// Vertex-shaped fake bodies.
func declaredToolSchema(sc *scenarioContext, tool, server string) (string, error) {
	f := sc.onlyFake()
	if f == nil {
		return "", fmt.Errorf("no provider fake recorded a request")
	}
	body := f.LastBody()
	want := "mcp_" + server + "_" + tool
	if !strings.Contains(body, want) {
		return "", fmt.Errorf("the request did not offer %q", want)
	}
	return body, nil
}

// thenOfferedToolNoServerMark asserts the offered declaration carries no
// vendor-extension mark — the family-agnostic floor (S-6).
func thenOfferedToolNoServerMark(ctx context.Context, tool, server string) error {
	sc := scenarioFrom(ctx)
	body, err := declaredToolSchema(sc, tool, server)
	if err != nil {
		return err
	}
	for _, mark := range []string{"x-mcp-header", `"$schema"`, "x-"} {
		if strings.Contains(body, mark) {
			return fmt.Errorf("the offered declaration still carries the server-side mark %q", mark)
		}
	}
	return nil
}

// thenOfferedToolNoUnsupportedKeyword asserts the offered declaration carries no
// keyword the (closed-wire) provider cannot read — the provider-side projection
// (S-1/S-2). The set is the live probe's reject list (ADR 0031).
func thenOfferedToolNoUnsupportedKeyword(ctx context.Context, tool, server string) error {
	sc := scenarioFrom(ctx)
	body, err := declaredToolSchema(sc, tool, server)
	if err != nil {
		return err
	}
	unsupported := []string{
		"x-mcp-header", "$schema", "$ref", "$defs", "definitions", "const",
		"examples", "deprecated", "readOnly", "writeOnly", "multipleOf",
		"uniqueItems", "anyOf",
	}
	for _, k := range unsupported {
		if strings.Contains(body, `"`+k+`"`) {
			return fmt.Errorf("the offered declaration still carries the unsupported keyword %q", k)
		}
	}
	return nil
}

// thenOfferedToolStillDescribesArguments asserts the projection was a carve-out:
// the listed argument names (and their descriptions) survive in the offered
// declaration.
func thenOfferedToolStillDescribesArguments(ctx context.Context, tool, server, args string) error {
	sc := scenarioFrom(ctx)
	body, err := declaredToolSchema(sc, tool, server)
	if err != nil {
		return err
	}
	for _, arg := range strings.Split(args, ",") {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		if !strings.Contains(body, `"`+arg+`"`) {
			return fmt.Errorf("the offered declaration lost the argument %q", arg)
		}
	}
	return nil
}
