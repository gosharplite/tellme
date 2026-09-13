package history

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

// T024 (round 008) — the widened history entry (with tool steps) round-trips and
// stays byte-deterministic (research Decision 5 / NFR-002).

func TestFileStore_WidenedEntryRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewFileStore(dir)
	entry := domainhistory.Entry{
		Prompt: "read notes.txt",
		Answer: "ORANGE",
		Steps:  []domainhistory.Step{{Tool: "read_files", Arguments: `{"path":"notes.txt"}`, Result: "ORANGE"}},
	}
	if err := s.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0], entry) {
		t.Fatalf("Load = %+v, want %+v", got, entry)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "history.jsonl"))
	wantLine := `{"prompt":"read notes.txt","answer":"ORANGE","steps":[{"tool":"read_files","arguments":"{\"path\":\"notes.txt\"}","result":"ORANGE"}]}` + "\n"
	if string(data) != wantLine {
		t.Fatalf("file = %q, want %q", string(data), wantLine)
	}
}

func TestFileStore_EntryWithoutStepsOmitsSteps(t *testing.T) {
	dir := t.TempDir()
	s := NewFileStore(dir)
	if err := s.Append(domainhistory.Entry{Prompt: "q", Answer: "a"}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "history.jsonl"))
	if string(data) != `{"prompt":"q","answer":"a"}`+"\n" {
		t.Fatalf("file = %q, want the no-steps form (omitempty)", string(data))
	}
}
