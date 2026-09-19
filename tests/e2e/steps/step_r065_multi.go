package steps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 065 (ADR 0035; closes #132) E2E steps: a Gemini model round that asks for
// MORE THAN ONE tool at once. The run must complete and the recorded Vertex
// request must carry the round's tool results in ONE `user` turn (the batched
// `functionResponse` shape Vertex requires) — the pre-065 shape split them across
// turns and Vertex rejected the request with a 400.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured Gemini provider "([^"]*)" whose endpoint asks tellme, in one step, to read "([^"]*)" and "([^"]*)" and then answers with "([^"]*)"$`, givenGeminiReadsTwo)
		ctx.Given(`^a configured Gemini provider "([^"]*)" that can take images and whose endpoint asks tellme, in one step, to read the image files "([^"]*)" and "([^"]*)" and then answers with "([^"]*)"$`, givenGeminiVisionReadsTwoImages)
		ctx.Then(`^the Gemini provider received the answer to both read requests together$`, thenGeminiReceivedBothTogether)
	})
}

// round065ReadArgs builds a read_files tool's argument object (filepaths + a required reason).
func round065ReadArgs(paths ...string) string {
	b, _ := json.Marshal(map[string]any{"filepaths": paths, "reason": "read for the round-065 check"})
	return string(b)
}

// round065ImageArgs builds a read_image tool's argument object (a SINGLE
// `filepath` + a required reason).
func round065ImageArgs(path string) string {
	b, _ := json.Marshal(map[string]any{"filepath": path, "reason": "look at the picture for the round-065 check"})
	return string(b)
}

// arrangeGemini scripts the given replies on a Vertex-shaped fake, writes the
// gemini config + service-account key, and optionally declares VISION.
func arrangeGemini(sc *scenarioContext, provider, answer string, vision bool, replies ...fakeprovider.Reply) error {
	f := sc.newFake()
	f.VertexMode()
	f.Script(append(replies, fakeprovider.Reply{Answer: unescapeText(answer)})...)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	keyPath, err := sc.writeServiceAccountKey("secrets/key.json", f.URL()+"/token")
	if err != nil {
		return err
	}
	if err := sc.writeGeminiConfig(provider, "gemini-3.8-flash", f.URL(), keyPath, 40960); err != nil {
		return err
	}
	return sc.setSelectedProviderVision(vision)
}

// givenGeminiReadsTwo arranges a Vertex provider whose ONE model response carries
// TWO read_files tool calls (the #132 reproduction on the non-media path).
func givenGeminiReadsTwo(ctx context.Context, provider, first, second, answer string) error {
	sc := scenarioFrom(ctx)
	return arrangeGemini(sc, provider, answer, false, fakeprovider.Reply{Tools: []fakeprovider.ToolRequest{
		{Name: "read_files", Arguments: round065ReadArgs(first)},
		{Name: "read_files", Arguments: round065ReadArgs(second)},
	}})
}

// givenGeminiVisionReadsTwoImages arranges a Vertex provider (VISION: true) whose
// ONE model response carries TWO read_image tool calls.
func givenGeminiVisionReadsTwoImages(ctx context.Context, provider, first, second, answer string) error {
	sc := scenarioFrom(ctx)
	return arrangeGemini(sc, provider, answer, true, fakeprovider.Reply{Tools: []fakeprovider.ToolRequest{
		{Name: "read_image", Arguments: round065ImageArgs(first)},
		{Name: "read_image", Arguments: round065ImageArgs(second)},
	}})
}

// thenGeminiReceivedBothTogether asserts the recorded Vertex request carries the
// round's tool results in ONE `user` turn whose `functionResponse` parts number
// TWO (the batched shape ADR 0035 requires). It counts the parts per turn so a
// regression to the pre-065 split (two single-part turns) reds.
func thenGeminiReceivedBothTogether(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	var decoded struct {
		Contents []struct {
			Role  string           `json:"role"`
			Parts []map[string]any `json:"parts"`
		} `json:"contents"`
	}
	if err := json.Unmarshal([]byte(lastBody(sc)), &decoded); err != nil {
		return fmt.Errorf("the recorded Vertex request is unreadable: %w", err)
	}
	// Find the turn(s) carrying functionResponse parts.
	batched := 0
	turns := 0
	for _, turn := range decoded.Contents {
		fr := 0
		for _, part := range turn.Parts {
			if _, ok := part["functionResponse"]; ok {
				fr++
			}
		}
		if fr > 0 {
			turns++
			if fr == 2 {
				batched++
			}
		}
	}
	if turns != 1 || batched != 1 {
		return fmt.Errorf("the round's tool results must share ONE turn with TWO functionResponse parts; found %d response turn(s), %d batched (want 1 and 1): %s", turns, batched, lastBody(sc))
	}
	return nil
}
