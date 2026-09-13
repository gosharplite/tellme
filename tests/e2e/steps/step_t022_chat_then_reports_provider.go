package steps

import "github.com/cucumber/godog"

// T022 — Then: the interactive prompt reports the active provider "{provider}"
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T022.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
