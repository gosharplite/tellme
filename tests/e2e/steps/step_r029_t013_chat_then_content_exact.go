package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T013 [BDD-RED] — Then: the content of "{name}" is exactly "{content}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the content of "([^"]*)" is exactly "([^"]*)"$`, thenContentExact)
	})
}

// thenContentExact (必查 權威狀態): the bytes of file {name} equal {content} exactly.
func thenContentExact(ctx context.Context, name, content string) error {
	sc := scenarioFrom(ctx)
	data, err := os.ReadFile(filepath.Join(sc.workDir, filepath.FromSlash(name)))
	if err != nil {
		return fmt.Errorf("could not read %q: %w", name, err)
	}
	want := unescapeText(content)
	if string(data) != want {
		return fmt.Errorf("the content of %q = %q; want exactly %q", name, string(data), want)
	}
	return nil
}
