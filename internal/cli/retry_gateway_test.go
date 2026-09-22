package cli

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/domain/llm"
)

// scriptedGateway is a Gateway whose Complete returns the scripted error for
// call i (nil = success). It counts calls so a test can assert the retry count.
type scriptedGateway struct {
	calls int
	errs  []error
}

func (g *scriptedGateway) Complete(_ context.Context, _ llm.Request) (llm.Response, error) {
	i := g.calls
	g.calls++
	if i < len(g.errs) && g.errs[i] != nil {
		return llm.Response{}, g.errs[i]
	}
	return llm.Response{Text: "ok"}, nil
}

func transportErr() error {
	return &llm.ProviderError{Provider: "p", Err: errors.New("connection reset by peer"), Transport: true}
}

func statusErr(code int) error {
	return &llm.ProviderError{Provider: "p", Err: errors.New("scripted status"), Status: code}
}

// immediateSleep records the delays it was asked to wait and never blocks.
type immediateSleep struct{ waited []time.Duration }

func (s *immediateSleep) sleep(_ context.Context, d time.Duration) bool {
	s.waited = append(s.waited, d)
	return true
}

func newTestRetry(g llm.Gateway, delays []time.Duration, sl func(context.Context, time.Duration) bool, notify func(int, int, time.Duration, error)) retryingGateway {
	return retryingGateway{inner: g, delays: delays, sleep: sl, notify: notify}
}

// TestRetryDelaysLiteral pins the operator-locked schedule literally.
func TestRetryDelaysLiteral(t *testing.T) {
	want := []time.Duration{1 * time.Second, 3 * time.Second}
	if len(retryDelays) != len(want) {
		t.Fatalf("retryDelays has %d entries, want %d", len(retryDelays), len(want))
	}
	for i, w := range want {
		if retryDelays[i] != w {
			t.Fatalf("retryDelays[%d] = %v, want %v", i, retryDelays[i], w)
		}
	}
}

func TestRetryingGateway_OneDropThenSuccess(t *testing.T) {
	g := &scriptedGateway{errs: []error{transportErr()}}
	sl := &immediateSleep{}
	rg := newTestRetry(g, []time.Duration{0, 0}, sl.sleep, nil)
	resp, err := rg.Complete(context.Background(), llm.Request{})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Text != "ok" {
		t.Fatalf("answer = %q, want ok", resp.Text)
	}
	if g.calls != 2 {
		t.Fatalf("calls = %d, want 2", g.calls)
	}
	if len(sl.waited) != 1 {
		t.Fatalf("waits = %d, want 1", len(sl.waited))
	}
}

func TestRetryingGateway_TwoDropsThenSuccess(t *testing.T) {
	g := &scriptedGateway{errs: []error{transportErr(), statusErr(503)}}
	sl := &immediateSleep{}
	rg := newTestRetry(g, []time.Duration{0, 0}, sl.sleep, nil)
	if _, err := rg.Complete(context.Background(), llm.Request{}); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if g.calls != 3 {
		t.Fatalf("calls = %d, want 3", g.calls)
	}
}

func TestRetryingGateway_Exhausts(t *testing.T) {
	g := &scriptedGateway{errs: []error{transportErr(), transportErr(), transportErr()}}
	rg := newTestRetry(g, []time.Duration{0, 0}, (&immediateSleep{}).sleep, nil)
	if _, err := rg.Complete(context.Background(), llm.Request{}); err == nil {
		t.Fatal("Complete: want error after exhausting retries")
	}
	if g.calls != 3 {
		t.Fatalf("calls = %d, want 3 (2 retries, 3 attempts)", g.calls)
	}
}

func TestRetryingGateway_NonRetryableFailsAtOnce(t *testing.T) {
	g := &scriptedGateway{errs: []error{statusErr(400)}}
	sl := &immediateSleep{}
	rg := newTestRetry(g, []time.Duration{0, 0}, sl.sleep, nil)
	if _, err := rg.Complete(context.Background(), llm.Request{}); err == nil {
		t.Fatal("Complete: want error")
	}
	if g.calls != 1 {
		t.Fatalf("calls = %d, want 1 (a 400 is not retried)", g.calls)
	}
	if len(sl.waited) != 0 {
		t.Fatalf("waits = %d, want 0 (no wait for a non-retryable failure)", len(sl.waited))
	}
}

