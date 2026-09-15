package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T006 [BDD-RED] — Then: the run reported one tool-loop log line for each tool call
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the run reported one tool-loop log line for each tool call$`, thenOneLinePerCall)
	})
}

// thenOneLinePerCall (必查 呈現結果): the captured stderr carries exactly one
// `[HH:MM:SS] [Tool] …` line per tool call the run executed (the count of `[Tool]`
// lines equals the number of role:"tool" messages in the final provider request).
func thenOneLinePerCall(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	want := executedToolCount(sc.onlyFake())
	if want == 0 {
		return fmt.Errorf("scenario executed no tool call (no role:\"tool\" message recorded); stderr=%q", sc.stderr)
	}
	got := len(toolLogs(sc.stderr))
	if got != want {
		return fmt.Errorf("expected one tool-loop log line per tool call: got %d for %d calls; stderr=%q", got, want, sc.stderr)
	}
	return nil
}
