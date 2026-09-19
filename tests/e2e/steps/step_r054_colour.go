package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 054 (ADR 0023) — the `-l` count default and the green chrome accents.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint asks tellme to read "([^"]*)" with the reason "([^"]*)" and then answers with "([^"]*)" and reports the token usage:$`, givenProviderReadReasonUsage)
		ctx.When(`^the operator asks tellme to list the last messages without a count$`, whenListLastNoCount)
		ctx.Then(`^the session chrome accents the tool reason, the mode, the measured tokens, and the session cost in green$`, thenChromeGreenAccents)
		ctx.Then(`^the session chrome carries no colour$`, thenChromeNoColour)
		ctx.Then(`^the session turn log carries no decoration$`, thenTurnLogNoDecoration)
	})
}

// thenTurnLogNoDecoration (必查 權威狀態) — round 054 fold B-54-1: the persisted
// session turn log stays control-free even when the stderr chrome was coloured.
func thenTurnLogNoDecoration(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	data, err := os.ReadFile(filepath.Join(sc.historyDir(), "turns.log"))
	if err != nil {
		return fmt.Errorf("the session turn log must exist: %w", err)
	}
	if strings.ContainsRune(string(data), '\x1b') {
		return fmt.Errorf("the turn log must carry no control bytes; got %q", string(data))
	}
	return nil
}

// givenProviderReadReasonUsage scripts a usage-reporting fake that returns a
// `read_files` call carrying `reason` then the answer — so a turn renders a
// `[Tool Reason]` line AND measured/ready lines (the round-054 colour positive).
func givenProviderReadReasonUsage(ctx context.Context, provider, path, reason, answer string, table *godog.Table) error {
	sc := scenarioFrom(ctx)
	answer = unescapeText(answer)
	prompt, cached, completion, thinking, err := usageFromTable(table)
	if err != nil {
		return err
	}
	f := sc.newFake()
	f.ReportUsageDetails(prompt, cached, completion, thinking)
	f.Script(
		fakeprovider.Reply{ToolName: "read_files", Arguments: readArgsWithReason(path, reason)},
		fakeprovider.Reply{Answer: answer},
	)
	sc.scriptedAnswer = answer
	sc.scriptedAnswerSet = true
	sc.scriptedTool = "read_files"
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

// whenListLastNoCount runs `tellme -l` (no count, no -c, no positional prompt).
func whenListLastNoCount(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	sc.args = []string{"-l"}
	sc.run()
	return nil
}

// The reference's green SGR pair (tell-me-go colors.go). Assertions read the RAW
// stderr bytes; the colour is emitted only when the diagnostic stream is a
// terminal (TELL_ME_FORCE_STDERR_TTY) and `-r` is off.
var (
	reGreenReason     = regexp.MustCompile("\x1b\\[0;32m\\[[0-9]{2}:[0-9]{2}:[0-9]{2}\\] \\[Tool Reason\\] ")
	reGreenMode       = regexp.MustCompile(" tokens - \x1b\\[0;32m[^\x1b]+\x1b\\[0m - ")
	reGreenMeasTokens = regexp.MustCompile("Payload: \x1b\\[0;32m[0-9]+\x1b\\[0m/[0-9]+ tokens")
	reGreenSessionCst = regexp.MustCompile("\x1b\\[0;32m\\$[0-9.]+\\x1b\\[0m - M: ")
)

// thenChromeGreenAccents (必查 呈現結果): the captured stderr carries the four
// round-054 green accents.
func thenChromeGreenAccents(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	for name, re := range map[string]*regexp.Regexp{
		"tool reason":   reGreenReason,
		"payload mode":  reGreenMode,
		"measured toks": reGreenMeasTokens,
		"session cost":  reGreenSessionCst,
	} {
		if !re.MatchString(sc.stderr) {
			return &missingAccent{name: name, stderr: sc.stderr}
		}
	}
	return nil
}

type missingAccent struct {
	name   string
	stderr string
}

func (m *missingAccent) Error() string {
	return "the chrome is missing the green " + m.name + " accent; stderr=" + quoteForLog(m.stderr)
}

// thenChromeNoColour (必查 呈現結果): the captured stderr carries no chrome
// accent — neither the round-054 green nor the round-057 grey/yellow.
func thenChromeNoColour(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	for _, code := range []string{"\x1b[0;32m", "\x1b[0;90m", "\x1b[0;33m"} {
		if strings.Contains(sc.stderr, code) {
			return &unexpectedColour{stderr: sc.stderr}
		}
	}
	return nil
}

type unexpectedColour struct{ stderr string }

func (u *unexpectedColour) Error() string {
	return "the chrome must carry no colour off a terminal; stderr=" + quoteForLog(u.stderr)
}

// quoteForLog renders a bounded, escaped view of a captured stream for a failure
// message (a raw stream may be long and carry control bytes).
func quoteForLog(s string) string {
	const max = 600
	if len(s) > max {
		s = s[:max] + "…"
	}
	b, err := json.Marshal(s)
	if err != nil {
		return "<unprintable>"
	}
	return string(b)
}
