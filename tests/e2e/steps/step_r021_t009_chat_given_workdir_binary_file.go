package steps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// T009 [BDD-RED] — Given: the working directory contains a binary file "{name}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains a binary file "([^"]*)"$`, givenWorkdirBinaryFile)
	})
}

// givenWorkdirBinaryFile (怎麼做 / 權威狀態落地): create {name} whose bytes are not
// valid UTF-8 text (a NUL byte is present), so the reader must report it binary.
func givenWorkdirBinaryFile(ctx context.Context, name string) error {
	sc := scenarioFrom(ctx)
	return os.WriteFile(filepath.Join(sc.workDir, filepath.FromSlash(name)), []byte{0x00, 0x01, 0x02, 0xff}, 0o644)
}
