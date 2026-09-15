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

// T007 (round 014) — the per-step provider signature round-trips, and a
// signature-less step stays byte-identical to the round-008 shape (omitempty).

func TestFileStore_SignedStepRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewFileStore(dir)
	entry := domainhistory.Entry{
		Prompt: "read notes.txt",
		Answer: "ORANGE",
		Steps:  []domainhistory.Step{{Tool: "read_files", Arguments: `{"path":"notes.txt"}`, Result: "ORANGE", Signature: "sig-abc"}},
	}
	if err := s.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || len(got[0].Steps) != 1 || got[0].Steps[0].Signature != "sig-abc" {
		t.Fatalf("Load = %+v, want the signature round-tripped", got)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "history.jsonl"))
	wantLine := `{"prompt":"read notes.txt","answer":"ORANGE","steps":[{"tool":"read_files","arguments":"{\"path\":\"notes.txt\"}","result":"ORANGE","signature":"sig-abc"}]}` + "\n"
	if string(data) != wantLine {
		t.Fatalf("file = %q, want %q", string(data), wantLine)
	}
}

func TestFileStore_UnsignedStepOmitsSignature(t *testing.T) {
	dir := t.TempDir()
	s := NewFileStore(dir)
	entry := domainhistory.Entry{
		Prompt: "q",
		Answer: "a",
		Steps:  []domainhistory.Step{{Tool: "read_files", Arguments: `{}`, Result: "r"}},
	}
	if err := s.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "history.jsonl"))
	wantLine := `{"prompt":"q","answer":"a","steps":[{"tool":"read_files","arguments":"{}","result":"r"}]}` + "\n"
	if string(data) != wantLine {
		t.Fatalf("file = %q, want the signature-less shape (omitempty)", string(data))
	}
}

// T007 (round 027) — the per-turn AI-call count round-trips, and a line without
// it (a legacy/arranged plain entry) loads as zero, which the counter treats as
// one.

func TestFileStore_CallsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewFileStore(dir)
	entry := domainhistory.Entry{
		Prompt: "read notes.txt",
		Answer: "ORANGE",
		Calls:  2,
		Steps:  []domainhistory.Step{{Tool: "read_files", Arguments: `{"path":"notes.txt"}`, Result: "ORANGE"}},
	}
	if err := s.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || got[0].Calls != 2 {
		t.Fatalf("Load = %+v, want Calls=2 round-tripped", got)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "history.jsonl"))
	wantLine := `{"prompt":"read notes.txt","answer":"ORANGE","calls":2,"steps":[{"tool":"read_files","arguments":"{\"path\":\"notes.txt\"}","result":"ORANGE"}]}` + "\n"
	if string(data) != wantLine {
		t.Fatalf("file = %q, want %q", string(data), wantLine)
	}
}

func TestFileStore_LegacyLineLoadsWithoutCalls(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "history.jsonl"), []byte(`{"prompt":"q","answer":"a"}`+"\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := NewFileStore(dir).Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || got[0].Calls != 0 {
		t.Fatalf("Load = %+v, want a legacy line (Calls=0)", got)
	}
}
