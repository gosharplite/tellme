package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T003 [BDD-ALIGN] — Then: the pre-flight payload line is separated from the answer by a single blank line
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the pre-flight payload line is separated from the answer by a single blank line$`, thenPayloadSingleBlank)
	})
}

// thenPayloadSingleBlank (必查 呈現結果): in the MERGED (stdout+stderr) capture,
// exactly ONE blank line separates the pre-flight payload line from the answer —
// on a no-tool turn the single blank is the round-017 frame gap (undoubled); a
// stray round-022 blank would make it two. Carried on the chrome `-i` surface
// (round 023; PR #50 review B1 re-anchor).
func thenPayloadSingleBlank(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	if sc.scriptedAnswer == "" {
		return fmt.Errorf("scenario did not record a scripted answer (single-blank witness)")
	}
	m := mergedView(sc)
	lines := strings.Split(m, "\n")

	payload := -1
	for i, ln := range lines {
		if strings.Contains(ln, "Payload: ~") {
			payload = i
		}
	}
	if payload < 0 {
		return fmt.Errorf("merged capture carried no pre-flight payload line; merged=%q", m)
	}
	answerLine := -1
	for i := payload + 1; i < len(lines); i++ {
		if strings.Contains(lines[i], sc.scriptedAnswer) {
			answerLine = i
			break
		}
	}
	if answerLine < 0 {
		return fmt.Errorf("merged capture carried no answer %q after the payload line; merged=%q", sc.scriptedAnswer, m)
	}
	blanks := 0
	for i := payload + 1; i < answerLine; i++ {
		if strings.TrimSpace(lines[i]) == "" {
			blanks++
			continue
		}
		return fmt.Errorf("non-blank content between the payload line and the answer (%q); merged=%q", lines[i], m)
	}
	if blanks != 1 {
		return fmt.Errorf("want exactly one blank line between the payload line and the answer, got %d; merged=%q", blanks, m)
	}
	return nil
}
