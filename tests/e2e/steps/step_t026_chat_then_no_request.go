package steps

import "github.com/cucumber/godog"

// T026 — Then: tellme sends no request to the provider "{provider}"
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T026.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
