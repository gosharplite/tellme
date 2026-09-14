package steps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Round-015 shared prompt-log helpers: the shared, append-only JSON-Lines log at
// the `output/` ROOT of the runtime home ($TELL_ME_HOME/output/global_prompts.jsonl),
// shared across modes/personas with tell-me-go. Kept in ONE file so the
// per-sentence step files stay independent (Zero Shared Edits), mirroring
// wire_tools.go / payload_status.go.

// promptLogPath is the shared log path for a scenario.
func promptLogPath(sc *scenarioContext) string {
	return filepath.Join(sc.home, "output", "global_prompts.jsonl")
}

// promptLogRecord is one shared-log line; the field order matches the frozen
// shape `{"timestamp":"<RFC3339>","prompt":"<text>"}` (data-model.dbml,
// table `prompt_log_entry`).
type promptLogRecord struct {
	Timestamp string `json:"timestamp"`
	Prompt    string `json:"prompt"`
}

// appendPromptLog appends one {timestamp,prompt} record to the shared log
// (Given: the shared prompt log already holds "…").
func appendPromptLog(sc *scenarioContext, prompt string) error {
	p := promptLogPath(sc)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	rec, err := json.Marshal(promptLogRecord{Timestamp: "2026-01-01T00:00:00Z", Prompt: prompt})
	if err != nil {
		return err
	}
	rec = append(rec, '\n')
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(rec); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// readPromptLog returns the prompts recorded in the shared log, in file order. A
// missing file yields an empty slice.
func readPromptLog(sc *scenarioContext) ([]string, error) {
	data, err := os.ReadFile(promptLogPath(sc))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var prompts []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec promptLogRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return nil, err
		}
		prompts = append(prompts, rec.Prompt)
	}
	return prompts, nil
}
