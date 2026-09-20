package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Round 071 (ADR 0043): search_files — the bounded, deterministic in-file content
// search (the missing half of the reader trio).

func TestSearchFilesFindsMatchesDeterministically(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("first\nhas needle here\nlast\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "b.txt"), []byte("also needle\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := searchFiles{}.Execute(context.Background(), `{"path":"`+dir+`","query":"needle","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatalf("search_files: %v", err)
	}
	if !strings.Contains(got, "a.txt:2: has needle here") {
		t.Errorf("search_files missing the a.txt:2 match; got %q", got)
	}
	if !strings.Contains(got, "b.txt:1: also needle") {
		t.Errorf("search_files missing the sub/b.txt:1 match; got %q", got)
	}
	if strings.Index(got, "a.txt:2") > strings.Index(got, "b.txt:1") {
		t.Errorf("search_files order is not path-ascending; got %q", got)
	}
}

func TestSearchFilesNoMatchIsAResult(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("nothing here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := searchFiles{}.Execute(context.Background(), `{"path":"`+dir+`","query":"absent","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatalf("a no-match search must not error: %v", err)
	}
	if !strings.Contains(got, "0 matches found") {
		t.Errorf("search_files no-match result = %q; want 0 matches found", got)
	}
}

func TestSearchFilesEmptyQueryErrors(t *testing.T) {
	if _, err := (searchFiles{}).Execute(context.Background(), `{"query":""}`, testBudget); err == nil {
		t.Error("an empty query must error")
	}
	if _, err := (searchFiles{}).Execute(context.Background(), `{"reason":"r"}`, testBudget); err == nil {
		t.Error("a missing query must error")
	}
}

func TestSearchFilesInvalidRegexErrors(t *testing.T) {
	dir := t.TempDir()
	_, err := searchFiles{}.Execute(context.Background(), `{"path":"`+dir+`","query":"(","is_regex":true,"reason":"r"}`, testBudget)
	if err == nil || !strings.Contains(err.Error(), "invalid regex") {
		t.Errorf("an invalid regex must be a stable error naming the pattern; got %v", err)
	}
}

func TestSearchFilesLiteralByDefault(t *testing.T) {
	dir := t.TempDir()
	// A literal "." matches only the line that contains a dot; as a regex it
	// would match every line — the literal-default contract (Q1).
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("has.dot\nno dot\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	literal, err := searchFiles{}.Execute(context.Background(), `{"path":"`+dir+`","query":".","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(literal, "has.dot") || strings.Contains(literal, "no dot") {
		t.Errorf("a literal '.' must match only the dotted line; got %q", literal)
	}
	regex, err := searchFiles{}.Execute(context.Background(), `{"path":"`+dir+`","query":".","is_regex":true,"reason":"r"}`, testBudget)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(regex, "has.dot") || !strings.Contains(regex, "no dot") {
		t.Errorf("a regex '.' must match every line; got %q", regex)
	}
}

func TestSearchFilesSkipsBinary(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bin"), []byte("needle\x00hidden"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := searchFiles{}.Execute(context.Background(), `{"path":"`+dir+`","query":"needle","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "0 matches found") {
		t.Errorf("a binary file must be skipped, leaving no matches; got %q", got)
	}
}

func TestSearchFilesTruncatesAtBudget(t *testing.T) {
	dir := t.TempDir()
	var sb strings.Builder
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&sb, "needle line %03d padding padding padding padding\n", i)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := searchFiles{}.Execute(context.Background(), `{"path":"`+dir+`","query":"needle","reason":"r"}`, 120)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) > 120 {
		t.Errorf("search_files result %d bytes exceeds the 120 budget", len(got))
	}
	if !strings.HasSuffix(strings.TrimRight(got, "\n"), "... (truncated)") {
		t.Errorf("search_files did not truncate at the budget; tail=%q", tail(got, 60))
	}
}

func TestSearchFilesCapsMatches(t *testing.T) {
	dir := t.TempDir()
	var sb strings.Builder
	for i := 0; i < searchMaxMatches+50; i++ {
		sb.WriteString("needle\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := searchFiles{}.Execute(context.Background(), `{"path":"`+dir+`","query":"needle","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(got, "a.txt:") != searchMaxMatches {
		t.Errorf("search_files reported %d matches; want the %d cap", strings.Count(got, "a.txt:"), searchMaxMatches)
	}
	if !strings.Contains(got, "(truncated)") {
		t.Errorf("a capped search must carry the truncation marker; tail=%q", tail(got, 40))
	}
}

func TestSearchFilesTrimsLongLines(t *testing.T) {
	dir := t.TempDir()
	long := "needle" + strings.Repeat("x", searchLineCap+100)
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte(long+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := searchFiles{}.Execute(context.Background(), `{"path":"`+dir+`","query":"needle","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatal(err)
	}
	// The reported line: `a.txt:1: ` + the capped text.
	idx := strings.Index(got, ": ")
	line := strings.TrimRight(got[idx+2:], "\n")
	if len(line) > searchLineCap {
		t.Errorf("matched line %d bytes exceeds the %d cap", len(line), searchLineCap)
	}
}

// TestSearchFilesDeterministicOrderAcrossDirsAndFiles pins the sort: a root file
// "a.go" sorts BEFORE a sub-folder's "a/x.txt" ('.' < '/'), which the raw
// depth-first walk order does NOT produce (the folder "a" is visited first). This
// is the round-071 determinism witness.
func TestSearchFilesDeterministicOrderAcrossDirsAndFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("needle\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a", "x.txt"), []byte("needle\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := searchFiles{}.Execute(context.Background(), `{"path":"`+dir+`","query":"needle","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(got, "a.go:1") < 0 || strings.Index(got, "a/x.txt:1") < 0 {
		t.Fatalf("search_files missing a match; got %q", got)
	}
	if strings.Index(got, "a.go:1") > strings.Index(got, "a/x.txt:1") {
		t.Errorf("search_files is not sorted path-ascending (a.go must precede a/x.txt); got %q", got)
	}
}
