package ui

import (
	"io"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// Round 040 (issue #82) — the WS-A idle-gap coordinator stress (ADR 0009 D3/D4):
// the resume/clear race, the anti-vacuity case, the Begin-seeded zero-output case,
// the no-residue close, the gated-off + no-label no-ops, the `\r`-only case, and
// the stalled-writer / End-while-write-in-flight safety. All deterministic
// (injected clock + tickers; no time.Sleep).

// uiTestClock is a mutex-guarded mutable clock shared by the writer + the spinner.
type uiTestClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *uiTestClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *uiTestClock) add(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

// uiSpinnerStatusRe matches a live spinner status line (a braille frame + a phase
// status) — the same shape the E2E helper uses.
var uiSpinnerStatusRe = regexp.MustCompile(`[\x{2800}-\x{28FF}] (Thinking|Executing)`)

// uiVisible simulates the carriage-return redraw: within each newline-delimited
// line only the text after the last `\r` survives, and the erase-to-end-of-line is
// dropped. A frame cleared synchronously leaves no visible residue.
func uiVisible(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if i := strings.LastIndexByte(line, '\r'); i >= 0 {
			line = line[i+1:]
		}
		line = strings.ReplaceAll(line, "\x1b[K", "")
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func newIdleSpinner(w *signalWriter, clock *uiTestClock, spinTick <-chan time.Time) *Spinner {
	s := NewSpinner(w, "m", clock.now(), nil, nil)
	s.now = clock.now
	s.newTicker = func() (<-chan time.Time, func()) { return spinTick, func() {} }
	return s
}

// drainWrites consumes every currently-queued write signal (Begin emits three
// writes — the clear, the header, the separator — so a single await would leave
// stale signals queued and race a later assertion).
func drainWrites(w *signalWriter) {
	for {
		select {
		case <-w.ch:
		default:
			return
		}
	}
}

func newTestCoordinator(w io.Writer, clock *uiTestClock, sp *Spinner, idle time.Duration) (*ToolOutputCoordinator, chan time.Time) {
	c := NewToolOutputCoordinator(w, clock.now, sp, idle)
	tick := make(chan time.Time, 16)
	c.newTicker = func() (<-chan time.Time, func()) { return tick, func() {} }
	return c, tick
}

// spinnerRunning reads the spinner's live flag under its mutex (N-40-5), so the
// test assertion itself cannot race a concurrent writer.
func spinnerRunning(s *Spinner) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// TestCoordinatorResumeClearsBeforeLineAndLeavesNoResidue covers (b′) the
// zero-output resume (the Begin-seeded lastLine), (a) the clear-before-line, and
// (c) the no-residue close.
func TestCoordinatorResumeClearsBeforeLineAndLeavesNoResidue(t *testing.T) {
	sw := newSignalWriter()
	clock := &uiTestClock{t: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)}
	s := newIdleSpinner(sw, clock, make(chan time.Time))
	s.OnToolsStart([]string{"execute_command"})
	awaitWrite(t, sw) // the pre-Begin synchronous frame

	coord, tick := newTestCoordinator(sw, clock, s, 50*time.Millisecond)
	coord.Begin()
	drainWrites(sw) // drain Begin's clear + header + separator writes
	afterBegin := len(sw.String())

	// (b′) A command that prints NOTHING at all still resumes after N (lastLine is
	// seeded at Begin).
	clock.add(100 * time.Millisecond)
	tick <- clock.now()
	awaitWrite(t, sw) // the resumed first frame (drawn by the goroutine)
	if !uiSpinnerStatusRe.MatchString(sw.String()[afterBegin:]) {
		t.Fatalf("no resumed frame after the idle gap (zero-output case): %q", sw.String()[afterBegin:])
	}
	if !spinnerRunning(s) {
		t.Fatal("the resume did not start the presenter")
	}

	// (a) An output line arriving while the indicator is live is cleared BEFORE it.
	beforeLine := len(sw.String())
	if _, err := coord.Writer().Write([]byte("hello\n")); err != nil {
		t.Fatal(err)
	}
	seg := sw.String()[beforeLine:]
	if spinnerRunning(s) {
		t.Fatal("the spinner is still live after an output line")
	}
	wantLine := FormatToolOutputLine(clock.now(), "hello") + "\n"
	if !strings.HasSuffix(seg, wantLine) {
		t.Fatalf("the output line was not the last write: %q", seg)
	}
	if !strings.Contains(seg, clearControl) {
		t.Fatalf("the indicator was not cleared before the line: %q", seg)
	}

	// (c) The block closes with the separator (a resume follows the close, so the
	// separator is not the last write), and a final stop leaves no residue.
	coord.End()
	if !strings.Contains(sw.String(), ToolOutputSeparator+"\n") {
		t.Fatalf("the block did not close with the separator: %q", sw.String())
	}
	s.Stop()
	if uiSpinnerStatusRe.MatchString(uiVisible(sw.String())) {
		t.Fatalf("spinner residue survived: %q", uiVisible(sw.String()))
	}
}

// TestCoordinatorContinuousOutputDrawsNoFrame covers (b) the anti-vacuity case:
// with continuous output no idle gap elapses, so NO frame appears between lines —
// the resume cannot decay into the rejected per-line yield.
func TestCoordinatorContinuousOutputDrawsNoFrame(t *testing.T) {
	sw := newSignalWriter()
	clock := &uiTestClock{t: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)}
	s := newIdleSpinner(sw, clock, make(chan time.Time))
	s.OnToolsStart([]string{"execute_command"})
	awaitWrite(t, sw)

	coord, tick := newTestCoordinator(sw, clock, s, time.Second) // N = 1s
	coord.Begin()
	afterBegin := len(sw.String())
	w := coord.Writer()

	// 20 lines, 10 ms apart: the total advance (200 ms) never reaches N between
	// lines, so the watcher never admits.
	for i := 0; i < 20; i++ {
		clock.add(10 * time.Millisecond)
		tick <- clock.now()
		if _, err := w.Write([]byte("burst\n")); err != nil {
			t.Fatal(err)
		}
	}
	if uiSpinnerStatusRe.MatchString(sw.String()[afterBegin:]) {
		t.Fatalf("a frame appeared between continuous output lines: %q", sw.String()[afterBegin:])
	}
	if spinnerRunning(s) {
		t.Fatal("the spinner remained live during continuous output")
	}
	coord.End()
	s.Stop()
}

