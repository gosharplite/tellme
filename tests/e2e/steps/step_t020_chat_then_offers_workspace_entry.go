package steps

import "github.com/cucumber/godog"

// T020 — Then: the interactive prompt offers the workspace entry "{entry}"
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T020.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
