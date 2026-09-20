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
	want := "[USER]\nhello\n\n[MODEL]\nNoted.\n\n"
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
	want := "[USER]\nq\n\n[MODEL]\na1\n\n"
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
