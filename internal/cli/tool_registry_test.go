package cli

import "testing"

// Round 024 T042 / round 029: the production registry factory offers exactly the
// six agent tools — the three filesystem readers, the write pair (write_file,
// replace_text), and the command tool — and never summarize_history or
// pipe_commands.

func TestNewToolRegistryOffersAgentTools(t *testing.T) {
	reg := newToolRegistry()
	got := map[string]bool{}
	for _, tl := range reg.Tools() {
		got[tl.Name()] = true
	}
	want := map[string]bool{
		"list_files":      true,
		"read_files":      true,
		"get_tree":        true,
		"write_file":      true,
		"replace_text":    true,
		"execute_command": true,
	}
	if len(got) != len(want) {
		t.Fatalf("registry tools = %v; want exactly list_files, read_files, get_tree, write_file, replace_text, execute_command", got)
	}
	for name := range want {
		if !got[name] {
			t.Errorf("registry is missing %q", name)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("registry offers unexpected tool %q", name)
		}
	}
}
