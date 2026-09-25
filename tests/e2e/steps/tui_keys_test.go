package steps

import "testing"

// TestTypeTextDeliversTheTerminalEnterByte pins the round-093 seam rule: a typed
// logical line break (`\n`) is encoded as the terminal Enter byte (CR), NOT a raw
// line feed. bubbletea decodes LF as `ctrl+j`, which the bubbles textarea does not
// treat as a newline, so a raw LF was silently dropped and the typed lines joined
// onto one row (issue #191). A revert to LF reddens this pin.
func TestTypeTextDeliversTheTerminalEnterByte(t *testing.T) {
	if typeText("line one\nline two") != "line one\rline two" {
		t.Fatalf("typeText did not encode the line break as Enter (CR): %q", typeText("line one\nline two"))
	}
	if typeText("hi") != "hi" {
		t.Fatalf("typeText altered a single-line value: %q", typeText("hi"))
	}
	if got := tuiKeysTypeAndAbort("line one\nline two"); got != "line one\rline two"+tuiKeyAbort {
		t.Fatalf("tuiKeysTypeAndAbort = %q, want the CR line break + abort", got)
	}
	if got := tuiKeysTypeAndSubmit("a\nb"); got != "a\rb"+tuiKeySubmit {
		t.Fatalf("tuiKeysTypeAndSubmit = %q, want the CR line break + submit", got)
	}
}
