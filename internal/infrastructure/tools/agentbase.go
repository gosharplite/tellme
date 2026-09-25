package tools

import (
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// NewAgentBaseTools returns the BASE agent tool set in offer order — the tools
// offered to EVERY provider, independent of capability:
//
//	the read-only filesystem readers (list_files, read_files, get_tree)
//	the in-file content search            (search_files — round 071; ADR 0043)
//	the write pair                        (write_file, replace_text — round 029)
//	the bash-first command tool           (execute_command — round 024)
//	the read-only skills listing tool     (list_skills — round 033)
//
// The command tool's `[Tool Output]` sink is injected at CONSTRUCTION (round 052;
// ADR 0021); pass nil on a path that never executes a tool (the offline
// `--tool-usage` report, the round-031 schema gate, the e2e enumerator).
//
// Round 092 (issue #189; ADR 0062): this is the SINGLE canonical owner of the
// base-set composition. Both the composition root (cmd/tellme assembleAgentTools)
// and the e2e harness (tests/e2e/steps registeredToolNames) derive their base set
// from here, so the two can no longer be edited out of step (the round-090
// R-090-1 mirror hazard). The capability-gated `read_image` append and the
// family-aware inline-ceiling resolution deliberately stay in the composition
// root — this function carries the BASE set only, so no capability policy enters
// the tools layer (ADR 0039 D2/D3).
func NewAgentBaseTools(sink domaintools.OutputSink) []domaintools.Tool {
	tools := NewFilesystemTools()
	tools = append(tools, NewSearchTool()...)
	tools = append(tools, NewWriteTools()...)
	tools = append(tools, NewCommandTool(sink))
	tools = append(tools, NewSkillsTool(nil))
	return tools
}
