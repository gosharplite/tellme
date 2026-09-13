// Package suggestions implements tellme's multi-source suggestion engine — the
// application coordinator over the suggestions.Service domain port (round 015).
// It aggregates recent prompts, workspace entries (path-like queries only), and
// registered tool names, matching by subsequence, deduplicating, and capping the
// result (research Decision 2). The behaviour lands with the Feature phase
// (round-015 T032); this is the landing skeleton with the injected source seams.
package suggestions

import (
	"context"

	domainsuggestions "github.com/gosharplite/tellme/internal/domain/suggestions"
)

// PromptSource yields recent operator prompts, newest-first (the shared global
// log seeded first, then the active session's prompts).
type PromptSource interface {
	RecentPrompts(ctx context.Context, n int) []string
}

// WorkspaceSource yields workspace entries for a path-like query, scoped to the
// query's directory, read in bounded batches, excluding ignored directories
// (round-015 PR #38 review directive ③).
type WorkspaceSource interface {
	Entries(ctx context.Context, dir, prefix string, limit int) []string
}

// ToolSource yields tellme's registered tool names (the tool-suggestion source).
type ToolSource interface {
	ToolNames() []string
}

// Service is the multi-source suggestion engine. It implements
// domainsuggestions.Service.
type Service struct {
	prompts   PromptSource
	workspace WorkspaceSource
	tools     ToolSource
}

// New builds the engine over the three injected sources.
func New(prompts PromptSource, workspace WorkspaceSource, tools ToolSource) *Service {
	return &Service{prompts: prompts, workspace: workspace, tools: tools}
}

// Suggest returns the deduplicated, capped suggestions for query. Skeleton: the
// matching / debounce / aggregation logic lands with the Feature phase
// (round-015 T032). The sources and constants the real body will use are wired
// so the coordinator is complete and free of unused symbols.
func (s *Service) Suggest(ctx context.Context, query string) []domainsuggestions.Suggestion {
	_, _, _ = s.prompts, s.workspace, s.tools
	_ = ctx
	_ = query
	return nil
}

// Close drains background work before the process exits. Skeleton: nothing to
// drain yet (round-015 research Decision 3 / PR #38 review directive ⑤).
func (s *Service) Close(ctx context.Context) error { return nil }
