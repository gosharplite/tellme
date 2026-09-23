package steps

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Round 073 (ADR 0045) — the `-l` listing's per-message presentation: role
// header lines, the rendered model body, the verbatim operator body, the blank
// separator, and the stdout-gated header accents.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the output is shown at a terminal$`, givenStdoutTerminal)
		ctx.When(`^the operator asks tellme to list the last (\d+) messages as raw output$`, whenListLastRaw)
		ctx.Then(`^tellme heads each listed message with its role$`, thenHeadsEachMessageWithItsRole)
		ctx.Then(`^the listing heads each message with its backward turn index$`, thenListingHeadsWithBackwardTurnIndex)
		ctx.Then(`^the listed model answer is presented as formatted prose$`, thenModelAnswerRendered)
		ctx.Then(`^the listed operator prompt is shown verbatim$`, thenOperatorPromptVerbatim)
		ctx.Then(`^the listed model answer is shown as its raw source$`, thenModelAnswerRawSource)
		ctx.Then(`^the listed messages are separated by a blank line$`, thenMessagesSeparatedByBlankLine)
		ctx.Then(`^the listing accents the operator role in blue and the model role in magenta$`, thenListingAccentsRoles)
		ctx.Then(`^the listing carries no accents$`, thenListingCarriesNoAccents)
	})
}

// givenStdoutTerminal forces tellme's STDOUT terminal probe to report a terminal
// via the round-073 diagnostic seam, so the listing's header accents are
// E2E-drivable over a pipe without a pty (怎麼做: the seam is set for the run).
func givenStdoutTerminal(ctx context.Context) error {
	scenarioFrom(ctx).setEnv("TELL_ME_FORCE_STDOUT_TTY", "1")
	return nil
}

// whenListLastRaw runs `tellme -l {count} -r` (no -c, no positional prompt).
func whenListLastRaw(ctx context.Context, count int) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-l", strconv.Itoa(count), "-r"}
	sc.run()
	return nil
}

// listingBlock is one listed message: the role header text (the bare role), the
// message's backward turn index (round 082; 0 when the header carries no index),
// and the (visible) body.
type listingBlock struct {
	role  string
	index int
	body  string
}

// parseListingHeader parses a role header line, optionally suffixed with the
// round-082 backward turn index: `[USER]` / `[MODEL]` / `[USER] - N` /
// `[MODEL] - N`. It returns the bare role, the index (0 when absent), and
// whether the line is a header at all.
func parseListingHeader(line string) (role string, index int, ok bool) {
	for _, r := range []string{"[USER]", "[MODEL]"} {
		if line == r {
			return r, 0, true
		}
		if v, found := strings.CutPrefix(line, r+" - "); found {
			if n, err := strconv.Atoi(v); err == nil {
				return r, n, true
			}
		}
	}
	return "", 0, false
}

// listingBlocks splits a listing capture into its messages by the role headers.
// Escape sequences are stripped first: the rendered model body's glamour style
// output is not the listing contract (the round-006 predicate precedent), only
// the structure — header, body, separator — is.
func listingBlocks(out string) []listingBlock {
	lines := strings.Split(stripANSI(out), "\n")
	var blocks []listingBlock
	for _, ln := range lines {
		t := strings.TrimRight(ln, " \t")
		if role, idx, ok := parseListingHeader(t); ok {
			blocks = append(blocks, listingBlock{role: role, index: idx})
			continue
		}
		if len(blocks) == 0 {
			continue
		}
		if strings.TrimSpace(t) == "" {
			continue // the separator (or glamour padding)
		}
		b := &blocks[len(blocks)-1]
		if b.body != "" {
			b.body += "\n"
		}
		b.body += t
	}
	return blocks
}

// stripANSI removes SGR escape sequences (CSI … m) so the listing's visible text
// can be asserted.
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

// expectedListingRoles is the role sequence the listing must show for the
// arranged exchanges, truncated to the last {count} messages (operator prompt,
// model answer per exchange).
func expectedListingRoles(sc *scenarioContext, count int) []string {
	var roles []string
	for range sc.arrangedExchanges {
		roles = append(roles, "[USER]", "[MODEL]")
	}
	if len(roles) > count {
		roles = roles[len(roles)-count:]
	}
	return roles
}

