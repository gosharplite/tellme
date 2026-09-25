package tools

import "time"

// MCPToolCacheEntry is one server's cached MCP tool enumeration (round 087;
// ADR 0058). It records the server DECLARATION it was fetched under — the URL and
// the effective auth mode (NEVER the token, round-032 FR-017) — the fetch time
// (for the TTL), and the server's advertised tool definitions (name, description,
// input_schema) exactly as they would be offered.
//
// A cached entry is valid for a live config only when URL and Auth still match
// the declaration; a mismatch is a cold key (the declaration changed).
type MCPToolCacheEntry struct {
	URL       string              `json:"url"`
	Auth      string              `json:"auth"`
	FetchedAt time.Time           `json:"fetched_at"`
	Tools     []MCPToolDefinition `json:"tools"`
}

// MCPToolCache is the cross-invocation MCP tool cache seam (round 087; ADR 0058).
// The prompt prelude consults it before dialing any enabled remote MCP server, so
// a warm, fresh entry lets the offered tools be served with no network contact.
//
// It is BEST-EFFORT: an implementation MUST NOT turn a read or write failure into
// a run failure — a failure simply degrades to live discovery. The interface is
// domain-typed so the cache-aware orchestration is unit-testable with a fake.
type MCPToolCache interface {
	// Load returns the cache keyed by server key. A missing file is (nil, nil);
	// a corrupt file is an error the caller treats as "every key cold".
	Load() (map[string]MCPToolCacheEntry, error)
	// Save writes the whole cache atomically (a temp file + fsync + rename, the
	// ADR 0056 discipline) so a reader never sees a half-written payload.
	Save(map[string]MCPToolCacheEntry) error
}
