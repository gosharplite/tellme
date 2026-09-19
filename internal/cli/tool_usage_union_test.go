package cli

import (
	"reflect"
	"testing"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// TestUnionToolNames pins the recordable-set enumerator (round 062; PR #129 fold
// R-062-2): first-seen order across the registries, de-duplicated by name.
func TestUnionToolNames(t *testing.T) {
	base := domaintools.NewRegistry(noopTool{name: "list_files"}, noopTool{name: "read_image"})
	extra := domaintools.NewRegistry(noopTool{name: "read_image"}, noopTool{name: "get_tree"})

	got := unionToolNames(base, extra)
	want := []string{"list_files", "read_image", "get_tree"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unionToolNames = %v, want %v", got, want)
	}

	if got := unionToolNames(nil, base); !reflect.DeepEqual(got, []string{"list_files", "read_image"}) {
		t.Errorf("unionToolNames with a nil registry = %v", got)
	}
	if got := unionToolNames(); len(got) != 0 {
		t.Errorf("unionToolNames() = %v, want empty", got)
	}
}
