package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T008 — Given: a previous run with the persona "{persona}" and the prompt "{prompt}" reported an estimated payload status
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a previous run with the persona "([^"]*)" and the prompt "([^"]*)" reported an estimated payload status$`, givenPreviousRunEstimate)
	})
}

// givenPreviousRunEstimate (安排): against a throwaway COPY of the runtime home,
// set PERSON to {persona}, run `tellme "{prompt}"` once, parse the pre-flight
// `~<n>` estimate, and store it on the scenario; the real home (history, config)
// is left untouched.
func givenPreviousRunEstimate(ctx context.Context, persona, prompt string) error {
	return scenarioFrom(ctx).previousRunEstimate(unescapeText(persona), unescapeText(prompt))
}
