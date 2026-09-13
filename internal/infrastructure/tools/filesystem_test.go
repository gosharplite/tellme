package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestListFiles exercises the read-only directory listing.
func TestListFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := listFiles{}.Execute(context.Background(), `{"path":"`+dir+`"}`)
	if err != nil {
		t.Fatalf("list_files: %v", err)
	}
	// sorted, directories suffixed with "/"
	if got != "a.txt\nb.txt\nsub/" {
		t.Errorf("list_files = %q; want %q", got, "a.txt\nb.txt\nsub/")
	}
}

// TestListFilesErrors covers missing/invalid arguments and a bad path.
func TestListFilesErrors(t *testing.T) {
	if _, err := (listFiles{}).Execute(context.Background(), `{}`); err == nil {
		t.Error("list_files: missing path must error")
	}
	if _, err := (listFiles{}).Execute(context.Background(), `not-json`); err == nil {
		t.Error("list_files: invalid arguments must error")
	}
	if _, err := (listFiles{}).Execute(context.Background(), `{"path":"/no/such/dir/xyz"}`); err == nil {
		t.Error("list_files: nonexistent directory must error")
	}
}

// TestReadFiles reads a small file verbatim.
func TestReadFiles(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(p, []byte("the launch code is ORANGE"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readFiles{}.Execute(context.Background(), `{"path":"`+p+`"}`)
	if err != nil {
		t.Fatalf("read_files: %v", err)
	}
	if got != "the launch code is ORANGE" {
		t.Errorf("read_files = %q", got)
	}
}

// TestReadFilesMissingPath is the non-terminal tool-error case (a missing file
// is returned as an error, and the loop feeds it back).
func TestReadFilesMissingPath(t *testing.T) {
	if _, err := (readFiles{}).Execute(context.Background(), `{"path":"/no/such/file.txt"}`); err == nil {
		t.Error("read_files: missing file must error")
	}
}

// TestReadFilesCap verifies the 1 MiB size bound (NFR-006).
func TestReadFilesCap(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "big.txt")
	big := strings.Repeat("x", readCap+100)
	if err := os.WriteFile(p, []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readFiles{}.Execute(context.Background(), `{"path":"`+p+`"}`)
	if err != nil {
		t.Fatalf("read_files: %v", err)
	}
	if !strings.HasSuffix(got, "(truncated at 1 MiB)") {
		t.Errorf("read_files: want truncation marker, got suffix %q", got[max(0, len(got)-40):])
	}
	if len(got) > readCap+64 {
		t.Errorf("read_files: result %d bytes exceeds the cap", len(got))
	}
}

// TestToolNamesWireValid pins the wire-valid snake_case identifiers
// (round-008 BLOCKER-1 / FR-001).
func TestToolNamesWireValid(t *testing.T) {
	for _, tool := range NewFilesystemTools() {
		name := tool.Name()
		if !wireNameOK(name) {
			t.Errorf("tool name %q is not wire-valid", name)
		}
		if err := json.Unmarshal(tool.Parameters(), new(map[string]any)); err != nil {
			t.Errorf("tool %q parameters are not valid JSON: %v", name, err)
		}
	}
}

func wireNameOK(s string) bool {
	if len(s) == 0 || len(s) > 64 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
