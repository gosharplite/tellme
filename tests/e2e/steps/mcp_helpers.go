package steps

// Round-032 MCP test helpers, kept in ONE file so the per-sentence step files
// stay independent (Zero Shared Edits).

// nativeAgentTools is the native (non-MCP) agent tool set a prompt run offers
// (round 029: the reader trio + the write pair + execute_command). The MCP
// offered-set Thens assert the MCP tool is offered ALONGSIDE these, unchanged.
var nativeAgentTools = []string{
	"list_files", "read_files", "get_tree", "write_file", "replace_text", "execute_command",
}

// containsString reports whether ss contains s.
func containsString(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
