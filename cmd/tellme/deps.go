package main

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/gosharplite/tellme/internal/agent"
	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/cli"
	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/metrics"
	"github.com/gosharplite/tellme/internal/domain/render"
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
		NewTurnsLogStore: func(workspace string) history.TurnsLogStore {
			return infrhistory.NewTurnsLogStore(workspace)
		},
		NewToolUsageStore: func(userHome func() (string, error)) history.ToolUsageStore {
			return infrhistory.NewToolUsageStore(userHome)
		},
		NewPromptTracker: func(home string, userHome func() (string, error)) history.PromptTracker {
			return infrhistory.NewGlobalPromptTracker(home, userHome)
		},
		NewToolRegistry: newToolRegistry,
		NewTUIRegistry:  newTUIRegistry,
		BindSkillsCatalog: func(reg domaintools.Registry, skillsDir string) {
			infratools.BindSkillsCatalog(reg, func() ([]domainskills.Skill, error) {
				return infrskills.Load(skillsDir)
			})
		},
		NewMetricsProvider: func() metrics.SystemMetricsProvider { return infratelemetry.NewSystemMetricsProvider() },
		MCPDiscoverer: func(ctx context.Context, servers map[string]config.MCPServerConfig) deps.Discovery {
			tools, warnings, closeFn := mcp.Discover(ctx, servers, mcpDiscoveryBound, di.NewRemoteClient, di.NewGhTokenResolver(mcpDiscoveryBound))
			// F-7: the close is a visible field (round 051 / ADR 0020). Adapt the
			// bare func() into an io.Closer.
			var closer io.Closer
			if closeFn != nil {
				closer = closeFunc(closeFn)
			}
			return deps.Discovery{Tools: tools, Warnings: warnings, Closer: closer}
		},
		LoopFactory:  func(spec agentport.LoopSpec) agentport.Loop { return agent.NewLoop(spec) },
		NewLines:     func(colour bool) render.Lines { return ui.NewLines(colour) },
		NewToolLines: func(colour bool) agentport.ToolLineRenderer { return ui.ToolLines(colour) },
		NewAnswer:    func() render.Answer { return ui.NewAnswer() },
		NewListing:   func() render.Listing { return ui.NewListing() },
		NewProgress: func(spec render.ProgressSpec) render.TurnProgress {
			return ui.NewTurnProgress(spec, infratelemetry.NewSystemMetricsProvider())
		},
		UserHomeDir: os.UserHomeDir,
	}
}

// agentTools assembles the agent tool set in offer order: the read-only
// filesystem readers, the write pair (round 029), the bash-first command tool,
// the read-only skills listing tool (round 033), and — only for a vision-enabled
// provider (round 062; ADR 0032) — the image reader. It stays PARAMETERLESS for
// the no-vision default (the round-031 well-formedness gate iterates it here):
// the `list_skills` catalog source is unbound here and bound on the prompt path
// only (round 044). `read_image` is NOT part of this default set, so its ceiling
// input is unused (0).
func agentTools() []domaintools.Tool { return assembleAgentTools(deps.ToolSetSpec{}) }

// assembleAgentTools builds the agent tool set from ONE named capability value
// (round 069; ADR 0039), with the `[Tool Output]` sink injected into the command
// tool at CONSTRUCTION (round 052, closing #115 R-2; ADR 0021). agentTools()
// passes the zero spec (a nil sink — the round-031 gate and the offline
// `--tool-usage` path never execute a tool); the prompt path passes the live
// `prog.ToolOutput` sink. The `vision` gate (round 062) appends the `read_image`
// tool only when the selected provider declares the capability, so the offered
// set tells the model the truth; the family-aware inline ceiling the tool
// enforces (round 063; ADR 0033 D4) is resolved HERE (lazily, in the vision
// branch) from the spec's provider label via the single-owned `infrallm.Family`
// classifier — the resolution stays out of internal/cli (ADR 0039 D2).
func assembleAgentTools(spec deps.ToolSetSpec) []domaintools.Tool {
	tools := infratools.NewFilesystemTools()
	tools = append(tools, infratools.NewSearchTool()...)
	tools = append(tools, infratools.NewWriteTools()...)
	tools = append(tools, infratools.NewCommandTool(spec.Sink))
	tools = append(tools, infratools.NewSkillsTool(nil))
	if spec.Vision {
		tools = append(tools, infratools.NewReadImageTool(resolveImageCeiling(spec)))
	}
	return tools
}

// resolveImageCeiling maps a ToolSetSpec's provider label to the family-aware
// inline image ceiling the `read_image` tool enforces (round 063; ADR 0033 D4).
// It is a NAMED seam (round 069 / ADR 0039) so the family→ceiling CONSUMPTION is
// pinnable on its own (PR #141 fold F2): the classifier (`infrallm.Family`) and
// the ceiling table (`infratools.ImageCeilingForFamily`) are both single-owned,
// and this helper is their one composition point in the agent registry path.
func resolveImageCeiling(spec deps.ToolSetSpec) int {
	return infratools.ImageCeilingForFamily(infrallm.Family(spec.ProviderType))
}

// newToolRegistry builds the agent registry (the agent loop's set) from ONE
// named capability value (round 069; ADR 0039) — the former positional scalars
// `(sink, vision, providerType)` are gone, so a new capability is a ToolSetSpec
// FIELD, never a fourth positional argument.
func newToolRegistry(spec deps.ToolSetSpec) domaintools.Registry {
	return domaintools.NewRegistry(assembleAgentTools(spec)...)
}

// newTUIRegistry builds the three-reader registry the `-i` suggestion source
// consumes (round-044 fix-1) — a narrower set than newToolRegistry.
func newTUIRegistry() domaintools.Registry {
	return domaintools.NewRegistry(infratools.NewFilesystemTools()...)
}

// closeFunc adapts a bare close func() into an io.Closer (round 051 / F-7).
type closeFunc func()

func (f closeFunc) Close() error { f(); return nil }
