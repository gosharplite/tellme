package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T006 [BDD-RED] — Then: the interactive prompt is cleared
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the interactive prompt is cleared$`, thenPromptCleared)
	})
}

// thenPromptCleared (必查 呈現結果): after reducing the captured diagnostic
// stream to its FINAL screen (applying carriage-return, erase, and cursor-up
// sequences — PR #51 principal review TD2), the editor frame (a `┌…└` block) does
// NOT survive the submit/abort. A naive raw-substring check would fail because
// earlier draw cycles remain in the buffer; the model-level pin in T007 is the
// authoritative clear.
func thenPromptCleared(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	rows := reduceTerminal(sc.stderr)
	if terminalHasBorder(rows) {
		return fmt.Errorf("the editor frame survived the submit/abort; final screen=%q", strings.Join(rows, "\n"))
	}
	return nil
}
