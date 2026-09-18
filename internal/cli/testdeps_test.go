package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/config"
	agentport "github.com/gosharplite/tellme/internal/domain/agent"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/metrics"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	domaintui "github.com/gosharplite/tellme/internal/domain/tui"
)

// fakeLoop is an in-package agentport.Loop double (round 050; R5.4 of #92;
// ADR 0019 — Q4 → A). internal/cli's turn orchestration is tested against it;
// the REAL loop's behaviour is covered by internal/agent unit tests + the godog
// E2E. RULE-E evaluates the merged (production + test) graph, so a _test.go
// import of internal/agent would keep the `cli -> agent` edge alive — hence the
// local double instead.
//
// The double records what the CLI handed it (the LoopSpec + the Run arguments)
// and, when the spec carries an Observer, fires the call hooks exactly as the
// real loop would, so the CLI's per-call renderer (frame + deferred tail) is
// exercised.
type fakeLoop struct {
	result    agentport.Result
	runErr    error
	writes    []string // lines emitted to spec.Stderr during Run (loop diagnostic output)
	gotSpec   agentport.LoopSpec
	gotPrompt string
	gotPrior  []history.Entry
}

func (f *fakeLoop) Run(_ context.Context, prompt string, prior []history.Entry) (agentport.Result, error) {
	f.gotPrompt, f.gotPrior = prompt, prior
	if f.gotSpec.Observer != nil {
		f.gotSpec.Observer.OnCallBegin(0, []llm.Message{{Role: "user", Content: prompt}})
	}
	for _, w := range f.writes {
		if f.gotSpec.Stderr != nil {
			_, _ = fmt.Fprintln(f.gotSpec.Stderr, w)
		}
	}
	if f.runErr != nil {
		return agentport.Result{}, f.runErr
	}
	if f.gotSpec.Observer != nil {
		f.gotSpec.Observer.OnCallEnd(0, f.result.Usage, nil, true)
	}
	return f.result, nil
}

// defaultFakeLoop is the fixture's default loop double: a completed one-call run
// with a reported usage, so renderer/persistence paths are exercised without a
// per-test script.
func defaultFakeLoop() *fakeLoop {
	return &fakeLoop{result: agentport.Result{Answer: "ok", Usage: llm.Usage{Reported: true}}}
}

// depsWithLoop returns a fixture whose LoopFactory yields lp (the CLI builds the
// loop from the spec through the injected factory). Extra mods apply after.
func depsWithLoop(lp *fakeLoop, mods ...func(*deps.Dependencies)) deps.Dependencies {
	all := append([]func(*deps.Dependencies){
		func(d *deps.Dependencies) {
			d.LoopFactory = func(spec agentport.LoopSpec) agentport.Loop { lp.gotSpec = spec; return lp }
		},
	}, mods...)
	return defaultTestDeps(all...)
}

// defaultTestDeps returns a fully-populated, in-memory, no-op
// deps.Dependencies so a unit test never hand-stubs a partial bag (a nil func
// field would panic). It imports NO internal/infrastructure/* package (NFR-003:
// the layer gate merges .TestImports, so an infra test import would keep a
// baseline edge alive). RF-1 (PR #102 review 5243331406).
func defaultTestDeps(mods ...func(*deps.Dependencies)) deps.Dependencies {
	d := deps.Dependencies{
		NewGateway: func(config.Provider, string, string) (llm.Gateway, error) {
			return &fakeGateway{text: "ok"}, nil
		},
		NewHistoryStore: func(string) history.Store { return &fakeStore{} },
		NewUsageStore:   func(string) history.UsageStore { return &capturingUsageStore{} },
		NewToolUsageStore: func(func() (string, error)) history.ToolUsageStore {
			return noopUsageStore{}
		},
		NewPromptTracker: func(string, func() (string, error)) history.PromptTracker {
			return noopPromptTracker{}
		},
		NewToolRegistry:    func() domaintools.Registry { return domaintools.NewRegistry() },
		NewTUIRegistry:     func() domaintools.Registry { return domaintools.NewRegistry() },
		BindToolOutput:     func(domaintools.Registry, domaintools.OutputSink) {},
		BindSkillsCatalog:  func(domaintools.Registry, string) {},
		NewMetricsProvider: func() metrics.SystemMetricsProvider { return nil },
		MCPDiscoverer: func(context.Context, map[string]config.MCPServerConfig) ([]domaintools.Tool, []string, func()) {
			return nil, nil, func() {}
		},
		LoopFactory: func(spec agentport.LoopSpec) agentport.Loop {
			l := defaultFakeLoop()
			l.gotSpec = spec
			return l
		},
		UserHomeDir: func() (string, error) { return "/tmp/tellme-test-home", nil },
	}
	for _, m := range mods {
		m(&d)
	}
	if err := d.Validate(); err != nil {
		panic("defaultTestDeps: incomplete fixture: " + err.Error())
	}
	return d
}

// testOptions returns an Options with the default test deps (Prompter left nil —
// a non-TUI test never reaches the interactive prompt; the TUI dispatch tests
// inject fakePrompter).
func testOptions() Options { return Options{Deps: defaultTestDeps()} }

// fakePrompter is an in-package double for the domain tui.Prompter port
// (round 048 / ADR 0017). It returns a canned result and records invocation when
// called is non-nil.
type fakePrompter struct {
	result string
	ok     bool
	err    error
	called *bool
}

// fakePrompterDebounce is the fake's debounce sentinel — deliberately distinct
// from prompt.DefaultDebounceDuration so a test can detect a hard-coded value
// (round-048 review TD-1).
const fakePrompterDebounce = 7 * time.Millisecond

// Run satisfies domaintui.Prompter.
func (f fakePrompter) Run(_ context.Context, _ io.Reader, _ io.Writer, _ domaintui.Source, _ time.Duration) (string, bool, error) {
	if f.called != nil {
		*f.called = true
	}
	return f.result, f.ok, f.err
}

// DefaultDebounceDuration satisfies domaintui.Prompter. It returns a distinctive
// sentinel (NOT the production ~100 ms) so a test can tell the injected
// fallback apart from a hard-coded value (round-048 review TD-1: a mirror-valued
// fake hides the direction).
func (fakePrompter) DefaultDebounceDuration() time.Duration { return fakePrompterDebounce }

// depsWithGateway returns a test deps with the gateway seam bound to gw/err.
func depsWithGateway(gw llm.Gateway, err error) deps.Dependencies {
	return defaultTestDeps(func(d *deps.Dependencies) {
		d.NewGateway = func(config.Provider, string, string) (llm.Gateway, error) { return gw, err }
	})
}

// noopPromptTracker is a no-op history.PromptTracker double.
type noopPromptTracker struct{}

func (noopPromptTracker) Append(context.Context, string) error { return nil }
func (noopPromptTracker) Recent(context.Context, int) ([]history.PromptLogEntry, error) {
	return nil, nil
}
func (noopPromptTracker) Close(context.Context) error { return nil }
