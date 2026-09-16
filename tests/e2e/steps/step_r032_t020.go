package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/infrastructure/mcp"
)

// T020 [BDD-RED] — Then: tellme reported on stderr that the MCP server "{server}" could not be reached
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme reported on stderr that the MCP server "([^"]*)" could not be reached$`, thenMCPUnreachableWarned)
	})
}

// thenMCPUnreachableWarned (必查 呈現結果): stderr carries the single-sourced
// skip warning naming {server}, and it is a warning — NOT the frozen class
// phrase (the run must not fail because a server was unreachable).
func thenMCPUnreachableWarned(ctx context.Context, server string) error {
	sc := scenarioFrom(ctx)
	want := mcp.UnreachableWarning(server)
	if !strings.Contains(sc.stderr, want) {
		return fmt.Errorf("stderr did not carry the MCP skip warning %q; stderr=%q", want, sc.stderr)
	}
	if strings.Contains(sc.stderr, "the tool request failed") || strings.Contains(sc.stderr, "the provider request failed") {
		return fmt.Errorf("the skip was reported as a class phrase, not a warning; stderr=%q", sc.stderr)
	}
	return nil
}
