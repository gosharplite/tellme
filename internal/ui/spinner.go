package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gosharplite/tellme/internal/domain/metrics"
)

// Round-019 turn progress spinner (specs/truth/features/cli/chat/presenting-the-progress-spinner.feature).
// A live, animated indicator on the diagnostic stream (stderr) while a non-TUI
// prompt turn waits: a braille frame advancing on a ~200 ms ticker, a phase
// status label naming the model or the tool(s), and a whole-seconds elapsed
// counter; the tool-execution state also reports the machine's CPU/memory. It is
// hand-written (no dependency) and its frame updates share one I/O mutex, so
// Stop()/Clear() are synchronous and no in-flight frame survives a clear
// (round-019 research Decisions 1, 2, 4, 10).
//
// The elapsed counter is TURN-scoped (research D4): it counts from the turn's
// prompt-capture epoch and is NEVER reset — an in-place relabel, or a clear +
// resume around interleaved output, keeps counting from that same epoch.
//
// Round 025 makes the line WIDTH-SAFE (research Decisions 1–3): the several-tool
// status label is BOUNDED (` Executing tools [<first> and <N-1> more]...`), and
// the presenter tracks the rendered-row count of its last frame and erases EVERY
// occupied row on redraw and on clear, so an over-wide (soft-wrapped) frame
// leaves no residue. The terminal width comes from an injected `columns` seam.

// SpinnerFrames is the reference's braille frame set.
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinnerInterval is the reference's ~200 ms redraw cadence.
const SpinnerInterval = 200 * time.Millisecond

// clearControl is the carriage-return + ANSI erase-to-end-of-line redraw prefix
// (round-019 research Decision 2). It is a cursor control, not colour. For a
// single-row frame it is also the whole clear (round 025 erases more rows only
// when the last frame wrapped).
const clearControl = "\r\x1b[K"

// cursorUp is the ANSI "move the cursor up one row" control, used by the
// round-025 row-aware clear to erase every row a soft-wrapped frame occupied.
const cursorUp = "\x1b[1A"

// ThinkingLabel renders the awaiting-the-model status label (leading space); the
// `[<model>]` bracket is omitted when model is empty.
func ThinkingLabel(model string) string {
	if model == "" {
		return " Thinking..."
	}
	return " Thinking [" + model + "]..."
}

// ExecutingLabel renders the single-tool status label (leading space).
func ExecutingLabel(tool string) string { return " Executing [" + tool + "]..." }

// ExecutingToolsLabel renders the tool-phase status label: the single-tool form
// for one name, the bare form when no names are available, and — round 025 — a
// BOUNDED several-tool form for two or more names that names the FIRST tool and
// counts the rest (` Executing tools [<first> and <N-1> more]...`), so the label
// cannot overrun the terminal (issue #55). The label no longer enumerates every
// name.
func ExecutingToolsLabel(names []string) string {
	switch len(names) {
	case 0:
		return " Executing tools..."
	case 1:
		return ExecutingLabel(names[0])
	default:
		return fmt.Sprintf(" Executing tools [%s and %d more]...", names[0], len(names)-1)
	}
}

// FormatSpinnerLine renders one spinner line `{frame}{status} ({n}s){resource}`.
func FormatSpinnerLine(frame, status string, elapsed int, resource string) string {
	return fmt.Sprintf("%s%s (%ds)%s", frame, status, elapsed, resource)
}

// FormatResourceSegment renders the tool-execution resource segment
// ` [CPU: <c>% | MEM: <m>%]` (one decimal each).
func FormatResourceSegment(cpu, mem float64) string {
	return fmt.Sprintf(" [CPU: %.1f%% | MEM: %.1f%%]", cpu, mem)
}

// eraseRows renders the ANSI sequence that erases n terminal rows, bottom-up,
// leaving the cursor at column 0 of the top row (round 025). n < 1 is treated as
// 1, so the single-row case is exactly the round-019 clearControl.
//
// Known bound (accepted): the row count is captured at draw time, so a mid-frame
// terminal resize leaves it stale and a clear may over-erase one row of prior
// output. Residue (the round-019 defect) is eliminated; over-erase-under-reflow
// is a recorded limitation — the reference has no resize handling at all.
func eraseRows(n int) string {
	if n < 1 {
		n = 1
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString(clearControl)
		if i < n-1 {
			b.WriteString(cursorUp)
		}
	}
	return b.String()
}

