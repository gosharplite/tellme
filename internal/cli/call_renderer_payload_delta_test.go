package cli

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/config"
	"github.com/gosharplite/tellme/internal/domain/llm"
)

// Round 057 (ADR 0027): the estimated payload line shows the increment over the
// previous estimate emitted in the same process (in-memory, session-scoped); the
// first estimate has no predecessor (`+0`).
func TestOnCallBeginEmitsIncrement(t *testing.T) {
	var buf bytes.Buffer
	r := &callRenderer{
		env:     runtimeEnv{stderr: &buf, clock: func() time.Time { return r036Clock }},
		res:     resolution{Mode: "butler", Selected: "test", Provider: config.Provider{Model: "test-model"}},
		lines:   fakeLines{},
		pricing: llm.Pricing{},
	}
	small := []llm.Message{{Role: "user", Content: "hi"}}
	big := []llm.Message{{Role: "user", Content: "hi"}, {Role: "user", Content: strings.Repeat("x", 400)}}

	r.OnCallBegin(0, small)
	r.OnCallBegin(1, big)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want exactly two estimate lines; got %d: %q", len(lines), buf.String())
	}
	if !strings.Contains(lines[0], "+0 ") {
		t.Errorf("the first estimate must show +0; got %q", lines[0])
	}
	if strings.Contains(lines[1], "+0 ") {
		t.Errorf("the second estimate must show a positive increment; got %q", lines[1])
	}
}

// F-057-1 (round-057 review): the CLI's signed arithmetic must be asked to
// produce a NEGATIVE increment. Through the CLI the in-turn message set only
// grows, so a shrinking estimate is not reachable end-to-end; this unit pin
// falsifies the subtraction (a clamp-to-+0 mutant reds it) and is the named
// carrier for the shrink form (see ADR 0027 §Forward).
func TestOnCallBeginEmitsNegativeIncrement(t *testing.T) {
	var buf bytes.Buffer
	r := &callRenderer{
		env:     runtimeEnv{stderr: &buf, clock: func() time.Time { return r036Clock }},
		res:     resolution{Mode: "butler", Selected: "test", Provider: config.Provider{Model: "test-model"}},
		lines:   fakeLines{},
		pricing: llm.Pricing{},
	}
	big := []llm.Message{{Role: "user", Content: strings.Repeat("x", 2000)}}
	small := []llm.Message{{Role: "user", Content: "hi"}}

	r.OnCallBegin(0, big)
	r.OnCallBegin(1, small)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want exactly two estimate lines; got %d: %q", len(lines), buf.String())
	}
	// The fake renders `<estimate <tokens> <delta> <mode> <model>>` with `%+d`, so a
	// shrinking payload must show a signed-negative delta field.
	if !reNegativeEstimate.MatchString(lines[1]) {
		t.Errorf("the second estimate must show a negative increment; got %q", lines[1])
	}
}

var reNegativeEstimate = regexp.MustCompile(`^<estimate [0-9]+ -[0-9]+ butler test-model>$`)