// TestCoordinatorGatedOffIsANoop covers FR-011: with no spinner (gated off) the
// coordinator's resume/clear is a no-op, yet the block still renders.
func TestCoordinatorGatedOffIsANoop(t *testing.T) {
	sw := newSignalWriter()
	clock := &uiTestClock{t: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)}
	coord, tick := newTestCoordinator(sw, clock, nil, 50*time.Millisecond)
	coord.Begin()
	clock.add(100 * time.Millisecond)
	tick <- clock.now()
	if _, err := coord.Writer().Write([]byte("out\n")); err != nil {
		t.Fatal(err)
	}
	coord.End()
	if uiSpinnerStatusRe.MatchString(sw.String()) {
		t.Fatalf("a gated-off coordinator drew a frame: %q", sw.String())
	}
	if !strings.Contains(sw.String(), ToolOutputSeparator) {
		t.Fatalf("the block was not rendered: %q", sw.String())
	}
}

// TestSpinnerAdmitResumeNoLabelNoop covers RF-3: a resume with no phase label is a
// defined no-op (the presenter has nothing to draw).
func TestSpinnerAdmitResumeNoLabelNoop(t *testing.T) {
	sw := newSignalWriter()
	s := NewSpinner(sw, "m", time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC), nil, nil)
	s.newTicker = func() (<-chan time.Time, func()) { return make(chan time.Time), func() {} }
	s.now = func() time.Time { return time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC) }
	s.AdmitResume() // status "" → no-op
	if spinnerRunning(s) {
		t.Fatal("AdmitResume started a label-less presenter")
	}
	if sw.String() != "" {
		t.Fatalf("AdmitResume wrote %q, want nothing", sw.String())
	}
}

