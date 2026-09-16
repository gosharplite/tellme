package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// mcpServerKeyPattern is the required server-key shape (round-032 FR-002): 1-24
// lowercase letters, digits, or hyphens.
var mcpServerKeyPattern = regexp.MustCompile(`^[a-z0-9-]{1,24}$`)

// MCPServerConfig is one entry under `MCP_SERVERS` (round-032 FR-001). tellme
// reads it from configuration only; it is NOT persisted system state (the data
// truth is NOOP this round).
//
// Decoding is tolerant (matching round 003's deliberate non-strict top-level
// decode): unmodelled sub-keys carried by a real-world `tell-me-go` block —
// `REQUIRES_CONSENT`, `ARGS`, `DIR`, `ENV`, … — are simply ignored, so an
// existing configuration never makes tellme refuse to start (round-032 FR-002,
// review fold TD4).
type MCPServerConfig struct {
	URL      string `yaml:"URL"`
	Token    string `yaml:"TOKEN"`
	Username string `yaml:"USERNAME"`
	Auth     string `yaml:"AUTH"`
	Timeout  int    `yaml:"TIMEOUT"`
	// Enabled is a per-server availability switch (round-032 FR-011/FR-012). A
	// pointer distinguishes "absent" (default true) from an explicit false.
	Enabled *bool `yaml:"ENABLED"`
	// Command marks a local stdio server (`COMMAND`), which this round does NOT
	// support: such an entry is classified first and warn+skipped, never judged a
	// malformed remote entry (round-032 FR-013).
	Command string `yaml:"COMMAND"`
}

// IsRemote reports whether the entry is a remote (Streamable HTTP) server. An
// entry with a COMMAND (and no URL) is a stdio server — out of scope this round.
func (s MCPServerConfig) IsRemote() bool { return strings.TrimSpace(s.Command) == "" }

// IsEnabled reports the effective availability: ENABLED defaults to true
// (FR-012).
func (s MCPServerConfig) IsEnabled() bool { return s.Enabled == nil || *s.Enabled }

// EffectiveAuth resolves the effective auth mode: AUTH when set, else "auto"
// (round-032 research Decision 4).
func (s MCPServerConfig) EffectiveAuth() string {
	if a := strings.TrimSpace(s.Auth); a != "" {
		return a
	}
	return "auto"
}

// MCPValidation is the outcome of validating the MCP_SERVERS registry:
// non-fatal warnings (e.g. a skipped COMMAND entry) plus, delimited by the
// returned error, any fatal malformed-remote-entry failure.
type MCPValidation struct {
	// Warnings are messages the caller surfaces on the diagnostic stream (a
	// warn+skip), never run failures.
	Warnings []string
}

