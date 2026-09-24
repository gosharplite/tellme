// Package render declares the CLI's presentation seams (round 051; R5.5 of
// #92; ADR 0020). It is the domain-owned contract through which internal/cli
// obtains its status/tail lines, its answer renderer, and its turn progress
// indicator — so the CLI names no internal/ui type. The BYTES (sanitize / fold /
// cap / format policy) stay single-owned in internal/ui, which implements these
// ports; the domain owns only the shapes. RULE-C-pure: stdlib + domain types
// only.
package render

import (
	"io"
	"time"

	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/metrics"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Lines renders the CLI's status/tail diagnostic lines. It mirrors the
// internal/ui formatters; the adapter delegates to them verbatim.
type Lines interface {
	InputCaptured(t time.Time) string
	TurnOpening(turn int, mode string) string
	TurnGap() string
	// PayloadMeasured renders the MEASURED payload line (`<tokens>/<budget> tokens
	// - <mode> - <model>`; round 009). Round 057 (ADR 0027; TD-057-1): the retired
	// `~`-estimated shape no longer exists here — the estimate is its own method
	// (PayloadEstimate), so ONE shape per line and no unreachable branch.
	PayloadMeasured(t time.Time, tokens, budget int, mode, model string) string
	// PayloadEstimate renders the ESTIMATED pre-flight payload line (round 057;
	// ADR 0027): `+<delta> ~<tokens> tokens - <mode> - <model>`. The delta is the
	// increment over the previous estimate (`0` when there is none) and the budget
	// is not shown.
	PayloadEstimate(t time.Time, tokens, delta int, mode, model string) string
	Metrics(t time.Time, provider string, u metrics.UsageCounts) string
	Ready(lastCallCost, turnCost, sessionCost float64, sessionMiss, sessionHit, sessionOut int, hitRate float64) string
	ToolReason(t time.Time, reason string) string
	// UnpairedCalls renders the round-068 unpaired-call diagnostic (ADR 0038):
	// a `[Tool …]`-class line naming the tool calls a round left unanswered.
	UnpairedCalls(t time.Time, ids []string) string
	ToolUsage(rows []history.ToolUsageRow) string
	// DefaultToolOutputIdleGap is the `[Tool Output]` block's idle-gap default
	// (ADR 0009 D3).
	DefaultToolOutputIdleGap() time.Duration
}

// Answer renders a Markdown answer to ANSI (round 006). It mirrors the CLI's
// former in-package `answerRenderer` interface.
type Answer interface {
	// Render returns the rendered answer and whether rendering degraded (in which
	// case the returned string is the sanitized raw fallback).
	Render(markdown string, width int) (text string, degraded bool)
	// WarnDegraded emits the one-time degradation warning.
	WarnDegraded(w io.Writer)
}

// ListingRole is the presented role of a listed history message (round 073;
// ADR 0045). It is a DOMAIN enum — the adapter owns the header text
// (`[USER]` / `[MODEL]`) — so the CLI never hands the stored/derived wire role
// (`user`/`assistant`) to a presentation seam.
type ListingRole int

const (
	// ListingOperator is the operator's own prompt.
	ListingOperator ListingRole = iota
	// ListingModel is the model's answer.
	ListingModel
)

// ListingMessage is one listed history message: its presented role, its
// (unrendered) body text, its backward turn index, and — for a model message —
// the turn's tool-call count.
//
// TurnIndex is the distance of the message's turn from the MOST RECENT turn of
// the loaded history, counting back: 1 = the most recent history entry, 2 = the
// one before it, and so on. Both messages of one entry carry the SAME index (a
// turn is atomic under `-b`), and it is the TRUE distance from the end of the
// loaded history — never derived from the (possibly truncated) message slice, so
// a lone leading message of an odd `-l N` keeps its true distance. It is
// computed by the caller (round 082; ADR 0054). A value <= 0 (a non-history
// caller) makes the adapter print the bare role label with no suffix.
//
// ToolCount is the number of tool steps the message's turn recorded
// (`len(history.Entry.Steps)` — round 086; ADR 0057), carried on the turn's
// MODEL message. The adapter prints a `[TOOLS] - <TurnIndex> (<ToolCount> calls)`
// line immediately before the `[MODEL]` header (between the turn's `[USER]` block
// and its `[MODEL]` header), so a partially-listed turn still shows it. It is the
// TOOL-call count, NOT `history.Entry.Calls` (the AI-endpoint inference-round
// count), and only the count crosses this seam — never any tool content.
type ListingMessage struct {
	Role      ListingRole
	Body      string
	TurnIndex int
	ToolCount int
}

// ListingSpec carries the listing's presentation inputs. Colour is the
// stdout-terminal gate (the STDOUT stream is a terminal — a distinct axis from
// the stderr-chrome gate — and `-r` is off; round 073); Raw is the `-r`/`--raw`
// flag (the model
// body is printed verbatim and no colour is emitted); Width is the resolved
// rendered width (0 = the renderer's built-in default); Warn, when non-nil,
// receives the one-time markdown-degradation warning.
type ListingSpec struct {
	Colour bool
	Raw    bool
	Width  int
	Warn   io.Writer
}

// Listing renders the offline session history listing (`-l`/`--list`; round
// 073; ADR 0045): for each message a role header line (`[USER]` / `[MODEL]`),
// the body (the MODEL body rendered as Markdown unless Raw; the operator body
// verbatim), and one blank line after every message. Round 082 (ADR 0054): the
// role header carries the backward turn index. Round 086 (ADR 0057): each turn's
// MODEL message is prefixed by a `[TOOLS] - <index> (<count> calls)` line, its
// own block between the turn's `[USER]` block and its `[MODEL]` header. Colour,
// when enabled, accents the header lines and the tool line.
type Listing interface {
	Render(w io.Writer, msgs []ListingMessage, spec ListingSpec)
}

// Indicator is the turn progress indicator (the spinner) as the CLI drives it.
// It is the loop's waiting-phase observer (agentport.LoopObserver) plus a Stop
// that clears it for the rest of the run.
type Indicator interface {
	agentport.LoopObserver
	Stop()
}

// TurnProgress bundles the coupled turn-scoped presentation objects a turn
// builds together: the progress Indicator (nil when the spinner is gated off)
// and the `[Tool Output]` sink (always present; it renders unconditionally).
type TurnProgress struct {
	Indicator  Indicator
	ToolOutput domaintools.OutputSink
}

// ProgressSpec is the progress factory's named-field input (round 057 fold
// TD-057-2; the repo's `LoopSpec` pattern, ADR 0019). It replaces a positional
// call whose two trailing bools meant DIFFERENT policies — the round-054 spinner
// gate and the round-054/057 chrome-colour gate — which a silent argument swap
// could not be caught on.
type ProgressSpec struct {
	Stream  io.Writer
	Now     func() time.Time
	Model   string
	Epoch   time.Time
	Columns func() int
	// IdleGap is the `[Tool Output]` block's resume threshold (ADR 0009 D3).
	IdleGap time.Duration
	// Spinner is the round-054 spinner gate (a bool-returning probe).
	Spinner bool
	// Colour is the chrome-colour gate (ADR 0023/0027): the diagnostic stream is a
	// terminal AND `-r` is off. Today it resolves to the same predicate as Spinner,
	// but it is a distinct policy — naming it keeps a future divergence honest.
	Colour bool
}

// ProgressFactory builds a turn's TurnProgress from a ProgressSpec. It is the
// injected seam (a func-typed deps field) that keeps internal/cli free of
// internal/ui.
type ProgressFactory func(spec ProgressSpec) TurnProgress
