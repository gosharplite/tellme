package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T015 — Then: tellme used no tool
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme used no tool$`, thenUsedNoTool)
	})
}

// thenUsedNoTool (必查 呈現結果 / 權威狀態): the run invoked no tool — no recorded
// request carried a tool result or an assistant tool-call.
func thenUsedNoTool(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if anyToolActivity(f) {
		return fmt.Errorf("the run invoked a tool, but a prompt that needs no tool must not")
	}
	return nil
}
