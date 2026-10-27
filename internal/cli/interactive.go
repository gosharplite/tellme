package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	appsuggestions "github.com/gosharplite/tellme/internal/app/suggestions"
	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/domain/history"
	domaintui "github.com/gosharplite/tellme/internal/domain/tui"
)

// runInteractiveTUI runs the interactive TUI prompt (round 015) for one
// invocation: it announces on the diagnostic stream, builds the multi-source
// suggestion engine over the injected Dependencies, and drives the prompt bound
// to stderr/stdin through the injected domain port. It returns the composed
// prompt (ok) on submit and never writes to stdout. Round 048 (ADR 0017): the
// terminal runtime is reached only through p — internal/cli no longer imports
// internal/ui/tui/prompt.
func runInteractiveTUI(ctx context.Context, res resolution, env runtimeEnv, in cliDeps) (string, bool, error) {
	_, _ = fmt.Fprintln(env.stderr, TUIHint)

	p := in.prompter
	tracker := in.dp.NewPromptTracker(res.Home, in.dp.UserHomeDir)
	defer func() { _ = tracker.Close(context.Background()) }()
	// Round 028: the first-use seed runs once, here, before the first suggestion
	// read. It is the segregated Seeder capability (NOT part of PromptTracker —
	// interface segregation, matching the round-026 port split); construction stays
	// pure (the tracker is built twice per `-i` run, so the seed is not a
	// constructor side effect — TD-2); the seed is best-effort.
	if seeder, ok := tracker.(history.Seeder); ok {
		_ = seeder.Seed(ctx)
	}
	reg := in.dp.NewTUIRegistry()
	engine := appsuggestions.New(
		appsuggestions.TrackerPrompts{Tracker: tracker},
		appsuggestions.OSSWorkspace{},
		appsuggestions.RegistryTools{Registry: reg},
	)
	src := tuiSource{svc: engine}
	// No startup disk I/O: the dashboard header was retired (round-016 FR-004 /
	// architect D3), so the history store is no longer read to build the prompt.
	return p.Run(ctx, env.stdin, env.stderr, src, tuiDebounceDuration(p))
}

// tuiDebounceDuration resolves the suggestion-refresh debounce. The hermetic E2E
// sets TELL_ME_TUI_DEBOUNCE=0 so the scripted keys observe suggestions
// synchronously (round-016); otherwise the port's default (~100 ms) applies.
func tuiDebounceDuration(p domaintui.Prompter) time.Duration {
	if v := strings.TrimSpace(os.Getenv("TELL_ME_TUI_DEBOUNCE")); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms >= 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return p.DefaultDebounceDuration()
}

// tuiSource adapts the suggestion engine to the prompt's Source seam, carrying
// ctx so a superseded fetch can be cancelled (round-016 architect D1).
type tuiSource struct {
	svc *appsuggestions.Service
}

// Suggest returns the engine's suggestion texts for the query.
func (s tuiSource) Suggest(ctx context.Context, query string) []string {
	sugg := s.svc.Suggest(ctx, query)
	out := make([]string, 0, len(sugg))
	for _, x := range sugg {
		out = append(out, x.Text)
	}
	return out
}

// tuiRequested reports whether the opt-in interactive prompt is enabled: the
// `-i`/`--interactive` flag, or the config `USE_TUI_PROMPT` key (round-015
// FR-001). The terminal-stdin requirement is enforced separately by the caller.
func tuiRequested(homeDir string, f *flags) bool {
	if f.interactive {
		return true
	}
	if cfg, err := config.Load(defaultConfigPath(homeDir)); err == nil {
		return cfg.UseTUIPrompt
	}
	return false
}

// runTUIPrompt resolves the setup and runs the interactive TUI prompt (round
// 015). A submitted prompt is recorded in the shared log — only for the
// interactive prompt — and then runs exactly one reasoning turn; an
// aborted/empty submission sends no request and exits success.
func runTUIPrompt(homeDir string, f *flags, env runtimeEnv, opts Options) int {
	dp := opts.Deps
	// Best-effort resolution: the TUI engages even when the setup is unresolved
	// (an aborted/empty submission owes no request — the round-012 ordering
	// rationale). The resolution feeds the dashboard and the submit path.
	res, _ := resolve(sessionRef{home: homeDir, config: f.configPath})
	p := opts.Prompter
	if p == nil {
		// Unreachable in production: the composition root always injects the port
		// (asserted by Options.Validate; see cmd/tellme's smoke test).
		// [round-048 divergence — ADR 0017] Reuse the environment class phrase so the
		// closed phrase vocabulary is not widened; the trailing detail is
		// contract-free. The CAUSE here is a composition-root wiring fault, not an
		// unusable home — a deliberate recorded mismatch for a future vocabulary
		// round.
		_, _ = fmt.Fprintf(env.stderr, "tellme: the runtime home is not usable (interactive prompt: no runner injected)\n")
		return EnvironmentError
	}
	text, ok, err := runInteractiveTUI(context.Background(), res, env, cliDepsFrom(opts))
	if err != nil {
		return emitProviderError(env.stderr, err)
	}
	if !ok || text == "" {
		return Success
	}
	// Record in the shared log (round-015 FR-009) — only the interactive prompt
	// writes it.
	tracker := dp.NewPromptTracker(res.Home, dp.UserHomeDir)
	_ = tracker.Append(context.Background(), text)
	_ = tracker.Close(context.Background())
	return renderTurn(sessionRef{home: homeDir, config: f.configPath}, text, turnOptions{raw: f.raw, chrome: true, echo: true}, env, dp)
}
