package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// Round-012 review RF1 — Then: the reading announcement is reported on the
// diagnostic output.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the reading announcement is reported on the diagnostic output$`, thenReadingAnnouncementOnStderr)
	})
}

// thenReadingAnnouncementOnStderr asserts the interactive multi-line hint is
// reported on standard error — the diagnostic stream — and NOT on standard
// output (必查 呈現結果; the hint literal is single-sourced from the product).
func thenReadingAnnouncementOnStderr(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stderr, cli.MultiLineHint) {
		return fmt.Errorf("the diagnostic output carries no reading announcement; stderr=%q", sc.stderr)
	}
	if strings.Contains(sc.stdout, cli.MultiLineHint) {
		return fmt.Errorf("the reading announcement leaked onto standard output; stdout=%q", sc.stdout)
	}
	return nil
}
