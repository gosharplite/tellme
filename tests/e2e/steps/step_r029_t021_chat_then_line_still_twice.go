package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// T021 [BDD-RED] — Then: the file "{name}" still contains the line "{line}" twice
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the file "([^"]*)" still contains the line "([^"]*)" twice$`, thenLineStillTwice)
	})
}

// thenLineStillTwice (必查 權威狀態): the file {name} still contains the line {line}
// twice — the refused edit did not touch it.
func thenLineStillTwice(ctx context.Context, name, line string) error {
	sc := scenarioFrom(ctx)
	data, err := os.ReadFile(filepath.Join(sc.workDir, filepath.FromSlash(name)))
	if err != nil {
		return fmt.Errorf("could not read %q: %w", name, err)
	}
	want := unescapeText(line)
	count := 0
	body := strings.TrimSuffix(string(data), "\n")
	if body != "" {
		for _, l := range strings.Split(body, "\n") {
			if l == want {
				count++
			}
		}
	}
	if count != 2 {
		return fmt.Errorf("the file %q contains the line %q %d times; want 2", name, want, count)
	}
	return nil
}
