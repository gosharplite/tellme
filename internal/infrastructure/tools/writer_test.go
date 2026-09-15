package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Round-029 T022 — write-tools unit tests.
//
// Observable behaviour is asserted through the public Execute; the atomicity
// guarantee (FR-009) is a UNIT-TIER invariant (no E2E fault injection exists) and
// is witnessed by overriding the writeTempContent seam to fail a partial write.
// These tests are RED until the tools are implemented in Phase 4A (T024).

const unitBudget = 1 << 20

func runWriteFile(args string) (string, error) {
	return writeFile{}.Execute(context.Background(), args, unitBudget)
}

func runReplaceText(args string) (string, error) {
	return replaceText{}.Execute(context.Background(), args, unitBudget)
}

func writeArgs(path, content string) string {
	b, _ := json.Marshal(map[string]any{"filepath": path, "content": content, "reason": "test"})
	return string(b)
}

func replaceArgs(path, oldText, newText string) string {
	b, _ := json.Marshal(map[string]any{"filepath": path, "old_text": oldText, "new_text": newText, "reason": "test"})
	return string(b)
}

// assertNoTempResidue fails if the directory still holds a `*.tmp` file (a
// torn-write residue).
func assertNoTempResidue(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temporary-file residue left after a failed write: %q", e.Name())
		}
	}
}

func TestWriteFileCreatesNewFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "notes.txt")
	if _, err := runWriteFile(writeArgs(dest, "hello")); err != nil {
		t.Fatalf("write_file returned an error: %v", err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read created file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("created content = %q; want %q", string(data), "hello")
	}
}

func TestWriteFileRefusesExistingFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(dest, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runWriteFile(writeArgs(dest, "replacement")); err == nil {
		t.Fatalf("write_file over an existing file must fail")
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Errorf("the existing file was modified: %q; want %q", string(data), "original")
	}
}

func TestWriteFileCreatesMissingParents(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "drafts", "todo.txt")
	if _, err := runWriteFile(writeArgs(dest, "buy milk")); err != nil {
		t.Fatalf("write_file returned an error: %v", err)
	}
	if info, err := os.Stat(filepath.Dir(dest)); err != nil || !info.IsDir() {
		t.Fatalf("the parent directory was not created: info=%v err=%v", info, err)
	}
	if data, err := os.ReadFile(dest); err != nil || string(data) != "buy milk" {
		t.Fatalf("the file was not created with the exact content: %q err=%v", string(data), err)
	}
}

func TestWriteFileCreatedFileModeIs0644(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "notes.txt")
	if _, err := runWriteFile(writeArgs(dest, "hello")); err != nil {
		t.Fatalf("write_file returned an error: %v", err)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Errorf("created file mode = %o; want 0644", perm)
	}
}

func TestWriteFileCreatedParentDirModeIs0755(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "drafts")
	dest := filepath.Join(parent, "todo.txt")
	if _, err := runWriteFile(writeArgs(dest, "x")); err != nil {
		t.Fatalf("write_file returned an error: %v", err)
	}
	info, err := os.Stat(parent)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o755 {
		t.Errorf("created parent dir mode = %o; want 0755", perm)
	}
}

func TestWriteFileRejectsMissingContentKey(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "notes.txt")
	args, _ := json.Marshal(map[string]any{"filepath": dest, "reason": "test"}) // no "content"
	if _, err := runWriteFile(string(args)); err == nil {
		t.Fatalf("write_file with a missing content key must fail")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Errorf("a missing content key must not create the file (stat err=%v)", err)
	}
}

func TestWriteFileRejectsParentPathComponentFile(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "a")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(blocker, "b.txt")
	if _, err := runWriteFile(writeArgs(dest, "y")); err == nil {
		t.Fatalf("a regular-file parent path component must fail")
	}
	data, err := os.ReadFile(blocker)
	if err != nil || string(data) != "x" {
		t.Errorf("the blocking file was modified: %q err=%v", string(data), err)
	}
}

