// Package suggestions implements tellme's multi-source suggestion engine — the
// application coordinator over the suggestions.Service domain port (round 015).
// It aggregates recent prompts, workspace entries (path-like queries only), and
// registered tool names, matching by subsequence, deduplicating, and capping the
// result (research Decision 2).
package suggestions

import (
	"context"
	"path/filepath"
	"strings"

	domainsuggestions "github.com/gosharplite/tellme/internal/domain/suggestions"
)

// maxSuggestions is the surface cap: the prompt never shows more than this many
// suggestions (round-015 research Decision 2).
const maxSuggestions = 10

// promptPoolDepth is how many newest distinct recent prompts the history source
// is asked for — the candidate pool the engine matches against. Round 064
// (ADR 0034): deepened from 10 to the reference's newest-50 pool, so a match
// older than the newest 10 is offered again; the surfaced list stays capped at
// maxSuggestions. These two constants are deliberately DISTINCT (the depth and
// the cap were conflated before this round).
const promptPoolDepth = 50

// dirBatch bounds a single directory read batch, so a query never triggers an
// unbounded scan (round-015 PR #38 review directive ③).
const dirBatch = 100

// PromptSource yields recent operator prompts, newest-first. The production
// source is the user-global shared prompt log (`~/.tellme/global_prompts.jsonl`,
// round 028); there is no separate session source (round 064 / ADR 0034 corrects
// the earlier "seeded first, then the active session's prompts" note — the shared
// log is the only prompt source).
type PromptSource interface {
	RecentPrompts(ctx context.Context, n int) []string
}

// WorkspaceSource yields workspace entries for a path-like query, scoped to the
// query's directory, read in bounded batches, excluding ignored directories.
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

// Suggest returns the deduplicated, capped suggestions for query. An empty query
// yields the newest recent prompts. Workspace entries are consulted only for a
// path-like query; the workspace source honours ctx and skips noisy directories.
func (s *Service) Suggest(ctx context.Context, query string) []domainsuggestions.Suggestion {
	query = strings.TrimSpace(query)
	acc := newAccumulator(maxSuggestions)
	s.addPrompts(ctx, query, acc)
	if acc.full() || query == "" {
		return acc.out // an empty query shows only the recent prompts
	}
	s.addWorkspace(ctx, query, acc)
	if acc.full() {
		return acc.out
	}
	s.addTools(query, acc)
	return acc.out
}

// accumulator collects deduplicated suggestions up to a cap.
type accumulator struct {
	out   []domainsuggestions.Suggestion
	seen  map[string]bool
	limit int
}

func newAccumulator(limit int) *accumulator {
	return &accumulator{seen: map[string]bool{}, limit: limit}
}

func (a *accumulator) add(text string) {
	if text == "" || a.seen[text] {
		return
	}
	a.seen[text] = true
	a.out = append(a.out, domainsuggestions.Suggestion{Text: text})
}

func (a *accumulator) full() bool { return len(a.out) >= a.limit }

// addPrompts adds the recent prompts matching the query (all, when empty). The
// source is asked for promptPoolDepth (the deepened candidate pool, round 064);
// the accumulator still caps the surfaced list at maxSuggestions.
func (s *Service) addPrompts(ctx context.Context, query string, acc *accumulator) {
	if s.prompts == nil {
		return
	}
	for _, p := range s.prompts.RecentPrompts(ctx, promptPoolDepth) {
		if ctx.Err() != nil || acc.full() {
			return
		}
		if query == "" || isSubsequence(query, p) {
			acc.add(p)
		}
	}
}

// addWorkspace adds the workspace entries for a path-like query.
func (s *Service) addWorkspace(ctx context.Context, query string, acc *accumulator) {
	if !isPathLike(query) || s.workspace == nil {
		return
	}
	dir, prefix := splitQuery(query)
	for _, e := range s.workspace.Entries(ctx, dir, prefix, maxSuggestions) {
		if ctx.Err() != nil || acc.full() {
			return
		}
		acc.add(e)
	}
}

// addTools adds the registered tool names matching the query.
func (s *Service) addTools(query string, acc *accumulator) {
	if s.tools == nil {
		return
	}
	for _, t := range s.tools.ToolNames() {
		if acc.full() {
			return
		}
		if isSubsequence(query, t) {
			acc.add(t)
		}
	}
}

// Close drains background work. The engine owns none yet (round-015 research
// Decision 3 / PR #38 review directive ⑤).
func (s *Service) Close(ctx context.Context) error { return nil }

// isSubsequence reports whether q (case-insensitive) appears in s as an ordered
// subsequence.
func isSubsequence(q, s string) bool {
	q = strings.ToLower(q)
	s = strings.ToLower(s)
	i := 0
	for j := 0; j < len(s) && i < len(q); j++ {
		if s[j] == q[i] {
			i++
		}
	}
	return i == len(q)
}

// isPathLike reports whether a query looks like a path (a separator or a
// leading dot), which is when workspace entries are offered (round-015 FR-004).
func isPathLike(q string) bool {
	return strings.Contains(q, "/") || strings.HasPrefix(q, ".")
}

// splitQuery splits a path-like query into its directory and the name prefix.
func splitQuery(q string) (dir, prefix string) {
	dir, prefix = filepath.Split(filepath.ToSlash(q))
	if dir == "" {
		dir = "."
	}
	return dir, prefix
}
