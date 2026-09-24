package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gosharplite/tellme/internal/domain/render"
)

// Round 073 (ADR 0045) — the `-l` listing adapter's bytes.

// TestListingReportsRoleHeadersAndSeparators pins the per-message shape: a role
// header line, the body, and exactly one blank line after every message.
func TestListingReportsRoleHeadersAndSeparators(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{Role: render.ListingOperator, Body: "hello"},
		{Role: render.ListingModel, Body: "Noted."},
	}, render.ListingSpec{Raw: true}) // raw ⇒ no rendering, so the bytes are exact
	want := "[USER]\nhello\n\n[TOOLS] (0 calls)\n\n[MODEL]\nNoted.\n\n"
	if out.String() != want {
		t.Fatalf("listing = %q, want %q", out.String(), want)
	}
}

// TestListingSeparatesEveryMessageOnce pins that a body already ending in a
// newline cannot produce a double gap, and that the listing ends with a blank
// line.
func TestListingSeparatesEveryMessageOnce(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{Role: render.ListingOperator, Body: "q\n"},
		{Role: render.ListingModel, Body: "a1\n\n"},
	}, render.ListingSpec{Raw: true})
	want := "[USER]\nq\n\n[TOOLS] (0 calls)\n\n[MODEL]\na1\n\n"
	if out.String() != want {
		t.Fatalf("listing = %q, want %q", out.String(), want)
	}
}

// TestListingRendersOnlyTheModelBody pins the clarify Q3 → B divergence: the
// model body is rendered (its literal Markdown markers are gone), the operator
// body is verbatim (its markers stay).
func TestListingRendersOnlyTheModelBody(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{Role: render.ListingOperator, Body: "What is **Alice**?"},
		{Role: render.ListingModel, Body: "**Alice**"},
	}, render.ListingSpec{})
	got := stripANSI(out.String())

	if !strings.Contains(got, "What is **Alice**?") {
		t.Errorf("the operator prompt must be verbatim; got %q", got)
	}
	if strings.Contains(got, "**Alice**\n") {
		t.Errorf("the model body must be rendered (markers absent); got %q", got)
	}
	if !strings.Contains(got, "Alice") {
		t.Errorf("the model body's words must be present; got %q", got)
	}
	if !strings.HasPrefix(got, "[USER]\n") {
		t.Errorf("the first header must be [USER]; got %q", got)
	}
	if !strings.Contains(got, "\n[MODEL]\n") {
		t.Errorf("the model header must follow the separator; got %q", got)
	}
}

// TestListingColourGate pins the stdout terminal gate: colour on ⇒ the reference
// blue/magenta SGR around the header lines; off ⇒ zero escape bytes.
func TestListingColourGate(t *testing.T) {
	msgs := []render.ListingMessage{
		{Role: render.ListingOperator, Body: "hi"},
		{Role: render.ListingModel, Body: "ok"},
	}
	var on bytes.Buffer
	NewListing().Render(&on, msgs, render.ListingSpec{Colour: true})
	if !strings.Contains(on.String(), colorBlue+"[USER]"+colorReset) {
		t.Errorf("the operator header must be blue; got %q", on.String())
	}
	if !strings.Contains(on.String(), colorMagenta+"[MODEL]"+colorReset) {
		t.Errorf("the model header must be magenta; got %q", on.String())
	}

	var off bytes.Buffer
	NewListing().Render(&off, msgs, render.ListingSpec{})
	// The header ACCENT is gated; the model body's glamour style output is the
	// answer path's behaviour (gated by -r alone, round 006) and is not an accent.
	if strings.Contains(off.String(), colorBlue) || strings.Contains(off.String(), colorMagenta) {
		t.Errorf("a non-colour listing must carry no header accent; got %q", off.String())
	}
	if !strings.Contains(off.String(), "[USER]") || !strings.Contains(off.String(), "[MODEL]") {
		t.Errorf("the headers must be printed without colour; got %q", off.String())
	}
}

