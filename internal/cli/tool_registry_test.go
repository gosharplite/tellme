package cli

import "testing"

// Round 021 T034: the production registry factory offers exactly the three
// filesystem reader tools — list_files, read_files, get_tree — and never
// summarize_history (RED until Phase 4E/4F).

func TestNewToolRegistryOffersReaderToolsOnly(t *testing.T) {
	reg := newToolRegistry()
	got := map[string]bool{}
	for _, tl := range reg.Tools() {
		got[tl.Name()] = true
	}
	want := map[string]bool{"list_files": true, "read_files": true, "get_tree": true}
	if len(got) != len(want) {
		t.Fatalf("registry tools = %v; want exactly list_files, read_files, get_tree", got)
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
