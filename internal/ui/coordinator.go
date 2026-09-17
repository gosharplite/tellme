package ui

import (
	"io"
	"sync"
	"time"
)

// Round 040 WS-A — the `[Tool Output]` block coordinator (ADR 0009 D3/D4). It is
// the single `internal/ui` object that owns the block writer AND the progress
// spinner, so `internal/cli` binds ONE value to the command tool's sink. It
// consolidates the **block-scoped** yield only (the `[Tool Output]` resume/clear):
// the other spinner yields remain where they are — the loop's `withToolLog`
// (Before/AfterToolLog) and `compositeObserver.yieldIndicatorBeforeTail` (round
// 035) — so the spinner-yield policy still has those homes (a `#69` ledger item;
// this change does not claim to have collapsed all of them).
//
// The writer stays the SOLE owner of the block mutex + row state (T002 B1: one
// lock owner, three entry points — WriteWith, EndWith, withLock); the coordinator
// never reaches into the writer's mutex. The resume/clear invariant is mutual
// exclusion + join: the watcher admits the resume under the writer's mutex
// (withLock), and the per-line clear is the presenter's synchronous,
// goroutine-joined deactivate() (via WriteWith's beforeLine), so lock order is
// block-writer mutex → spinner mutex. The block critical section never spans a
// frame write: the resume draws its first frame on the redraw goroutine
// (Spinner.AdmitResume).
//
// ACCEPTED RESIDUAL (N-40-4, recorded here so the next reader finds it): `End`
// while an output-line write is stuck on a wedged `stderr` cannot be resolved —
// `stopWatcher` joins the watcher, which may itself be parked inside `w.mu`
// (`withLock`), so nothing can abandon a blocked underlying write. The behaviour
// is bounded-then-accepted, not repaired: the stream is assumed to make progress
// (a real terminal / pipe does). The unit stress
// `TestCoordinatorEndWhileLineWriteInFlight` exercises the **released** writer,
// i.e. the recoverable shape of this case.
//
// The coordinator models ONE concurrent block (one writer + one watcher); if
// concurrent tool execution ever lands, a second open block plus the composite's
// unconditional `AfterToolLog` resume would break the idle-gap invariant (frames
// between output lines) — a `#69` forward item.

// DefaultToolOutputIdleGap is the default idle-gap threshold for the WS-A
// resume (ADR 0009 D3): the indicator reappears after N seconds with no new
// output line. The hermetic seam TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS overrides it
// (milliseconds; 0 = admit immediately; unset/invalid = this default), resolved
// once at construction by internal/cli.
const DefaultToolOutputIdleGap = 3 * time.Second

// ToolOutputCoordinator owns the `[Tool Output]` block writer and the progress
// spinner, and coordinates the WS-A idle-gap liveness between them.
type ToolOutputCoordinator struct {
	w       *ToolOutputWriter
	sp      *Spinner // nil when the spinner is gated off (a no-op coordinator)
	idleGap time.Duration

	// newTicker is the watcher's poll seam (the spinner's ~200 ms cadence); a
	// test injects a controllable channel. nil falls back to the real ticker.
	newTicker func() (<-chan time.Time, func())

	// watcher lifecycle (guarded by mu).
	mu              sync.Mutex
	watching        bool
	watchStop       chan struct{}
	watchDone       chan struct{}
	watchStopTicker func()
}

// NewToolOutputCoordinator builds the coordinator over the diagnostic stream,
// the writer's clock seam, the turn spinner (nil when gated off), and the
// resolved idle gap.
func NewToolOutputCoordinator(stream io.Writer, now func() time.Time, sp *Spinner, idleGap time.Duration) *ToolOutputCoordinator {
	return &ToolOutputCoordinator{
		w:         &ToolOutputWriter{W: stream, Now: now},
		sp:        sp,
		idleGap:   idleGap,
		newTicker: defaultToolOutputTicker,
	}
}

