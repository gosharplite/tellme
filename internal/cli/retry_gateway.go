package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gosharplite/tellme/internal/domain/llm"
	"github.com/gosharplite/tellme/internal/domain/render"
)

// retryDelays is the fixed, bounded retry schedule (round 078; ADR 0050 D3):
// after a retryable provider failure, wait 1 s and retry, then wait 3 s and
// retry — at most TWO retries / THREE attempts — then fail with the frozen
// provider phrase + exit 6. Deliberately a fixed constant pair (the operator's
// "simple" ask), not config; a LITERAL unit pin fixes the values (the
// round-064/076 literal-pin precedent).
var retryDelays = []time.Duration{1 * time.Second, 3 * time.Second}

// retryDelayEnv collapses both delays to a single millisecond value for the
// hermetic E2E (round 078; ADR 0050 D3) — mirroring TELL_ME_FORCE_TOOLOUTPUT_IDLE_MS.
// Unset/invalid ⇒ the real delays above. The E2E sets it to 0 so a retry asserts
// COUNT/ORDER, never wall-clock.
const retryDelayEnv = "TELL_ME_FORCE_RETRY_DELAY_MS"

// resolveRetryDelays returns the schedule for a run: the hermetic override when
// TELL_ME_FORCE_RETRY_DELAY_MS is a valid non-negative integer (milliseconds),
// else the real retryDelays.
func resolveRetryDelays() []time.Duration {
	if v := strings.TrimSpace(os.Getenv(retryDelayEnv)); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms >= 0 {
			d := time.Duration(ms) * time.Millisecond
			return []time.Duration{d, d}
		}
	}
	return retryDelays
}

// retryingGateway is the round-078 (ADR 0050) transport-retry decorator. It
// wraps the resolved provider gateway and, on a RETRYABLE failure (classified by
// the domain-owned llm.Retryable — transport, or HTTP 429/5xx), waits the fixed
// schedule and re-sends the SAME request. A non-retryable failure is returned at
// once (one attempt). Exhaustion returns the last error unchanged, so the CLI's
// provider phrase + exit 6 are byte-identical to a no-retry failure.
//
// It is the INNER wrapper (withProviderRetry withUnpairedDiagnostic) so the
// round-068 unpaired diagnostic inspects the POST-retry response. Both provider
// families are covered by this ONE seam (FR-006/I-7).
//
// Accounting invariant (I-4 / D7): a successful retry returns a SINGLE
// llm.Response, so the loop sees one Complete return — one `calls` entry, one
// usage record, one turn frame; a failed attempt writes no history (the loop
// persists only a completed turn). Cancellation (I-3 / D6): a parent-context
// cancellation aborts immediately — never sleep through, never retry after.
type retryingGateway struct {
	inner llm.Gateway
	// delays is the schedule; len(delays) retries, len(delays)+1 attempts.
	delays []time.Duration
	// notify announces one scheduled retry on the diagnostic stream; nil = silent.
	notify func(attempt, total int, delay time.Duration, err error)
	// sleep waits delay, returning false when ctx is cancelled during the wait.
	sleep func(ctx context.Context, delay time.Duration) bool
}

func (g retryingGateway) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	total := len(g.delays) + 1
	var lastErr error
	for i := 0; i < total; i++ {
		// A parent cancellation (SIGINT/SIGTERM) aborts at once — before the call
		// and before any wait — surfacing the last classified failure (or ctx err).
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return llm.Response{}, lastErr
			}
			return llm.Response{}, err
		}
		resp, err := g.inner.Complete(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if i == len(g.delays) {
			break // the last attempt failed — no more retries
		}
		if !llm.Retryable(err) {
			return llm.Response{}, err // non-retryable — fail immediately (one attempt)
		}
		delay := g.delays[i]
		if g.notify != nil {
			g.notify(i+2, total, delay, err)
		}
		if !g.sleep(ctx, delay) {
			return llm.Response{}, lastErr // cancelled during the wait
		}
	}
	return llm.Response{}, lastErr
}

// realRetrySleep waits delay and reports whether the wait completed; a context
// cancellation during the wait returns false (the retry is abandoned). A
// non-positive delay returns immediately (a cancelled context yields false).
func realRetrySleep(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// withProviderRetry wraps a gateway with the bounded transport retry. A nil
// gateway, or an empty schedule, returns the gateway unchanged.
func withProviderRetry(gw llm.Gateway, notify func(attempt, total int, delay time.Duration, err error)) llm.Gateway {
	delays := resolveRetryDelays()
	if gw == nil || len(delays) == 0 {
		return gw
	}
	return retryingGateway{inner: gw, delays: delays, notify: notify, sleep: realRetrySleep}
}

// retryNotifier builds the diagnostic emit callback for a retry: a single plain
// line on the diagnostic stream (`stderr`), control-free (newlines folded). The
// line is chrome-styled (`[HH:MM:SS] …`) and deliberately NOT prefixed with the
// frozen class phrase — so `tellme: the provider request failed` appears exactly
// once, on final failure (ADR 0050 D5; the round-017 "exactly one `tellme:` line"
// contract for a failed request holds). When a live indicator is supplied its
// frame is YIELDED (cleared) before the line and RESTORED after, so the retry
// line never tears the spinner frame (ADR 0050 D5; ADR 0014's yield route). A nil
// stderr disables it.
func retryNotifier(env runtimeEnv, ind render.Indicator) func(attempt, total int, delay time.Duration, err error) {
	if env.stderr == nil {
		return nil
	}
	return func(attempt, total int, delay time.Duration, err error) {
		if ind != nil {
			ind.YieldIndicator()
		}
		detail := strings.ReplaceAll(err.Error(), "\n", " ")
		_, _ = fmt.Fprintf(env.stderr, "[%s] retrying the provider request in %s (attempt %d of %d): %s\n", env.now().Format("15:04:05"), delay, attempt, total, detail)
		if ind != nil {
			ind.RestoreIndicator()
		}
	}
}
