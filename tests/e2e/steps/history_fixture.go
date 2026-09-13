package steps

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

// historyEntry mirrors internal/domain/history.Entry for file assertions in the
// E2E steps (the steps package reads the store as a black box).
type historyEntry struct {
	Prompt string `json:"prompt"`
	Answer string `json:"answer"`
}

// readHistoryEntries reads the JSON-Lines history file at path. A missing file
// is an empty history (nil), not an error.
func readHistoryEntries(path string) ([]historyEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var entries []historyEntry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e historyEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// wireMessage is one entry of the conversation recorded by the fake, normalized
// across wire families.
type wireMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// decodeWireMessages extracts the conversation from a recorded request body,
// normalized across wire families (round 013):
//   - OpenAI-compatible: the top-level `messages` array.
//   - Vertex/Gemini: a leading `system` message synthesised from
//     `systemInstruction`, then the `contents` array with text parts concatenated
//     and role "model" mapped to "assistant".
//
// Normalizing both keeps the wire-family-agnostic Then steps (persona, prior
// exchange) working for a gemini provider.
func decodeWireMessages(body string) ([]wireMessage, error) {
	var probe struct {
		Messages []wireMessage `json:"messages"`
		Contents []struct {
			Role  string `json:"role"`
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"contents"`
		SystemInstruction *struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"systemInstruction"`
	}
	if err := json.Unmarshal([]byte(body), &probe); err != nil {
		return nil, err
	}
	if len(probe.Messages) > 0 {
		return probe.Messages, nil
	}
	out := make([]wireMessage, 0, len(probe.Contents)+1)
	if probe.SystemInstruction != nil {
		var sb strings.Builder
		for _, p := range probe.SystemInstruction.Parts {
			sb.WriteString(p.Text)
		}
		if sb.Len() > 0 {
			out = append(out, wireMessage{Role: "system", Content: sb.String()})
		}
	}
	for _, c := range probe.Contents {
		role := c.Role
		if role == "model" {
			role = "assistant"
		}
		var sb strings.Builder
		for _, p := range c.Parts {
			sb.WriteString(p.Text)
		}
		out = append(out, wireMessage{Role: role, Content: sb.String()})
	}
	return out, nil
}
