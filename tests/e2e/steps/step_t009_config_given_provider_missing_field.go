package steps

import (
	"context"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"
)

// T009 — Given: a configuration "{config_path}" where selected provider "{provider}" is missing "{field}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configuration "([^"]*)" where selected provider "([^"]*)" is missing "([^"]*)"$`, givenProviderMissingField)
	})
}

// givenProviderMissingField writes a configuration where the selected provider omits the specified mandatory field.
func givenProviderMissingField(ctx context.Context, configPath, provider, field string) error {
	sc := scenarioFrom(ctx)
	provMap := map[string]interface{}{
		"TYPE":  "deepseek",
		"MODEL": "deepseek-v4-flash",
		"URL":   "https://api.deepseek.com",
	}
	delete(provMap, field)
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
