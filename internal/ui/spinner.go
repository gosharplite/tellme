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

// SpinnerFrames is the reference's braille frame set.
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// SpinnerInterval is the reference's ~200 ms redraw cadence.
const SpinnerInterval = 200 * time.Millisecond

// clearControl is the carriage-return + ANSI erase-to-end-of-line redraw prefix
// (round-019 research Decision 2). It is a cursor control, not colour.
const clearControl = "\r\x1b[K"

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
// for one name, the named several-tool form for several, and the bare form when
// no names are available.
func ExecutingToolsLabel(names []string) string {
	switch len(names) {
	case 0:
		return " Executing tools..."
	case 1:
		return ExecutingLabel(names[0])
	default:
		return " Executing tools [" + strings.Join(names, ", ") + "]..."
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

// Spinner is the live progress presenter. It implements the agent-loop observer
// port structurally (internal/domain/agent.LoopObserver) and writes only to the
// diagnostic stream.
type Spinner struct {
	mu      sync.Mutex
	w       io.Writer
	model   string
	metrics metrics.SystemMetricsProvider

	// epoch is the turn's prompt-capture time: the elapsed counter measures
	// now − epoch for the WHOLE turn and is never reset (research D4).
	epoch time.Time

	// now and newTicker are the injected time seams (round-019 research D8); a
	// nil newTicker falls back to the real ~200 ms ticker.
	now       func() time.Time
	newTicker func() (<-chan time.Time, func())

	// live state (guarded by mu).
	running    bool
	toolPhase  bool
	frameIdx   int
	status     string
	stopCh     chan struct{}
	doneCh     chan struct{}
	stopTicker func()
}

// NewSpinner builds a spinner writing to w, labelling the model, counting the
// elapsed from epoch (the turn's prompt-capture time), and sampling machine
// resources from m (nil disables the resource segment).
func NewSpinner(w io.Writer, model string, epoch time.Time, m metrics.SystemMetricsProvider) *Spinner {
	return &Spinner{
		w:       w,
		model:   model,
		metrics: m,
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
// measured from the turn epoch (turn-scoped — research D4).
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
	_, _ = io.WriteString(s.w, clearControl+FormatSpinnerLine(frame, s.status, elapsed, resource))
}

// clearLocked writes the clear frame (the mutex must be held).
func (s *Spinner) clearLocked() {
	if s.w == nil {
		return
	}
	_, _ = io.WriteString(s.w, clearControl)
}
