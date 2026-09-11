package steps

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"
)

// T005 — Given: the provider "{provider}" in configuration "{config_path}" includes custom headers:
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the provider "([^"]*)" in configuration "([^"]*)" includes custom headers:$`, givenProviderHeaders)
	})
}

// givenProviderHeaders updates the specified provider in the configuration file with the custom headers.
func givenProviderHeaders(ctx context.Context, provider, configPath string, table *godog.Table) error {
	sc := scenarioFrom(ctx)
	raw, err := os.ReadFile(sc.homePath(configPath))
	if err != nil {
		return err
	}
	var doc map[string]interface{}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return err
	}
	provs, ok := doc["PROVIDERS"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("PROVIDERS section not found or invalid in %s", configPath)
	}
	provEntry, ok := provs[provider].(map[string]interface{})
	if !ok {
		return fmt.Errorf("provider %s not found in PROVIDERS", provider)
	}
	headers := make(map[string]interface{})
	for _, row := range table.Rows[1:] {
		if len(row.Cells) >= 2 {
			k := strings.TrimSpace(row.Cells[0].Value)
			v := strings.TrimSpace(row.Cells[1].Value)
			headers[k] = v
		}
	}
	provEntry["HEADERS"] = headers
	data, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	return sc.writeFile(configPath, data)
}