// rowsForLine returns how many terminal rows line occupies at the given column
// width (round 025). A width <= 0 (unknown) or an empty line is a single row.
//
// It measures RUNES, not display columns: correct for the ASCII labels + the
// single-width braille frames used today, but a wide (CJK/emoji) or zero-width
// (combining) operator-configured model name would mis-measure — a documented
// ASCII/single-width boundary (R-1; swap for a display-width lib if the label
// ever admits wide runes).
func rowsForLine(line string, columns int) int {
	if columns <= 0 {
		return 1
	}
	w := len([]rune(line))
	if w <= 0 {
		return 1
	}
	rows := (w + columns - 1) / columns
	if rows < 1 {
		rows = 1
	}
	return rows
}

// Spinner is the live progress presenter. It implements the agent-loop observer
// port structurally (internal/domain/agent.LoopObserver) and writes only to the
// diagnostic stream.
type Spinner struct {
	mu      sync.Mutex
	w       io.Writer
	model   string
	metrics metrics.SystemMetricsProvider

	// columns reports the terminal width in columns for the round-025 row-aware
	// clear; nil or <= 0 means the width is unknown (single-row best effort).
	columns func() int

	// epoch is the turn's prompt-capture time: the elapsed counter measures
	// now − epoch for the WHOLE turn and is never reset (research D4).
	epoch time.Time

	// now and newTicker are the injected time seams (round-019 research D8); a
	// nil newTicker falls back to the real ~200 ms ticker.
	now       func() time.Time
	newTicker func() (<-chan time.Time, func())

	// live state (guarded by mu).
	running   bool
	toolPhase bool
	frameIdx  int
	status    string
	// lastRows is the number of terminal rows the last drawn frame occupied
	// (round 025); the next redraw / the clear erases all of them.
	lastRows   int
	stopCh     chan struct{}
	doneCh     chan struct{}
	stopTicker func()
}

// NewSpinner builds a spinner writing to w, labelling the model, counting the
// elapsed from epoch (the turn's prompt-capture time), sampling machine resources
// from m (nil disables the resource segment), and reading the terminal width from
// columns (nil or <= 0 = unknown, round 025) for the row-aware clear.
//
// A functional-options constructor is a recorded forward nit (R-2) if a sixth
// seam ever lands; five positional params are within the repo's current norm.
func NewSpinner(w io.Writer, model string, epoch time.Time, m metrics.SystemMetricsProvider, columns func() int) *Spinner {
	return &Spinner{
		w:       w,
		model:   model,
		metrics: m,
		columns: columns,
		epoch:   epoch,
		now:     time.Now,
		newTicker: func() (<-chan time.Time, func()) {
			t := time.NewTicker(SpinnerInterval)
			return t.C, t.Stop
		},
	}
}

// OnInferenceStart starts (or relabels) the model-phase indicator. The first
// frame is drawn synchronously; an in-place relabel preserves the elapsed counter
// (which is turn-scoped — research D4).
func (s *Spinner) OnInferenceStart() { s.activate(ThinkingLabel(s.model), false) }

// OnInferenceEnd leaves the indicator running until the next phase or the final
// clear (the CLI calls Stop before any interleaved write).
func (s *Spinner) OnInferenceEnd() {}

// OnToolsStart starts (or relabels) the tool-phase indicator, which also reports
// the machine's resource usage.
func (s *Spinner) OnToolsStart(names []string) {
	s.activate(ExecutingToolsLabel(names), true)
}

// OnToolsEnd leaves the indicator running until the next phase.
func (s *Spinner) OnToolsEnd() {}

// BeforeToolLog yields the line to a tool-loop log write (a synchronous clear).
func (s *Spinner) BeforeToolLog() { s.deactivate() }

