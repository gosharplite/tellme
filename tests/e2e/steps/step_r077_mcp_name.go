package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// Round 077 (issue #155 / ADR 0049) — the offered MCP declaration makes the
// tool's CALLABLE wire name positively discoverable: its description carries
// tellme's call-name note naming `mcp_<server>_<tool>` (the server's own text is
// relayed unchanged, added-to not rewritten). Observed on the fake-provider wire.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the offered tool "([^"]*)" from the MCP server "([^"]*)" names its callable wire name in the description$`, thenOfferedMCPToolNamesCallableName)
	})
}

// thenOfferedMCPToolNamesCallableName asserts the recorded declaration for
// `mcp_{server}_{tool}` carries the callable wire name in its description.
func thenOfferedMCPToolNamesCallableName(ctx context.Context, tool, server string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no provider fake recorded a request")
	}
	var body any
	if err := json.Unmarshal([]byte(f.LastBody()), &body); err != nil {
		return fmt.Errorf("could not decode the recorded request body: %w", err)
	}
	want := "mcp_" + server + "_" + tool
	decl := findDeclarationNode(body, want)
	if decl == nil {
		return fmt.Errorf("the request did not offer the declaration %q", want)
	}
	name, _ := decl["name"].(string)
	desc, _ := decl["description"].(string)
	if name != want {
		return fmt.Errorf("declaration name = %q, want %q", name, want)
	}
	if !strings.Contains(desc, `"`+want+`"`) {
		return fmt.Errorf("the offered declaration's description does not name the callable wire name %q; description=%q", want, desc)
	}
	return nil
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
