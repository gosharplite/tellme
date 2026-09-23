package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 083 (ADR 0055; closes #167) — a step that reads SEVERAL pictures. The
// loop folds a round's media ONCE, AFTER the per-call loop, so a round's `tool`
// results are CONTIGUOUS on the wire (the OpenAI-compatible family rejects a
// `tool` block interrupted by a media/user message). These steps script a
// single-response multi-tool-call round and assert the recorded request's layout.
//
// The contiguity half is single-owned by the SHARED `toolExchangeChronologyOK`
// (wire_tools.go), which this multi-image round is the carrier for (F-083-6) —
// no second contiguity predicate is defined here (TD-083-1).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" that can take images and whose endpoint asks tellme, in one step, to read "([^"]*)", "([^"]*)" and "([^"]*)" and then answers with "([^"]*)"$`, givenVisionProviderReadsThreeImages)
		ctx.Given(`^the session history already holds a turn in which the agent read "([^"]*)", "([^"]*)" and "([^"]*)"$`, givenHistoryHoldsImageReadTurn)

		ctx.Then(`^the tool results of the step were answered together, before the pictures were shown$`, thenToolResultsAnsweredTogether)
		ctx.Then(`^the step showed the three pictures together after the results$`, thenThreePicturesOneMessageAfterResults)
		ctx.Then(`^the replayed step answered every tool result together$`, thenReplayedStepContiguous)
	})
}

// round083ImageArgs builds a read_image argument object (a single `filepath` + a
// required reason).
func round083ImageArgs(path string) string {
	b, _ := json.Marshal(map[string]any{"filepath": path, "reason": "look at the picture"})
	return string(b)
}

// givenVisionProviderReadsThreeImages arranges a vision-enabled provider whose
// fake scripts ONE response carrying THREE `read_image` tool calls (for the three
// paths), then answers {answer} — the multi-media round of issue #167.
func givenVisionProviderReadsThreeImages(ctx context.Context, provider, first, second, third, answer string) error {
	sc := scenarioFrom(ctx)
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{Tools: []fakeprovider.ToolRequest{
			{Name: "read_image", Arguments: round083ImageArgs(first)},
			{Name: "read_image", Arguments: round083ImageArgs(second)},
			{Name: "read_image", Arguments: round083ImageArgs(third)},
		}},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_image"
	sc.registerFake(provider, f)
	if err := sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()}); err != nil {
		return err
	}
	return sc.setSelectedProviderVision(true)
}

