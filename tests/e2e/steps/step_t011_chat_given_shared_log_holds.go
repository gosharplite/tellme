package steps

import "github.com/cucumber/godog"

// T011 — Given: the shared prompt log already holds "{prompt}"
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T011.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
