package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T009 [BDD-RED] — Then: the tool loop added no blank line before the answer
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the tool loop added no blank line before the answer$`, thenNoBlankBeforeAnswer)
	})
}

// thenNoBlankBeforeAnswer (必查 呈現結果): in the MERGED (stdout+stderr) capture,
// the answer is NOT immediately preceded by a blank line — the line before the
// answer is the last diagnostic line, not an empty one (FR-007 negative; carried
// on the non-chrome `-i` surface so it isolates round-022's blank from the
// round-017 frame gap).
func thenNoBlankBeforeAnswer(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.merged == "" {
		sc.captureMerged()
	}
	if sc.scriptedAnswer == "" {
		return fmt.Errorf("scenario did not record a scripted answer (blank-line witness)")
	}
	m := mergedView(sc)
	lines := strings.Split(m, "\n")
	answerLine := -1
	for i, ln := range lines {
		if strings.Contains(ln, sc.scriptedAnswer) {
			answerLine = i
			break
		}
	}
	if answerLine < 0 {
		return fmt.Errorf("merged capture carried no answer %q; merged=%q", sc.scriptedAnswer, m)
	}
	if answerLine > 0 && strings.TrimSpace(lines[answerLine-1]) == "" {
		return fmt.Errorf("a blank line preceded the answer; merged=%q", m)
	}
	return nil
}
