package steps

import (
	"encoding/json"
	"strings"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round-008 tool-loop helpers: decode the fake provider's recorded request
// bodies and inspect the assistant tool-call / tool-result messages the agent
// loop sent. The fake records raw OpenAI-compatible bodies; these helpers expose
// only the fields the tool-loop Then steps assert on. Kept in ONE file so the
// per-sentence step files stay independent (Zero Shared Edits) and never collide
// on a top-level helper name.

// wireToolCall is one assistant tool-call request in a recorded message.
type wireToolCall struct {
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// wireToolMessage is one message in a recorded request body.
type wireToolMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []wireToolCall `json:"tool_calls"`
	ToolCallID string         `json:"tool_call_id"`
}

// decodeToolMessages decodes the `messages` array from a recorded request body.
func decodeToolMessages(body string) []wireToolMessage {
	var decoded struct {
		Messages []wireToolMessage `json:"messages"`
	}
	_ = json.Unmarshal([]byte(body), &decoded)
	return decoded.Messages
}

// toolRounds returns, per recorded request, its decoded messages.
func toolRounds(f *fakeprovider.Provider) [][]wireToolMessage {
	n := f.RequestCount()
	rounds := make([][]wireToolMessage, 0, n)
	for i := 0; i < n; i++ {
		rounds = append(rounds, decodeToolMessages(f.BodyAt(i)))
	}
	return rounds
}

// countToolIterations counts the recorded requests that carry a tool result (a
// role:"tool" message) — one per executed tool-iteration round.
func countToolIterations(f *fakeprovider.Provider) int {
	count := 0
	for _, msgs := range toolRounds(f) {
		for _, m := range msgs {
			if m.Role == "tool" {
				count++
				break
			}
		}
	}
	return count
}

// hasToolCall reports whether any recorded request carried an assistant
// tool-call naming tool whose arguments contain argSubstr ("" matches any).
func hasToolCall(f *fakeprovider.Provider, tool, argSubstr string) bool {
	for _, msgs := range toolRounds(f) {
		for _, m := range msgs {
			if m.Role != "assistant" {
				continue
			}
			for _, tc := range m.ToolCalls {
				if tc.Function.Name == tool && strings.Contains(tc.Function.Arguments, argSubstr) {
					return true
				}
			}
		}
	}
	return false
}

// lastToolResult returns the content of the last role:"tool" message recorded
// across all requests, or "" when none was recorded.
func lastToolResult(f *fakeprovider.Provider) string {
	result := ""
	for _, msgs := range toolRounds(f) {
		for _, m := range msgs {
			if m.Role == "tool" {
				result = m.Content
			}
		}
	}
	return result
}

// anyToolActivity reports whether any recorded request carried tool activity —
// an assistant tool-call or a tool result.
func anyToolActivity(f *fakeprovider.Provider) bool {
	for _, msgs := range toolRounds(f) {
		for _, m := range msgs {
			if m.Role == "tool" || len(m.ToolCalls) > 0 {
				return true
			}
		}
	}
	return false
}

// readArgs builds the read_files tool arguments JSON for a path.
func readArgs(path string) string {
	b, _ := json.Marshal(map[string]string{"path": path})
	return string(b)
}
