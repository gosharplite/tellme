package steps

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cucumber/godog"
)

// T022 — Given: the session history already holds a tool-using exchange
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the session history already holds a tool-using exchange$`, givenHistoryToolUsingExchange)
	})
}

// widenedEntry is the persisted-turn shape carrying tool activity (round 008,
// specs/truth/data/data-model.dbml `history_entry` + `history_step`).
type widenedEntry struct {
	Prompt string        `json:"prompt"`
	Answer string        `json:"answer"`
	Calls  int           `json:"calls"`
	Steps  []widenedStep `json:"steps"`
}

type widenedStep struct {
	Tool      string `json:"tool"`
	Arguments string `json:"arguments"`
	Result    string `json:"result"`
}

// givenHistoryToolUsingExchange (怎麼做 / 權威狀態落地 / 回寫): create the per-mode
// session workspace and append one widened JSON line carrying prompt, answer, and
// an ordered steps array with one step. The arranged exchange is recorded so the
// operator-only Then can compute the expected listing.
func givenHistoryToolUsingExchange(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if err := os.MkdirAll(sc.historyDir(), 0o755); err != nil {
		return err
	}
	const prompt = "My name is Alice."
	const answer = "Noted."
	entry := widenedEntry{
		Prompt: prompt,
		Answer: answer,
		Calls:  2,
		Steps: []widenedStep{{
			Tool:      "read_files",
			Arguments: readArgs("notes.txt"),
			Result:    "the launch code is ORANGE",
		}},
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := os.WriteFile(sc.historyFilePath(), append(line, '\n'), 0o644); err != nil {
		return err
	}
	sc.recordExchange(prompt, answer)
	return nil
}
