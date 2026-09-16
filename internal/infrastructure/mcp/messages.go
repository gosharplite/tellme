package mcp

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// The single-sourced MCP diagnostic messages (round-032 FR-008/FR-019). They are
// the ONE definition of the operator-facing warn+skip text, referenced by both
// the discovery layer (which emits them on stderr) and the E2E suite (which
// asserts on the exact product text — so the assertion cannot go vacuous if the
// product message changes, mirroring the round-012 cli.MultiLineHint pattern).
//
// They are NOT frozen class phrases: the frozen `tellme: <phrase>` vocabulary and
// the exit-code set are unchanged (FR-015) — a skip is a warning, not a failure.

// UnreachableWarning is the warn+skip message for a server that did not answer
// within the fixed fast-fail bound (or otherwise failed discovery).
func UnreachableWarning(server string) string {
	return fmt.Sprintf("[mcp] the server %q could not be reached; skipping its tools", server)
}

// UnreachableWarningWithHint is UnreachableWarning extended with a short, SAFE
// cause hint derived from the discovery error (round-032 SC-002 review): the
// base sentence stays a prefix, so callers/assertions matching UnreachableWarning
// still hold, and the hint only appends — e.g.
// `… could not be reached; skipping its tools (401 unauthorized)`.
//
// The hint is a fixed classification token only (an HTTP status, "timed out",
// …) — never any part of the raw error — so a resolved credential can never leak
// (FR-017). An unrecognised error yields the un-hinted base text.
func UnreachableWarningWithHint(server string, err error) string {
	base := UnreachableWarning(server)
	if h := reachabilityHint(err); h != "" {
		return base + " (" + h + ")"
	}
	return base
}

// statusCodePattern matches an HTTP 4xx/5xx status code anywhere in a transport
// error string (a token is never a bare 3-digit 4xx/5xx code, so this is safe).
var statusCodePattern = regexp.MustCompile(`\b([45][0-9]{2})\b`)

// reachabilityHint derives a SHORT, SAFE cause classification from a discovery
// error, so the skip warning is self-diagnosing rather than a single opaque
// "could not be reached" covering auth rejection, an unreachable host, a timeout,
// and the like. It returns only a fixed token — never the raw error text — so no
// credential can leak (FR-017). An unrecognised error yields "".
func reachabilityHint(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "timed out"
	}
	msg := strings.ToLower(err.Error())
	if code := statusCodePattern.FindString(msg); code != "" {
		switch code {
		case "401":
			return "401 unauthorized"
		case "403":
			return "403 forbidden"
		case "404":
			return "404 not found"
		case "429":
			return "429 rate limited"
		default:
			return code + " http error"
		}
	}
	switch {
	case strings.Contains(msg, "deadline exceeded"), strings.Contains(msg, "timeout"), strings.Contains(msg, "timed out"):
		return "timed out"
	case strings.Contains(msg, "connection refused"):
		return "connection refused"
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "name resolution"):
		return "name resolution failed"
	}
	return ""
}

// SchemaSkippedWarning is the warn+skip message for a tool whose advertised
// input schema could not be normalized to a safe object schema (FR-019/B1).
func SchemaSkippedWarning(server, tool string) string {
	return fmt.Sprintf("[mcp] the tool %q from server %q has an unusable schema; skipping it", tool, server)
}

// NameSkippedWarning is the warn+skip message for a tool whose namespaced wire
// name would not satisfy the provider tool-name grammar (round-032 F5).
func NameSkippedWarning(server, tool string) string {
	return fmt.Sprintf("[mcp] the tool %q from server %q has a name that cannot be offered safely; skipping it", tool, server)
}
