package mcp

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 087 (ADR 0058) — the file-backed cache store.

func sampleEntry() domaintools.MCPToolCacheEntry {
	return domaintools.MCPToolCacheEntry{
		URL:       "https://mcp.example/mcp",
		Auth:      "auto",
		FetchedAt: time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC),
		Tools: []domaintools.MCPToolDefinition{{
			Name:        "lookup_price",
			Description: "look up a price",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"sku":{"type":"string"}}}`),
		}},
	}
}

func TestFileToolCache_SaveLoadRoundTrip(t *testing.T) {
	home := t.TempDir()
	c := NewFileToolCache(home)
	in := map[string]domaintools.MCPToolCacheEntry{"shop": sampleEntry()}
	if err := c.Save(in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	out, err := c.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got, ok := out["shop"]
	if !ok {
		t.Fatalf("Load did not return the saved key; got %v", out)
	}
	if got.URL != in["shop"].URL || got.Auth != in["shop"].Auth {
		t.Fatalf("declaration round-trip lost: got %q/%q", got.URL, got.Auth)
	}
	if !got.FetchedAt.Equal(in["shop"].FetchedAt) {
		t.Fatalf("fetched_at round-trip lost: got %v", got.FetchedAt)
	}
	if len(got.Tools) != 1 || got.Tools[0].Name != "lookup_price" {
		t.Fatalf("tools round-trip lost: got %+v", got.Tools)
	}
	if compact(got.Tools[0].InputSchema) != compact(in["shop"].Tools[0].InputSchema) {
		t.Fatalf("input_schema round-trip lost: got %s", got.Tools[0].InputSchema)
	}
}

// compact normalises a JSON value so a round-trip comparison ignores whitespace.
func compact(raw json.RawMessage) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return string(raw)
	}
	return buf.String()
}

func TestFileToolCache_MissingFileIsCold(t *testing.T) {
	c := NewFileToolCache(t.TempDir())
	m, err := c.Load()
	if err != nil {
		t.Fatalf("a missing file must not be an error; got %v", err)
	}
	if m != nil {
		t.Fatalf("a missing file must load as nil (every key cold); got %v", m)
	}
}

func TestFileToolCache_CorruptFileIsAnError(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, MCPToolCacheFileName), []byte("{ not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFileToolCache(home).Load(); err == nil {
		t.Fatalf("a corrupt cache must be an error (the caller treats every key cold)")
	}
}

// TestFileToolCache_SaveIsAtomic pins that Save writes through a temp file and
// renames (no partial payload is ever visible, and no temp file is left behind).
func TestFileToolCache_SaveIsAtomic(t *testing.T) {
	home := t.TempDir()
	c := NewFileToolCache(home)
	if err := c.Save(map[string]domaintools.MCPToolCacheEntry{"shop": sampleEntry()}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != MCPToolCacheFileName {
			t.Fatalf("Save left a stray file %q (temp file not renamed/removed)", e.Name())
		}
	}
}

// TestFileToolCache_NeverStoresCredential is the FR-009 / round-032 FR-017 pin:
// the persisted payload must never carry a token, a credential, or an
// Authorization header.
func TestFileToolCache_NeverStoresCredential(t *testing.T) {
	home := t.TempDir()
	c := NewFileToolCache(home)
	if err := c.Save(map[string]domaintools.MCPToolCacheEntry{"shop": sampleEntry()}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(home, MCPToolCacheFileName))
	if err != nil {
		t.Fatal(err)
	}
	lower := string(data)
	for _, banned := range []string{"token", "Token", "TOKEN", "Authorization", "Bearer", "secret"} {
		if contains(lower, banned) {
			t.Fatalf("the cache file must never persist a credential; found %q in:\n%s", banned, data)
		}
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// recordingCacheFS is a cacheFS that records the lifecycle events and can be
// made to fail the rename — the F-087-4 mechanism seam.
type recordingCacheFS struct {
	events  []string
	failRen bool
}

func (r *recordingCacheFS) CreateTemp(dir, pattern string) (*os.File, error) {
	r.events = append(r.events, "createtemp")
	return os.CreateTemp(dir, pattern)
}
func (r *recordingCacheFS) Rename(oldpath, newpath string) error {
	r.events = append(r.events, "rename")
	if r.failRen {
		return os.ErrPermission
	}
	return os.Rename(oldpath, newpath)
}
func (r *recordingCacheFS) Remove(name string) error {
	r.events = append(r.events, "remove")
	return os.Remove(name)
}

// TestFileToolCache_SaveUsesTempThenRename is the F-087-4 carrier: Save MUST go
// through a same-directory temp file that is renamed over the active path — an
// in-place os.WriteFile calls neither method, so the recorded events are empty.
func TestFileToolCache_SaveUsesTempThenRename(t *testing.T) {
	home := t.TempDir()
	fs := &recordingCacheFS{}
	c := &fileToolCache{path: filepath.Join(home, MCPToolCacheFileName), fs: fs}
	if err := c.Save(map[string]domaintools.MCPToolCacheEntry{"shop": sampleEntry()}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := strings.Join(fs.events, ","); got != "createtemp,rename" {
		t.Fatalf("Save must create a temp file then rename it (never write in place); events = %q", got)
	}
}

// TestFileToolCache_SaveFailureKeepsPrior is the F-087-4 second half: a failed
// rename leaves the PRIOR cache file byte-intact (no in-place truncate).
func TestFileToolCache_SaveFailureKeepsPrior(t *testing.T) {
	home := t.TempDir()
	if err := NewFileToolCache(home).Save(map[string]domaintools.MCPToolCacheEntry{"shop": sampleEntry()}); err != nil {
		t.Fatalf("seed Save: %v", err)
	}
	prior, err := os.ReadFile(filepath.Join(home, MCPToolCacheFileName))
	if err != nil {
		t.Fatal(err)
	}
	fs := &recordingCacheFS{failRen: true}
	c := &fileToolCache{path: filepath.Join(home, MCPToolCacheFileName), fs: fs}
	if err := c.Save(map[string]domaintools.MCPToolCacheEntry{"other": sampleEntry()}); err == nil {
		t.Fatalf("a failed rename must surface an error")
	}
	after, err := os.ReadFile(filepath.Join(home, MCPToolCacheFileName))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(prior) {
		t.Fatalf("a failed Save must leave the prior cache byte-intact")
	}
}
