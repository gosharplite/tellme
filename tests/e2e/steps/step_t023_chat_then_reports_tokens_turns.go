package steps

import "github.com/cucumber/godog"

// T023 — Then: the interactive prompt reports the session's token usage and turn count
// Foundational landing skeleton (round-015 T009); the registration + handler
// land with T023.
func init() {
	registrars = append(registrars, func(_ *godog.ScenarioContext) {})
}
