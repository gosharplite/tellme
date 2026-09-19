package steps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

// Round 057 (ADR 0027) — the payload increment (the estimated line shows
// `+<delta> ~<tokens>`, no budget) and the two new chrome accents (the
// `[Tool Output]` frame grey, the `[Tool Action]` line yellow).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the estimated payload line does not show the allowance$`, thenEstimatedNoAllowance)
		ctx.Then(`^the first estimated payload line shows an increase of zero$`, thenFirstEstimateZero)
		ctx.Then(`^a later estimated payload line shows a positive increase$`, thenLaterEstimatePositive)
		ctx.Then(`^the tool output frame is shown in grey$`, thenToolOutputGrey)
		ctx.Then(`^the action line is shown in yellow$`, thenActionYellow)
		ctx.Then(`^the saved turn log carries the pre-flight payload with its increment and no allowance$`, thenTurnLogPayloadParity)
	})
}

// The reference's grey/yellow SGR pairs (tell-me-go colors.go).
var (
	reGreyHeader    = regexp.MustCompile("\x1b\\[0;90m\\[[0-9]{2}:[0-9]{2}:[0-9]{2}\\] \\[Tool Output\\] ")
	reGreySeparator = regexp.MustCompile("\x1b\\[0;90m-{60}\x1b\\[0m")
	// reGreyToolOutputPrefix matches a grey-wrapped `[Tool Output] …` line (header
	// OR content); round 058 counts them to prove the content lines are grey.
	reGreyToolOutputPrefix = regexp.MustCompile("\x1b\\[0;90m\\[[0-9]{2}:[0-9]{2}:[0-9]{2}\\] \\[Tool Output\\] ")
	reYellowAction         = regexp.MustCompile("\x1b\\[0;33m\\[[0-9]{2}:[0-9]{2}:[0-9]{2}\\] \\[Tool Action\\] ")
	reOldEstimateCur       = regexp.MustCompile(`Payload: ~[0-9]+/[0-9]+ tokens`)
)

// thenEstimatedNoAllowance: the estimated line no longer carries the `/budget`
// allowance (it now shows a signed increment before the estimated size).
func thenEstimatedNoAllowance(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !hasEstimatedPayloadStatus(sc.stderr) {
		return fmt.Errorf("standard error carried no estimated payload status line; stderr=%q", sc.stderr)
	}
	if reOldEstimateCur.MatchString(sc.stderr) {
		return fmt.Errorf("the estimated payload line must not show the /budget allowance; stderr=%q", sc.stderr)
	}
	return nil
}

// thenFirstEstimateZero: the FIRST pre-flight estimate of the run shows a zero
// increment (no predecessor in this process).
func thenFirstEstimateZero(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	d, ok := estimatedPayloadDelta(sc.stderr)
	if !ok {
		return fmt.Errorf("standard error carried no estimated payload status line; stderr=%q", sc.stderr)
	}
	if d != 0 {
		return fmt.Errorf("the first estimated payload increment = %d, want 0; stderr=%q", d, sc.stderr)
	}
	return nil
}

// thenLaterEstimatePositive: at least one estimate line after the first shows a
// positive increment (a later request in the same run grew the payload).
func thenLaterEstimatePositive(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	for _, m := range reEstimatedPayload.FindAllStringSubmatch(sc.stderr, -1) {
		if strings.HasPrefix(m[1], "+") && m[1] != "+0" {
			return nil
		}
	}
	return fmt.Errorf("no estimated payload line showed a positive increase; stderr=%q", sc.stderr)
}

// thenToolOutputGrey: EVERY line of the `[Tool Output]` block — the header,
// each streamed CONTENT line, and BOTH horizontal separators — is wrapped grey on
// the colour-enabled terminal (round 057 greyed the frame; round 058 (ADR 0028)
// extends it to the content lines, so the whole block reads as one grey region).
func thenToolOutputGrey(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !reGreyHeader.MatchString(sc.stderr) {
		return fmt.Errorf("the tool output header is not grey; stderr=%q", sc.stderr)
	}
	if n := len(reGreySeparator.FindAllString(sc.stderr, -1)); n < 2 {
		return fmt.Errorf("expected the opening and closing separators grey; found %d grey separator(s); stderr=%q", n, sc.stderr)
	}
	// The header shares the `[Tool Output]` prefix with a content line, so count
	// every grey-wrapped `[Tool Output] …` line: header (1) + at least one content
	// line ⇒ ≥ 2 total.
	if n := len(reGreyToolOutputPrefix.FindAllString(sc.stderr, -1)); n < 2 {
		return fmt.Errorf("expected the header AND at least one content line grey (>=2 grey `[Tool Output]` lines); found %d; stderr=%q", n, sc.stderr)
	}
	return nil
}

// thenActionYellow: the whole `[Tool Action]` line is wrapped yellow.
func thenActionYellow(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !reYellowAction.MatchString(sc.stderr) {
		return fmt.Errorf("the action line is not yellow; stderr=%q", sc.stderr)
	}
	return nil
}

// reTurnLogEstimate matches the round-057 estimated payload line as persisted in
// `turns.log` (no colour escapes — the file leg is plain by construction).
var reTurnLogEstimate = regexp.MustCompile(`Payload: [+-][0-9]+ ~[0-9]+ tokens`)
var reOldTurnLogEstimate = regexp.MustCompile(`Payload: ~[0-9]+/[0-9]+ tokens`)

// thenTurnLogPayloadParity (F-057-2, round-057 review): the saved turn log is
// CONTENT-EQUAL to the terminal chrome (Q3 → 1) but plain — so it must carry the
// new estimated shape (`Payload: +<delta> ~<n> tokens`) and must NOT carry the
// retired `Payload: ~<n>/<budget>` allowance form.
func thenTurnLogPayloadParity(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	data, err := os.ReadFile(filepath.Join(sc.historyDir(), "turns.log"))
	if err != nil {
		return fmt.Errorf("the saved turn log must exist: %w", err)
	}
	got := string(data)
	if !reTurnLogEstimate.MatchString(got) {
		return fmt.Errorf("the saved turn log must carry the estimated payload with its increment; got %q", got)
	}
	if reOldTurnLogEstimate.MatchString(got) {
		return fmt.Errorf("the saved turn log must not carry the retired allowance form; got %q", got)
	}
	if strings.ContainsRune(got, '\x1b') {
		return fmt.Errorf("the saved turn log must carry no colour; got %q", got)
	}
	return nil
}
