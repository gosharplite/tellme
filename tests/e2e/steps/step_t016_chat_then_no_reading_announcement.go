package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/cli"
)

// T016 — Then: no reading announcement is reported
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^no reading announcement is reported$`, thenNoReadingAnnouncement)
	})
}

// thenNoReadingAnnouncement (必查 呈現結果): neither stdout nor stderr carries the
// multi-line reading hint — the interactive reader never engages on a
// non-terminal input (round-012). The hint literal is single-sourced from the
// product (cli.MultiLineHint), so a product-side text change cannot silently make
// this guard vacuous (round-012 review TD1).
func thenNoReadingAnnouncement(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if strings.Contains(sc.stdout, cli.MultiLineHint) || strings.Contains(sc.stderr, cli.MultiLineHint) {
		return fmt.Errorf("a reading announcement was reported; stdout=%q stderr=%q", sc.stdout, sc.stderr)
	}
	return nil
}
