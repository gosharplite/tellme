package steps

import "github.com/cucumber/godog"

// T019 — Then: the interactive prompt offers the recent prompt "{prompt}"
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T019.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
