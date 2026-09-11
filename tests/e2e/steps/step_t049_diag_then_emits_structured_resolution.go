package steps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

// T049 — Then: tellme emits the resolution status as structured output
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme emits the resolution status as structured output$`, thenEmitsResolvedJSON)
	})
}

// diagnosticJSON is the pinned --json object contract
// (specs/truth/features/cli/diagnostics/dsl.md, grill #4 fix A2).
type diagnosticJSON struct {
	Status           string `json:"status"`
	Reason           string `json:"reason"`
	RuntimeHome      string `json:"runtime_home"`
	SessionWorkspace string `json:"session_workspace"`
}

// thenEmitsResolvedJSON (必查 呈現結果): stdout parses as the pinned JSON
// object with status resolved, and runtime_home / session_workspace equal the
// resolved values.
func thenEmitsResolvedJSON(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	var obj diagnosticJSON
	if err := json.Unmarshal([]byte(sc.stdout), &obj); err != nil {
		return fmt.Errorf("stdout is not the pinned JSON object: %v; got %q", err, sc.stdout)
	}
	if obj.Status != "resolved" {
		return fmt.Errorf("status = %q, want %q", obj.Status, "resolved")
	}
	if obj.RuntimeHome != sc.home {
		return fmt.Errorf("runtime_home = %q, want %q", obj.RuntimeHome, sc.home)
	}
	if want := sc.expectedWorkspace(); obj.SessionWorkspace != want {
		return fmt.Errorf("session_workspace = %q, want %q", obj.SessionWorkspace, want)
	}
	return nil
}
