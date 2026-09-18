package ui

import (
	"testing"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 051 (R5.5 of #92; ADR 0020; review R-51-3): a TYPED-NIL
// *ToolOutputCoordinator placed in a non-nil domaintools.OutputSink interface
// must report disabled and must not panic — the nil-receiver guards keep the
// interface form as safe as the former struct-of-funcs zero value.
func TestOutputSinkTypedNilIsDisabled(t *testing.T) {
	var c *ToolOutputCoordinator
	// Build the interface through a helper so the typed-nil hazard is exercised
	// without tripping staticcheck SA4023 (a direct comparison is provably false).
	sink := asOutputSink(c)
	if !isNonNilInterface(sink) {
		t.Fatalf("a typed-nil must not make the interface nil (the hazard this pin covers)")
	}
	if sink.Enabled() {
		t.Errorf("a typed-nil coordinator must report Enabled() == false")
	}
	// The lifecycle methods must be nil-safe (no panic).
	sink.Begin()
	sink.End()
	if sink.Writer() != nil {
		t.Errorf("a typed-nil coordinator must yield a nil Writer")
	}
}

// asOutputSink boxes a typed-nil *ToolOutputCoordinator into the interface.
func asOutputSink(c *ToolOutputCoordinator) domaintools.OutputSink { return c }

// isNonNilInterface reports whether the interface is non-nil (a typed-nil pointer
// still satisfies this — the hazard R-51-3 covers).
func isNonNilInterface(s domaintools.OutputSink) bool { return s != nil }
