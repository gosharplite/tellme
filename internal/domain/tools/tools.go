// Package tools defines the agent-tool domain port: the model-facing tool
// abstraction and the registry the agent loop dispatches against. It is
// network-free and knows nothing about execution — the concrete tools live in
// internal/infrastructure/tools (round-008 research Decision 1).
package tools

import (
	"context"
	"encoding/json"
)

// Tool is one capability the model may invoke during a prompt run. Execute runs
// the tool with the model's raw arguments string and returns its result text. A
// tool failure is returned as an error and is treated by the loop as a
// non-terminal tool result (the model may recover); it is NOT a process
// failure. Every implementation MUST be read-only/local for round 008.
//
// The context lets the loop bound each execution with a per-tool timeout
// (round-008 RF-2).
type Tool interface {
	Name() string
	Description() string
	Parameters() json.RawMessage
	Execute(ctx context.Context, arguments string) (string, error)
}

// Registry resolves a tool by its wire name and lists the registered tools.
type Registry interface {
	Lookup(name string) (Tool, bool)
	Tools() []Tool
}

// registry is the default Registry over a fixed, ordered tool set.
type registry struct {
	byName map[string]Tool
	order  []Tool
}

// NewRegistry builds a Registry from a set of tools. The tool order is the
// order they were supplied (the order offered to the model).
func NewRegistry(ts ...Tool) Registry {
	r := &registry{byName: make(map[string]Tool, len(ts)), order: make([]Tool, 0, len(ts))}
	for _, t := range ts {
		r.byName[t.Name()] = t
		r.order = append(r.order, t)
	}
	return r
}

// Lookup returns the tool with the given wire name.
func (r *registry) Lookup(name string) (Tool, bool) {
	t, ok := r.byName[name]
	return t, ok
}

// Tools returns the registered tools in offer order.
func (r *registry) Tools() []Tool { return r.order }
