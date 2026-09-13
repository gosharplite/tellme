package steps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

// T005 — Then: the request replayed the earlier tool step "{tool}" carrying the provider token "{token}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the request replayed the earlier tool step "([^"]*)" carrying the provider token "([^"]*)"$`, thenReplayedWithProviderToken)
	})
}

// replayToolCall is one replayed assistant tool call: its tool name and the
// provider token echoed on it (empty for the OpenAI-compatible family).
type replayToolCall struct {
	Name      string
	Signature string
}

// replayedToolCalls returns, from a recorded request body, the replayed assistant
// tool calls **in order** — the Vertex `contents` functionCall parts with their
// `thoughtSignature` (the OpenAI family carries none). Indexing by ORDER rather
// than by tool name keeps a turn that replays the same tool more than once
// unambiguous (review PR #35 forward consideration: name-keyed maps collapse
// repeated calls to one entry).
func replayedToolCalls(body string) []replayToolCall {
	var probe struct {
		Contents []struct {
			Parts []struct {
				FunctionCall *struct {
					Name string `json:"name"`
				} `json:"functionCall"`
				ThoughtSignature string `json:"thoughtSignature"`
			} `json:"parts"`
		} `json:"contents"`
	}
	_ = json.Unmarshal([]byte(body), &probe)
	var calls []replayToolCall
	for _, c := range probe.Contents {
		for _, p := range c.Parts {
			if p.FunctionCall != nil {
				calls = append(calls, replayToolCall{Name: p.FunctionCall.Name, Signature: p.ThoughtSignature})
			}
		}
	}
	return calls
}

// thenReplayedWithProviderToken (必查 呈現結果 / 權威狀態): a recorded request
// replays the earlier tool step {tool} with its provider token {token} echoed
// verbatim (the Gemini 3 `thoughtSignature`). The check scans the replayed calls
// in order and matches a call of {tool} whose token equals {token}, so a turn
// that replays several calls is handled independently of their names.
func thenReplayedWithProviderToken(ctx context.Context, tool, token string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	sawTool := false
	lastSignature := ""
	for i := 0; i < f.RequestCount(); i++ {
		for _, tc := range replayedToolCalls(f.BodyAt(i)) {
			if tc.Name != tool {
				continue
			}
			sawTool = true
			lastSignature = tc.Signature
			if tc.Signature == token {
				return nil
			}
		}
	}
	if sawTool {
		return fmt.Errorf("the replayed tool step %q carried token %q, want %q", tool, lastSignature, token)
	}
	return fmt.Errorf("the request did not replay the earlier tool step %q carrying the token; body=%s", tool, f.LastBody())
}
