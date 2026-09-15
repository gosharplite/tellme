package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T011 [BDD-RED] — Then: the tool usage shows no tool has been used
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the tool usage shows no tool has been used$`, thenLogEmpty)
	})
}

// thenLogEmpty (必查 權威狀態): the tool-usage log is absent or holds zero records
// (a tool-less turn writes nothing — lazy creation).
func thenLogEmpty(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	recs, err := sc.toolUsageLogRecords()
	if err != nil {
		return err
	}
	if len(recs) != 0 {
		return fmt.Errorf("the tool-usage log holds %d records, want 0", len(recs))
	}
	return nil
}
