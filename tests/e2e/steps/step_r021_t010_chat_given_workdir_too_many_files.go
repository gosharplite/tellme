package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T010 [BDD-RED] — Given: the working directory contains more files than tellme reads in one request
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains more files than tellme reads in one request$`, givenWorkdirTooManyFiles)
	})
}

// givenWorkdirTooManyFiles (怎麼做 / 權威狀態落地): create more than 50 files (the
// read_files per-call cap) in the subprocess working directory.
func givenWorkdirTooManyFiles(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	for i := 0; i < 51; i++ {
		p := filepath.Join(sc.workDir, r021TooManyFileName(i))
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// r021TooManyFileName names the i-th fixture file the too-many scenarios arrange
// and request (kept consistent between T010 and T012).
func r021TooManyFileName(i int) string { return fmt.Sprintf("file%02d.txt", i) }
