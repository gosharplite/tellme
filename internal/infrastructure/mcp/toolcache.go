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

// fileToolCache is the file-backed tools.MCPToolCache (round 087; ADR 0058). The
// write follows the round-084 durability discipline (a same-directory temp file,
// `fsync`ed and atomically renamed) so a reader never observes a half-written
// payload; the `fsync` is a MECHANISM here, not an asserted guarantee (a torn or
// absent cache is merely a cold key).
type fileToolCache struct {
	path string
}

// NewFileToolCache returns the cache store rooted at home ($TELL_ME_HOME).
func NewFileToolCache(home string) domaintools.MCPToolCache {
	return &fileToolCache{path: filepath.Join(home, MCPToolCacheFileName)}
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
// then `rename` over the active path. An empty map is a no-op.
func (c *fileToolCache) Save(m map[string]domaintools.MCPToolCacheEntry) error {
	if len(m) == 0 {
		return nil
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(c.path), ".mcp-toolcache-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil { // fsync BEFORE the rename (ADR 0056 discipline)
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, c.path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}
