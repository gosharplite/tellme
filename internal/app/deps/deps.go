// Package deps defines the injected, domain-typed Dependencies value the
// composition root (cmd/tellme) builds and hands to internal/cli (round 044,
// ADR 0013). It imports internal/domain/** + internal/config + stdlib only (the
// R1 tier-2 ceiling; a UI-typed seam would be RULE-A).
package deps

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/metrics"
	"github.com/gosharplite/tellme/internal/domain/render"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// ToolSetSpec is the named capability value the agent tool-registry build
// consumes (round 069; ADR 0039). It replaces the former positional scalars
// `(sink domaintools.OutputSink, vision bool, providerType string)` — the seam
// recorded twice (ADR 0032 F-062-4 / RF-062-10; ADR 0033 RF-063-6). Its
// invariant: a NEW capability is a NEW FIELD here, never a fourth positional
// argument (D5). It carries the RAW inputs; the composition root
// (cmd/tellme) resolves ProviderType into the family-aware inline image ceiling
// (the resolution must stay out of internal/cli, which may not name
// infrastructure — the RULE-B 0-violation baseline; ADR 0039 D2).
type ToolSetSpec struct {
	// Sink is the `[Tool Output]` sink injected into the command tool at
	// construction (ADR 0021); nil on the offline `--tool-usage` path.
	Sink domaintools.OutputSink
	// Vision is the selected provider's declared capability gate (config
	// Provider.Vision): when true, the `read_image` tool is offered (ADR 0032).
	Vision bool
	// ProviderType is the provider's TYPE label; the composition root resolves it
	// to the family-aware inline image ceiling the `read_image` tool enforces
	// (ADR 0033 D4).
	ProviderType string
}

// Discovery is the result of one MCP discovery pass (round 051 / ADR 0020; F-7):
// the discovered tools, the warn+skip messages, and a closer for the opened
// clients — a NAMED type so the close is a visible field, not a bare trailing
// func() (PR #104 review deferral F-7).
type Discovery struct {
	Tools    []domaintools.Tool
	Warnings []string
	Closer   io.Closer
	// Refresh, when non-nil, is the round-087 post-answer stale-cache refresh: the
	// CLI runs it once the turn has produced its answer (best-effort) and prints
	// any returned warnings on the diagnostic stream. Nil when no cached MCP tool
	// entry was stale.
	Refresh func() []string
}