// TestCoordinatorCarriageReturnOnlyStreamResumes covers (g): a `\r`-only stream
// emits NO complete output line, so under FR-001 (line-scoped) the indicator MUST
// reappear — a `\r`-only stream is visually silent.
func TestCoordinatorCarriageReturnOnlyStreamResumes(t *testing.T) {
	sw := newSignalWriter()
	clock := &uiTestClock{t: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)}
	s := newIdleSpinner(sw, clock, make(chan time.Time))
	s.OnToolsStart([]string{"execute_command"})
	awaitWrite(t, sw)

	coord, tick := newTestCoordinator(sw, clock, s, 50*time.Millisecond)
	coord.Begin()
	drainWrites(sw) // drain Begin's clear + header + separator writes
	afterBegin := len(sw.String())
	if _, err := coord.Writer().Write([]byte("half a line\r")); err != nil { // no newline → no line
		t.Fatal(err)
	}
	clock.add(100 * time.Millisecond)
	tick <- clock.now()
	awaitWrite(t, sw)
	if !uiSpinnerStatusRe.MatchString(sw.String()[afterBegin:]) {
		t.Fatalf("a `\\r`-only quiet stream did not resume the indicator: %q", sw.String()[afterBegin:])
	}
	coord.End()
	s.Stop()
}

// gatedWriter blocks a gated Write until released (the stalled-stderr case). The
// first gated Write signals `entered`, then waits on `gate`.
type gatedWriter struct {
	mu      sync.Mutex
	buf     strings.Builder
	gate    chan struct{}
	entered chan struct{}
	arm     bool
	once    sync.Once
}

func newGatedWriter() *gatedWriter {
	return &gatedWriter{gate: make(chan struct{}), entered: make(chan struct{})}
}

func (g *gatedWriter) armGate() { g.mu.Lock(); g.arm = true; g.mu.Unlock() }

func (g *gatedWriter) release() { close(g.gate) }

func (g *gatedWriter) Write(p []byte) (int, error) {
	g.mu.Lock()
	g.buf.Write(p)
	block := g.arm
	g.mu.Unlock()
	if block {
		g.once.Do(func() { close(g.entered) })
		<-g.gate
	}
	return len(p), nil
}

func (g *gatedWriter) String() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.buf.String()
}

// TestCoordinatorEndWhileLineWriteInFlight covers (f)/(h): an output-line write is
// in flight (holding the block mutex, blocked on a stalled stderr) when End runs;
// the writer is RELEASED, and both the line path and End complete with the closing
// separator written and no stranded frame (the R-8 stop-join-clear order).
func TestCoordinatorEndWhileLineWriteInFlight(t *testing.T) {
	gw := newGatedWriter()
	clock := &uiTestClock{t: time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)}
	s := NewSpinner(gw, "m", clock.now(), nil, nil)
	s.newTicker = func() (<-chan time.Time, func()) { return make(chan time.Time), func() {} }
	s.now = clock.now
	s.OnToolsStart([]string{"execute_command"})

	coord, _ := newTestCoordinator(gw, clock, s, 50*time.Millisecond)
	coord.Begin()
	gw.armGate() // the NEXT write (the clear + the output line) blocks until released

	lineDone := make(chan struct{})
	go func() {
		_, _ = coord.Writer().Write([]byte("stalled line\n"))
		close(lineDone)
	}()
	select {
	case <-gw.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("the output-line write never entered the stalled writer")
	}

	endDone := make(chan struct{})
	go func() {
		coord.End()
		close(endDone)
	}()

	gw.release()
	select {
	case <-lineDone:
	case <-time.After(2 * time.Second):
		t.Fatal("the output-line write did not complete after the writer was released")
	}
	select {
	case <-endDone:
	case <-time.After(2 * time.Second):
		t.Fatal("End did not complete after the writer was released")
	}
	if !strings.Contains(gw.String(), ToolOutputSeparator) {
		t.Fatalf("the block did not close with the separator: %q", gw.String())
	}
	s.Stop()
	if uiSpinnerStatusRe.MatchString(uiVisible(gw.String())) {
		t.Fatalf("a stranded frame survived the stalled-writer close: %q", uiVisible(gw.String()))
	}
}
