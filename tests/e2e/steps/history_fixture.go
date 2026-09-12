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

// wireMessage is one entry of the OpenAI `messages` array recorded by the fake.
type wireMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// decodeWireMessages extracts the `messages` array from a recorded request body.
func decodeWireMessages(body string) ([]wireMessage, error) {
	var decoded struct {
		Messages []wireMessage `json:"messages"`
	}
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		return nil, err
	}
	return decoded.Messages, nil
}
