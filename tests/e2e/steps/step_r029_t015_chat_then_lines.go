package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// T015 [BDD-RED] — Then: the lines of "{name}" are:
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the lines of "([^"]*)" are:$`, thenLines)
	})
}

// thenLines (必查 權威狀態): the lines of file {name} equal the table's rows (in
// order).
func thenLines(ctx context.Context, name string, table *godog.Table) error {
	sc := scenarioFrom(ctx)
	data, err := os.ReadFile(filepath.Join(sc.workDir, filepath.FromSlash(name)))
	if err != nil {
		return fmt.Errorf("could not read %q: %w", name, err)
	}
	body := strings.TrimSuffix(string(data), "\n")
	var got []string
	if body != "" {
		got = strings.Split(body, "\n")
	}
	want := make([]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		want = append(want, row.Cells[0].Value)
	}
	if len(got) != len(want) {
		return fmt.Errorf("the lines of %q = %v; want %v", name, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			return fmt.Errorf("the lines of %q = %v; want %v", name, got, want)
		}
	}
	return nil
}
