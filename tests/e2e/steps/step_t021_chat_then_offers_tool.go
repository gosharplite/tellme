package steps

import "github.com/cucumber/godog"

// T021 — Then: the interactive prompt offers the available tool "{tool}"
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T021.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