// Dependencies carries every seam the CLI used to hold as a package-level
// factory var (round 044 / ADR 0013). Each field is domain/config-typed so the
// application layer (internal/cli) names no infrastructure package. Every field
// has a named consumer (round-044 fix requirement: no field without an owner).
//
// The AGGREGATE is deliberately wide (a recorded accepted smell — G1); interface
// segregation is a deferred alternative recorded in ADR 0013.
type Dependencies struct {
	// NewGateway builds the provider gateway for the resolved provider entry.
	NewGateway func(prov config.Provider, name, persona string) (llm.Gateway, error)

	// NewHistoryStore builds the session-history store for a workspace.
	NewHistoryStore func(workspace string) history.Store
	// NewUsageStore builds the per-mode usage-log store for a workspace.
	NewUsageStore func(workspace string) history.UsageStore
	// NewTurnsLogStore builds the per-session turn-log store for a workspace
	// (round 053; ADR 0022). The CLI tees the rendered turn chrome into it on the
	// prompt path and reads/archives it from the offline session commands
	// (`-t`, `--new`).
	NewTurnsLogStore func(workspace string) history.TurnsLogStore
	// NewToolUsageStore builds the user-global tool-usage log adapter. It takes
	// the user-home resolver as an argument (the resolver it previously captured
	// from the deleted userHomeDir var) — so UserHomeDir has a real consumer.
	NewToolUsageStore func(userHome func() (string, error)) history.ToolUsageStore
	// NewPromptTracker builds the shared global prompt log adapter; it also takes
	// the user-home resolver.
	NewPromptTracker func(home string, userHome func() (string, error)) history.PromptTracker

	// NewToolRegistry builds the seven-tool agent registry (the agent loop's tool
	// set) from ONE named capability value — the caller passes the live
	// `[Tool Output]` sink on the prompt path and nil on the offline
	// `--tool-usage` path (round 052, closing #115 R-2; ADR 0021). Round 069
	// (ADR 0039) replaces the former positional scalars
	// (`sink domaintools.OutputSink, vision bool, providerType string`) with
	// `ToolSetSpec`, so a NEW capability is a field, never a fourth positional
	// argument (the seam recorded twice: ADR 0032 F-062-4 / RF-062-10 and
	// ADR 0033 RF-063-6).
	// NewTUIRegistry builds the three-reader registry the `-i` suggestion source
	// consumes — a DISTINCT, narrower set (round-044 fix-1; reusing the agent
	// registry would change the suggested tool names).
	NewToolRegistry func(spec ToolSetSpec) domaintools.Registry
	NewTUIRegistry  func() domaintools.Registry

	// BindSkillsCatalog rebinds the list_skills catalog source over the given
	// skills directory (the loader stays behind the composition root).
	BindSkillsCatalog func(reg domaintools.Registry, skillsDir string)

	// NewMetricsProvider builds the machine-wide system-metrics provider the
	// turn spinner's resource segment reads.
	NewMetricsProvider func() metrics.SystemMetricsProvider

	// MCPDiscoverer runs the round-032 prompt-path MCP discovery for a server
	// registry, consulting the round-087 cross-invocation tool cache (ADR 0058),
	// which round 088 (ADR 0059) places in the resolved per-mode WORKSPACE
	// (output/<mode>/), not the runtime home. Round 051 (F-7): it returns a NAMED
	// Discovery whose Closer is a visible field (rather than a bare trailing
	// func()); round 087 adds the post-answer Refresh hook.
	MCPDiscoverer func(ctx context.Context, workspace string, servers map[string]config.MCPServerConfig) Discovery

	// LoopFactory builds the agent loop for a turn (round 050; R5.4 of #92;
	// ADR 0019). It is the construction seam that replaced the (now-removed)
	// direct `&agent.AgentLoop{…}` build in internal/cli with an injected domain
	// port, so the application layer names no internal/agent type. The precise
	// field type is the domain port agentport.LoopFactory (LoopSpec -> Loop).
	LoopFactory agentport.LoopFactory

	// NewLines builds the status/tail line renderer (round 051; R5.5 of #92;
	// ADR 0020). The bytes stay owned by internal/ui, bound here as a domain
	// render.Lines port, so internal/cli names no internal/ui type. Round 054
	// (ADR 0023): the caller passes the chrome-colour flag (the diagnostic stream
	// is a terminal AND -r is off); the adapter applies the round-054 green
	// accents only when true.
	NewLines func(colour bool) render.Lines
	// NewToolLines builds the loop's four-line renderer (agentport.ToolLineRenderer)
	// with the round-054 chrome-colour flag.
	NewToolLines func(colour bool) agentport.ToolLineRenderer
	// NewAnswer builds the markdown answer renderer (round 006).
	NewAnswer func() render.Answer
	// NewListing builds the offline history-listing renderer (`-l`; round 073;
	// ADR 0045). Like NewAnswer, the bytes stay single-owned in internal/ui,
	// bound as a domain render.Listing port so internal/cli names no internal/ui
	// type.
	NewListing func() render.Listing
	// NewProgress builds a turn's coupled progress indicator + `[Tool Output]`
	// sink (round 051). It keeps the spinner/coordinator wiring out of internal/cli.
	NewProgress render.ProgressFactory

	// UserHomeDir resolves the user home (the `~/.tellme` root). It must be wired
	// unconditionally to a non-nil resolver (round-044 RF-2).
	UserHomeDir func() (string, error)
}

// Validate reports the first unbound (nil) function seam, so a missed wiring site
// fails with a message instead of a nil-func dereference deep inside
// internal/cli. It reflects over the struct's func fields, so a future 13th seam
// is covered by construction (F-5, PR #104 review 5243584043). It is called from
// the composition root's test and the CLI test fixture; `Run`'s zero value is
// still the caller's responsibility.
func (d Dependencies) Validate() error {
	v := reflect.ValueOf(d)
	for i := 0; i < v.NumField(); i++ {
		if f := v.Type().Field(i); f.Type.Kind() == reflect.Func && v.Field(i).IsNil() {
			return fmt.Errorf("deps: seam %s is not wired", f.Name)
		}
	}
	return nil
}
