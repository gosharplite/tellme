package ui

import (
	"strings"
	"sync"
	"testing"
	"time"
)

// Round-019 spinner unit tests (T013): the label/elapsed/resource formatters, the
// frame advancement over injected ticks (no time.Sleep — a signalling writer plus
// a bounded select), the turn-scoped elapsed counter, and the synchronous clear.

func TestSpinnerLabels(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"thinking with model", ThinkingLabel("gpt"), " Thinking [gpt]..."},
		{"thinking without model", ThinkingLabel(""), " Thinking..."},
		{"executing one tool", ExecutingLabel("read_files"), " Executing [read_files]..."},
		{"executing several tools (round 025: bounded)", ExecutingToolsLabel([]string{"a", "b"}), " Executing tools [a and 1 more]..."},
		{"executing one name", ExecutingToolsLabel([]string{"a"}), " Executing [a]..."},
		{"executing no names", ExecutingToolsLabel(nil), " Executing tools..."},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestFormatSpinnerLine(t *testing.T) {
	if got := FormatSpinnerLine("A", " Thinking [m]...", 3, ""); got != "A Thinking [m]... (3s)" {
		t.Errorf("FormatSpinnerLine = %q, want %q", got, "A Thinking [m]... (3s)")
	}
	if got := FormatSpinnerLine("A", " Executing [t]...", 0, " [CPU: 1.0% | MEM: 2.0%]"); got != "A Executing [t]... (0s) [CPU: 1.0% | MEM: 2.0%]" {
		t.Errorf("FormatSpinnerLine with resource = %q", got)
	}
}

func TestFormatResourceSegment(t *testing.T) {
	if got := FormatResourceSegment(12.34, 56.78); got != " [CPU: 12.3% | MEM: 56.8%]" {
		t.Errorf("FormatResourceSegment = %q, want %q", got, " [CPU: 12.3% | MEM: 56.8%]")
	}
}

// signalWriter appends to a buffer and signals once per Write, so a test can wait
// for a draw deterministically without time.Sleep.
type signalWriter struct {
	mu  sync.Mutex
	buf strings.Builder
	ch  chan struct{}
}

func newSignalWriter() *signalWriter { return &signalWriter{ch: make(chan struct{}, 64)} }

func (w *signalWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	w.buf.Write(p)
	w.mu.Unlock()
	w.ch <- struct{}{}
	return len(p), nil
}

func (w *signalWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// awaitWrite blocks until a write lands, or fails the test after a bounded wait
// (so a missing write is an assertion failure, not a hang). No time.Sleep.
func awaitWrite(t *testing.T, w *signalWriter) {
	t.Helper()
	select {
	case <-w.ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("no spinner write within the deadline; buffer=%q", w.String())
	}
}

func TestSpinnerAdvancesFramesAndClearsSynchronously(t *testing.T) {
	w := newSignalWriter()
	fixed := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	s := NewSpinner(w, "m", fixed, nil, nil)

	tick := make(chan time.Time, 8)
	stopped := false
	s.newTicker = func() (<-chan time.Time, func()) { return tick, func() { stopped = true } }
	s.now = func() time.Time { return fixed }

	// The first frame is drawn synchronously on start.
	s.OnInferenceStart()
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "⠋ Thinking [m]... (0s)") {
		t.Fatalf("first frame = %q, want the synchronous ⠋ Thinking frame", w.String())
	}

	// A tick advances the frame.
	tick <- fixed
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "⠙ Thinking [m]... (0s)") {
		t.Fatalf("after a tick = %q, want the advanced ⠙ frame", w.String())
	}

	// Stop is synchronous: the ticker is stopped and the clear frame is written
	// before Stop returns.
	s.Stop()
	if !stopped {
		t.Error("Stop did not stop the ticker")
	}
	if !strings.HasSuffix(w.String(), "\r\x1b[K") {
		t.Errorf("Stop did not write a trailing clear: %q", w.String())
	}
	// The clear's own write is the last signal; nothing else follows.
	awaitWrite(t, w)
	if len(w.ch) != 0 {
		t.Errorf("Stop returned with %d in-flight write(s) beyond the clear", len(w.ch))
	}
	// Stop is idempotent.
	s.Stop()
}

func TestSpinnerStopIdempotentAndNoopWhenUnstarted(t *testing.T) {
	w := newSignalWriter()
	s := NewSpinner(w, "m", time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC), nil, nil)
	s.Stop() // never started → no write
	if w.String() != "" {
		t.Errorf("Stop on an unstarted spinner wrote %q, want nothing", w.String())
	}
}

// TestSpinnerElapsedIsTurnScoped pins round-019 research D4: the elapsed counter
// counts from the turn's prompt-capture epoch and is NEVER reset — a phase
// relabel and a clear→resume around interleaved output both keep counting from
// that same epoch.
func TestSpinnerElapsedIsTurnScoped(t *testing.T) {
	w := newSignalWriter()
	base := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	nowV := base
	s := NewSpinner(w, "m", base, nil, nil)
	s.newTicker = func() (<-chan time.Time, func()) { return make(chan time.Time), func() {} }
	s.now = func() time.Time { return nowV }

	s.OnInferenceStart() // (0s)
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "Thinking [m]... (0s)") {
		t.Fatalf("first frame = %q, want (0s)", w.String())
	}

	// A relabel while waiting preserves the turn-scoped counter (5s).
	nowV = base.Add(5 * time.Second)
	s.OnToolsStart([]string{"read_files"})
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "Executing [read_files]... (5s)") {
		t.Fatalf("relabel = %q, want the turn-scoped (5s)", w.String())
	}

	// Interleaved output: clear, then resume. The counter continues from the turn
	// epoch (7s), NOT from 0 — this is the pin for research D4.
	nowV = base.Add(6 * time.Second)
	s.BeforeToolLog()
	awaitWrite(t, w) // the clear write
	nowV = base.Add(7 * time.Second)
	s.AfterToolLog()
	awaitWrite(t, w)
	if !strings.Contains(w.String(), "Executing [read_files]... (7s)") {
		t.Fatalf("resume = %q, want the turn-scoped (7s), not a reset to (0s)", w.String())
	}
	s.Stop()
}
