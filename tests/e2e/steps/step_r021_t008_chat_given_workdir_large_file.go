package steps

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// r024ReadBoundBytes exceeds the default tool result BYTE bound (= effectiveBudget
// = MAX_HISTORY_TOKENS, default 1000000) so the reader must truncate (round 024;
// the former 100000-byte per-file cap is retired).
const r024ReadBoundBytes = 1200000

// T011 [BDD-ALIGN] — Given: the working directory contains a file "{name}" whose text is longer than the read bound
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains a file "([^"]*)" whose text is longer than the read bound$`, givenWorkdirLargeFile)
	})
}

// givenWorkdirLargeFile (怎麼做 / 權威狀態落地): create {name} with content larger
// than the effective result byte bound so tellme truncates it.
func givenWorkdirLargeFile(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	content := strings.Repeat("x", r024ReadBoundBytes+1)
	return os.WriteFile(filepath.Join(sc.workDir, filepath.FromSlash(name)), []byte(content), 0o644)
}
