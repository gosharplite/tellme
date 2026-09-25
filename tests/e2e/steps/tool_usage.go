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

// registeredToolNames enumerates the LIVE agent-tool BASE set — the tools a
// non-vision provider is offered — so the all-zero Then cannot pass vacuously
// when a tool is added or removed (round-026 review F8). It derives the set from
// the SINGLE canonical owner of the base composition,
// `infratools.NewAgentBaseTools` (round 092; ADR 0062) — the same owner the
// production assembler `cmd/tellme.assembleAgentTools` uses — so this harness and
// the production registry can no longer be edited out of step (the round-090
// R-090-1 mirror hazard). The capability-gated `read_image` is NOT here; see
// recordableToolNames.
func registeredToolNames() []string {
	return flattenToolNames(infratools.NewAgentBaseTools(nil))
}

// recordableToolNames enumerates every tool that can be RECORDED in the tool-usage
// log — the UNION of the base set and the capability-gated set (round 062;
// PR #129 fold F-062-1). The offline `--tool-usage` report lists this union, so
// the all-zero guard reads the same authority. The base half is the canonical
// owner (round 092; ADR 0062); the capability-gated `read_image` is appended here
// exactly as the composition root appends it under the `vision` gate.
func recordableToolNames() []string {
	return flattenToolNames(append(infratools.NewAgentBaseTools(nil), infratools.NewReadImageTool(0)))
}

// flattenToolNames flattens the given tool groups into their wire names, in order.
func flattenToolNames(groups ...[]domaintools.Tool) []string {
	var all []domaintools.Tool
	for _, g := range groups {
		all = append(all, g...)
	}
	tools := domaintools.NewRegistry(all...).Tools()
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		names = append(names, t.Name())
	}
	return names
}
