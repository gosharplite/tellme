package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// MCPToolCacheFileName is the cross-invocation cache file's name under
// $TELL_ME_HOME (round 087; ADR 0058). The home root (not output/<mode>/) keeps
// it shared across modes and untouched by `--new` (it is not session state).
const MCPToolCacheFileName = "mcp-toolcache.json"

// cacheFS is the unexported durability seam for the cache write (round 087 fold
// F-087-4; the round-084 `durableFS` precedent / ADR 0056). It exists so a unit
// pin can witness that Save writes through a SAME-DIRECTORY TEMP FILE that is
// renamed over the active path — an in-place `os.WriteFile` calls neither method
// and therefore reddens the pin, closing the "asserted mechanism with no carrier"
// gap. The `fsync` itself is a mechanism (a torn cache is merely a cold key,
// ADR 0058 D3), so it is not part of the seam.
type cacheFS interface {
	// CreateTemp creates a temp file in dir (the active file's directory).
	CreateTemp(dir, pattern string) (*os.File, error)
	// Rename atomically renames oldpath to newpath.
	Rename(oldpath, newpath string) error
	// Remove deletes name (temp-file cleanup on a write error).
	Remove(name string) error
}

// osCacheFS is the production cacheFS (os-backed).
type osCacheFS struct{}

func (osCacheFS) CreateTemp(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}
func (osCacheFS) Rename(oldpath, newpath string) error { return os.Rename(oldpath, newpath) }
func (osCacheFS) Remove(name string) error             { return os.Remove(name) }

// fileToolCache is the file-backed tools.MCPToolCache (round 087; ADR 0058). The
// write goes through cacheFS (a same-directory temp file, renamed over the active
// path) so a reader never observes a half-written payload; the `fsync` is a
// MECHANISM here, not an asserted guarantee (a torn or absent cache is merely a
// cold key).
type fileToolCache struct {
	path string
	fs   cacheFS
}

// NewFileToolCache returns the cache store rooted at home ($TELL_ME_HOME).
func NewFileToolCache(home string) domaintools.MCPToolCache {
	return &fileToolCache{path: filepath.Join(home, MCPToolCacheFileName), fs: osCacheFS{}}
}

// Load reads and decodes the cache. A missing file is (nil, nil) — every key is
// cold; a corrupt file is an error the caller treats the same way (best-effort).
func (c *fileToolCache) Load() (map[string]domaintools.MCPToolCacheEntry, error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var m map[string]domaintools.MCPToolCacheEntry
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Save atomically writes the whole cache: a same-directory temp file, `fsync`,
// then rename over the active path (never an in-place truncate). An empty map is
// a no-op; a write/rename failure removes the temp file and returns the error
// (the caller treats it best-effort — the prior cache survives).
func (c *fileToolCache) Save(m map[string]domaintools.MCPToolCacheEntry) error {
	if len(m) == 0 {
		return nil
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := c.fs.CreateTemp(filepath.Dir(c.path), ".mcp-toolcache-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = c.fs.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil { // fsync BEFORE the rename (ADR 0056 discipline)
		_ = tmp.Close()
		_ = c.fs.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = c.fs.Remove(tmpName)
		return err
	}
	if err := c.fs.Rename(tmpName, c.path); err != nil {
		_ = c.fs.Remove(tmpName)
		return err
	}
	return nil
}
