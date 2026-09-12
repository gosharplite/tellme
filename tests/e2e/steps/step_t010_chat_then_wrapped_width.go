package steps

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// T010 — Then: the captured standard output is wrapped so that no line is wider than {width} columns
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the captured standard output is wrapped so that no line is wider than (\d+) columns$`, thenWrappedWidth)
	})
}

// thenWrappedWidth (必查 呈現結果): every line of the captured standard output is
// at most {width} display columns wide, after ANSI stripping and normalising the
// renderer's leading margin whitespace (round-006 research Decision 7).
func thenWrappedWidth(ctx context.Context, widthStr string) error {
	sc := scenarioFrom(ctx)
	width, err := strconv.Atoi(widthStr)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(harness.StripANSI(sc.stdout), "\n") {
		normalized := strings.TrimLeft(line, " ")
		if n := len([]rune(normalized)); n > width {
			return fmt.Errorf("line %q is %d columns wide, want <= %d", normalized, n, width)
		}
	}
	return nil
}
