package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// T017 — Then: tellme explains on stderr that "{reason}"
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme explains on stderr that "([^"]*)"$`, thenExplainsOnStderr)
	})
}

// thenExplainsOnStderr (必查 呈現結果, grill round #6): exactly ONE stderr line
// begins with the frozen prefix `tellme: `, and that line starts with the
// frozen class phrase `tellme: {reason}` — the trailing detail (a path, a
// provider name, an OS error string) is contract-free (prefix, never equality).
//
// Each line is read as a terminal would: a round-019 spinner leaves the cursor at
// column 0 of an erased line via `\r`+erase-to-end-of-line and then writes the
// class phrase on it, so a physical line may be `\r<erase>…\r<erase>tellme: …`.
// Reducing each `\n`-delimited line to the text after its last carriage-return
// (with the erase control dropped) recognizes that phrase; a plain line (no `\r`)
// is unchanged, so every other failure assertion is unaffected.
func thenExplainsOnStderr(ctx context.Context, reason string) error {
	sc := scenarioFrom(ctx)

	var matched []string
	for _, line := range strings.Split(sc.stderr, "\n") {
		if i := strings.LastIndexByte(line, '\r'); i >= 0 {
			line = line[i+1:]
		}
		line = strings.ReplaceAll(line, "\x1b[K", "")
		if strings.HasPrefix(line, "tellme: ") {
			matched = append(matched, line)
		}
	}
	if len(matched) != 1 {
		return fmt.Errorf("want exactly one stderr line prefixed %q, got %d: %q", "tellme: ", len(matched), matched)
	}

	want := "tellme: " + reason
	if !strings.HasPrefix(matched[0], want) {
		return fmt.Errorf("stderr line %q does not start with the frozen class phrase %q", matched[0], want)
	}
	return nil
}
