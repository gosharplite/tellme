package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T013 — Then: the captured standard output carries no terminal decoration
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the captured standard output carries no terminal decoration$`, thenCapturedStdoutNoDecoration)
	})
}

// thenCapturedStdoutNoDecoration (必查 呈現結果): the captured standard output
// contains no ANSI escape sequences (presentation suppressed when stdout is not
// a terminal).
func thenCapturedStdoutNoDecoration(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if strings.Contains(sc.stdout, "\x1b[") {
		return fmt.Errorf("stdout carried terminal decoration: %q", sc.stdout)
	}
	return nil
}
