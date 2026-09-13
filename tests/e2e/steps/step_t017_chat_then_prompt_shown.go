package steps

import "github.com/cucumber/godog"

// T017 — Then: the interactive prompt is shown
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T017.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
