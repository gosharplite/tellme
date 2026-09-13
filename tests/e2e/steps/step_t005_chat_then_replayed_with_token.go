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

// replaySignatureByTool extracts, from a recorded request body, the provider
// token echoed on each assistant tool call — the Vertex `thoughtSignature` on a
// `contents` functionCall part (the OpenAI family carries none).
func replaySignatureByTool(body string) map[string]string {
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
	out := map[string]string{}
	for _, c := range probe.Contents {
		for _, p := range c.Parts {
			if p.FunctionCall != nil {
				out[p.FunctionCall.Name] = p.ThoughtSignature
			}
		}
	}
	return out
}

// thenReplayedWithProviderToken (必查 呈現結果 / 權威狀態): a recorded request
// replays the earlier tool step {tool} with its provider token {token} echoed
// verbatim (the Gemini 3 `thoughtSignature`).
func thenReplayedWithProviderToken(ctx context.Context, tool, token string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	for i := 0; i < f.RequestCount(); i++ {
		sigs := replaySignatureByTool(f.BodyAt(i))
		if sig, ok := sigs[tool]; ok {
			if sig != token {
				return fmt.Errorf("the replayed tool step %q carried token %q, want %q", tool, sig, token)
			}
			return nil
		}
	}
	return fmt.Errorf("the request did not replay the earlier tool step %q carrying the token; body=%s", tool, f.LastBody())
}
