package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T007 [BDD-RED] — Then: the tool report is separated from the answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the tool report is separated from the answer$`, thenToolReportSeparated)
	})
}

// thenToolReportSeparated (必查 呈現結果): in the MERGED (stdout+stderr) capture,
// exactly one blank line separates the last `[HH:MM:SS] [Tool] …` line from the
// answer bytes (the round-010 cross-stream witness).
func thenToolReportSeparated(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	if sc.scriptedAnswer == "" {
		return fmt.Errorf("scenario did not record a scripted answer (tool-separation witness)")
	}
	m := mergedView(sc)
	lines := strings.Split(m, "\n")

	lastTool := -1
	for i, ln := range lines {
		if _, ok := parseToolLog(ln); ok {
			lastTool = i
		}
	}
	if lastTool < 0 {
		return fmt.Errorf("merged capture carried no tool report; merged=%q", m)
	}
	answerLine := -1
	for i := lastTool + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], sc.scriptedAnswer) {
			answerLine = i
			break
		}
	}
	if answerLine < 0 {
		return fmt.Errorf("merged capture carried no answer %q after the tool report; merged=%q", sc.scriptedAnswer, m)
	}
	if lines[lastTool+1] != "" || answerLine != lastTool+2 {
		return fmt.Errorf("expected exactly one blank line between the tool report and the answer; merged=%q", m)
	}
	return nil
}
