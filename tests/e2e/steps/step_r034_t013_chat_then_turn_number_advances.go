package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// T013 [BDD-RED] — Then: the turn number advances from one frame to the next
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the turn number advances from one frame to the next$`, thenTurnNumberAdvances)
	})
}

// thenTurnNumberAdvances (必查 呈現結果): consecutive `╭─⠿ Turn N` frames carry
// strictly increasing, consecutive numbers (N, N+1, …) — the number advances
// within a prompt (round 034 FR-008).
func thenTurnNumberAdvances(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	nums := turnFrameNumbers(sc.stderr)
	if len(nums) < 2 {
		return fmt.Errorf("expected at least two frames to advance the number, got %d; stderr=%q", len(nums), sc.stderr)
	}
	for i := 1; i < len(nums); i++ {
		if nums[i] != nums[i-1]+1 {
			return fmt.Errorf("the frame number did not advance by one: %d then %d; stderr=%q", nums[i-1], nums[i], sc.stderr)
		}
	}
	return nil
}
