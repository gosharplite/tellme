package steps

import (
	"context"
	"strconv"

	"github.com/cucumber/godog"
)

// T012 — When: the operator asks tellme to list the last {count} messages
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.When(`^the operator asks tellme to list the last (\d+) messages$`, whenListLastMessages)
	})
}

// whenListLastMessages runs `tellme -l {count}` (no -c, no positional prompt)
// under the current environment (怎麼做 / 回寫: capture exit code, stdout, stderr).
func whenListLastMessages(ctx context.Context, count int) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-l", strconv.Itoa(count)}
	sc.run()
	return nil
}
