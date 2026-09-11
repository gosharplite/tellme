// Package cli owns command-line flag parsing, exit-code classification, and
// diagnostic dispatch for the tellme CLI.
package cli

// Process exit codes.
//
// FR-014 requires distinct, deterministic codes for at least: success, usage
// error, configuration error, and environment error. The diagnostic
// "unresolved" code is additive (research.md Decision 2 / clarify Q1). The
// numeric values are an implementation choice — only distinctness and
// determinism are contractual (see TestExitCodesAreDistinct).
const (
	Success                   = 0 // a successful boot / reporting run
	UsageError                = 2 // unrecognized flag / invalid command-line usage
	ConfigError               = 3 // missing, malformed, or invalid configuration
	EnvironmentError          = 4 // runtime home unset/unusable, workspace unusable
	DiagnosticUnresolvedError = 5 // the -d report was produced but setup did not resolve
)
