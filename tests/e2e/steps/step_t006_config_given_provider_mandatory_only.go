package steps

import (
	"context"
	"strings"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"
)

// T006 — Given: a well-formed configuration "{config_path}" where provider "{provider}" specifies only mandatory fields:
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a well-formed configuration "([^"]*)" where provider "([^"]*)" specifies only mandatory fields:$`, givenProviderMandatoryOnly)
	})
}

// givenProviderMandatoryOnly writes a configuration file with only mandatory provider fields.
func givenProviderMandatoryOnly(ctx context.Context, configPath, provider string, table *godog.Table) error {
	sc := scenarioFrom(ctx)
	provMap := make(map[string]interface{})
	for _, row := range table.Rows[1:] {
		if len(row.Cells) >= 2 {
			k := strings.TrimSpace(row.Cells[0].Value)
			v := strings.TrimSpace(row.Cells[1].Value)
			provMap[k] = v
		}
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
