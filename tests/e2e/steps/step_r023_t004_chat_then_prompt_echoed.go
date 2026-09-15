package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T004 [BDD-RED] — Then: the submitted prompt "{prompt}" is echoed on the diagnostic output
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the submitted prompt "([^"]*)" is echoed on the diagnostic output$`, thenPromptEchoed)
	})
}

// thenPromptEchoed (必查 呈現結果): the captured stderr carries the submitted
// prompt AFTER the editor frame's last border rune (the frame was cleared on
// submit) and BEFORE the input-capture acknowledgement. It must not be written to
// standard output.
func thenPromptEchoed(ctx context.Context, prompt string) error {
	sc := scenarioFrom(ctx)
	p := unescapeText(prompt)
	ack := strings.Index(sc.stderr, "Input captured")
	if ack < 0 {
		return fmt.Errorf("no input-capture acknowledgement; stderr=%q", sc.stderr)
	}
	base := 0
	tail := sc.stderr
	if last := strings.LastIndex(sc.stderr, "└"); last >= 0 {
		base = last
		tail = sc.stderr[last:]
	}
	rel := strings.Index(tail, p)
	if rel < 0 {
		return fmt.Errorf("the submitted prompt %q was not echoed after the editor frame cleared; stderr=%q", p, sc.stderr)
	}
	if base+rel > ack {
		return fmt.Errorf("the echoed prompt did not precede the acknowledgement; stderr=%q", sc.stderr)
	}
	if strings.Contains(sc.stdout, p) {
		return fmt.Errorf("the echoed prompt leaked onto standard output; stdout=%q", sc.stdout)
	}
	return nil
}
