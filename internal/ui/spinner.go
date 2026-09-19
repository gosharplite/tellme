package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/metrics"
)

// Round-019 turn progress spinner (specs/truth/features/cli/chat/presenting-the-progress-spinner.feature).
// A live, animated indicator on the diagnostic stream (stderr) while a non-TUI
// prompt turn waits: a braille frame advancing on a ~200 ms ticker, a phase
// status label naming the model or the tool(s), and a DUAL whole-seconds elapsed
// counter; the tool-execution state also reports the machine's CPU/memory. It is
// hand-written (no dependency) and its frame updates share one I/O mutex, so
// Stop()/Clear() are synchronous and no in-flight frame survives a clear
// (round-019 research Decisions 1, 2, 4, 10).
//
// The elapsed display is a DUAL timer (round 040, issue #83; amends round-019 D4;
// ADR 0009 D1): the TOTAL since the turn's prompt-capture epoch — turn-scoped and
// NEVER reset — plus the CURRENT MODEL CALL's elapsed, reset at each AI-endpoint
// call (stamped in OnInferenceStart). Terminology: a turn is the whole prompt
// exchange (1..k model calls); the second figure measures the current model call.
//
// Round 059 (ADR 0029) THROTTLES the tool-phase resource segment: the metrics
// provider is sampled at most once per ResourceSampleInterval (1 s), the last
// pair reused between samples, while the frame keeps its 200 ms cadence — so the
// numbers are readable instead of refreshing 5×/s.
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

// ResourceSampleInterval is the round-059 throttle for the tool-phase CPU/MEM
// figures (reference parity — `metrics_shouldSample`): the metrics provider is
// sampled at most once per second while the braille frame keeps advancing on
// SpinnerInterval, so the numbers are readable and the animation stays smooth.
const ResourceSampleInterval = time.Second

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

// wholeSecondsSince returns the whole seconds from t to now, floored at 0
// (round 040 T012: the single elapsed-formatting site for both figures).
func wholeSecondsSince(t, now time.Time) int {
	s := int(now.Sub(t) / time.Second)
	if s < 0 {
		return 0
	}
	return s
}

// FormatSpinnerLine renders one spinner line
// `{frame}{status} ({total}s {call}s){resource}`. Round 040 (ADR 0009 D1): the
// elapsed segment carries TWO unlabelled whole-second figures — the total since
// prompt capture, then the current model call's elapsed.
func FormatSpinnerLine(frame, status string, total, call int, resource string) string {
	return fmt.Sprintf("%s%s (%ds %ds)%s", frame, status, total, call, resource)
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
// diagnostic stream. Round 045 (R3 of #92): its two yield hooks
// (YieldIndicator/RestoreIndicator) are thin adapters that delegate to the yield
// policy's single owner, ui.YieldController (see yield.go) — the clear/resume
// MECHANISM below (deactivate/resume/AdmitResume) is unchanged.
type Spinner struct {
	mu      sync.Mutex
	w       io.Writer
	model   string
	metrics metrics.SystemMetricsProvider

	// columns reports the terminal width in columns for the round-025 row-aware
	// clear; nil or <= 0 means the width is unknown (single-row best effort).
	columns func() int

	// epoch is the turn's prompt-capture time: the TOTAL elapsed figure measures
	// now − epoch for the WHOLE turn and is never reset (round-019 D4).
	epoch time.Time

	// callEpoch is the CURRENT model call's start (round 040, ADR 0009 D2): the
	// SECOND elapsed figure measures now − callEpoch and is reset at each
	// AI-endpoint call (OnInferenceStart). Initialised to epoch so the figure is
	// valid before the first call. It is an INTERNAL epoch — NOT a sixth
	// constructor seam (round-019 R-2).
	callEpoch time.Time

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

	// resource-sample throttle (round 059): the tool-phase provider is sampled at
	// most once per ResourceSampleInterval, and the last pair is reused between
	// samples (the frames in between redraw the cached figures).
	haveSample bool
	lastSample time.Time
	cachedCPU  float64
	cachedMem  float64
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
		// Round 040: the model-call epoch starts equal to the turn epoch, so the
		// dual figures coincide until the first OnInferenceStart.
		callEpoch: epoch,
		now:       time.Now,
		newTicker: func() (<-chan time.Time, func()) {
			t := time.NewTicker(SpinnerInterval)
			return t.C, t.Stop
		},
	}
}

// OnInferenceStart starts (or relabels) the model-phase indicator and STAMPS the
// current model call's epoch (round 040, ADR 0009 D2). The loop fires it exactly
// once per AI-endpoint call, so the second figure resets here; the total figure is
// unaffected (turn-scoped, never reset). The first frame is drawn synchronously.
func (s *Spinner) OnInferenceStart() {
	s.mu.Lock()
	s.callEpoch = s.now()
	s.mu.Unlock()
	s.activate(ThinkingLabel(s.model), false)
}

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

// OnCallBegin is a no-op on the spinner: the round-034 per-call rendering is
// composed by the CLI's compositeObserver, not the spinner (ADR 0005 D1). The
// method exists so *Spinner still satisfies the extended agentport.LoopObserver.
func (s *Spinner) OnCallBegin(callIndex int, messages []llm.Message) {}

