package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T011 — Then: the captured standard output contains the answer on a single line
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the captured standard output contains the answer on a single line$`, thenSingleLine)
	})
}

// thenSingleLine (必查 呈現結果): under -r the raw answer is emitted unwrapped —
// the captured stdout, with the single CLI-appended terminating newline removed,
// is one line (round-006 FR-007: the width is ignored under -r).
func thenSingleLine(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := strings.Split(strings.TrimRight(sc.stdout, "\n"), "\n")
	if len(lines) != 1 {
		return fmt.Errorf("want the answer on a single line, got %d lines: %q", len(lines), sc.stdout)
	}
	return nil
}
