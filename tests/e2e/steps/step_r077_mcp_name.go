package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	mcptest "github.com/gosharplite/tellme/internal/infrastructure/mcp/mcptest"
)

// Round 077 (issue #155 / ADR 0049) — the offered MCP declaration makes the
// tool's CALLABLE wire name positively discoverable: its description carries
// tellme's call-name note naming `mcp_<server>_<tool>` (the server's own text is
// relayed unchanged, added-to not rewritten). Observed on the fake-provider wire.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a remote MCP server "([^"]*)" that offers a tool "([^"]*)" with no description answering "([^"]*)"$`, givenMCPServerOffersToolNoDescription)
		ctx.Then(`^the offered tool "([^"]*)" from the MCP server "([^"]*)" names its callable wire name in the description$`, thenOfferedMCPToolNamesCallableName)
		ctx.Then(`^the offered tool "([^"]*)" from the MCP server "([^"]*)" falls back to its callable name, not the bare one$`, thenOfferedMCPToolFallbackNamesCallable)
	})
}

// givenMCPServerOffersToolNoDescription arranges a reachable fake MCP server whose
// tool advertises a genuinely EMPTY description — the E2E carrier for tellme's
// empty-description fallback (TD-1 / FR-004).
func givenMCPServerOffersToolNoDescription(ctx context.Context, server, tool, result string) error {
	sc := scenarioFrom(ctx)
	fake := sc.startMCPFake(server, mcptest.Options{Tool: tool, Result: unescapeText(result), EmptyDescription: true})
	sc.addMCPServer(server, mcpServerEntry{URL: fake.URL()})
	return nil
}

// thenOfferedMCPToolNamesCallableName asserts the recorded declaration for
// `mcp_{server}_{tool}` carries the callable wire name in its description.
func thenOfferedMCPToolNamesCallableName(ctx context.Context, tool, server string) error {
	sc := scenarioFrom(ctx)
	desc, name, err := offeredMCPDescription(sc, tool, server)
	if err != nil {
		return err
	}
	want := "mcp_" + server + "_" + tool
	if name != want {
		return fmt.Errorf("declaration name = %q, want %q", name, want)
	}
	if !strings.Contains(desc, `"`+want+`"`) {
		return fmt.Errorf("the offered declaration's description does not name the callable wire name %q; description=%q", want, desc)
	}
	return nil
}

// thenOfferedMCPToolFallbackNamesCallable is the FORCE-BEARING carrier for the
// empty-description fallback (TD-1 / FR-004): the description's FALLBACK body
// must name the `mcp_{server}_{tool}` wire name and the source server, and must
// NOT present the bare upstream name as callable.
func thenOfferedMCPToolFallbackNamesCallable(ctx context.Context, tool, server string) error {
	sc := scenarioFrom(ctx)
	desc, want, err := offeredMCPDescription(sc, tool, server)
	if err != nil {
		return err
	}
	// Strip tellme's call-name note prefix, leaving the fallback body the round
	// synthesizes for an empty server description.
	note := `Call this tool as "` + want + `". `
	body := strings.TrimPrefix(desc, note)
	if !strings.Contains(body, want) {
		return fmt.Errorf("the fallback body must name the callable wire name %q; body=%q", want, body)
	}
	if !strings.Contains(body, "from server "+server) {
		return fmt.Errorf("the fallback body must state the source server %q; body=%q", server, body)
	}
	if strings.Contains(body, "MCP tool "+tool+" from server") {
		return fmt.Errorf("the fallback body must NOT present the bare upstream name %q as callable; body=%q", tool, body)
	}
	return nil
}

// offeredMCPDescription returns the recorded declaration's description and name
// for `mcp_{server}_{tool}` (shape-agnostic across the OpenAI/Vertex fake bodies).
func offeredMCPDescription(sc *scenarioContext, tool, server string) (desc, name string, err error) {
	f := sc.onlyFake()
	if f == nil {
		return "", "", fmt.Errorf("no provider fake recorded a request")
	}
	var body any
	if jerr := json.Unmarshal([]byte(f.LastBody()), &body); jerr != nil {
		return "", "", fmt.Errorf("could not decode the recorded request body: %w", jerr)
	}
	want := "mcp_" + server + "_" + tool
	decl := findDeclarationNode(body, want)
	if decl == nil {
		return "", "", fmt.Errorf("the request did not offer the declaration %q", want)
	}
	name, _ = decl["name"].(string)
	desc, _ = decl["description"].(string)
	return desc, name, nil
}

// findDeclarationNode returns the declaration object whose `name` is `want` and
// which carries `parameters` — the shape-agnostic match used by the round-061
// helper, here returning the node itself (so its description is visible).
func findDeclarationNode(node any, want string) map[string]any {
	switch n := node.(type) {
	case map[string]any:
		if name, ok := n["name"].(string); ok && name == want {
			if _, has := n["parameters"]; has {
				return n
			}
		}
		for _, v := range n {
			if found := findDeclarationNode(v, want); found != nil {
				return found
			}
		}
	case []any:
		for _, e := range n {
			if found := findDeclarationNode(e, want); found != nil {
				return found
			}
		}
	}
	return nil
}
