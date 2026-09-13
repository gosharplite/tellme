package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T016 — Then: no reading announcement is reported
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^no reading announcement is reported$`, thenNoReadingAnnouncement)
	})
}

// readingHint is the POSIX multi-line hint the interactive reader prints (pinned
// in specs/truth/features/cli/chat/dsl.md).
const readingHint = "[Reading multi-line input. Press Ctrl+C to cancel, or Ctrl+D to send]"

// thenNoReadingAnnouncement (必查 呈現結果): neither stdout nor stderr carries the
// multi-line reading hint — the interactive reader never engages on a
// non-terminal input (round-012).
func thenNoReadingAnnouncement(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if strings.Contains(sc.stdout, readingHint) || strings.Contains(sc.stderr, readingHint) {
		return fmt.Errorf("a reading announcement was reported; stdout=%q stderr=%q", sc.stdout, sc.stderr)
	}
	return nil
}