// arrangedListCount reads the `-l N` count from the arranged arguments (1 when
// the count is omitted — the round-054 default). It lets a Then derive the
// EXPECTED listing length from the request, not from the observed output, so a
// dropped message reddens (round-073 fold TD-5).
func arrangedListCount(sc *scenarioContext) int {
	for i, a := range sc.args {
		if a == "-l" || a == "--list" {
			if i+1 < len(sc.args) {
				if n, err := strconv.Atoi(sc.args[i+1]); err == nil {
					return n
				}
			}
			return 1
		}
		if v, ok := strings.CutPrefix(a, "-l="); ok {
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
		if v, ok := strings.CutPrefix(a, "--list="); ok {
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
	}
	return 1
}

// thenHeadsEachMessageWithItsRole (必查 呈現結果): every listed message carries its
// role header, in order. The expected length comes from the `-l N` REQUEST (fold
// TD-5), so a dropped message fails.
func thenHeadsEachMessageWithItsRole(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	blocks := listingBlocks(sc.stdout)
	want := expectedListingRoles(sc, arrangedListCount(sc))
	if len(blocks) != len(want) {
		return fmt.Errorf("the listing showed %d messages, want %d; stdout=%q", len(blocks), len(want), sc.stdout)
	}
	for i := range want {
		if blocks[i].role != want[i] {
			return fmt.Errorf("message %d header = %q, want %q; stdout=%q", i, blocks[i].role, want[i], sc.stdout)
		}
	}
	return nil
}

// thenModelAnswerRendered (必查 呈現結果): every listed model body is the RENDERED
// form — its raw Markdown markers are absent — while at least one listed body
// carries an arranged answer's words. Only the LISTED messages are inspected (a
// truncated `-l N` may omit earlier exchanges).
func thenModelAnswerRendered(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	var models []listingBlock
	for _, b := range listingBlocks(sc.stdout) {
		if b.role == "[MODEL]" {
			models = append(models, b)
		}
	}
	if len(models) == 0 {
		return fmt.Errorf("the listing showed no [MODEL] message; stdout=%q", sc.stdout)
	}
	markerAnswerArranged := false
	for _, x := range sc.arrangedExchanges {
		if strings.Contains(x.answer, "**") {
			markerAnswerArranged = true
		}
	}
	if !markerAnswerArranged {
		return fmt.Errorf("the scenario arranged no marker-bearing answer to witness rendering")
	}
	// No listed model body may carry raw Markdown (the render must have run) …
	for _, b := range models {
		if strings.Contains(b.body, "**") {
			return fmt.Errorf("the model body still carries its raw Markdown (%q); stdout=%q", b.body, sc.stdout)
		}
	}
	// … and at least one LISTED body must be the rendered form of a marker-bearing
	// arranged answer (fold F-1: it's not enough that some answer was arranged).
	matched := false
	for _, x := range sc.arrangedExchanges {
		if !strings.Contains(x.answer, "**") {
			continue
		}
		plain := strings.ReplaceAll(x.answer, "**", "")
		for _, b := range models {
			if plain != "" && strings.Contains(b.body, plain) {
				matched = true
			}
		}
	}
	if !matched {
		return fmt.Errorf("no listed model body shows a RENDERED marker-bearing answer; stdout=%q", sc.stdout)
	}
	return nil
}

// thenOperatorPromptVerbatim (必查 呈現結果): the operator body is the prompt
// EXACTLY as written (its raw Markdown markers survive).
func thenOperatorPromptVerbatim(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	var userBlocks []listingBlock
	for _, b := range listingBlocks(sc.stdout) {
		if b.role == "[USER]" {
			userBlocks = append(userBlocks, b)
		}
	}
	if len(userBlocks) == 0 {
		return fmt.Errorf("the listing showed no [USER] message; stdout=%q", sc.stdout)
	}
	for _, x := range sc.arrangedExchanges {
		found := false
		for _, b := range userBlocks {
			if strings.Contains(b.body, x.prompt) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("the operator prompt %q is not shown verbatim; stdout=%q", x.prompt, sc.stdout)
		}
	}
	return nil
}

// thenModelAnswerRawSource (必查 呈現結果): under `-r`, the model body is the
// answer EXACTLY as written (its raw Markdown markers present, no escape byte).
func thenModelAnswerRawSource(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if strings.ContainsRune(sc.stdout, '\x1b') {
		return fmt.Errorf("a raw listing must carry no escape byte; stdout=%q", sc.stdout)
	}
	if len(sc.arrangedExchanges) == 0 {
		return fmt.Errorf("no arranged exchange to witness the raw source")
	}
	answer := sc.arrangedExchanges[len(sc.arrangedExchanges)-1].answer
	if !strings.Contains(sc.stdout, answer) {
		return fmt.Errorf("the raw model body must be the answer %q verbatim; stdout=%q", answer, sc.stdout)
	}
	// Round-073 fold N-1: one blank line closes every message, so the listing
	// ends with a blank line (the reference's Fprintln-after-every-block shape).
	if !strings.HasSuffix(sc.stdout, "\n\n") {
		return fmt.Errorf("the listing must end with a blank line (a trailing \"\\n\\n\"); stdout=%q", sc.stdout)
	}
	return nil
}

// thenMessagesSeparatedByBlankLine (必查 呈現結果): the line immediately preceding
// each role header (after the first) is blank — exactly one separator, no run-on.
func thenMessagesSeparatedByBlankLine(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	lines := strings.Split(stripANSI(sc.stdout), "\n")
	headers := 0
	for i, ln := range lines {
		t := strings.TrimRight(ln, " \t")
		if _, _, ok := parseListingHeader(t); !ok {
			continue
		}
		headers++
		if headers == 1 || i == 0 {
			continue
		}
		if strings.TrimRight(lines[i-1], " \t") != "" {
			return fmt.Errorf("no blank line separates the messages (line %d = %q); stdout=%q", i-1, lines[i-1], sc.stdout)
		}
	}
	if headers == 0 {
		return fmt.Errorf("the listing showed no message; stdout=%q", sc.stdout)
	}
	return nil
}

// thenListingAccentsRoles (必查 呈現結果): on a terminal stdout the [USER] header is
// wrapped in the reference's bright blue and the [MODEL] header in bright magenta.
// Round 082 (ADR 0054): the wrap encloses the WHOLE label, including the ` - N`
// backward turn index (`\x1b[1;34m[USER] - N\x1b[0m`).
func thenListingAccentsRoles(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !strings.Contains(sc.stdout, "\x1b[1;34m[USER] - ") {
		return fmt.Errorf("the whole [USER] label (incl. the turn index) must be accented bright blue; stdout=%q", sc.stdout)
	}
	if !strings.Contains(sc.stdout, "\x1b[1;35m[MODEL] - ") {
		return fmt.Errorf("the whole [MODEL] label (incl. the turn index) must be accented bright magenta; stdout=%q", sc.stdout)
	}
	if !strings.Contains(sc.stdout, "\x1b[0m") {
		return fmt.Errorf("the accent must be reset with \\x1b[0m; stdout=%q", sc.stdout)
	}
	return nil
}

// thenListingCarriesNoAccents (必查 呈現結果): no header accent appears on a
// redirected stdout or under `-r`; a `-r` listing carries no escape byte at all
// (nothing renders), and the headers still read [USER] / [MODEL].
func thenListingCarriesNoAccents(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	for _, code := range []string{"\x1b[1;34m", "\x1b[1;35m"} {
		if strings.Contains(sc.stdout, code) {
			return fmt.Errorf("the listing must carry no header accent %q; stdout=%q", code, sc.stdout)
		}
	}
	raw := false
	for _, a := range sc.args {
		if a == "-r" || a == "--raw" {
			raw = true
		}
	}
	if raw && strings.ContainsRune(sc.stdout, '\x1b') {
		return fmt.Errorf("a -r listing must carry no escape byte; stdout=%q", sc.stdout)
	}
	// Round 082: the plain label still carries the backward turn index.
	for _, h := range []string{"[USER] - ", "[MODEL] - "} {
		if !strings.Contains(sc.stdout, h) {
			return fmt.Errorf("the header %q must still be printed (plain, with its turn index); stdout=%q", h, sc.stdout)
		}
	}
	return nil
}

// thenListingHeadsWithBackwardTurnIndex (必查 呈現結果; round 082, ADR 0054): the
// role header sequence equals the expected `[USER] - K` / `[MODEL] - K` lines,
// with K the turn's TRUE distance from the end of the loaded history (1 = the
// most recent turn). The expectation is derived from the ARRANGED exchanges and
// the `-l N` REQUEST (the last N MESSAGES), never from the observed output — so a
// dropped message, a missing index, or a re-numbered leading message reddens.
func thenListingHeadsWithBackwardTurnIndex(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	blocks := listingBlocks(sc.stdout)

	var want []string
	for i := range sc.arrangedExchanges {
		turnIndex := len(sc.arrangedExchanges) - i
		want = append(want,
			fmt.Sprintf("[USER] - %d", turnIndex),
			fmt.Sprintf("[MODEL] - %d", turnIndex))
	}
	if n := arrangedListCount(sc); len(want) > n {
		want = want[len(want)-n:]
	}
	if len(blocks) != len(want) {
		return fmt.Errorf("the listing showed %d messages, want %d; stdout=%q", len(blocks), len(want), sc.stdout)
	}
	for i, w := range want {
		got := blocks[i].role
		if blocks[i].index > 0 {
			got = fmt.Sprintf("%s - %d", blocks[i].role, blocks[i].index)
		}
		if got != w {
			return fmt.Errorf("message %d header = %q, want %q; stdout=%q", i, got, w, sc.stdout)
		}
	}
	return nil
}
