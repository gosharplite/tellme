package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/tools"
)

// round083MediaTool is a media-producing fake: it echoes the requested filepath
// as the media bytes, so the round's media parts can be asserted in CALL order.
type round083MediaTool struct{}

func (round083MediaTool) Name() string        { return "read_image" }
func (round083MediaTool) Description() string { return "fake media tool" }
func (round083MediaTool) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (round083MediaTool) Contract() tools.ToolContract { return tools.ToolContract{} }

func (round083MediaTool) Execute(_ context.Context, _ string, _ tools.ByteBudget) (string, error) {
	return "ok", nil
}

func (round083MediaTool) ExecuteMedia(_ context.Context, arguments string, _ tools.ByteBudget) (string, []tools.MediaPart, error) {
	var probe struct {
		Filepath string `json:"filepath"`
	}
	_ = json.Unmarshal([]byte(arguments), &probe)
	return "read " + probe.Filepath, []tools.MediaPart{{MIMEType: "image/png", Data: []byte(probe.Filepath)}}, nil
}

// mediaCall builds a read_image tool call for the given filepath.
func mediaCall(id, filepath string) llm.ToolCall {
	args, _ := json.Marshal(map[string]string{"filepath": filepath, "reason": "look"})
	return llm.ToolCall{ID: id, Name: "read_image", Arguments: string(args)}
}

// TestRunRoundScopedMedia_ThreeCalls_FoldedOnceAfterResults pins round 083
// (ADR 0055; closes #167): a round that makes THREE media-producing calls emits
// the round's `tool` results CONTIGUOUSLY and the round's media in ONE `user`
// message AFTER them — the order the OpenAI-compatible family requires (an
// assistant `tool_calls` block must be answered by its `tool` messages with no
// interleaved message). Pre-083 the media interleaved between the results.
func TestRunRoundScopedMedia_ThreeCalls_FoldedOnceAfterResults(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{mediaCall("c1", "a.png"), mediaCall("c2", "b.png"), mediaCall("c3", "c.png")}},
		{Text: "three squares"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(round083MediaTool{})}
	res, err := a.Run(context.Background(), "what do these show?", nil)
	if err != nil {
		t.Fatalf("a multi-media round must not fail: %v", err)
	}
	if res.Answer != "three squares" {
		t.Fatalf("answer = %q, want the final answer", res.Answer)
	}
	if len(gw.calls) != 2 {
		t.Fatalf("provider calls = %d, want 2", len(gw.calls))
	}
	assertRound083ThreeMediaLayout(t, gw.calls[1].Messages)
}

// assertRound083ThreeMediaLayout pins the emitted order for a 3-media round:
// [user(prompt), assistant(tool_calls:3), tool, tool, tool, user(media:3)] — the
// round's `tool` results CONTIGUOUS, the round's media in ONE trailing message.
func assertRound083ThreeMediaLayout(t *testing.T, msgs []llm.Message) {
	t.Helper()
	if len(msgs) != 6 {
		t.Fatalf("emitted messages = %d, want 6: %+v", len(msgs), msgs)
	}
	if msgs[0].Role != "user" {
		t.Errorf("msgs[0].Role = %q, want user (the prompt)", msgs[0].Role)
	}
	if msgs[1].Role != "assistant" || len(msgs[1].ToolCalls) != 3 {
		t.Fatalf("msgs[1] = role %q with %d tool calls, want assistant with 3", msgs[1].Role, len(msgs[1].ToolCalls))
	}
	// Contiguity: messages 2..4 are ALL `tool` (no interleaved user/media).
	for i := 2; i <= 4; i++ {
		if msgs[i].Role != "tool" {
			t.Fatalf("msgs[%d].Role = %q, want tool — a round's tool results must be contiguous: %+v", i, msgs[i].Role, msgs)
		}
	}
	if msgs[2].ToolCallID != "c1" || msgs[3].ToolCallID != "c2" || msgs[4].ToolCallID != "c3" {
		t.Errorf("tool results out of call order: %q,%q,%q", msgs[2].ToolCallID, msgs[3].ToolCallID, msgs[4].ToolCallID)
	}
	// The round's media rides ONE `user` message AFTER the results.
	if msgs[5].Role != "user" || len(msgs[5].Media) != 3 {
		t.Fatalf("msgs[5] = role %q with %d media parts, want user with 3", msgs[5].Role, len(msgs[5].Media))
	}
	for i, want := range []string{"a.png", "b.png", "c.png"} {
		if got := string(msgs[5].Media[i].Data); got != want {
			t.Errorf("round media part %d = %q, want %q (call order)", i, got, want)
		}
	}
}

// TestRunRoundScopedMedia_SingleCall_ByteIdentical pins the N == 1 case (spec
// I-3): the emitted order stays exactly `assistant(tool_calls), tool(result),
// user(media)` — the shipped single-image shape, unchanged by round 083.
func TestRunRoundScopedMedia_SingleCall_ByteIdentical(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{mediaCall("c1", "shot.png")}},
		{Text: "a red square"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(round083MediaTool{})}
	if _, err := a.Run(context.Background(), "what is in shot.png?", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	msgs := gw.calls[1].Messages
	// [user(prompt), assistant(tool_calls:1), tool, user(media:1)]
	if len(msgs) != 4 {
		t.Fatalf("emitted messages = %d, want 4: %+v", len(msgs), msgs)
	}
	if msgs[1].Role != "assistant" || len(msgs[1].ToolCalls) != 1 {
		t.Fatalf("msgs[1] = %+v, want assistant with 1 tool call", msgs[1])
	}
	if msgs[2].Role != "tool" {
		t.Fatalf("msgs[2].Role = %q, want tool", msgs[2].Role)
	}
	if msgs[3].Role != "user" || len(msgs[3].Media) != 1 {
		t.Fatalf("msgs[3] = role %q with %d media, want user with 1", msgs[3].Role, len(msgs[3].Media))
	}
}

// TestRunRoundScopedMedia_NoMedia_Unchanged pins that a media-free round emits no
// media message (spec edge case).
func TestRunRoundScopedMedia_NoMedia_Unchanged(t *testing.T) {
	gw := &fakeGateway{responses: []llm.Response{
		{ToolCalls: []llm.ToolCall{
			{ID: "c1", Name: "read_files", Arguments: `{"filepaths":["a.txt"]}`},
			{ID: "c2", Name: "read_files", Arguments: `{"filepaths":["b.txt"]}`},
		}},
		{Text: "done"},
	}}
	a := &AgentLoop{Gateway: gw, Registry: tools.NewRegistry(fakeTool{name: "read_files", result: "x"})}
	if _, err := a.Run(context.Background(), "read them", nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, m := range gw.calls[1].Messages {
		if len(m.Media) > 0 {
			t.Fatalf("a media-free round must emit no media message: %+v", gw.calls[1].Messages)
		}
	}
}