func TestWriteFileAtomicityLeavesNoPartialOrTemp(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.txt")
	orig := writeTempContent
	writeTempContent = func(f *os.File, data []byte) error {
		if _, err := f.Write(data[:3]); err != nil {
			return err
		}
		return errors.New("injected write failure")
	}
	defer func() { writeTempContent = orig }()

	_, err := runWriteFile(writeArgs(dest, "hello"))
	if err == nil || !strings.Contains(err.Error(), "injected write failure") {
		t.Fatalf("want the injected write failure surfaced; got err=%v", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Errorf("destination exists after a failed atomic write (stat err=%v)", statErr)
	}
	assertNoTempResidue(t, dir)
}

func TestReplaceTextReplacesUniqueBlock(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "config.txt")
	if err := os.WriteFile(dest, []byte("alpha\nBETA\ngamma\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runReplaceText(replaceArgs(dest, "BETA", "beta")); err != nil {
		t.Fatalf("replace_text returned an error: %v", err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "alpha\nbeta\ngamma\n" {
		t.Errorf("edited content = %q; want %q", string(data), "alpha\nbeta\ngamma\n")
	}
}

func TestReplaceTextRefusesAbsentBlock(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "config.txt")
	if err := os.WriteFile(dest, []byte("alpha\nBETA\ngamma\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runReplaceText(replaceArgs(dest, "DELTA", "delta")); err == nil {
		t.Fatalf("replace_text with an absent block must fail")
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "alpha\nBETA\ngamma\n" {
		t.Errorf("the file changed on a refused edit: %q", string(data))
	}
}

func TestReplaceTextRefusesAmbiguousBlock(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "log.txt")
	if err := os.WriteFile(dest, []byte("note\nnote\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := runReplaceText(replaceArgs(dest, "note", "done"))
	if err == nil {
		t.Fatalf("replace_text with a non-unique block must fail")
	}
	lower := strings.ToLower(err.Error())
	if !strings.Contains(lower, "unique") && !strings.Contains(lower, "more than once") {
		t.Errorf("the ambiguity error should name the match count; got %v", err)
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "note\nnote\n" {
		t.Errorf("the file changed on a refused edit: %q", string(data))
	}
}

func TestReplaceTextRejectsEmptyOldText(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "config.txt")
	if err := os.WriteFile(dest, []byte("alpha\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runReplaceText(replaceArgs(dest, "", "x")); err == nil {
		t.Fatalf("replace_text with an empty old_text must fail")
	}
}

func TestReplaceTextMissingFileDoesNotCreate(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "missing.txt")
	if _, err := runReplaceText(replaceArgs(dest, "a", "b")); err == nil {
		t.Fatalf("replace_text on a missing file must fail")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Errorf("replace_text must not create the file (stat err=%v)", err)
	}
}

func TestReplaceTextNoOpShortCircuits(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "config.txt")
	if err := os.WriteFile(dest, []byte("alpha\nBETA\ngamma\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runReplaceText(replaceArgs(dest, "BETA", "BETA")); err != nil {
		t.Fatalf("a no-op replace_text must succeed: %v", err)
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "alpha\nBETA\ngamma\n" {
		t.Errorf("the file changed on a no-op replace: %q", string(data))
	}
}

func TestReplaceTextAtomicityLeavesNoPartialOrTemp(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "config.txt")
	if err := os.WriteFile(dest, []byte("alpha\nBETA\ngamma\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := writeTempContent
	writeTempContent = func(f *os.File, data []byte) error { return errors.New("injected write failure") }
	defer func() { writeTempContent = orig }()

	_, err := runReplaceText(replaceArgs(dest, "BETA", "beta"))
	if err == nil || !strings.Contains(err.Error(), "injected write failure") {
		t.Fatalf("want the injected write failure surfaced; got err=%v", err)
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "alpha\nBETA\ngamma\n" {
		t.Errorf("destination changed after a failed atomic edit: %q", string(data))
	}
	assertNoTempResidue(t, dir)
}
