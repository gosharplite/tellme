// Round 097 (ADR 0067) — Then: the input capture acknowledgement is shown in blue.
//
// The chrome's FIFTH accent (an input-capture acknowledgement in a bright blue,
// whole line, timestamp included) is a raw-byte contract on the diagnostic
// stream, so this Then reads the captured stderr directly — the same discipline
// as the sibling round-054/057/058 assertions in step_r054_colour.go.
//
// The accent is gated by ADR 0023 D2 (chromeColour: terminal stderr + `-r` off +
// a prompt-bearing surface), which the Given `the diagnostics are shown at a
// terminal` drives through the TELL_ME_FORCE_STDERR_TTY seam.
package steps

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cucumber/godog"
)

// Round 097 (ADR 0067) — the blue acknowledgement accent.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the input capture acknowledgement is shown in blue$`, thenAckShownInBlue)
	})
}

// reBlueAck pins the WHOLE line inside the blue pair, timestamp included — so a
// wrap that started only at "Input captured" (leaving the `[HH:MM:SS] ` prefix
// plain) would redden, which is ADR 0067's whole-line clause.
var reBlueAck = regexp.MustCompile("\x1b\\[1;34m\\[[0-9]{2}:[0-9]{2}:[0-9]{2}\\] Input captured\\. Processing\\.\\.\\.\x1b\\[0m")

// thenAckShownInBlue (必查 呈現結果): the captured stderr carries the bright-blue
// SGR pair around the whole acknowledgement line.
//
// Should-not-happen (asserted by the sibling negative rows, not here): the accent
// must not be green, must not reach stdout, and must not enter `turns.log`.
func thenAckShownInBlue(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !reBlueAck.MatchString(sc.stderr) {
		return fmt.Errorf("the acknowledgement must be wrapped in the bright-blue pair "+
			"(\\x1b[1;34m…\\x1b[0m), timestamp included; stderr=%s", quoteForLog(sc.stderr))
	}
	return nil
}
