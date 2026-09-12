package steps

import (
	"context"

	"github.com/cucumber/godog"
)

// T012 — Given: a well-formed configuration "{config_path}" whose rendered width is "{width}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a well-formed configuration "([^"]*)" whose rendered width is "([^"]*)"$`, givenWellFormedConfigRenderedWidth)
	})
}

// givenWellFormedConfigRenderedWidth writes a resolvable YAML file at
// {home}/{config_path} whose top-level WRAP_WIDTH is {width} (怎麼做 / 權威狀態落地:
// the file carries the rendered width value; 回寫: the configuration file).
func givenWellFormedConfigRenderedWidth(ctx context.Context, configPath, width string) error {
	sc := scenarioFrom(ctx)
	body := "MODE: butler\n" +
		"PERSON: \"e2e persona\"\n" +
		"SELECTED_PROVIDER: test-model\n" +
		"WRAP_WIDTH: " + width + "\n" +
		"PROVIDERS:\n" +
		"  test-model:\n" +
		"    TYPE: deepseek\n" +
		"    MODEL: deepseek-v4-flash\n" +
		"    URL: https://api.deepseek.com\n"
	return sc.writeFile(configPath, []byte(body))
}
