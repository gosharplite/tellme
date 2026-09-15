package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cucumber/godog"
)

// T003 — Given: the session history already holds a tool-using exchange carrying the provider token "{token}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the session history already holds a tool-using exchange carrying the provider token "([^"]*)"$`, givenSignedToolUsingExchange)
	})
}

// signedStepFixture / signedEntryFixture mirror the persisted record shape
// (specs/truth/data/data-model.dbml, `history_step`) so the Given writes a turn
// whose tool step carries a provider signature.
type signedStepFixture struct {
	Tool      string `json:"tool"`
	Arguments string `json:"arguments"`
	Result    string `json:"result"`
	Signature string `json:"signature,omitempty"`
}

type signedEntryFixture struct {
	Prompt string              `json:"prompt"`
	Answer string              `json:"answer"`
	Calls  int                 `json:"calls"`
	Steps  []signedStepFixture `json:"steps,omitempty"`
}

// givenSignedToolUsingExchange (怎麼做 / 權威狀態落地 / 回寫): append one JSON line
// into the active session history whose single tool step carries {token}.
func givenSignedToolUsingExchange(ctx context.Context, token string) error {
	return writeToolUsingExchange(ctx, token)
}

// writeToolUsingExchange appends a tool-using turn to the active history; an
// empty token omits the signature field (the provider-neutral shape).
func writeToolUsingExchange(ctx context.Context, token string) error {
	sc := scenarioFrom(ctx)
	if sc == nil {
		return fmt.Errorf("no scenario context")
	}
	if err := os.MkdirAll(sc.historyDir(), 0o755); err != nil {
		return err
	}
	entry := signedEntryFixture{
		Prompt: "What is the launch code? Read notes.txt to find out.",
		Answer: "The launch code is ORANGE",
		Calls:  2,
		Steps: []signedStepFixture{{
			Tool:      "read_files",
			Arguments: `{"path":"notes.txt"}`,
			Result:    "the launch code is ORANGE",
			Signature: token,
		}},
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(sc.historyFilePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
