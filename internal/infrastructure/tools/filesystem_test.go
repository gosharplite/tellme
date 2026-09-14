package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Round 021: the filesystem tools take the reference contracts — list_files emits
// `Contents of <path>:` with `[d]`/`[f]` lines; read_files takes a multi-file
// `filepaths` array with per-file framing and the reference limits (100000 B/file,
// binary/directory/≤50 handling) plus tellme's aggregate 1 MiB result cap. These
// assertions are the T032 pins (they fail RED until the Phase 4 implementations).

func TestListFilesShape(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := listFiles{}.Execute(context.Background(), `{"path":"`+dir+`","reason":"r"}`)
	if err != nil {
		t.Fatalf("list_files: %v", err)
	}
	if !strings.Contains(got, "Contents of "+dir+":") {
		t.Errorf("list_files header = %q; want Contents of <path>:", got)
	}
	if !strings.Contains(got, "[f] a.txt") || !strings.Contains(got, "[d] sub") {
		t.Errorf("list_files entries = %q; want [f] a.txt and [d] sub", got)
	}
}

func TestListFilesDefaultPath(t *testing.T) {
	got, err := listFiles{}.Execute(context.Background(), `{"reason":"r"}`)
	if err != nil {
		t.Fatalf("list_files (default path): %v", err)
	}
	if !strings.HasPrefix(got, "Contents of .:") {
		t.Errorf("list_files default header = %q; want prefix Contents of .:", got)
	}
}

func TestReadFilesMulti(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "left.txt")
	p2 := filepath.Join(dir, "right.txt")
	if err := os.WriteFile(p1, []byte("LEFT"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("RIGHT"), 0o644); err != nil {
		t.Fatal(err)
	}
	args, _ := json.Marshal(map[string]any{"filepaths": []string{p1, p2}, "reason": "r"})
	got, err := readFiles{}.Execute(context.Background(), string(args))
	if err != nil {
		t.Fatalf("read_files: %v", err)
	}
	if !strings.Contains(got, "--- File: "+p1+" ---") || !strings.Contains(got, "--- File: "+p2+" ---") {
		t.Errorf("read_files framing = %q", got)
	}
	if strings.Index(got, p1) > strings.Index(got, p2) {
		t.Errorf("read_files did not preserve request order: %q", got)
	}
}

func TestReadFilesTruncates(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "big.txt")
	if err := os.WriteFile(p, []byte(strings.Repeat("x", 100001)), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readFiles{}.Execute(context.Background(), `{"filepaths":["`+p+`"]}`)
	if err != nil {
		t.Fatalf("read_files: %v", err)
	}
	if !strings.HasSuffix(strings.TrimRight(got, "\n"), "(truncated)") {
		t.Errorf("read_files did not truncate at the per-file cap; tail=%q", tail(got, 40))
	}
	if len(got) > 100000+200 {
		t.Errorf("read_files result %d bytes is not bounded by the 100000 cap", len(got))
	}
}

func TestReadFilesBinary(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bin")
	if err := os.WriteFile(p, []byte{0x00, 0x01, 0x02}, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readFiles{}.Execute(context.Background(), `{"filepaths":["`+p+`"]}`)
	if err != nil {
		t.Fatalf("read_files: %v", err)
	}
	if !strings.Contains(got, "(Binary file, cannot display as text)") {
		t.Errorf("read_files binary marker missing; got %q", got)
	}
}

func TestReadFilesDirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := readFiles{}.Execute(context.Background(), `{"filepaths":["`+sub+`"]}`)
	if err != nil {
		t.Fatalf("read_files: %v", err)
	}
	if !strings.Contains(got, "ERROR: path is a directory, use list_files instead") {
		t.Errorf("read_files directory error missing; got %q", got)
	}
}

func TestReadFilesTooMany(t *testing.T) {
	paths := make([]string, 51)
	for i := range paths {
		paths[i] = fmt.Sprintf("/tmp/r021/file%02d.txt", i)
	}
	args, _ := json.Marshal(map[string]any{"filepaths": paths, "reason": "r"})
	got, err := readFiles{}.Execute(context.Background(), string(args))
	if err != nil {
		t.Fatalf("too-many-files must be a non-fatal result, got error %v", err)
	}
	if !strings.Contains(got, "requested too many files") || !strings.Contains(got, "Maximum is 50") {
		t.Errorf("read_files too-many message = %q", got)
	}
}

func TestReadFilesEmptyArgs(t *testing.T) {
	if _, err := (readFiles{}).Execute(context.Background(), `{"filepaths":[]}`); err == nil {
		t.Error("empty filepaths must error")
	}
	if _, err := (readFiles{}).Execute(context.Background(), `{}`); err == nil {
		t.Error("missing filepaths must error")
	}
}

func TestReadFilesAggregateCap(t *testing.T) {
	dir := t.TempDir()
	// 11 × 100000 bytes ≈ 1.05 MB > the 1 MiB aggregate cap.
	paths := make([]string, 0, 11)
	for i := 0; i < 11; i++ {
		p := filepath.Join(dir, fmt.Sprintf("f%02d.txt", i))
		if err := os.WriteFile(p, []byte(strings.Repeat("y", 100000)), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}
	args, _ := json.Marshal(map[string]any{"filepaths": paths, "reason": "r"})
	got, err := readFiles{}.Execute(context.Background(), string(args))
	if err != nil {
		t.Fatalf("read_files: %v", err)
	}
	if len(got) > (1<<20)+200 {
		t.Errorf("the aggregate result %d bytes exceeds the 1 MiB cap", len(got))
	}
	if !strings.Contains(got, "truncated at the read budget") {
		t.Errorf("the aggregate cap marker is missing; tail=%q", tail(got, 60))
	}
}

func TestToolSchemasRequireReason(t *testing.T) {
	for _, params := range []json.RawMessage{listFiles{}.Parameters(), readFiles{}.Parameters(), getTree{}.Parameters()} {
		var s struct {
			Required []string `json:"required"`
		}
		if err := json.Unmarshal(params, &s); err != nil {
			t.Fatalf("tool schema is not valid JSON: %v", err)
		}
		found := false
		for _, r := range s.Required {
			if r == "reason" {
				found = true
			}
		}
		if !found {
			t.Errorf("tool schema does not require reason: %s", params)
		}
	}
}

// TestToolNamesWireValid pins the wire-valid snake_case identifiers.
func TestToolNamesWireValid(t *testing.T) {
	for _, tool := range []interface {
		Name() string
		Parameters() json.RawMessage
	}{listFiles{}, readFiles{}, getTree{}} {
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

// tail returns the last n bytes of s (for diagnostics).
func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
