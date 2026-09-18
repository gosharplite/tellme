package main

import (
	"context"
	"os"
	"time"

	"github.com/gosharplite/tellme/internal/agent"
	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/cli"
	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/metrics"
	domainskills "github.com/gosharplite/tellme/internal/domain/skills"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	"github.com/gosharplite/tellme/internal/infrastructure/di"
	infrhistory "github.com/gosharplite/tellme/internal/infrastructure/history"
	infrallm "github.com/gosharplite/tellme/internal/infrastructure/llm"
	"github.com/gosharplite/tellme/internal/infrastructure/mcp"
	infrskills "github.com/gosharplite/tellme/internal/infrastructure/skills"
	infratelemetry "github.com/gosharplite/tellme/internal/infrastructure/telemetry"
	infratools "github.com/gosharplite/tellme/internal/infrastructure/tools"
	"github.com/gosharplite/tellme/internal/ui"
)

// mcpDiscoveryBound is the fixed, small per-server MCP fast-fail deadline
// (round-032 FR-008/NFR-002; relocated here by round 044). It also bounds the
// credential token source (FR-020).
const mcpDiscoveryBound = 3 * time.Second

// buildOptions constructs the single injected value internal/cli consumes
// (round 044 / ADR 0013). It is the composition root: the concrete adapters are
// built here (this package is exempt from the R1 tier table) and injected into
// the application layer as domain-typed seams.
func buildOptions() cli.Options {
	return cli.Options{Deps: buildDeps(), Prompter: ui.TUIPrompter{}}
}

// buildDeps assembles the domain-typed Dependencies value.
func buildDeps() deps.Dependencies {
	return deps.Dependencies{
		NewGateway:      infrallm.NewGateway,
		NewHistoryStore: func(workspace string) history.Store { return infrhistory.NewFileStore(workspace) },
		NewUsageStore:   func(workspace string) history.UsageStore { return infrhistory.NewUsageStore(workspace) },
		NewToolUsageStore: func(userHome func() (string, error)) history.ToolUsageStore {
			return infrhistory.NewToolUsageStore(userHome)
		},
		NewPromptTracker: func(home string, userHome func() (string, error)) history.PromptTracker {
			return infrhistory.NewGlobalPromptTracker(home, userHome)
		},
		NewToolRegistry: newToolRegistry,
		NewTUIRegistry:  newTUIRegistry,
		BindToolOutput:  infratools.BindToolOutput,
		BindSkillsCatalog: func(reg domaintools.Registry, skillsDir string) {
			infratools.BindSkillsCatalog(reg, func() ([]domainskills.Skill, error) {
				return infrskills.Load(skillsDir)
			})
		},
		NewMetricsProvider: func() metrics.SystemMetricsProvider { return infratelemetry.NewSystemMetricsProvider() },
		MCPDiscoverer: func(ctx context.Context, servers map[string]config.MCPServerConfig) ([]domaintools.Tool, []string, func()) {
			return mcp.Discover(ctx, servers, mcpDiscoveryBound, di.NewRemoteClient, di.NewGhTokenResolver(mcpDiscoveryBound))
		},
		LoopFactory: func(spec agentport.LoopSpec) agentport.Loop { return agent.NewLoop(spec) },
		UserHomeDir: os.UserHomeDir,
	}
}

// agentTools assembles the agent tool set in offer order: the read-only
// filesystem readers, the write pair (round 029), the bash-first command tool,
// and the read-only skills listing tool (round 033). It stays PARAMETERLESS and
// READ-FREE (round-033 FR-009): the `list_skills` catalog source is unbound here
// and bound on the prompt path only (round 044). Moved here from internal/cli by
// round 044 / ADR 0013; the round-031 well-formedness gate iterates it here.
func agentTools() []domaintools.Tool {
	tools := infratools.NewFilesystemTools()
	tools = append(tools, infratools.NewWriteTools()...)
	tools = append(tools, infratools.NewCommandTool())
	tools = append(tools, infratools.NewSkillsTool(nil))
	return tools
}

// newToolRegistry builds the seven-tool agent registry (the agent loop's set).
func newToolRegistry() domaintools.Registry {
	return domaintools.NewRegistry(agentTools()...)
}

// newTUIRegistry builds the three-reader registry the `-i` suggestion source
// consumes (round-044 fix-1) — a narrower set than newToolRegistry.
func newTUIRegistry() domaintools.Registry {
	return domaintools.NewRegistry(infratools.NewFilesystemTools()...)
}
