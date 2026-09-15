package steps

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// r021ReadLimitBytes is the read_files per-file cap the fixture must exceed
// (100000 bytes) to force the truncation branch.
const r021ReadLimitBytes = 100000

// T008 [BDD-RED] — Given: the working directory contains a file "{name}" whose text is longer than the read limit
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains a file "([^"]*)" whose text is longer than the read limit$`, givenWorkdirLargeFile)
	})
}

// givenWorkdirLargeFile (怎麼做 / 權威狀態落地): create {name} with content larger
// than the read cap so tellme must truncate it.
func givenWorkdirLargeFile(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	content := strings.Repeat("x", r021ReadLimitBytes+1)
	return os.WriteFile(filepath.Join(sc.workDir, filepath.FromSlash(name)), []byte(content), 0o644)
}
