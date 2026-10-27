package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/config"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	"github.com/gosharplite/tellme/internal/home"
)

// emitBootError maps a resolve failure to its actionable stderr message + code.
func emitBootError(stderr io.Writer, res resolution, rerr *resolveError) int {
	switch rerr.Reason {
	case reasonConfigMissing:
		if res.Explicit {
			_, _ = fmt.Fprintf(stderr, "tellme: the configuration could not be found at %s\n", res.Path)
		} else {
			_, _ = fmt.Fprintf(stderr, "tellme: no configuration could be found at %s\n", res.Path)
		}
		return ConfigError
	case reasonConfigInvalid:
		// A present-but-invalid value (e.g. a negative rendered width) uses the
		// general configuration-invalid class phrase (round-006 FR-006); a YAML
		// parse failure keeps the parse phrase. The sentinel prefix is stripped so
		// the message does not double the wording.
		if errors.Is(rerr.Err, config.ErrInvalidValue) {
			detail := strings.TrimPrefix(rerr.Err.Error(), config.ErrInvalidValue.Error()+": ")
			_, _ = fmt.Fprintf(stderr, "tellme: the configuration is invalid: %s\n", detail)
		} else {
			_, _ = fmt.Fprintf(stderr, "tellme: the configuration could not be parsed at %s\n", res.Path)
		}
		return ConfigError
	case reasonProviderMismatch:
		_, _ = fmt.Fprintf(stderr, "tellme: the selected provider is not in the registry (%q)\n", res.Selected)
		return ConfigError
	case reasonProviderInvalid:
		_, _ = fmt.Fprintf(stderr, "tellme: the provider configuration is invalid: provider %q: %v\n", res.Selected, rerr.Err)
		return ConfigError
	case reasonHomeUnusable:
		if errors.Is(rerr.Err, home.ErrNotDirectory) {
			_, _ = fmt.Fprintf(stderr, "tellme: the workspace path is not a directory (%s)\n", res.Workspace)
		} else {
			_, _ = fmt.Fprintf(stderr, "tellme: the runtime home is not usable (%v)\n", rerr.Err)
		}
		return EnvironmentError
	default: // reasonHomeUnset
		_, _ = fmt.Fprintln(stderr, "tellme: the runtime home is not usable")
		return EnvironmentError
	}
}

// emitProviderError maps a provider/transport failure to its frozen class phrase
// and dedicated exit code (round-004 FR-006..FR-008, Clarify Q3). The trailing
// detail is contract-free; newlines are folded so exactly one line carries the
// phrase.
func emitProviderError(w io.Writer, err error) int {
	detail := strings.ReplaceAll(err.Error(), "\n", " ")
	_, _ = fmt.Fprintf(w, "tellme: the provider request failed: %s\n", detail)
	return ProviderError
}

// augmentRegistryWithMCP performs the round-032 prompt-path MCP discovery: it
// discovers each enabled remote MCP server's tools (bounded, non-stall), offers
// them ALONGSIDE the native tools, and surfaces any warn+skip messages on the
// diagnostic stream. Discovery runs ONLY on the prompt path (this function is
// called from runTurn), so an offline run makes no MCP network contact. The
// returned close hook tears down the discovered clients when the turn ends.
func augmentRegistryWithMCP(ctx context.Context, res resolution, reg domaintools.Registry, stderr io.Writer, dp deps.Dependencies) (domaintools.Registry, func()) {
	d := dp.MCPDiscoverer(ctx, res.Workspace, res.MCPServers)
	for _, w := range res.MCPWarnings {
		_, _ = fmt.Fprintln(stderr, w)
	}
	for _, w := range d.Warnings {
		_, _ = fmt.Fprintln(stderr, w)
	}
	if len(d.Tools) > 0 {
		reg = domaintools.NewRegistry(append(reg.Tools(), d.Tools...)...)
	}
	// The returned hook runs when the turn ends (post-answer): first the round-087
	// stale-cache refresh (best-effort; a stale entry is revalidated off the
	// critical path, and a failure keeps the prior entry), then the client
	// teardown. It never affects the turn's outcome.
	refresh := d.Refresh
	closer := d.Closer
	closeFn := func() {
		if refresh != nil {
			for _, w := range refresh() {
				_, _ = fmt.Fprintln(stderr, w)
			}
		}
		if closer != nil {
			_ = closer.Close()
		}
	}
	return reg, closeFn
}

// emitToolError maps an incomplete tool loop to the frozen tool class phrase and
// dedicated exit code (round-008 FR-010 / Clarify R2 Q1). The trailing detail is
// contract-free; newlines are folded so exactly one line carries the phrase.
func emitToolError(w io.Writer, err error) int {
	detail := strings.ReplaceAll(err.Error(), "\n", " ")
	_, _ = fmt.Fprintf(w, "tellme: the tool request failed: %s\n", detail)
	return ToolError
}

// emitHistoryError maps a session-history read/write failure to the environment
// class phrase and exit code (round-007 Decision 6 / Clarify Q3). The trailing
// detail is contract-free; newlines are folded so exactly one line carries the
// phrase, keeping the frozen vocabulary at ten.
func emitHistoryError(w io.Writer, err error) int {
	detail := strings.ReplaceAll(err.Error(), "\n", " ")
	_, _ = fmt.Fprintf(w, "tellme: the runtime home is not usable (session history: %s)\n", detail)
	return EnvironmentError
}

// renderDiagnostic runs the -d path. It always emits a plain report and returns 0
// when resolution succeeded, else the dedicated diagnostic "unresolved" code.
// Round 002: the machine-readable `--json` form was removed; `--json` is no
// longer a flag, so any use of it is an unrecognized-flag usage error.
func renderDiagnostic(target sessionRef, stdout io.Writer) int {
	res, rerr := resolve(target)
	emitDiagnosticText(stdout, res, rerr)
	if rerr != nil {
		return DiagnosticUnresolvedError
	}
	return Success
}

// emitDiagnosticText writes the plain-text report.
func emitDiagnosticText(w io.Writer, res resolution, rerr *resolveError) {
	_, _ = fmt.Fprintln(w, "tellme diagnostic")
	if rerr == nil {
		_, _ = fmt.Fprintln(w, "configuration: resolved")
		_, _ = fmt.Fprintln(w, "runtime_home: "+res.Home)
		_, _ = fmt.Fprintln(w, "session_workspace: "+res.Workspace)
		return
	}
	_, _ = fmt.Fprintln(w, "configuration: unresolved")
	_, _ = fmt.Fprintln(w, "reason: "+rerr.Reason)
}

// defaultConfigPath is the default configuration path for the effective mode
// seed: $TELL_ME_HOME/configs/<seed>.yaml, where the seed is TELL_ME_MODE or
// "butler" (FR-007 discovery seed only).
func defaultConfigPath(homeDir string) string {
	seed := os.Getenv("TELL_ME_MODE")
	if seed == "" {
		seed = "butler"
	}
	return filepath.Join(homeDir, "configs", seed+".yaml")
}

// emitUsageError writes the usage-error stderr message and returns the usage
// error code (FR-014).
func emitUsageError(w io.Writer) int {
	_, _ = fmt.Fprintln(w, "tellme: the command-line usage is invalid")
	return UsageError
}
