package steps

import "github.com/cucumber/godog"

// T018 — Then: the interactive prompt is not shown
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T018.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
