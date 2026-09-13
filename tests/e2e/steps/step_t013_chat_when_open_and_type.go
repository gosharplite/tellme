package steps

import "github.com/cucumber/godog"

// T013 — When: the operator opens the interactive prompt and types "{query}"
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T013.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
