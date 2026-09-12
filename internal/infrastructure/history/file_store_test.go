package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	domainhistory "github.com/gosharplite/tellme/internal/domain/history"
)

// Round-007 T018: history store path/append/reload/archive unit tests.

func TestFileStore_AppendLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewFileStore(dir)
	if err := s.Append(domainhistory.Entry{Prompt: "q1", Answer: "a1"}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := s.Append(domainhistory.Entry{Prompt: "q2", Answer: "a2"}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := []domainhistory.Entry{{Prompt: "q1", Answer: "a1"}, {Prompt: "q2", Answer: "a2"}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("Load = %+v, want %+v", got, want)
	}
	// One deterministic line per entry (no id/timestamp).
	data, _ := os.ReadFile(filepath.Join(dir, "history.jsonl"))
	wantFile := `{"prompt":"q1","answer":"a1"}` + "\n" + `{"prompt":"q2","answer":"a2"}` + "\n"
	if string(data) != wantFile {
		t.Fatalf("file = %q, want %q", string(data), wantFile)
	}
}

func TestFileStore_LoadMissingIsEmpty(t *testing.T) {
	s := NewFileStore(t.TempDir())
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Load = %+v, want empty", got)
	}
}

func TestFileStore_ArchiveRetainsAndClears(t *testing.T) {
	dir := t.TempDir()
	s := NewFileStore(dir)
	if err := s.Append(domainhistory.Entry{Prompt: "q", Answer: "a"}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := s.Archive(); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	active, _ := s.Load()
	if len(active) != 0 {
		t.Fatalf("active after archive = %+v, want empty", active)
	}
	arch, _ := os.ReadFile(filepath.Join(dir, "history.archive.jsonl"))
	if !strings.Contains(string(arch), `"prompt":"q"`) {
		t.Fatalf("archive = %q, want the prior entry", string(arch))
	}
}

func TestFileStore_ArchiveMissingIsNoop(t *testing.T) {
	s := NewFileStore(t.TempDir())
	if err := s.Archive(); err != nil {
		t.Fatalf("Archive on empty = %v, want nil (RF-1)", err)
	}
}

func TestFileStore_ArchiveIsAppendAcrossResets(t *testing.T) {
	dir := t.TempDir()
	s := NewFileStore(dir)
	_ = s.Append(domainhistory.Entry{Prompt: "first", Answer: "1"})
	if err := s.Archive(); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	_ = s.Append(domainhistory.Entry{Prompt: "second", Answer: "2"})
	if err := s.Archive(); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	arch, _ := os.ReadFile(filepath.Join(dir, "history.archive.jsonl"))
	if !strings.Contains(string(arch), `"first"`) || !strings.Contains(string(arch), `"second"`) {
		t.Fatalf("archive = %q, want both archived sessions", string(arch))
	}
}

func TestFileStore_LargePromptRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewFileStore(dir)
	big := strings.Repeat("x", 200_000) // 200 KB — well over bufio.Scanner's 64 KB default
	if err := s.Append(domainhistory.Entry{Prompt: big, Answer: "ok"}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load (TD-2): %v", err)
	}
	if len(got) != 1 || got[0].Prompt != big {
		t.Fatalf("large prompt round-trip failed (entries=%d)", len(got))
	}
}
