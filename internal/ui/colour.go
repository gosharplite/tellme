package ui

// Round-054 chrome colour policy (ADR 0023). The diagnostic chrome gains GREEN
// accents on four elements — the whole `[Tool Reason]` line, the `MODE` token in
// both `Payload` lines, the measured token number in the measured `Payload` line,
// and the third (session) cost in `╰─⠿ Ready`. The COLOUR CODE is the reference's
// (`tell-me-go/internal/ui/colors.go`: `colorGreen = "\033[0;32m"`); the ELEMENT
// SET is tellme's own — a recorded divergence (the reference greens only the
// session cost, and prints its reason line / payload mode in gray).
//
// The colour is emitted only when the caller passes enabled=true — the CLI gates
// it on the diagnostic stream being a terminal AND `-r` being off (the round-019
// spinner gate), so redirected/piped output, `-r`, and the offline paths stay
// byte-identical to the plain form. Colour never enters `stdout` or `turns.log`
// (the round-053 turn-log tee keeps the plain chrome: ADR 0022 RF-53-1 / FR-006).

const (
	// colorGreen is the reference's 8-colour SGR green (tell-me-go colors.go).
	colorGreen = "\033[0;32m"
	// colorReset clears the SGR state, returning the terminal to default.
	colorReset = "\033[0m"
)

// green wraps s in the reference's green SGR pair when enabled. An empty string
// is never wrapped (so a format slot for an empty field — e.g. an empty MODE —
// contributes no stray escape), and a disabled call returns s verbatim (the
// byte-identical plain form).
func green(s string, enabled bool) string {
	if !enabled || s == "" {
		return s
	}
	return colorGreen + s + colorReset
}
