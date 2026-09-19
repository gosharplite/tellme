package cli

import (
	"bytes"
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
