package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T008 — Given: the working directory contains a file "{name}" whose text is "{content}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains a file "([^"]*)" whose text is "([^"]*)"$`, givenWorkdirFile)
	})
}

// givenWorkdirFile (怎麼做 / 權威狀態落地 / 回寫): create a file named {name} whose
// text is {content} in the subprocess working directory, so the read_files tool
// can read it. {content} is decoded through the round-005 escape convention.
func givenWorkdirFile(ctx context.Context, name, content string) error {
	sc := scenarioFrom(ctx)
	return sc.writeWorkFile(name, unescapeText(content))
}
