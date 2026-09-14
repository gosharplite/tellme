package steps

import (
	"context"
	"strconv"

	"github.com/cucumber/godog"
)

// T005 — Given: the configuration prices the active model with hit "{hit}", miss "{miss}", and completion "{comp}" per million tokens
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the configuration prices the active model with hit "([^"]*)", miss "([^"]*)", and completion "([^"]*)" per million tokens$`, givenConfigurationPrices)
	})
}

// givenConfigurationPrices (怎麼做 / 權威狀態落地 / 回寫): write `MODELS: { <active
// model>: { PRICING: { HIT, MISS, COMP } } }` into the default configuration.
func givenConfigurationPrices(ctx context.Context, hit, miss, comp string) error {
	sc := scenarioFrom(ctx)
	model, err := activeModel(sc)
	if err != nil {
		return err
	}
	h, err := strconv.ParseFloat(hit, 64)
	if err != nil {
		return err
	}
	m, err := strconv.ParseFloat(miss, 64)
	if err != nil {
		return err
	}
	c, err := strconv.ParseFloat(comp, 64)
	if err != nil {
		return err
	}
	return addPricing(sc, model, h, m, c)
}
