package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// Round 024: the filesystem reader tools take the reference contracts — list_files
// emits `Contents of <path>:` with `[d]`/`[f]` lines; read_files takes a
// multi-file `filepaths` array with per-file framing (binary/directory/≤50
// handling) — and are bounded by the resolved BYTE budget (the former fixed
// 100000-byte / 1 MiB caps are retired).

// testBudget is the byte budget the reader unit tests pass to Execute.
const testBudget = 1 << 20

func TestListFilesShape(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := listFiles{}.Execute(context.Background(), `{"path":"`+dir+`","reason":"r"}`, testBudget)
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
	got, err := listFiles{}.Execute(context.Background(), `{"reason":"r"}`, testBudget)
	if err != nil {
		t.Fatalf("list_files (default path): %v", err)
	}
	if !strings.HasPrefix(got, "Contents of .:") {
		t.Errorf("list_files default header = %q; want prefix Contents of .:", got)
	}
}

func TestListFilesTruncatesAtBudget(t *testing.T) {
	dir := t.TempDir()
	// Many entries so the listing overflows a small budget.
	for i := 0; i < 200; i++ {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("file%03d.txt", i)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := listFiles{}.Execute(context.Background(), `{"path":"`+dir+`","reason":"r"}`, 200)
	if err != nil {
		t.Fatalf("list_files: %v", err)
	}
	if len(got) > 200 {
		t.Errorf("list_files result %d bytes exceeds the 200 budget", len(got))
	}
	if !strings.HasSuffix(strings.TrimRight(got, "\n"), "... (truncated)") {
		t.Errorf("list_files did not truncate at the budget; tail=%q", tail(got, 40))
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
	got, err := readFiles{}.Execute(context.Background(), string(args), testBudget)
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

func TestReadFilesTruncatesAtBudget(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "big.txt")
	if err := os.WriteFile(p, []byte(strings.Repeat("x", testBudget+100)), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readFiles{}.Execute(context.Background(), `{"filepaths":["`+p+`"]}`, testBudget)
	if err != nil {
		t.Fatalf("read_files: %v", err)
	}
	if !strings.HasSuffix(strings.TrimRight(got, "\n"), "(truncated)") {
		t.Errorf("read_files did not truncate at the budget; tail=%q", tail(got, 40))
	}
	if len(got) > testBudget {
		t.Errorf("read_files result %d bytes is not bounded by the %d budget", len(got), testBudget)
	}
}

func TestReadFilesBinary(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bin")
	if err := os.WriteFile(p, []byte{0x00, 0x01, 0x02}, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readFiles{}.Execute(context.Background(), `{"filepaths":["`+p+`"]}`, testBudget)
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
	got, err := readFiles{}.Execute(context.Background(), `{"filepaths":["`+sub+`"]}`, testBudget)
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
		paths[i] = fmt.Sprintf("/tmp/r024/file%02d.txt", i)
	}
	args, _ := json.Marshal(map[string]any{"filepaths": paths, "reason": "r"})
	got, err := readFiles{}.Execute(context.Background(), string(args), testBudget)
	if err != nil {
		t.Fatalf("too-many-files must be a non-fatal result, got error %v", err)
	}
	if !strings.Contains(got, "requested too many files") || !strings.Contains(got, "Maximum is 50") {
		t.Errorf("read_files too-many message = %q", got)
	}
}

func TestReadFilesEmptyArgs(t *testing.T) {
	if _, err := (readFiles{}).Execute(context.Background(), `{"filepaths":[]}`, testBudget); err == nil {
		t.Error("empty filepaths must error")
	}
	if _, err := (readFiles{}).Execute(context.Background(), `{}`, testBudget); err == nil {
		t.Error("missing filepaths must error")
	}
}

func TestReadFilesAggregateBudgetSkips(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "big.txt")
	extra := filepath.Join(dir, "extra.txt")
	// big alone (≈0.9 MB) fits; adding extra overflows the 1 MB budget, so extra
	// is named as not read (no header).
	if err := os.WriteFile(big, []byte(strings.Repeat("y", 900000)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(extra, []byte(strings.Repeat("z", 900000)), 0o644); err != nil {
		t.Fatal(err)
	}
	args, _ := json.Marshal(map[string]any{"filepaths": []string{big, extra}, "reason": "r"})
	got, err := readFiles{}.Execute(context.Background(), string(args), testBudget)
	if err != nil {
		t.Fatalf("read_files: %v", err)
	}
	if len(got) > testBudget {
		t.Errorf("the aggregate result %d bytes exceeds the %d budget", len(got), testBudget)
	}
	if !strings.Contains(got, "truncated at the read budget") {
		t.Errorf("the aggregate budget marker is missing; tail=%q", tail(got, 60))
	}
	if !strings.Contains(got, "not read") || !strings.Contains(got, extra) {
		t.Errorf("the skip marker must name extra.txt; tail=%q", tail(got, 80))
	}
	if strings.Contains(got, "--- File: "+extra+" ---") {
		t.Errorf("a skipped file must get no header; got %q", tail(got, 80))
	}
}

// TestTruncateToBudgetIsBoundedAndUTF8Safe witnesses the budget cut: a
// multi-byte-heavy over-budget result is truncated to within the budget AND
// remains valid UTF-8 (no split rune).
func TestTruncateToBudgetIsBoundedAndUTF8Safe(t *testing.T) {
	big := strings.Repeat("├── x\n", testBudget) // 3-byte glyphs, ≫ budget
	got := truncateToBudget(big, testBudget)
	if len(got) > testBudget {
		t.Fatalf("truncateToBudget result %d bytes exceeds the %d budget", len(got), testBudget)
	}
	if !utf8.ValidString(got) {
		t.Fatal("truncateToBudget split a UTF-8 rune")
	}
	if !strings.HasSuffix(strings.TrimRight(got, "\n"), "... (truncated)") {
		t.Fatalf("truncateToBudget is missing the cap marker; tail=%q", tail(got, 40))
	}
	if small := "short"; truncateToBudget(small, testBudget) != small {
		t.Fatalf("truncateToBudget(%q) = %q; want unchanged", small, truncateToBudget(small, testBudget))
	}
}

// TestTruncateToBudgetDropsSplitRune makes the cut land MID-rune, so the
// ToValidUTF8 guard — not mere arithmetic — keeps the result valid UTF-8. The
// input is a run of 4-byte runes; the cut at budget - len(capMarker) is not a
// multiple of 4, so a byte cut alone would split a rune.
func TestTruncateToBudgetDropsSplitRune(t *testing.T) {
	big := strings.Repeat("\U0001F600", testBudget) // 4-byte rune, ≫ budget
	got := truncateToBudget(big, testBudget)
	if len(got) > testBudget {
		t.Fatalf("truncateToBudget result %d bytes exceeds the %d budget", len(got), testBudget)
	}
	if !utf8.ValidString(got) {
		t.Fatal("truncateToBudget left a split rune — the ToValidUTF8 guard is missing")
	}
}

// TestToolSchemasRequireReason pins the `reason` requirement of every
// builder-backed tool (the five tools the shared `resourceSchema` builds; the
// `execute_command` tool builds its schema inline). Round 031 strengthened it to
// assert that `reason` is a declared PROPERTY — not merely listed in `required` —
// which is the defect this round fixes (issue #64): a `required` name with no
// matching property is rejected by a strict provider (Vertex/Gemini). Completeness
// across the whole registry is owned by the cli-level gate
// (`TestAgentToolSchemasAreWellFormed`); this test keeps the `reason`-specific
// prose and must not re-enumerate the full tool set.
func TestToolSchemasRequireReason(t *testing.T) {
	backed := []struct {
		name   string
		params json.RawMessage
	}{
		{"list_files", listFiles{}.Parameters()},
		{"read_files", readFiles{}.Parameters()},
		{"get_tree", getTree{}.Parameters()},
		{"write_file", writeFile{}.Parameters()},
		{"replace_text", replaceText{}.Parameters()},
	}
	for _, tc := range backed {
		var s struct {
			Properties map[string]json.RawMessage `json:"properties"`
			Required   []string                   `json:"required"`
		}
		if err := json.Unmarshal(tc.params, &s); err != nil {
			t.Fatalf("%s: tool schema is not valid JSON: %v", tc.name, err)
		}
		if _, ok := s.Properties["reason"]; !ok {
			t.Errorf("%s: schema does not declare the `reason` property (required ⊆ properties would be violated): %s", tc.name, tc.params)
		}
		found := false
		for _, r := range s.Required {
			if r == "reason" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: tool schema does not require reason: %s", tc.name, tc.params)
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