// TestListingRawSuppressesColour pins that -r disables the accents even when the
// colour gate is on.
func TestListingRawSuppressesColour(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{Role: render.ListingOperator, Body: "hi"},
	}, render.ListingSpec{Colour: true, Raw: true})
	if strings.ContainsRune(out.String(), '\x1b') {
		t.Errorf("a raw listing must carry no escape byte; got %q", out.String())
	}
}

// stripANSI removes SGR escape sequences so a rendered body can be asserted by
// its visible words (the round-006 predicate precedent: glamour output carries
// colour/padding, so exact bytes are not the contract).
func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			j := i + 1
			for j < len(s) && s[j] != 'm' {
				j++
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// Round 082 (ADR 0054) — the backward turn index in the role headers.

// TestListingReportsBackwardTurnIndex pins the suffixed, whole-label header
// bytes: `[USER] - N` / `[MODEL] - N`.
func TestListingReportsBackwardTurnIndex(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{Role: render.ListingOperator, Body: "q2", TurnIndex: 2},
		{Role: render.ListingModel, Body: "a2", TurnIndex: 2},
		{Role: render.ListingOperator, Body: "q1", TurnIndex: 1},
		{Role: render.ListingModel, Body: "a1", TurnIndex: 1},
	}, render.ListingSpec{Raw: true}) // raw ⇒ exact bytes
	want := "[USER] - 2\nq2\n\n[TOOLS] - 2 (0 calls)\n\n[MODEL] - 2\na2\n\n[USER] - 1\nq1\n\n[TOOLS] - 1 (0 calls)\n\n[MODEL] - 1\na1\n\n"
	if out.String() != want {
		t.Fatalf("listing = %q, want %q", out.String(), want)
	}
}

// TestListingAccentsTheWholeTurnIndexLabel pins that the colour unit encloses the
// WHOLE label — role word AND the ` - N` suffix.
func TestListingAccentsTheWholeTurnIndexLabel(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{Role: render.ListingOperator, Body: "hi", TurnIndex: 1},
		{Role: render.ListingModel, Body: "ok", TurnIndex: 1},
	}, render.ListingSpec{Colour: true})
	if !strings.Contains(out.String(), colorBlue+"[USER] - 1"+colorReset) {
		t.Errorf("the whole operator label (incl. the index) must be blue; got %q", out.String())
	}
	if !strings.Contains(out.String(), colorMagenta+"[MODEL] - 1"+colorReset) {
		t.Errorf("the whole model label (incl. the index) must be magenta; got %q", out.String())
	}
}

// TestListingHeaderBareLabelWithoutTurnIndex pins the non-positive-index fallback
// (a non-history caller): the bare label, never a ` - 0` suffix.
func TestListingHeaderBareLabelWithoutTurnIndex(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{Role: render.ListingOperator, Body: "hi"}, // TurnIndex 0
	}, render.ListingSpec{Raw: true})
	if !strings.HasPrefix(out.String(), "[USER]\n") {
		t.Errorf("a non-positive index must fall back to the bare label; got %q", out.String())
	}
	if strings.Contains(out.String(), " - 0") {
		t.Errorf("the header must never print a ` - 0` suffix; got %q", out.String())
	}
}

// TestListingToolLineSuppressedUnderRawOnATerminal pins FR-005's clause for the
// TOOL line specifically: on a TERMINAL stdout (Colour on) under `-r` (Raw on)
// the listing must carry no escape byte at all — so the yellow `[TOOLS]` accent
// is suppressed even though the colour gate is on. This is the real production
// path `renderHistoryList` builds (`Colour: env.stdoutIsTerminal()`, `Raw: raw`);
// the round-073 `TestListingRawSuppressesColour` renders only a [USER] message and
// therefore never exercises the tool line (architect fold F-086-2).
func TestListingToolLineSuppressedUnderRawOnATerminal(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{Role: render.ListingOperator, Body: "q", TurnIndex: 1},
		{Role: render.ListingModel, Body: "a", TurnIndex: 1, ToolCount: 2},
	}, render.ListingSpec{Colour: true, Raw: true})
	if strings.ContainsRune(out.String(), '\x1b') {
		t.Errorf("a raw (-r) terminal listing must carry no escape byte; got %q", out.String())
	}
	if !strings.Contains(out.String(), "[TOOLS] - 1 (2 calls)") {
		t.Errorf("the tool line must still be printed plainly under -r; got %q", out.String())
	}
}

