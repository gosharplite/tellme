package steps

import (
	"context"
	"strings"

	"github.com/cucumber/godog"
)

// T008 [BDD-RED] — Given: the working directory contains a file "{name}" whose lines are:
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains a file "([^"]*)" whose lines are:$`, givenWorkdirLines)
	})
}

// givenWorkdirLines (怎麼做 / 權威狀態落地 / 回寫): create a file named {name} whose
// content is the table's rows joined by "\n", each row newline-terminated.
func givenWorkdirLines(ctx context.Context, name string, table *godog.Table) error {
	sc := scenarioFrom(ctx)
	var b strings.Builder
	for _, row := range table.Rows {
		b.WriteString(row.Cells[0].Value)
		b.WriteByte('\n')
	}
	return sc.writeWorkFile(name, b.String())
}
