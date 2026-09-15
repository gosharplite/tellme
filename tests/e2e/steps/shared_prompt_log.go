package steps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Round-028 shared prompt-log helpers: the shared, append-only JSON-Lines log at
// the USER-global `<HOME>/.tellme/global_prompts.jsonl` (the same user-global root
// as the round-026 tool-usage log), written only under `-i`. The
// environment-scoped `<TELL_ME_HOME>/output/global_prompts.jsonl` is the round-028
// seed SOURCE only. Kept in ONE file so the per-sentence step files stay
// independent (Zero Shared Edits), mirroring wire_tools.go / payload_status.go.

// promptLogPath is the user-global shared log path for a scenario
// (`<user-home>/.tellme/global_prompts.jsonl`) — the round-026 `userHomeDir` seam.
func promptLogPath(sc *scenarioContext) string {
	return filepath.Join(sc.userHomeDir, ".tellme", "global_prompts.jsonl")
}

// envPromptLogPath is the environment-scoped seed source for a scenario
// (`<TELL_ME_HOME>/output/global_prompts.jsonl`) — no longer read/written by the
// product; arranged only to exercise the first-use seed.
func envPromptLogPath(sc *scenarioContext) string {
	return filepath.Join(sc.home, "output", "global_prompts.jsonl")
}

// promptLogRecord is one shared-log line; the field order matches the frozen
// shape `{"timestamp":"<RFC3339>","prompt":"<text>"}` (data-model.dbml,
// table `prompt_log_entry`).
type promptLogRecord struct {
	Timestamp string `json:"timestamp"`
	Prompt    string `json:"prompt"`
}

// appendPromptLine appends one {timestamp,prompt} record to the log at path,
// creating its parent directory if absent.
func appendPromptLine(path, prompt string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	rec, err := json.Marshal(promptLogRecord{Timestamp: "2026-01-01T00:00:00Z", Prompt: prompt})
	if err != nil {
		return err
	}
	rec = append(rec, '\n')
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(rec); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// appendPromptLog appends one record to the user-global shared log
// (Given: the shared prompt log already holds "…").
func appendPromptLog(sc *scenarioContext, prompt string) error {
	return appendPromptLine(promptLogPath(sc), prompt)
}

// appendEnvPromptLog appends one record to the environment-scoped seed source
// (Given: the environment prompt log already holds "…").
func appendEnvPromptLog(sc *scenarioContext, prompt string) error {
	return appendPromptLine(envPromptLogPath(sc), prompt)
}

// ensureSharedPromptLogAbsent removes the user-global shared log so the next
// `-i` run's first-use seed engages (Given: the shared prompt log has not been
// created yet).
func ensureSharedPromptLogAbsent(sc *scenarioContext) error {
	if err := os.Remove(promptLogPath(sc)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
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
