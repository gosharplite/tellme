package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T014 [BDD-RED] — Then: the content of "{name}" is still exactly "{content}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the content of "([^"]*)" is still exactly "([^"]*)"$`, thenContentUnchanged)
	})
}

// thenContentUnchanged (必查 權威狀態): the bytes of file {name} still equal
// {content} exactly — a refused create must not modify the file.
func thenContentUnchanged(ctx context.Context, name, content string) error {
	sc := scenarioFrom(ctx)
	data, err := os.ReadFile(filepath.Join(sc.workDir, filepath.FromSlash(name)))
	if err != nil {
		return fmt.Errorf("could not read %q: %w", name, err)
	}
	want := unescapeText(content)
	if string(data) != want {
		return fmt.Errorf("the content of %q = %q; want still exactly %q", name, string(data), want)
	}
	return nil
}
