package cli

import (
	"context"
	"io"
	"time"

	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/metrics"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	domaintui "github.com/gosharplite/tellme/internal/domain/tui"
)

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

// Run satisfies domaintui.Prompter.
func (f fakePrompter) Run(_ context.Context, _ io.Reader, _ io.Writer, _ domaintui.Source, _ time.Duration) (string, bool, error) {
	if f.called != nil {
		*f.called = true
	}
	return f.result, f.ok, f.err
}

// DefaultDebounceDuration satisfies domaintui.Prompter.
func (fakePrompter) DefaultDebounceDuration() time.Duration { return 100 * time.Millisecond }

// depsWithGateway returns a test deps with the gateway seam bound to gw/err.
func depsWithGateway(gw llm.Gateway, err error) deps.Dependencies {
	return defaultTestDeps(func(d *deps.Dependencies) {
		d.NewGateway = func(config.Provider, string, string) (llm.Gateway, error) { return gw, err }
	})
}

// depsWithTools returns a test deps with the gateway bound to gw and the agent
// tool registry built from the given in-package tools (so a tool-round test does
// not import internal/infrastructure/* — NFR-003).
func depsWithTools(gw llm.Gateway, tools ...domaintools.Tool) deps.Dependencies {
	return defaultTestDeps(func(d *deps.Dependencies) {
		d.NewGateway = func(config.Provider, string, string) (llm.Gateway, error) { return gw, nil }
		d.NewToolRegistry = func() domaintools.Registry { return domaintools.NewRegistry(tools...) }
	})
}

// noopPromptTracker is a no-op history.PromptTracker double.
type noopPromptTracker struct{}

func (noopPromptTracker) Append(context.Context, string) error { return nil }
func (noopPromptTracker) Recent(context.Context, int) ([]history.PromptLogEntry, error) {
	return nil, nil
}
func (noopPromptTracker) Close(context.Context) error { return nil }
