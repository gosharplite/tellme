package steps

import "github.com/cucumber/godog"

// T024 — Then: the shared prompt log records the prompt "{prompt}"
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T024.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
