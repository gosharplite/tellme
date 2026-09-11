package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T007 — Given: the environment variable "{var_name}" is set to "{var_value}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the environment variable "([^"]*)" is set to "([^"]*)"$`, givenEnvVarSet)
	})
}

// givenEnvVarSet sets the given environment variable for the scenario subprocess.
func givenEnvVarSet(ctx context.Context, varName, varValue string) error {
	scenarioFrom(ctx).setEnv(varName, varValue)
	return nil
}