// OnCallEnd is a no-op on the spinner (ADR 0005 D1).
func (s *Spinner) OnCallEnd(callIndex int, usage llm.Usage, roundReasons []string, final bool) {}

// YieldIndicator is the loop-facing yield hook (internal/domain/agent.
// LoopObserver): it clears the indicator so a line the loop writes directly to
// the diagnostic stream starts on its own cleared row. The yield POLICY lives
// with the owner (ui.YieldController); this method only adapts the port to it
// (round 045 / ADR 0014).
func (s *Spinner) YieldIndicator() { NewYieldController(s).Yield() }

// RestoreIndicator is the loop-facing restore hook: it resumes the indicator
// after a yield. Delegates to the yield owner (ui.YieldController), which keeps
// the turn-scoped total and the current model call's elapsed (round-019 D4 /
// round-040 ADR 0009 D2).
func (s *Spinner) RestoreIndicator() { NewYieldController(s).Restore() }

// Stop synchronously clears the indicator and stops its redraw. It is idempotent.
func (s *Spinner) Stop() { s.deactivate() }

// activate starts the redraw (drawing the first frame synchronously) or relabels
// a running indicator. The total epoch is turn-scoped and never reset; activate
// preserves the current call's epoch (round 040).
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
	go s.loop(tick, s.stopCh, s.doneCh, false)
}

// AdmitResume is the round-040 WS-A resume-only admission path (ADR 0009 D4): it
// starts the redraw for a live-but-stopped indicator WITHOUT drawing the first
// frame synchronously — the redraw goroutine draws it immediately on start
// (renderFirst), so no blocking frame write is held inside the caller's block
// critical section (the coordinator admits under the block-writer mutex, TD-3).
// It is a defined no-op when the indicator is already running or has no phase
// label (RF-3), and never touches the turn/call epochs (they are preserved).
func (s *Spinner) AdmitResume() {
	s.mu.Lock()
	status := s.status
	if s.running || status == "" {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.frameIdx = 0
	// Capture the channels as LOCALS and pass them to the goroutine (R-40-1): a
	// later AdmitResume/activate sets the FIELDS under the mutex, so reading the
	// fields after unlocking would be an unsynchronized read — the harmful
	// interleaving being a mismatched stop/done pair (a wedged deactivate or a
	// "close of closed channel" panic). activate() already passes locals.
	stopCh, doneCh := make(chan struct{}), make(chan struct{})
	s.stopCh, s.doneCh = stopCh, doneCh
	tick, stopTicker := s.newTicker()
	s.stopTicker = stopTicker
	s.mu.Unlock()
	go s.loop(tick, stopCh, doneCh, true)
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

// loop advances and redraws the frame on each tick until stopped. renderFirst
// (round 040) draws one frame immediately on start — the WS-A resume path
// (AdmitResume) uses it so the resumed frame appears without waiting a full poll
// period; activate keeps renderFirst=false (its synchronous first frame is drawn
// before the goroutine starts).
func (s *Spinner) loop(tick <-chan time.Time, stop, done chan struct{}, renderFirst bool) {
	defer close(done)
	if renderFirst {
		s.mu.Lock()
		if !s.running {
			s.mu.Unlock()
			return
		}
		s.renderLocked()
		s.mu.Unlock()
	}
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

// renderLocked writes one redrawn frame (the mutex must be held). Round 040
// measures TWO elapsed figures: the TOTAL from the turn epoch (turn-scoped, never
// reset) and the CURRENT MODEL CALL's elapsed from callEpoch (reset per
// AI-endpoint call). Round 025 erases every row the PREVIOUS frame occupied before
// drawing, and records the new frame's row count so the next redraw / the clear
// can erase all of them.
func (s *Spinner) renderLocked() {
	if s.w == nil {
		return
	}
	frame := SpinnerFrames[s.frameIdx%len(SpinnerFrames)]
	now := s.now()
	resource := ""
	if s.toolPhase {
		cpu, mem := s.sampleResources(now)
		resource = FormatResourceSegment(cpu, mem)
	}
	line := FormatSpinnerLine(frame, s.status, wholeSecondsSince(s.epoch, now), wholeSecondsSince(s.callEpoch, now), resource)
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

// sampleResources returns the machine's CPU/memory percentages for the
// tool-phase resource segment, throttled to at most one provider Sample per
// ResourceSampleInterval (round 059). The frame redraws every 200 ms, but the
// figures refresh once per second — the last sampled pair is reused in between —
// so the numbers stay readable while the wheel keeps animating; a nil metrics
// provider disables the segment. The throttle reads the spinner's injected clock
// (this method runs under the spinner mu, with `now` already taken by
// renderLocked), so it is deterministic under test.
func (s *Spinner) sampleResources(now time.Time) (cpu, mem float64) {
	if s.metrics == nil {
		return 0, 0
	}
	if s.haveSample && now.Sub(s.lastSample) < ResourceSampleInterval {
		return s.cachedCPU, s.cachedMem
	}
	s.cachedCPU, s.cachedMem = s.metrics.Sample()
	s.lastSample = now
	s.haveSample = true
	return s.cachedCPU, s.cachedMem
}
