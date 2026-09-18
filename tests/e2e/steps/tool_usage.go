package steps

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	infratools "github.com/gosharplite/tellme/internal/infrastructure/tools"
)

// Round-026 tool-usage helpers: read the scenario-local user-global tool-usage
// log and parse the offline report. Kept in ONE file so the per-sentence step
// files stay independent (Zero Shared Edits).

// toolUsageLogRecord is one parsed line of the user-global tool-usage log.
type toolUsageLogRecord struct {
	Tool    string `json:"tool"`
	Outcome string `json:"outcome"`
}

// toolUsageLogRecords reads the scenario-local tool-usage log
// ($HOME/.tellme/tools-count.jsonl). A missing file is an empty slice.
func (sc *scenarioContext) toolUsageLogRecords() ([]toolUsageLogRecord, error) {
	data, err := os.ReadFile(sc.userToolUsagePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var recs []toolUsageLogRecord
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		var rec toolUsageLogRecord
		if err := json.Unmarshal([]byte(trimmed), &rec); err != nil {
			return nil, fmt.Errorf("malformed tool-usage log line %q: %w", trimmed, err)
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

// toolUsageCounts tallies the log records for one tool by outcome.
func toolUsageCounts(recs []toolUsageLogRecord, tool string) (ok, errCount, timeout int) {
	for _, r := range recs {
		if r.Tool != tool {
			continue
		}
		switch r.Outcome {
		case "ok":
			ok++
		case "error":
			errCount++
		case "timeout":
			timeout++
		}
	}
	return
}

// mapToolOutcomeWord maps a Gherkin outcome word to the persisted outcome value.
func mapToolOutcomeWord(word string) (string, bool) {
	switch word {
	case "succeeded":
		return "ok", true
	case "failed":
		return "error", true
	case "ran out of time":
		return "timeout", true
	default:
		return "", false
	}
}

// reportToolCounts is one parsed per-tool line of the --tool-usage report.
type reportToolCounts struct {
	total   int
	ok      int
	errN    int
	timeout int
}

// parseReportLine parses the `<tool>: total=N ok=N error=N timeout=N` report line
// for the given tool out of the captured stdout.
func parseReportLine(stdout, tool string) (reportToolCounts, bool) {
	prefix := tool + ": "
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimRight(line, "\r")
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		counts := reportToolCounts{}
		for _, field := range strings.Fields(strings.TrimPrefix(line, prefix)) {
			kv := strings.SplitN(field, "=", 2)
			if len(kv) != 2 {
				continue
			}
			n, err := strconv.Atoi(kv[1])
			if err != nil {
				return reportToolCounts{}, false
			}
			switch kv[0] {
			case "total":
				counts.total = n
			case "ok":
				counts.ok = n
			case "error":
				counts.errN = n
			case "timeout":
				counts.timeout = n
			}
		}
		return counts, true
	}
	return reportToolCounts{}, false
}

// registeredToolNames enumerates the LIVE agent-tool registry (the same
// constructors cli.newToolRegistry uses), so the all-zero Then cannot pass
// vacuously when a tool is added or removed (round-026 review F8). Round 029 adds
// the write pair (write_file, replace_text) and round 033 adds the read-only
// list_skills tool, keeping this in sync with cli.newToolRegistry.
func registeredToolNames() []string {
	reg := domaintools.NewRegistry(append(
		append(
			append(infratools.NewFilesystemTools(), infratools.NewWriteTools()...),
			infratools.NewCommandTool(nil),
		),
		infratools.NewSkillsTool(nil),
	)...)
	tools := reg.Tools()
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		names = append(names, t.Name())
	}
	return names
}
