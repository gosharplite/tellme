package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/gosharplite/tellme/tests/e2e/fakeprovider"
)

// Round 071 (ADR 0043) — the `search_files` tool E2E steps. Keep them in ONE file
// (the Zero-Shared-Edits split exists for parallel dispatch; these sentences are
// independent regardless).

func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Given(`^the working directory contains a file "([^"]*)" holding (\d+) lines containing "([^"]*)"$`, givenWorkdirMatchingLines)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint searches the working directory for "([^"]*)" and then answers with "([^"]*)"$`, givenSearchesWorkdir)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint searches the working directory with a small result budget and then answers with "([^"]*)"$`, givenSearchesSmallBudget)
		ctx.Given(`^a configured provider "([^"]*)" whose endpoint searches the working directory with an invalid pattern and then answers with "([^"]*)"$`, givenSearchesInvalidPattern)

		ctx.Then(`^tellme searched the working directory using its search_files tool$`, thenSearched)
		ctx.Then(`^the search result lists a match in "([^"]*)"$`, thenSearchMatchIn)
		ctx.Then(`^the search result reported no matches$`, thenSearchNoMatches)
		ctx.Then(`^the search result was trimmed to what the run can hold$`, thenSearchTrimmed)
		ctx.Then(`^the search result reported an invalid pattern$`, thenSearchInvalidPattern)
		ctx.Then(`^the search result lists the matches in path order$`, thenSearchPathOrder)
	})
}

// searchArgs marshals search_files arguments.
func searchArgs(m map[string]any) string {
	b, _ := json.Marshal(m)
	return string(b)
}

// scriptSearchThenAnswer scripts the fake to request one search_files call, then
// answer.
func scriptSearchThenAnswer(sc *scenarioContext, provider, arguments, answer string) error {
	f := sc.newFake()
	f.Script(
		fakeprovider.Reply{ToolName: "search_files", Arguments: arguments},
		fakeprovider.Reply{Answer: unescapeText(answer)},
	)
	sc.scriptedAnswer = unescapeText(answer)
	sc.scriptedAnswerSet = true
	sc.registerFake(provider, f)
	return sc.writeDefaultConfig(provider, map[string]string{provider: f.URL()})
}

// --- Givens ---

// givenWorkdirMatchingLines writes {name} with {count} lines each containing
// {text} — a fixture that overflows a small result budget.
func givenWorkdirMatchingLines(ctx context.Context, name string, count int, text string) error {
	sc := scenarioFrom(ctx)
	var sb strings.Builder
	for i := 0; i < count; i++ {
		fmt.Fprintf(&sb, "%s line %04d padding padding padding padding\n", text, i)
	}
	return os.WriteFile(filepath.Join(sc.workDir, filepath.FromSlash(name)), []byte(sb.String()), 0o644)
}

func givenSearchesWorkdir(ctx context.Context, provider, query, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptSearchThenAnswer(sc, provider, searchArgs(map[string]any{"path": ".", "query": query, "reason": "search the working directory"}), answer)
}

func givenSearchesSmallBudget(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptSearchThenAnswer(sc, provider, searchArgs(map[string]any{"path": ".", "query": "needle", "max_output_tokens": 100, "reason": "search"}), answer)
}

func givenSearchesInvalidPattern(ctx context.Context, provider, answer string) error {
	sc := scenarioFrom(ctx)
	return scriptSearchThenAnswer(sc, provider, searchArgs(map[string]any{"path": ".", "query": "(", "is_regex": true, "reason": "search"}), answer)
}

// --- Thens ---

func thenSearched(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !hasToolCall(f, "search_files", "") {
		return fmt.Errorf("the run did not make a search_files tool call")
	}
	if countToolIterations(f) == 0 {
		return fmt.Errorf("the search_files result was not fed back into the conversation")
	}
	return nil
}

func thenSearchMatchIn(ctx context.Context, path string) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	res := lastToolResult(f)
	if !strings.Contains(res, path+":") {
		return fmt.Errorf("the search result does not list a match in %q; result=%q", path, res)
	}
	return nil
}

func thenSearchNoMatches(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !strings.Contains(lastToolResult(f), "0 matches found") {
		return fmt.Errorf("the search result did not report no matches; result=%q", lastToolResult(f))
	}
	return nil
}

func thenSearchTrimmed(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !strings.Contains(lastToolResult(f), "(truncated)") {
		return fmt.Errorf("the search result was not trimmed; result=%q", lastToolResult(f))
	}
	return nil
}

func thenSearchInvalidPattern(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	if !strings.Contains(lastToolResult(f), "invalid regex") {
		return fmt.Errorf("the search result did not report an invalid pattern; result=%q", lastToolResult(f))
	}
	return nil
}

// thenSearchPathOrder (必查 權威狀態): the search result is sorted
// path-ascending. The assertion is GENERIC — it parses every `path:line: text`
// match line and requires the paths to be non-decreasing in file-path order (then
// line) — so the fixture can change without coupling the step to a file name. A
// fixture whose depth-first walk order differs from the sorted order (e.g.
// "a.go" vs "a/x.txt", where '.' < '/') makes the raw walk order fail this.
// This is the interface-truth carrier for the round-071 determinism invariant.
func thenSearchPathOrder(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	f := sc.onlyFake()
	if f == nil {
		return fmt.Errorf("no fake provider recorded a request")
	}
	res := lastToolResult(f)
	var prevPath string
	var prevLine int
	seen := 0
	for _, raw := range strings.Split(res, "\n") {
		line := strings.TrimRight(raw, "\r")
		// A match line is `<path>:<line>: <text>`; skip the truncation marker and
		// any non-matching line.
		colon := strings.Index(line, ":")
		if colon < 0 {
			continue
		}
		lineNoStr := line[colon+1:]
		sep := strings.Index(lineNoStr, ":")
		if sep < 0 {
			continue
		}
		lineNo, err := strconv.Atoi(lineNoStr[:sep])
		if err != nil {
			continue
		}
		path := line[:colon]
		if seen > 0 {
			if path < prevPath || (path == prevPath && lineNo < prevLine) {
				return fmt.Errorf("the search result is not in path order (%q:%d follows %q:%d); result=%q", path, lineNo, prevPath, prevLine, res)
			}
		}
		prevPath, prevLine = path, lineNo
		seen++
	}
	if seen < 2 {
		return fmt.Errorf("the search result carries fewer than two matches to order; result=%q", res)
	}
	return nil
}