// givenHistoryHoldsImageReadTurn appends ONE history entry whose steps are three
// `read_image` steps (one per file) — the resumed multi-media replay.
func givenHistoryHoldsImageReadTurn(ctx context.Context, first, second, third string) error {
	sc := scenarioFrom(ctx)
	steps := make([]history.Step, 0, 3)
	for _, p := range []string{first, second, third} {
		steps = append(steps, history.Step{
			Tool:      "read_image",
			Arguments: round083ImageArgs(p),
			Result:    "read " + p,
		})
	}
	entry := history.Entry{Prompt: "What do these pictures show?", Answer: "three squares", Calls: 1, Steps: steps}
	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(sc.historyDir(), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(sc.historyFilePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// r083Message is one recorded OpenAI-compatible message (content kept raw, so a
// media content ARRAY is preserved).
type r083Message struct {
	Role       string            `json:"role"`
	Content    json.RawMessage   `json:"content"`
	ToolCalls  []json.RawMessage `json:"tool_calls"`
	ToolCallID string            `json:"tool_call_id"`
}

// r083CountImages returns the number of inline `image_url` blocks in a message's
// content (0 for a plain string content).
func r083CountImages(content json.RawMessage) int {
	var parts []struct {
		Type     string `json:"type"`
		ImageURL struct {
			URL string `json:"url"`
		} `json:"image_url"`
	}
	if json.Unmarshal(content, &parts) != nil {
		return 0
	}
	n := 0
	for _, p := range parts {
		if p.Type == "image_url" && p.ImageURL.URL != "" {
			n++
		}
	}
	return n
}

// r083Messages decodes the recorded request's `messages` array (raw content).
func r083Messages(body string) ([]r083Message, error) {
	var req struct {
		Messages []r083Message `json:"messages"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return nil, fmt.Errorf("decode the recorded request: %w", err)
	}
	return req.Messages, nil
}

// thenToolResultsAnsweredTogether (必查 權威狀態): the round's `tool` results are
// contiguous (single-owned by the shared `toolExchangeChronologyOK`) AND any
// media message appears AFTER the last `tool` result.
func thenToolResultsAnsweredTogether(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if f := sc.onlyFake(); f == nil || !toolExchangeChronologyOK(f) {
		return fmt.Errorf("the round's `tool` results are not contiguous (a non-`tool` message sits between them)")
	}
	msgs, err := r083Messages(lastBody(sc))
	if err != nil {
		return err
	}
	lastTool := -1
	firstMedia := -1
	for i, m := range msgs {
		if m.Role == "tool" {
			lastTool = i
		}
		if firstMedia < 0 && r083CountImages(m.Content) > 0 {
			firstMedia = i
		}
	}
	if lastTool < 0 {
		return fmt.Errorf("the recorded request carries no tool results: %+v", msgs)
	}
	if firstMedia >= 0 && firstMedia < lastTool {
		return fmt.Errorf("a picture message at index %d precedes the last tool result (%d) — the pictures must be shown AFTER every result: %+v", firstMedia, lastTool, msgs)
	}
	return nil
}

// thenThreePicturesOneMessageAfterResults (必查 權威狀態): the round's three
// images ride ONE `user` message placed after the round's `tool` results.
func thenThreePicturesOneMessageAfterResults(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if f := sc.onlyFake(); f == nil || !toolExchangeChronologyOK(f) {
		return fmt.Errorf("the round's `tool` results are not contiguous (a non-`tool` message sits between them)")
	}
	msgs, err := r083Messages(lastBody(sc))
	if err != nil {
		return err
	}
	// The block opener: the last assistant carrying tool calls, and the run of
	// `tool` results that follows it.
	open := -1
	for i, m := range msgs {
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			open = i
		}
	}
	if open < 0 {
		return fmt.Errorf("no assistant tool-call block in the recorded request: %+v", msgs)
	}
	toolRun := 0
	for i := open + 1; i < len(msgs) && msgs[i].Role == "tool"; i++ {
		toolRun++
	}
	mediaIdx := open + 1 + toolRun
	if mediaIdx >= len(msgs) {
		return fmt.Errorf("no media message follows the round's %d tool result(s): %+v", toolRun, msgs)
	}
	mediaMsg := msgs[mediaIdx]
	if mediaMsg.Role != "user" {
		return fmt.Errorf("the message after the tool block is %q, want a user media message: %+v", mediaMsg.Role, msgs)
	}
	if n := r083CountImages(mediaMsg.Content); n != 3 {
		return fmt.Errorf("the round's media message carries %d picture(s), want 3 (one message, three images): %+v", n, msgs)
	}
	// No other media message: the three images must not be split across messages.
	total := 0
	for _, m := range msgs {
		total += r083CountImages(m.Content)
	}
	if total != 3 {
		return fmt.Errorf("the request carries %d pictures in total, want 3 (one message): %+v", total, msgs)
	}
	return nil
}

// thenReplayedStepContiguous (必查 權威狀態): the resumed request's replayed
// conversation answers the earlier multi-step turn's tool-call blocks
// contiguously (single-owned by the shared `toolExchangeChronologyOK`).
func thenReplayedStepContiguous(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil || !toolExchangeChronologyOK(f) {
		return fmt.Errorf("the replayed conversation's `tool` results are not contiguous")
	}
	msgs, err := r083Messages(lastBody(sc))
	if err != nil {
		return err
	}
	tools := 0
	for _, m := range msgs {
		if m.Role == "tool" {
			tools++
		}
	}
	if tools < 3 {
		return fmt.Errorf("the replayed request carries %d tool result(s), want the earlier turn's 3 steps: %+v", tools, msgs)
	}
	return nil
}
