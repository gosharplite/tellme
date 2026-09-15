package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/harness"
)

// T006 [BDD-RED] — Then: the interactive prompt is cleared
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt is cleared$`, thenPromptCleared)
	})
}

// thenPromptCleared (必查 呈現結果): reduce the captured diagnostic stream to its
// FINAL screen (applying carriage-return, erase, and cursor-up sequences) and
// assert the editor frame (a `┌…└` block) does NOT survive the submit/abort.
//
// A guard requires the frame to have PAINTED first (the raw stream carries a
// border rune), so a stream carrying no frame cannot be read as "cleared" — the
// assertion is not vacuous (PR #51 implementation-review REFACTOR). The
// model-level pin in T007 (`View() == ""`) is the authoritative clear.
func thenPromptCleared(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.ContainsAny(harness.StripANSI(sc.stderr), "┌└") {
		return fmt.Errorf("the editor frame was never painted (the cleared assertion would be vacuous); stderr=%q", sc.stderr)
	}
	rows := reduceTerminal(sc.stderr)
	if terminalHasBorder(rows) {
		return fmt.Errorf("the editor frame survived the submit/abort; final screen=%q", strings.Join(rows, "\n"))
	}
	return nil
}
