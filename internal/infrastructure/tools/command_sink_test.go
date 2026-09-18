package tools

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
)

// Round 052 (closes #115 R-2; ADR 0021): the `[Tool Output]` sink is injected at
// CONSTRUCTION, not rebound on the registry afterwards. These pins are the
// deterministic witness: a sink passed to NewCommandTool receives the block; a
// nil sink is a no-op. Reverting the injection (a nil sink) reds the first pin.

// fakeOutputSink is a test double for domaintools.OutputSink. It records the
// Begin/End brackets and captures the bytes the command tool tees into Writer.
type fakeOutputSink struct {
	mu      sync.Mutex
	begin   int
	end     int
	buf     bytes.Buffer
	enabled bool
}

func (f *fakeOutputSink) Begin() {
	f.mu.Lock()
	f.begin++
	f.mu.Unlock()
}

func (f *fakeOutputSink) End() {
	f.mu.Lock()
	f.end++
	f.mu.Unlock()
}

func (f *fakeOutputSink) Enabled() bool { return f.enabled }

func (f *fakeOutputSink) Writer() io.Writer { return f }

func (f *fakeOutputSink) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.buf.Write(p)
}

func (f *fakeOutputSink) counts() (begin, end int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.begin, f.end
}

func (f *fakeOutputSink) captured() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.buf.String()
}

// TestNewCommandToolCarriesInjectedOutputSink pins the construction seam: the
// sink handed to the constructor is the one the tool drives at Execute time.
func TestNewCommandToolCarriesInjectedOutputSink(t *testing.T) {
	sink := &fakeOutputSink{enabled: true}
	tool := NewCommandTool(sink)

	got, err := tool.Execute(context.Background(), `{"command":"echo hello-from-sink","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatalf("execute_command: %v", err)
	}
	if !strings.Contains(got, "hello-from-sink") {
		t.Fatalf("result %q does not carry the command output", got)
	}
	if begin, end := sink.counts(); begin != 1 || end != 1 {
		t.Fatalf("injected sink Begin/End = %d/%d, want 1/1", begin, end)
	}
	if !strings.Contains(sink.captured(), "hello-from-sink") {
		t.Fatalf("injected sink writer captured %q, want the child output", sink.captured())
	}
}

// TestNewCommandToolNilSinkIsNoOp pins the nil-sink path the round-031 assembler
// gate and the offline `--tool-usage` path use: no block, no panic.
func TestNewCommandToolNilSinkIsNoOp(t *testing.T) {
	tool := NewCommandTool(nil)

	got, err := tool.Execute(context.Background(), `{"command":"echo no-sink","reason":"r"}`, testBudget)
	if err != nil {
		t.Fatalf("execute_command: %v", err)
	}
	if !strings.Contains(got, "no-sink") {
		t.Fatalf("result %q does not carry the command output", got)
	}
}