// ValidateMCPServers validates the MCP_SERVERS registry deterministically
// (round-032 FR-002/FR-013), in sorted key order. Only a malformed REMOTE entry
// is a fatal error (the caller maps it to the existing configuration-invalid
// class phrase); a COMMAND-shaped (stdio) entry is warn+skipped, never fatal, so
// a real config carrying a stdio server does not invert the round's purpose.
func (c *Config) ValidateMCPServers() (MCPValidation, error) {
	var val MCPValidation
	if len(c.MCPServers) == 0 {
		return val, nil
	}
	keys := make([]string, 0, len(c.MCPServers))
	for name := range c.MCPServers {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {
		warn, err := validateMCPServerEntry(name, c.MCPServers[name])
		if err != nil {
			return val, err
		}
		if warn != "" {
			val.Warnings = append(val.Warnings, warn)
		}
	}
	return val, nil
}

// validateMCPServerEntry validates one entry. It returns a non-fatal warning for
// a skipped COMMAND (stdio) entry, or a fatal error for a malformed REMOTE entry
// (wrapping ErrInvalidValue), or ("", nil) for a valid remote entry.
func validateMCPServerEntry(name string, s MCPServerConfig) (string, error) {
	if !mcpServerKeyPattern.MatchString(name) {
		return "", fmt.Errorf("%w: MCP_SERVERS key %q must match ^[a-z0-9-]{1,24}$", ErrInvalidValue, name)
	}
	if !s.IsRemote() {
		return fmt.Sprintf("[mcp] the server %q uses a COMMAND (stdio) transport, which is not supported this round; skipping it", name), nil
	}
	if strings.TrimSpace(s.URL) == "" {
		return "", fmt.Errorf("%w: MCP_SERVERS.%s.URL is required for a remote server", ErrInvalidValue, name)
	}
	switch s.EffectiveAuth() {
	case "auto", "gh", "none":
	case "bearer", "basic":
		if strings.TrimSpace(s.Token) == "" {
			return "", fmt.Errorf("%w: MCP_SERVERS.%s requires TOKEN for AUTH %q", ErrInvalidValue, name, s.EffectiveAuth())
		}
		if s.EffectiveAuth() == "basic" && strings.TrimSpace(s.Username) == "" {
			return "", fmt.Errorf("%w: MCP_SERVERS.%s requires USERNAME for AUTH \"basic\"", ErrInvalidValue, name)
		}
	default:
		return "", fmt.Errorf("%w: MCP_SERVERS.%s.AUTH %q is not one of auto/gh/bearer/basic/none", ErrInvalidValue, name, s.EffectiveAuth())
	}
	if s.Timeout < 0 {
		return "", fmt.Errorf("%w: MCP_SERVERS.%s.TIMEOUT cannot be negative", ErrInvalidValue, name)
	}
	return "", nil
}

// ExpandMCPServers expands ${VAR} / ${VAR:-default} expressions in the
// MCP_SERVERS string fields, BEST-EFFORT (round-032 SC-002 / issue #67).
//
// The reference tell-me-go expands ${VAR} across MCP_SERVERS at config load, so
// an entry such as `TOKEN: "${GITHUB_TOKEN}"` authenticates. tellme previously
// sent the literal string (`Authorization: Bearer ${GITHUB_TOKEN}`), the endpoint
// rejected it (401), and the server was warn+skipped — a misleading "could not be
// reached" for what was actually an auth rejection.
//
// This is deliberately NON-FATAL: a variable that is unset without a default (or
// a malformed expression) leaves the field's ORIGINAL text in place. An
// unresolved ${VAR} therefore neither fails the load nor turns a would-be
// warn+skip into a startup failure — the server simply fails its handshake and is
// skipped, exactly as before. This mirrors round 003's tolerant-decode principle
// ("never make tellme refuse to start").
//
// Only the credential/endpoint surface is expanded here (URL, TOKEN, USERNAME,
// and COMMAND for surface consistency); the fields the deferred stdio transport
// adds (ARGS/DIR/ENV) join this list when that transport lands. The resolved
// values are NEVER logged (FR-017).
//
// It returns one non-fatal diagnostic warning per field whose expansion failed
// (an unset variable without a default, or a malformed expression), naming the
// server, field, and variable — so an unresolved credential is SELF-DIAGNOSING
// (round-032 SC-002 review) instead of surfacing only as an opaque
// "could not be reached". Note the deliberate divergence this makes explicit:
// `${VAR}` is STRICT for the selected provider entry (round-003 D2 errors on an
// unset variable) but BEST-EFFORT here, because an MCP server is optional and
// fail-open by design.
func (c *Config) ExpandMCPServers() []string {
	return c.expandMCPServersWithLookup(os.LookupEnv)
}

// expandMCPServersWithLookup is the injectable-lookup form (round-003 review F1
// precedent): tests drive an in-memory lookup so expansion stays a pure function
// of its inputs and the table runs without os.Setenv. Warnings are emitted in
// sorted server-key order, so the diagnostic output is deterministic.
func (c *Config) expandMCPServersWithLookup(lookup EnvLookupFunc) []string {
	var warnings []string
	names := make([]string, 0, len(c.MCPServers))
	for name := range c.MCPServers {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		s := c.MCPServers[name]
		for _, f := range []struct {
			field string
			value *string
		}{
			{"URL", &s.URL},
			{"TOKEN", &s.Token},
			{"USERNAME", &s.Username},
			{"COMMAND", &s.Command},
		} {
			out, warning := expandFieldWarning(name, f.field, *f.value, lookup)
			*f.value = out
			if warning != "" {
				warnings = append(warnings, warning)
			}
		}
		c.MCPServers[name] = s
	}
	return warnings
}

// expandFieldWarning expands ${VAR} / ${VAR:-default} in one MCP_SERVERS string
// field, returning the resolved value plus a non-fatal diagnostic warning that
// names the field and the offending variable when the expansion FAILS (an
// variable that is unset without a default, or a malformed expression). On
// failure the ORIGINAL text is returned unchanged, so expansion can never become
// a startup failure (issue #67); the warning is what makes the cause visible
// instead of an opaque "could not be reached" (round-032 SC-002 review). A value
// with nothing to expand is returned untouched with no warning.
func expandFieldWarning(server, field, value string, lookup EnvLookupFunc) (string, string) {
	if !strings.Contains(value, "${") {
		return value, ""
	}
	out, err := ExpandStringWithLookup(value, lookup)
	if err == nil {
		return out, ""
	}
	if errors.Is(err, ErrUnsetVariable) {
		name := strings.TrimPrefix(err.Error(), ErrUnsetVariable.Error()+": ")
		return value, fmt.Sprintf("[mcp] MCP_SERVERS.%s.%s references ${%s}, which is unset; sending the literal value", server, field, name)
	}
	return value, fmt.Sprintf("[mcp] MCP_SERVERS.%s.%s contains a malformed ${...} expression; sending the literal value", server, field)
}
