package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T019 [BDD-RED] — Then: the final model request reports no tool step marker
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the final model request reports no tool step marker$`, thenNoMarkerForFinalRequest)
	})
}

// thenNoMarkerForFinalRequest (必查 呈現結果): the bound-reached final request
// renders its frame but NO `[Tool Engine]` line — the last `[Tool Engine]` line
// precedes the last `╭─⠿ Turn N` frame (round 034 FR-001 / G9).
func thenNoMarkerForFinalRequest(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := stderrLines(sc.stderr)
	lastFrame := -1
	for i, l := range lines {
		if strings.Contains(l, turnHeaderMarker) {
			lastFrame = i
		}
	}
	if lastFrame < 0 {
		return fmt.Errorf("standard error carried no `╭─⠿ Turn N` frame; stderr=%q", sc.stderr)
	}
	lastEngine := lastToolEngineIndex(sc.stderr)
	if lastEngine < 0 {
		return fmt.Errorf("standard error carried no `[Tool Engine]` line to contrast the final request; stderr=%q", sc.stderr)
	}
	if lastEngine > lastFrame {
		return fmt.Errorf("the final model request reported a `[Tool Engine]` marker (line %d after frame %d); stderr=%q", lastEngine, lastFrame, sc.stderr)
	}
	return nil
}
