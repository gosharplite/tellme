// Package tools defines the agent-tool domain port: the model-facing tool
// abstraction and the registry the agent loop dispatches against. It is
// network-free and knows nothing about execution — the concrete tools live in
// internal/infrastructure/tools (round-008 research Decision 1).
package tools

import (
	"context"
	"encoding/json"
	"time"
)

// ToolContract is the UPWARD half of the two-way Tool port (round-024 Q2): the
// per-tool resource descriptors the loop needs to resolve a call's effective
// bound but cannot derive from the model's arguments (a *default* is what applies
// when the param is absent). It is a small descriptor (not a raft of methods) so
// it can grow without a second signature churn.
type ToolContract struct {
	// DefaultTimeout is the tool's protective per-call timeout default (FR-016:
	// shell 300 s, readers 30 s). It is the fallback the loop uses when neither
	// the call's `timeout` param nor a loop-level default supplies one.
	DefaultTimeout time.Duration
}

// ByteBudget is the DOWNWARD half of the two-way Tool port (round-024 D4): the
// resolved result BYTE budget the tool bounds its own output to at the source
// (the loop's raw-`len` clamp is a strictly-larger backstop). It is a typed value
// — not a context `any` — so the contract is explicit.
type ByteBudget int

// TruncationMarker terminates a result cut at its byte budget (round-024 D7). It
// is the SINGLE source for the marker emitted by both the tool adapters and the
// loop's backstop (review TD3), so the two copies cannot drift.
const TruncationMarker = "\n... (truncated)\n"

// ReasonArgKey is the SINGLE authoritative spelling of the call-level `reason`
// argument tellme renders as `[Tool Reason]` and (round 056 / ADR 0025) refuses a
// call without. It is shared so the wire key cannot drift across the loop's
// extraction, the native schema builders, the MCP envelope, and the
// presentation seam (round-056 review R-056-1). NOTE: a struct-tag
// (`json:"reason"`) and raw schema-template literals cannot interpolate it; those
// sites are documented in ADR 0025 §Forward.
const ReasonArgKey = "reason"

// PayloadArgKey is the MCP envelope property carrying a remote server's own
// arguments (round 056 / ADR 0025 D1/D2). tellme owns the envelope; the server
// never sees this key (only the payload object is forwarded).
const PayloadArgKey = "MCP_PAYLOAD"

// Tool is one capability the model may invoke during a prompt run. Execute runs
// the tool with the model's raw arguments string and the resolved byte budget,
// and returns its result text. A tool failure is returned as an error and is
// treated by the loop as a non-terminal tool result (the model may recover); it
// is NOT a process failure. A tool that OBSERVES its effective timeout returns a
// nil-error timeout result (round-024 FR-018), never an error.
//
// The context lets the loop bound each execution with a per-tool timeout
// (round-008 RF-2). Contract exposes the tool's own default timeout (round-024
// Q2) so the loop stays tool-name-agnostic.
type Tool interface {
	Name() string
	Description() string
	Parameters() json.RawMessage
	Contract() ToolContract
	Execute(ctx context.Context, arguments string, budget ByteBudget) (string, error)
}

// MediaTool is the OPTIONAL capability contract for a tool whose result also
// carries media (round 070; ADR 0040). The agent loop type-asserts it: a tool
// that satisfies it is executed via ExecuteMedia, which returns the text result
// AND the media it produced IN-BAND (no ambient context channel); every other
// tool keeps the plain Tool contract untouched. It is a SEGREGATED CAPABILITY
// interface (the history.Seeder / LoopObserver precedent), NOT a widening of
// Tool — media is the exception today (`read_image` is the sole producer). The
// rejected alternative (widening Execute for every tool) is recorded settled in
// ADR 0040.
type MediaTool interface {
	Tool
	// ExecuteMedia runs like Execute and additionally returns the media the call
	// produced. A failure is returned as an error (as in Execute); an empty media
	// slice means the call produced no media.
	ExecuteMedia(ctx context.Context, arguments string, budget ByteBudget) (text string, media []MediaPart, err error)
}

// Registry resolves a tool by its wire name and lists the registered tools.
type Registry interface {
	Lookup(name string) (Tool, bool)
	Tools() []Tool
}

// registry is the default Registry over a fixed, ordered tool set.
type registry struct {
	byName map[string]Tool
	order  []Tool
}

// NewRegistry builds a Registry from a set of tools. The tool order is the
// order they were supplied (the order offered to the model).
func NewRegistry(ts ...Tool) Registry {
	r := &registry{byName: make(map[string]Tool, len(ts)), order: make([]Tool, 0, len(ts))}
	for _, t := range ts {
		r.byName[t.Name()] = t
		r.order = append(r.order, t)
	}
	return r
}

// Lookup returns the tool with the given wire name.
func (r *registry) Lookup(name string) (Tool, bool) {
	t, ok := r.byName[name]
	return t, ok
}

// Tools returns the registered tools in offer order.
func (r *registry) Tools() []Tool { return r.order }
