package steps

import (
	"context"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"
)

// T011 — Given: a well-formed configuration "{config_path}" where provider "{provider}" references unset variable "{var_name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a well-formed configuration "([^"]*)" where provider "([^"]*)" references unset variable "([^"]*)"$`, givenProviderUnsetVar)
	})
}

// givenProviderUnsetVar writes a configuration referencing an unset environment variable in API_KEY.
func givenProviderUnsetVar(ctx context.Context, configPath, provider, varName string) error {
	sc := scenarioFrom(ctx)
	if sc.envUnset == nil {
		sc.envUnset = map[string]bool{}
	}
	sc.envUnset[varName] = true
	delete(sc.envOverrides, varName)

	provMap := map[string]interface{}{
		"TYPE":    "openai",
		"MODEL":   "gpt-5.5",
		"URL":     "https://api.openai.com/v1",
		"API_KEY": "${" + varName + "}",
	}
	cfg := map[string]interface{}{
		"MODE":              "butler",
		"PERSON":            "e2e persona",
		"SELECTED_PROVIDER": provider,
		"PROVIDERS": map[string]interface{}{
			provider: provMap,
		},
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return sc.writeFile(configPath, data)
}
