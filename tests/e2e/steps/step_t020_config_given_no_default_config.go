package steps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T020 — Given: no configuration exists at the default location
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^no configuration exists at the default location$`, givenNoDefaultConfig)
	})
}

// givenNoDefaultConfig ensures the default configuration path for the effective
// mode ($TELL_ME_HOME/configs/<mode>.yaml) is absent (怎麼做 / 權威狀態落地: no
// default configuration is discoverable).
func givenNoDefaultConfig(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	mode := "butler"
	if m, ok := sc.envOverrides["TELL_ME_MODE"]; ok && m != "" {
		mode = m
	}
	return os.RemoveAll(sc.homePath(filepath.Join("configs", mode+".yaml")))
}
