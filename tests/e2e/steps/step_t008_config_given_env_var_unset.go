package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T008 — Given: the environment variable "{var_name}" is unset
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the environment variable "([^"]*)" is unset$`, givenEnvVarUnset)
	})
}

// givenEnvVarUnset ensures the given environment variable is removed from the scenario subprocess environment.
func givenEnvVarUnset(ctx context.Context, varName string) error {
	sc := scenarioFrom(ctx)
	if sc.envUnset == nil {
		sc.envUnset = map[string]bool{}
	}
	sc.envUnset[varName] = true
	delete(sc.envOverrides, varName)
	return nil
}
