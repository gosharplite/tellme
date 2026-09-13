package steps

import "github.com/cucumber/godog"

// T025 — Then: the shared prompt log still holds only "{prompt}"
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T025.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
