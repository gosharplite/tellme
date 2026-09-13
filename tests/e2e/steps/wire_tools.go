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

// decodeToolMessages decodes a recorded request body into a NORMALIZED message
// list, regardless of the wire family the provider speaks:
//   - OpenAI-compatible: the top-level `messages` array (returned as-is).
//   - Vertex/Gemini: the `contents` array — assistant tool calls under
//     role:"model" functionCall, tool results under role:"user" functionResponse
//     — normalized to the assistant / tool shape the Then steps assert on.
//
// Round 013 keeps the tool-loop Then steps wire-family-agnostic.
func decodeToolMessages(body string) []wireToolMessage {
	var probe struct {
		Messages []wireToolMessage `json:"messages"`
		Contents []struct {
			Role  string `json:"role"`
			Parts []struct {
				Text         string `json:"text"`
				FunctionCall *struct {
					Name string          `json:"name"`
					Args json.RawMessage `json:"args"`
				} `json:"functionCall"`
				FunctionResponse *struct {
					Name     string `json:"name"`
					Response struct {
						Content string `json:"content"`
					} `json:"response"`
				} `json:"functionResponse"`
			} `json:"parts"`
		} `json:"contents"`
	}
	_ = json.Unmarshal([]byte(body), &probe)
	if len(probe.Messages) > 0 {
		return probe.Messages
	}
	out := make([]wireToolMessage, 0, len(probe.Contents))
	for _, c := range probe.Contents {
		msg := wireToolMessage{Role: c.Role}
		if msg.Role == "model" {
			msg.Role = "assistant"
		}
		for _, p := range c.Parts {
			switch {
			case p.FunctionCall != nil:
				args := strings.TrimSpace(string(p.FunctionCall.Args))
				if args == "" {
					args = "{}"
				}
				tc := wireToolCall{}
				tc.Function.Name = p.FunctionCall.Name
				tc.Function.Arguments = args
				msg.ToolCalls = append(msg.ToolCalls, tc)
			case p.FunctionResponse != nil:
				msg.Role = "tool"
				msg.Content = p.FunctionResponse.Response.Content
			default:
				msg.Content += p.Text
			}
		}
		out = append(out, msg)
	}
	return out
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

// toolExchangeChronologyOK reports whether every recorded request that carries a
// tool result keeps the OpenAI-mandated chronology: an assistant tool-call is
// immediately followed by its tool result, and a user message precedes them.
// This makes the BLOCKER-1 inversion (tool activity before the user prompt) fail
// the E2E suite (review PR #25 TD-2).
func toolExchangeChronologyOK(f *fakeprovider.Provider) bool {
	for _, msgs := range toolRounds(f) {
		for i, m := range msgs {
			if m.Role != "tool" {
				continue
			}
			if i < 1 || msgs[i-1].Role != "assistant" || len(msgs[i-1].ToolCalls) == 0 {
				return false
			}
			if !hasUserBefore(msgs, i) {
				return false
			}
		}
	}
	return true
}

// hasUserBefore reports whether a user-role message precedes index idx.
func hasUserBefore(msgs []wireToolMessage, idx int) bool {
	for j := 0; j < idx; j++ {
		if msgs[j].Role == "user" {
			return true
		}
	}
	return false
}

// readArgs builds the read_files tool arguments JSON for a path.
func readArgs(path string) string {
	b, _ := json.Marshal(map[string]string{"path": path})
	return string(b)
}
