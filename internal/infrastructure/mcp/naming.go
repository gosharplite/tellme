package mcp

import (
	"crypto/sha256"
	"encoding/hex"
)

// The deterministic namespaced tool-name contract (round-032 FR-004, research
// Decision 5). Kept a pure helper so it is unit-testable in isolation.

const (
	// mcpNamePrefix namespaces every MCP-provided tool so it cannot collide with
	// a native tool.
	mcpNamePrefix = "mcp_"
	// maxToolNameLen is the wire maximum for a tool name (the OpenAI
	// `tools[i].function.name` limit `^[a-zA-Z0-9_-]{1,64}$`).
	maxToolNameLen = 64
	// hashLen is the number of hex characters of the SHA-256 suffix appended when
	// a name must be truncated for uniqueness.
	hashLen = 8
	// prefixCap bounds the retained tool-name prefix so the fixed segments always
	// fit; the `{1,24}` server-key bound leaves room by construction.
	prefixCap = 40
)

// NamespacedName returns the deterministic, namespaced wire name for a tool a
// server advertises: `mcp_<server>_<tool>` when that fits the 64-byte wire
// maximum. When it does not, the tool segment is truncated to a DERIVED budget —
// `50 - len(server)` bytes, capped at 40, floored at 0 — and an 8-hex SHA-256
// prefix of the full tool name is appended, so the result is always ≤ 64 BYTES
// (measured with len(), not runes) and two tools that share a truncated prefix
// stay distinct (round-032 TD3).
func NamespacedName(server, tool string) string {
	full := mcpNamePrefix + server + "_" + tool
	if len(full) <= maxToolNameLen {
		return full
	}
	maxPrefix := maxToolNameLen - len(mcpNamePrefix) - len(server) - 1 - 1 - hashLen
	if maxPrefix > prefixCap {
		maxPrefix = prefixCap
	}
	if maxPrefix < 0 {
		maxPrefix = 0
	}
	sum := sha256.Sum256([]byte(tool))
	hash8 := hex.EncodeToString(sum[:4]) // 8 hex characters
	truncated := tool
	if len(truncated) > maxPrefix {
		truncated = truncated[:maxPrefix]
	}
	return mcpNamePrefix + server + "_" + truncated + "_" + hash8
}