// AfterToolLog restores the indicator after a tool-loop log write. The elapsed
// counter is turn-scoped, so it continues from the turn epoch (it does NOT reset
// — research D4).
func (s *Spinner) AfterToolLog() { s.resume() }

// Stop synchronously clears the indicator and stops its redraw. It is idempotent.
func (s *Spinner) Stop() { s.deactivate() }

// activate starts the redraw (drawing the first frame synchronously) or relabels
// a running indicator. The elapsed epoch is turn-scoped and never reset.
func (s *Spinner) activate(status string, toolPhase bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolPhase = toolPhase
	s.status = status
	if s.running {
		s.renderLocked()
		return
	}
	s.running = true
	s.frameIdx = 0
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	tick, stopTicker := s.newTicker()
	s.stopTicker = stopTicker
	s.renderLocked()
	go s.loop(tick, s.stopCh, s.doneCh)
}

// resume restores the indicator after interleaved output. The turn-scoped epoch
// is kept, so the counter continues rather than restarting.
func (s *Spinner) resume() {
	s.mu.Lock()
	status, toolPhase := s.status, s.toolPhase
	s.mu.Unlock()
	if status == "" {
		return
	}
	s.activate(status, toolPhase)
}

// deactivate stops the redraw goroutine, waits for it to exit, and writes the
// final clear frame — all synchronously, so no in-flight frame survives the
// clear (round-019 research D10).
//
// It claims the stop by clearing `running` UNDER the mutex before closing
// `stopCh`, so two concurrent Stop()s cannot both pass the guard and double-close
// the channel (a panic). It is idempotent.
func (s *Spinner) deactivate() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	stop, done, stopTicker := s.stopCh, s.doneCh, s.stopTicker
	s.stopTicker = nil
	s.mu.Unlock()
	close(stop)
	<-done
	if stopTicker != nil {
		stopTicker()
	}
	s.mu.Lock()
	s.clearLocked()
	s.mu.Unlock()
}

// loop advances and redraws the frame on each tick until stopped.
func (s *Spinner) loop(tick <-chan time.Time, stop, done chan struct{}) {
	defer close(done)
	for {
		select {
		case <-stop:
			return
		case <-tick:
			s.mu.Lock()
			if !s.running {
				s.mu.Unlock()
				return
			}
			s.frameIdx++
			s.renderLocked()
			s.mu.Unlock()
		}
	}
}

// renderLocked writes one redrawn frame (the mutex must be held). The elapsed is
// measured from the turn epoch (turn-scoped — research D4). Round 025 erases every
// row the PREVIOUS frame occupied before drawing, and records the new frame's row
// count so the next redraw / the clear can erase all of them.
func (s *Spinner) renderLocked() {
	if s.w == nil {
		return
	}
	frame := SpinnerFrames[s.frameIdx%len(SpinnerFrames)]
	elapsed := int(s.now().Sub(s.epoch) / time.Second)
	if elapsed < 0 {
		elapsed = 0
	}
	resource := ""
	if s.toolPhase {
		var cpu, mem float64
		if s.metrics != nil {
			cpu, mem = s.metrics.Sample()
		}
		resource = FormatResourceSegment(cpu, mem)
	}
	line := FormatSpinnerLine(frame, s.status, elapsed, resource)
	// Erase every row the previous frame occupied (round 025; a single-row clear
	// when lastRows <= 1), then draw the new frame and record its row count.
	_, _ = io.WriteString(s.w, eraseRows(s.lastRows)+line)
	s.lastRows = rowsForLine(line, s.columnsWidth())
}

// clearLocked writes the clear frame (the mutex must be held). Round 025 erases
// every row the last frame occupied (a single-row clear when lastRows <= 1).
func (s *Spinner) clearLocked() {
	if s.w == nil {
		return
	}
	_, _ = io.WriteString(s.w, eraseRows(s.lastRows))
	s.lastRows = 0
}

// columnsWidth reports the resolved terminal width (0 = unknown).
func (s *Spinner) columnsWidth() int {
	if s.columns == nil {
		return 0
	}
	return s.columns()
}
