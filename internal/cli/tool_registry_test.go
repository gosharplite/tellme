package cli

import "testing"

// Round 024 T042: the production registry factory offers exactly the four agent
// tools — the three filesystem readers plus the command tool — and never
// summarize_history or pipe_commands.

func TestNewToolRegistryOffersAgentTools(t *testing.T) {
	reg := newToolRegistry()
	got := map[string]bool{}
	for _, tl := range reg.Tools() {
		got[tl.Name()] = true
	}
	want := map[string]bool{"list_files": true, "read_files": true, "get_tree": true, "execute_command": true}
	if len(got) != len(want) {
		t.Fatalf("registry tools = %v; want exactly list_files, read_files, get_tree, execute_command", got)
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
