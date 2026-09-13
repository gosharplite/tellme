package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"
)

// T021 — Then: the earlier conversation records are unchanged
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the earlier conversation records are unchanged$`, thenRecordsUnchanged)
	})
}

// thenRecordsUnchanged (必查 權威狀態): the history file's previously stored lines
// are byte-unchanged — the file still begins with exactly the arranged lines
// (the run appends the current turn but must not rewrite earlier records).
func thenRecordsUnchanged(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	var wantPrefix strings.Builder
	for _, x := range sc.arrangedExchanges {
		line, err := json.Marshal(historyEntry{Prompt: x.prompt, Answer: x.answer})
		if err != nil {
			return err
		}
		wantPrefix.Write(line)
		wantPrefix.WriteByte('\n')
	}
	data, err := os.ReadFile(sc.historyFilePath())
	if err != nil {
		return fmt.Errorf("read session history: %w", err)
	}
	if !strings.HasPrefix(string(data), wantPrefix.String()) {
		return fmt.Errorf("the earlier conversation records changed: file=%q, want prefix %q", string(data), wantPrefix.String())
	}
	return nil
}