// defaultToolOutputTicker is the production watcher poll: the spinner's ~200 ms
// cadence.
func defaultToolOutputTicker() (<-chan time.Time, func()) {
	t := time.NewTicker(SpinnerInterval)
	return t.C, t.Stop
}

// Begin opens the block (ADR 0009 D3/D4): it clears the indicator, writes the
// header/opening separator (seeding the writer's idle clock), and starts the
// idle watcher. sink.Begin() runs BEFORE the child starts, so no drain goroutine
// exists yet and the header write cannot race a line or a frame (N-13) — that is
// why Begin needs no lock-scoped hook while End does.
func (c *ToolOutputCoordinator) Begin() {
	if c.w.W == nil {
		return
	}
	c.clearIndicator()
	c.w.Begin()
	c.startWatcher()
}

// Writer returns the sink's io.Writer: the writer's line path with the per-line
// clear hook wired (WriteWith).
func (c *ToolOutputCoordinator) Writer() io.Writer { return coordinatorSink{c} }

// End closes the block (ADR 0009 D4, R-8/R-1b): stop the watcher AND join it →
// clear (inside the writer's critical section, via EndWith's hook) → reset +
// closing separator → resume. The clear runs inside the writer's mutex because a
// drain goroutine can still be inside Write at End on the trim/timeout paths.
func (c *ToolOutputCoordinator) End() {
	if c.w.W == nil {
		return
	}
	c.stopWatcher()
	c.w.EndWith(c.clearIndicator)
	if c.sp != nil {
		// Resume the indicator for the rest of the turn (the standard AfterToolLog
		// path — a synchronous first frame, outside the block's critical section).
		c.sp.AfterToolLog()
	}
}

// clearIndicator synchronously clears the indicator (a no-op when gated off or
// already stopped). It is the goroutine-joined deactivate() (ADR 0009 D4).
func (c *ToolOutputCoordinator) clearIndicator() {
	if c.sp != nil {
		c.sp.Stop()
	}
}

// startWatcher starts the block-scoped idle watcher (a no-op when the spinner is
// gated off — FR-011).
func (c *ToolOutputCoordinator) startWatcher() {
	if c.sp == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.watching {
		return
	}
	tick, stopTicker := c.newTicker()
	c.watchStop = make(chan struct{})
	c.watchDone = make(chan struct{})
	c.watchStopTicker = stopTicker
	c.watching = true
	go c.watch(tick, c.watchStop, c.watchDone)
}

// stopWatcher stops AND joins the idle watcher (stopping the ticker alone is not
// stopping the goroutine), so an in-flight admitResume cannot race the close.
func (c *ToolOutputCoordinator) stopWatcher() {
	c.mu.Lock()
	if !c.watching {
		c.mu.Unlock()
		return
	}
	stop, done, stopTicker := c.watchStop, c.watchDone, c.watchStopTicker
	c.watching = false
	c.mu.Unlock()
	close(stop)
	<-done
	if stopTicker != nil {
		stopTicker()
	}
}

// watch polls on the ~200 ms cadence. On each poll it evaluates the writer-owned
// idle clock under the writer's mutex and admits the resume when the gap has
// elapsed — one critical section for check + admit (R-11), honouring lock order
// block-mutex → spinner-mutex.
func (c *ToolOutputCoordinator) watch(tick <-chan time.Time, stop, done chan struct{}) {
	defer close(done)
	for {
		select {
		case <-stop:
			return
		case <-tick:
			c.w.withLock(func(idle time.Duration) {
				if c.sp != nil && idle >= c.idleGap {
					c.sp.AdmitResume()
				}
			})
		}
	}
}

// coordinatorSink is the sink's io.Writer: it routes each write through the
// writer's line path with the per-line clear hook.
type coordinatorSink struct{ c *ToolOutputCoordinator }

// Write routes p through the writer's WriteWith with the per-line clear.
func (s coordinatorSink) Write(p []byte) (int, error) {
	return s.c.w.WriteWith(p, s.c.clearIndicator)
}
