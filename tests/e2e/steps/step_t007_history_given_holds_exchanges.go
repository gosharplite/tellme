package steps

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/cucumber/godog"
)

// T007 — Given: the session history already holds the exchanges:
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the session history already holds the exchanges:$`, givenHistoryHoldsExchanges)
	})
}

// givenHistoryHoldsExchanges arranges a prior conversation in the session
// history (怎麼做 / 權威狀態落地): create the per-mode session workspace and write
// one JSON line per table row into history.jsonl, in table order. The arranged
// exchanges are recorded for a listing Then.
func givenHistoryHoldsExchanges(ctx context.Context, table *godog.Table) error {
	sc := scenarioFrom(ctx)
	if err := os.MkdirAll(sc.historyDir(), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for i, row := range table.Rows {
		if i == 0 {
			continue // header: prompt | answer
		}
		prompt := row.Cells[0].Value
		answer := row.Cells[1].Value
		line, err := json.Marshal(historyEntry{Prompt: prompt, Answer: answer})
		if err != nil {
			return err
		}
		b.Write(line)
		b.WriteByte('\n')
		sc.recordExchange(prompt, answer)
	}
	return os.WriteFile(sc.historyFilePath(), []byte(b.String()), 0o644)
}
