package steps

import "github.com/cucumber/godog"

// T012 — When: the operator opens the interactive prompt
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T012.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