func TestRetryingGateway_NotifiesEachRetry(t *testing.T) {
	g := &scriptedGateway{errs: []error{transportErr(), transportErr()}}
	type note struct {
		attempt, total int
		delay          time.Duration
	}
	var notes []note
	notify := func(attempt, total int, d time.Duration, _ error) {
		notes = append(notes, note{attempt, total, d})
	}
	rg := newTestRetry(g, []time.Duration{0, 0}, (&immediateSleep{}).sleep, notify)
	if _, err := rg.Complete(context.Background(), llm.Request{}); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	want := []note{{2, 3, 0}, {3, 3, 0}}
	if len(notes) != len(want) {
		t.Fatalf("notes = %v, want %v", notes, want)
	}
	for i := range want {
		if notes[i] != want[i] {
			t.Fatalf("notes[%d] = %v, want %v", i, notes[i], want[i])
		}
	}
}

func TestRetryingGateway_CancellationDuringWaitAborts(t *testing.T) {
	g := &scriptedGateway{errs: []error{transportErr(), transportErr()}}
	cancelled := false
	cancelSleep := func(_ context.Context, _ time.Duration) bool {
		cancelled = true
		return false // the wait was abandoned
	}
	rg := newTestRetry(g, []time.Duration{0, 0}, cancelSleep, nil)
	if _, err := rg.Complete(context.Background(), llm.Request{}); err == nil {
		t.Fatal("Complete: want the classified failure when the wait is abandoned")
	}
	if !cancelled {
		t.Fatal("the interruptible sleep was not consulted")
	}
	if g.calls != 1 {
		t.Fatalf("calls = %d, want 1 (no retry after an abandoned wait)", g.calls)
	}
}

func TestRetryingGateway_ParentCancellationAborts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := &scriptedGateway{}
	rg := newTestRetry(g, []time.Duration{0, 0}, (&immediateSleep{}).sleep, nil)
	if _, err := rg.Complete(ctx, llm.Request{}); err == nil {
		t.Fatal("Complete: want an error on a cancelled context")
	}
	if g.calls != 0 {
		t.Fatalf("calls = %d, want 0 (never call the provider on a cancelled ctx)", g.calls)
	}
}

func TestRealRetrySleep(t *testing.T) {
	if !realRetrySleep(context.Background(), 0) {
		t.Fatal("a zero delay must return true")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if realRetrySleep(ctx, time.Second) {
		t.Fatal("a cancelled context must return false")
	}
}

func TestResolveRetryDelays_Override(t *testing.T) {
	t.Setenv(retryDelayEnv, "0")
	got := resolveRetryDelays()
	if len(got) != 2 || got[0] != 0 || got[1] != 0 {
		t.Fatalf("override delays = %v, want [0 0]", got)
	}
	t.Setenv(retryDelayEnv, "not-a-number")
	if got := resolveRetryDelays(); got[0] != time.Second || got[1] != 3*time.Second {
		t.Fatalf("invalid override must fall back to the real delays, got %v", got)
	}
}

func TestRetryNotifier_WritesPlainLine(t *testing.T) {
	var sb strings.Builder
	env := runtimeEnv{stderr: &sb}
	n := retryNotifier(env, nil)
	if n == nil {
		t.Fatal("notifier must be non-nil with a stderr")
	}
	n(2, 3, time.Second, errors.New("line1\nline2"))
	out := sb.String()
	if !strings.Contains(out, "retrying the provider request in 1s (attempt 2 of 3)") {
		t.Fatalf("unexpected line: %q", out)
	}
	if strings.Contains(out, "\nline2") {
		t.Fatalf("the detail must be folded to one line: %q", out)
	}
	if strings.Contains(out, "the provider request failed") {
		t.Fatalf("the retry line must NOT carry the frozen class phrase: %q", out)
	}
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.HasPrefix(line, "tellme: ") {
			t.Fatalf("the retry line must NOT carry the `tellme: ` class prefix (it would break the exactly-one-class-line contract): %q", line)
		}
	}
}
