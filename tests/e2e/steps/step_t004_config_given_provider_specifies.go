package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"
)

// T004 — Given: a well-formed configuration "{config_path}" where provider "{provider}" specifies:
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a well-formed configuration "([^"]*)" where provider "([^"]*)" specifies:$`, givenProviderSpecifies)
	})
}

// givenProviderSpecifies parses a DataTable of field-value pairs and writes a configuration YAML
// with the specified provider attributes.
func givenProviderSpecifies(ctx context.Context, configPath, provider string, table *godog.Table) error {
	sc := scenarioFrom(ctx)
	provMap := make(map[string]interface{})
	for _, row := range table.Rows[1:] {
		if len(row.Cells) >= 2 {
			k := strings.TrimSpace(row.Cells[0].Value)
			v := strings.TrimSpace(row.Cells[1].Value)
			var val interface{} = v
			var intVal int
			if _, err := fmt.Sscanf(v, "%d", &intVal); err == nil && (k == "MAX_TOKENS" || k == "THINKING_BUDGET") {
				val = intVal
			}
			provMap[k] = val
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