// Round 086 (ADR 0057) — the `[TOOLS] - M (N calls)` line.

// TestListingReportsTheToolActivityLine pins the line's bytes and placement: it
// reads `[TOOLS] - M (N calls)` as its OWN block between the turn's [USER] block
// and its [MODEL] header, with N the message's ToolCount (0 included).
func TestListingReportsTheToolActivityLine(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{Role: render.ListingOperator, Body: "q2", TurnIndex: 2},
		{Role: render.ListingModel, Body: "a2", TurnIndex: 2, ToolCount: 3},
		{Role: render.ListingOperator, Body: "q1", TurnIndex: 1},
		{Role: render.ListingModel, Body: "a1", TurnIndex: 1, ToolCount: 0},
	}, render.ListingSpec{Raw: true}) // raw ⇒ exact bytes
	want := "[USER] - 2\nq2\n\n[TOOLS] - 2 (3 calls)\n\n[MODEL] - 2\na2\n\n" +
		"[USER] - 1\nq1\n\n[TOOLS] - 1 (0 calls)\n\n[MODEL] - 1\na1\n\n"
	if out.String() != want {
		t.Fatalf("listing = %q, want %q", out.String(), want)
	}
}

// TestListingToolLineAccentsWholeLabelYellow pins the colour unit and its gate:
// the WHOLE `[TOOLS] …` label is wrapped yellow only when colour is on.
func TestListingToolLineAccentsWholeLabelYellow(t *testing.T) {
	msgs := []render.ListingMessage{
		{Role: render.ListingOperator, Body: "hi", TurnIndex: 1},
		{Role: render.ListingModel, Body: "ok", TurnIndex: 1, ToolCount: 2},
	}
	var on bytes.Buffer
	NewListing().Render(&on, msgs, render.ListingSpec{Colour: true})
	if !strings.Contains(on.String(), colorYellow+"[TOOLS] - 1 (2 calls)"+colorReset) {
		t.Errorf("the whole tool label must be yellow; got %q", on.String())
	}
	var off bytes.Buffer
	NewListing().Render(&off, msgs, render.ListingSpec{}) // no colour gate
	if strings.Contains(off.String(), colorYellow) {
		t.Errorf("a non-colour listing must carry no yellow accent; got %q", off.String())
	}
	if !strings.Contains(off.String(), "[TOOLS] - 1 (2 calls)") {
		t.Errorf("the tool line must still be printed plainly; got %q", off.String())
	}
}

// TestListingToolLineBareLabelWithoutTurnIndex pins the non-positive-index
// fallback: the bare `[TOOLS] (N calls)` label, never ` - 0` (the round-082
// bare-label fallback shape).
func TestListingToolLineBareLabelWithoutTurnIndex(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{Role: render.ListingModel, Body: "ok", ToolCount: 2}, // TurnIndex 0
	}, render.ListingSpec{Raw: true})
	if !strings.HasPrefix(out.String(), "[TOOLS] (2 calls)\n") {
		t.Errorf("a non-positive index must fall back to the bare tool label; got %q", out.String())
	}
	if strings.Contains(out.String(), " - 0 ") {
		t.Errorf("the tool line must never print a ` - 0` suffix; got %q", out.String())
	}
}

// TestListingToolLineRidesTheModelMessage pins EC-002: a partial listing whose
// first message is the answer still shows the turn's tool line (the line rides
// the MODEL message, not the [USER] one).
func TestListingToolLineRidesTheModelMessage(t *testing.T) {
	var out bytes.Buffer
	NewListing().Render(&out, []render.ListingMessage{
		{ // the lone listed message of an odd `-l 1` window: the answer
			Role: render.ListingModel, Body: "a1", TurnIndex: 1, ToolCount: 1,
		},
	}, render.ListingSpec{Raw: true})
	want := "[TOOLS] - 1 (1 calls)\n\n[MODEL] - 1\na1\n\n"
	if out.String() != want {
		t.Fatalf("listing = %q, want %q", out.String(), want)
	}
}
