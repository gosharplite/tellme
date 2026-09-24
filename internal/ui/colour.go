package ui

// Round-054 chrome colour policy (ADR 0023), extended by round 057 (ADR 0027).
// The diagnostic chrome gains GREEN accents on four elements — the whole
// `[Tool Reason]` line, the `MODE` token in both `Payload` lines, the measured
// token number in the measured `Payload` line, and the third (session) cost in
// `╰─⠿ Ready` — plus, since round 057, GREY on the whole `[Tool Output]` block (the
// header + both separators; round 058 (ADR 0028) extends it to the streamed
// content lines), and YELLOW on the whole `[Tool Action]` line. The COLOUR CODES are the reference's (`tell-me-go/internal/ui/colors.go`:
// `colorGreen` / `colorGray` / `colorYellow`); the ELEMENT SET is tellme's own — a
// recorded divergence.
//
// The colour is emitted only when the caller passes enabled=true — the CLI gates
// it on the diagnostic stream being a terminal AND `-r` being off (the round-019
// spinner gate), so redirected/piped output, `-r`, and the offline paths stay
// byte-identical to the plain form. Colour never enters `stdout` or `turns.log`
// (the round-053 turn-log tee keeps the plain chrome: ADR 0022 RF-53-1 / FR-006).

const (
	// colorGreen is the reference's 8-colour SGR green (tell-me-go colors.go).
	colorGreen = "\033[0;32m"
	// colorGray is the reference's bright-black (grey) SGR (tell-me-go
	// colors.go). Round 057 (ADR 0027) greyed the `[Tool Output]` header line and
	// both horizontal separators; round 058 (ADR 0028) extends the grey to the
	// streamed content lines, so the whole block is grey on a colour-enabled
	// terminal.
	colorGray = "\033[0;90m"
	// colorYellow is the reference's SGR yellow (tell-me-go colors.go). Round 057
	// (ADR 0027): the whole `[Tool Action]` line is wrapped yellow on a
	// colour-enabled terminal. Round 086 (ADR 0057): the same yellow wraps the
	// whole `[TOOLS] - M (N calls)` line of the `-l`/`--list` listing.
	colorYellow = "\033[0;33m"
	// colorBlue is the reference's bright-blue SGR (tell-me-go colors.go). Round
	// 073 (ADR 0045): the `[USER]` header of the `-l` listing, on a terminal
	// stdout. Round 082 (ADR 0054; N-082-1): the WHOLE label is the colour unit,
	// so the wrap encloses the suffix too (`[USER] - N`).
	colorBlue = "\033[1;34m"
	// colorMagenta is the reference's bright-magenta SGR (tell-me-go colors.go).
	// Round 073 (ADR 0045): the `[MODEL]` header of the `-l` listing, on a
	// terminal stdout. Round 082 (ADR 0054; N-082-1): the WHOLE label is the
	// colour unit, so the wrap encloses the suffix too (`[MODEL] - N`).
	colorMagenta = "\033[1;35m"
	// colorReset clears the SGR state, returning the terminal to default.
	colorReset = "\033[0m"
)

// green wraps s in the reference's green SGR pair when enabled. An empty string
// is never wrapped (so a format slot for an empty field — e.g. an empty MODE —
// contributes no stray escape), and a disabled call returns s verbatim (the
// byte-identical plain form).
func green(s string, enabled bool) string { return wrap(s, colorGreen, enabled) }

// grey wraps s in the reference's grey SGR pair when enabled (round 057; the
// same whole-value, empty-safe, plain-when-disabled rule as green).
func grey(s string, enabled bool) string { return wrap(s, colorGray, enabled) }

// yellow wraps s in the reference's yellow SGR pair when enabled (round 057).
func yellow(s string, enabled bool) string { return wrap(s, colorYellow, enabled) }

// blue wraps s in the reference's bright-blue SGR pair when enabled (round 073;
// the same whole-value, empty-safe, plain-when-disabled rule as green).
func blue(s string, enabled bool) string { return wrap(s, colorBlue, enabled) }

// magenta wraps s in the reference's bright-magenta SGR pair when enabled
// (round 073).
func magenta(s string, enabled bool) string { return wrap(s, colorMagenta, enabled) }

// wrap applies one SGR pair around s when enabled. An empty string is never
// wrapped (no stray escape for an empty slot) and a disabled call returns s
// verbatim (the byte-identical plain form).
func wrap(s, code string, enabled bool) string {
	if !enabled || s == "" {
		return s
	}
	return code + s + colorReset
}
