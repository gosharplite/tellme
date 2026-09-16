package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T009 [BDD-RED] — Given: the working directory contains a file "{name}" that contains the line "{line}" twice
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains a file "([^"]*)" that contains the line "([^"]*)" twice$`, givenWorkdirLineTwice)
	})
}

// givenWorkdirLineTwice (怎麼做 / 權威狀態落地 / 回寫): create a file named {name}
// whose content is the line {line} twice (so the block is not unique).
func givenWorkdirLineTwice(ctx context.Context, name, line string) error {
	sc := scenarioFrom(ctx)
	l := unescapeText(line)
	return sc.writeWorkFile(name, l+"\n"+l+"\n")
}
