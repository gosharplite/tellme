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

// recordingPrompts records the n the engine requests and returns a canned list.
type recordingPrompts struct {
	asked int
	items []string
}

func (r *recordingPrompts) RecentPrompts(_ context.Context, n int) []string {
	r.asked = n
	if n > 0 && len(r.items) > n {
		return r.items[:n]
	}
	return r.items
}

// TestServiceSuggestAsksForDeepenedPoolAndCapsAtTen (round-064 T005, ADR 0034):
// the engine asks the history source for the deepened candidate pool
// (promptPoolDepth = 50), not the shallow 10, AND still surfaces at most
// maxSuggestions (10) suggestions. Two claims: the pool depth is deepened, and
// the surfaced cap is unchanged and non-vacuous (20 matches collapse to 10).
func TestServiceSuggestAsksForDeepenedPoolAndCapsAtTen(t *testing.T) {
	// 20 matching prompts (subsequence "commit") — more than the cap.
	items := make([]string, 0, 20)
	for i := 0; i < 20; i++ {
		items = append(items, "commit workspace "+string(rune('a'+i)))
	}
	src := &recordingPrompts{items: items}
	svc := New(src, fakeWorkspace{}, fakeTools{})

	got := svc.Suggest(context.Background(), "commit")

	if src.asked != promptPoolDepth {
		t.Fatalf("history source asked for n = %d, want the deepened pool %d (round 064)", src.asked, promptPoolDepth)
	}
	// Literal pin (review R-4 recorded; the round-057 argValueCap = 500 pattern):
	// the parity depth itself is bound, so a silent drift to 51 (or any non-50 value)
	// is caught here, not only by the qualitative depth>cap relation.
	if promptPoolDepth != 50 {
		t.Fatalf("promptPoolDepth = %d, want the reference's 50 (ADR 0034)", promptPoolDepth)
	}
	if promptPoolDepth <= maxSuggestions {
		t.Fatalf("promptPoolDepth (%d) must be strictly deeper than the surface cap (%d)", promptPoolDepth, maxSuggestions)
	}
	if len(got) != maxSuggestions {
		t.Fatalf("Suggest(commit) surfaced %d suggestions, want exactly the cap %d", len(got), maxSuggestions)
	}
}
