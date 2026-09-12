// Package cli owns command-line flag parsing, exit-code classification, and
// dispatch for the tellme CLI (boot, diagnostic, and the reasoning turn).
package cli

// Process exit codes.
//
// FR-014 requires distinct, deterministic codes for at least: success, usage
// error, configuration error, and environment error; the diagnostic
// "unresolved" code is additive (research.md Decision 2 / clarify Q1).
//
// Round 002 / FR-005 pinned these numeric values (success 0, usage 2,
// configuration 3, environment 4, diagnostic-unresolved 5). Round 004 / Clarify
// Q3 adds the provider error code 6, distinct from every other class. All values
// are published in specs/truth/features/cli/**/dsl.md and pinned by
// TestExitCodesMatchPinnedContract (distinctness by TestExitCodesAreDistinct).
//
// Round 005 (grill Q7): the environment class (code 4) also covers a standard
// input read failure — an I/O environment condition, not a usage/config/provider
// fault — which reuses the existing `the runtime home is not usable` class
// phrase with a "(standard input: …)" detail.
const (
	Success                   = 0 // a successful boot / reporting run
	UsageError                = 2 // unrecognized flag / invalid command-line usage
	ConfigError               = 3 // missing, malformed, or invalid configuration
	EnvironmentError          = 4 // runtime home unset/unusable, workspace unusable, standard input unreadable
	DiagnosticUnresolvedError = 5 // the -d report was produced but setup did not resolve
	ProviderError             = 6 // the provider request (transport / status / body) failed
)
