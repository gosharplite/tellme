// Package deps defines the injected, domain-typed Dependencies value the
// composition root (cmd/tellme) builds and hands to internal/cli (round 044,
// ADR 0013). It imports internal/domain/** + internal/config + stdlib only (the
// R1 tier-2 ceiling; a UI-typed seam would be RULE-A).
package deps

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/metrics"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

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
	// NewToolUsageStore builds the user-global tool-usage log adapter. It takes
	// the user-home resolver as an argument (the resolver it previously captured
	// from the deleted userHomeDir var) — so UserHomeDir has a real consumer.
	NewToolUsageStore func(userHome func() (string, error)) history.ToolUsageStore
	// NewPromptTracker builds the shared global prompt log adapter; it also takes
	// the user-home resolver.
	NewPromptTracker func(home string, userHome func() (string, error)) history.PromptTracker

	// NewToolRegistry builds the seven-tool agent registry (the agent loop's tool
	// set). NewTUIRegistry builds the three-reader registry the `-i` suggestion
	// source consumes — a DISTINCT, narrower set (round-044 fix-1; reusing the
	// agent registry would change the suggested tool names).
	NewToolRegistry func() domaintools.Registry
	NewTUIRegistry  func() domaintools.Registry

	// BindToolOutput rebinds the live `[Tool Output]` sink on the command tool.
	BindToolOutput func(reg domaintools.Registry, sink domaintools.OutputSink)
	// BindSkillsCatalog rebinds the list_skills catalog source over the given
	// skills directory (the loader stays behind the composition root).
	BindSkillsCatalog func(reg domaintools.Registry, skillsDir string)

	// NewMetricsProvider builds the machine-wide system-metrics provider the
	// turn spinner's resource segment reads.
	NewMetricsProvider func() metrics.SystemMetricsProvider

	// MCPDiscoverer runs the round-032 prompt-path MCP discovery for a server
	// registry, returning the discovered tools, the warn+skip messages, and a
	// close hook. It is a func-typed port (the caller needs no new domain type).
	MCPDiscoverer func(ctx context.Context, servers map[string]config.MCPServerConfig) (tools []domaintools.Tool, warnings []string, close func())

	// LoopFactory builds the agent loop for a turn (round 050; R5.4 of #92;
	// ADR 0019). It is the construction seam that replaced the (now-removed)
	// direct `&agent.AgentLoop{…}` build in internal/cli with an injected domain
	// port, so the application layer names no internal/agent type. The precise
	// field type is the domain port agentport.LoopFactory (LoopSpec -> Loop).
	LoopFactory agentport.LoopFactory

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
