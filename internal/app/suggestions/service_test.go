package suggestions

import (
	"context"
	"testing"
)

// fakePrompts is a canned recent-prompt source.
type fakePrompts struct{ recent []string }

func (f fakePrompts) RecentPrompts(context.Context, int) []string { return f.recent }

// fakeWorkspace is a canned workspace-entry source keyed by the query's dir.
type fakeWorkspace struct{ byDir map[string][]string }

func (f fakeWorkspace) Entries(_ context.Context, dir, _ string, _ int) []string { return f.byDir[dir] }

// fakeTools is a canned tool-name source.
type fakeTools struct{ names []string }

func (f fakeTools) ToolNames() []string { return f.names }

// TestServiceSuggestMatchesRecentPrompt (round-015 T027): a query yields the
// matching recent prompt (subsequence), deduplicated and capped.
func TestServiceSuggestMatchesRecentPrompt(t *testing.T) {
	svc := New(
		fakePrompts{recent: []string{"deploy to staging with version 015", "review the last two commits"}},
		fakeWorkspace{},
		fakeTools{},
	)
	got := svc.Suggest(context.Background(), "deploy")
	if len(got) != 1 || got[0].Text != "deploy to staging with version 015" {
		t.Fatalf("Suggest(deploy) = %v, want exactly the matching recent prompt", got)
	}
}

// TestServiceSuggestEmptyQueryReturnsRecent (round-015 T027): an empty query
// yields the recent prompts, newest-first.
func TestServiceSuggestEmptyQueryReturnsRecent(t *testing.T) {
	svc := New(fakePrompts{recent: []string{"newest", "older"}}, fakeWorkspace{}, fakeTools{})
	got := svc.Suggest(context.Background(), "")
	if len(got) == 0 || got[0].Text != "newest" {
		t.Fatalf("Suggest(\"\") = %v, want the newest recent prompt first", got)
	}
}

// TestServiceSuggestOffersTool (round-015 T027): a matching query yields a
// registered tool name.
func TestServiceSuggestOffersTool(t *testing.T) {
	svc := New(fakePrompts{}, fakeWorkspace{}, fakeTools{names: []string{"read_files"}})
	got := svc.Suggest(context.Background(), "read")
	for _, s := range got {
		if s.Text == "read_files" {
			return
		}
	}
	t.Fatalf("Suggest(read) = %v, want it to offer the tool read_files", got)
}
