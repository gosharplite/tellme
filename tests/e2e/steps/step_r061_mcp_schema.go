package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"

	"github.com/gosharplite/tellme/internal/infrastructure/llm/gemini"
	mcptest "github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
)

// Round 061 (issue #127 / ADR 0031) — a remote server's argument marks must not
// reach a provider whose tool-definition reader is closed (Vertex/Gemini).
//
//   - Given: the server advertises its tool with an ANNOTATED schema (the GitHub
//     server's `x-mcp-header` shape) — the exact input that used to 400 every
//     Gemini turn.
//   - Then: the offered DECLARATION carries no server-side mark (any family), no
//     keyword outside the provider's supported surface (the owner set, folded
//     per review F-061-1), and still describes the server's real arguments.
//
// The live-request acceptance (that Vertex now ACCEPTS the projected payload) is
// proven by the round-061 research probe (ADR 0031) + the hermetic unit pins; the
// E2E observes the declaration the fake provider recorded.
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

// offeredDeclaration returns the recorded request's DECLARATION for
// `mcp_{server}_{tool}` (shape-agnostic across the OpenAI-shaped and
// Vertex-shaped fake bodies) as decoded JSON — the scope all three Thens judge
// (review F-061-3(3): a body-wide scan is a latent false-red).
func offeredDeclaration(sc *scenarioContext, tool, server string) (any, error) {
	f := sc.onlyFake()
	if f == nil {
		return nil, fmt.Errorf("no provider fake recorded a request")
	}
	var body any
	if err := json.Unmarshal([]byte(f.LastBody()), &body); err != nil {
		return nil, fmt.Errorf("could not decode the recorded request body: %w", err)
	}
	want := "mcp_" + server + "_" + tool
	if decl := findDeclaration(body, want); decl != nil {
		return decl, nil
	}
	return nil, fmt.Errorf("the request did not offer the declaration %q", want)
}

// findDeclaration searches a decoded request body for the object whose `name` is
// `want` and which carries `parameters`.
func findDeclaration(node any, want string) any {
	switch n := node.(type) {
	case map[string]any:
		if name, ok := n["name"].(string); ok && name == want {
			if _, has := n["parameters"]; has {
				return n["parameters"]
			}
		}
		for _, v := range n {
			if found := findDeclaration(v, want); found != nil {
				return found
			}
		}
	case []any:
		for _, e := range n {
			if found := findDeclaration(e, want); found != nil {
				return found
			}
		}
	}
	return nil
}

// walkKeywords visits every schema KEYWORD in a declaration (structure-aware:
// `properties` values are visited, their NAMES are opaque; data values are not).
func walkKeywords(node any, visit func(string)) {
	m, ok := node.(map[string]any)
	if !ok {
		return
	}
	for k, v := range m {
		visit(k)
		switch k {
		case "properties": // names are opaque
			if props, ok := v.(map[string]any); ok {
				for _, sub := range props {
					walkKeywords(sub, visit)
				}
			}
		case "items", "additionalProperties":
			walkKeywords(v, visit)
		case "oneOf", "allOf":
			if list, ok := v.([]any); ok {
				for _, e := range list {
					walkKeywords(e, visit)
				}
			}
		}
	}
}

// thenOfferedToolNoServerMark asserts the offered declaration carries no
// vendor-extension mark — the family-agnostic floor (S-6), scoped to the
// declaration.
func thenOfferedToolNoServerMark(ctx context.Context, tool, server string) error {
	sc := scenarioFrom(ctx)
	decl, err := offeredDeclaration(sc, tool, server)
	if err != nil {
		return err
	}
	var offender string
	walkKeywords(decl, func(k string) {
		if strings.HasPrefix(k, "x-") || k == "$schema" {
			offender = k
		}
	})
	if offender != "" {
		return fmt.Errorf("the offered declaration still carries the server-side mark %q", offender)
	}
	return nil
}

// thenOfferedToolNoUnsupportedKeyword asserts CONTAINMENT against the provider's
// named owner set (not a hardcoded deny-list — review F-061-1/TD-061-2).
func thenOfferedToolNoUnsupportedKeyword(ctx context.Context, tool, server string) error {
	sc := scenarioFrom(ctx)
	decl, err := offeredDeclaration(sc, tool, server)
	if err != nil {
		return err
	}
	var offenders []string
	walkKeywords(decl, func(k string) {
		if !gemini.SupportedSchemaKeys()[k] {
			offenders = append(offenders, k)
		}
	})
	if len(offenders) > 0 {
		sort.Strings(offenders)
		return fmt.Errorf("the offered declaration carries keyword(s) the provider cannot read: %v", offenders)
	}
	return nil
}

// thenOfferedToolStillDescribesArguments asserts the projection was a carve-out:
// each listed argument is a DECLARED property of the offered declaration and
// carries a description (the row's own claim — review F-061-3(2)).
func thenOfferedToolStillDescribesArguments(ctx context.Context, tool, server, args string) error {
	sc := scenarioFrom(ctx)
	decl, err := offeredDeclaration(sc, tool, server)
	if err != nil {
		return err
	}
	node, _ := decl.(map[string]any)
	// An MCP tool's offered declaration is tellme's envelope: the SERVER's own
	// arguments live inside the `MCP_PAYLOAD` property's subschema (round 056 /
	// ADR 0025 D1).
	envProps, _ := node["properties"].(map[string]any)
	if payload, ok := envProps["MCP_PAYLOAD"].(map[string]any); ok {
		node = payload
	}
	props, _ := node["properties"].(map[string]any)
	if props == nil {
		return fmt.Errorf("the offered declaration has no properties; got %v", decl)
	}
	for _, arg := range strings.Split(args, ",") {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		sub, ok := props[arg].(map[string]any)
		if !ok {
			return fmt.Errorf("the offered declaration does not declare the argument %q; declared=%v", arg, keysOf(props))
		}
		if desc, _ := sub["description"].(string); strings.TrimSpace(desc) == "" {
			return fmt.Errorf("the offered argument %q lost its description; got %v", arg, sub)
		}
	}
	return nil
}

// keysOf lists a map's keys (sorted, for deterministic messages).
func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
