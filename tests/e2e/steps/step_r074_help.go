package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// Round 074 (ADR 0046) — the `-h`/`--help` flag list on stdout, ending in
// success (empty stderr, exit code 0).
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^tellme prints its flag list$`, thenPrintsFlagList)
		ctx.Then(`^the help is reported as a success$`, thenHelpReportedAsSuccess)
	})
}

// thenPrintsFlagList (必查 呈現結果): stdout carries the flag list, including the
// new `-h`/`--help` and `-v`/`--version` flags plus the pre-existing surface.
func thenPrintsFlagList(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if strings.TrimSpace(sc.stdout) == "" {
		return fmt.Errorf("the help must print a flag list; stdout is empty (stderr=%q)", sc.stderr)
	}
	for _, want := range []string{"-h, --help", "-v, --version", "-c, --config", "-d, --diagnostics", "-l, --list", "-r, --raw"} {
		if !strings.Contains(sc.stdout, want) {
			return fmt.Errorf("the flag list is missing %q; stdout=%q", want, sc.stdout)
		}
	}
	return nil
}

// thenHelpReportedAsSuccess (必查 呈現結果): the exit code is 0 and stderr is
// empty — help is a success, not a usage error (no `tellme: …` phrase).
func thenHelpReportedAsSuccess(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if sc.exitCode != 0 {
		return fmt.Errorf("help must exit 0 (a success); got exit code %d (stderr=%q)", sc.exitCode, sc.stderr)
	}
	if sc.stderr != "" {
		return fmt.Errorf("help must carry nothing on stderr; got %q", sc.stderr)
	}
	return nil
}
