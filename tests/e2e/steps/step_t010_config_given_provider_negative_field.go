package steps

import (
	"context"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"
)

// T010 — Given: a configuration "{config_path}" where selected provider "{provider}" has negative "{field}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configuration "([^"]*)" where selected provider "([^"]*)" has negative "([^"]*)"$`, givenProviderNegativeField)
	})
}

// givenProviderNegativeField writes a configuration where the selected provider specifies a negative integer for a budget field.
func givenProviderNegativeField(ctx context.Context, configPath, provider, field string) error {
	sc := scenarioFrom(ctx)
	provMap := map[string]interface{}{
		"TYPE":  "deepseek",
		"MODEL": "deepseek-v4-flash",
		"URL":   "https://api.deepseek.com",
		field:   -1,
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
