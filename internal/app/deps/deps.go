// Package deps defines the injected, domain-typed Dependencies value the
// composition root (cmd/tellme) builds and hands to internal/cli (round 044,
// ADR 0013). It imports only internal/domain/**, internal/config and
// internal/home — never internal/infrastructure/**, internal/ui or
// internal/agent (the R1 tier-2 ceiling; a UI-typed seam would be RULE-A).
package deps

import (
	"context"

	"github.com/gosharplite/tellme/internal/config"
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

	// UserHomeDir resolves the user home (the `~/.tellme` root). It must be wired
	// unconditionally to a non-nil resolver (round-044 RF-2).
	UserHomeDir func() (string, error)
}
