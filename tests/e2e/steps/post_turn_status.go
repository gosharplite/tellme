package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"
)

// Round-018 post-turn status-line helpers. Kept in ONE file so the per-sentence
// step files stay independent (Zero Shared Edits), mirroring payload_status.go.
//
// The line shapes are pinned by specs/truth/features/cli/chat/dsl.md:
//
//	[HH:MM:SS] [<provider>] M: <miss> H: <cached> C: <completion> Th: <thinking>
//	╰─⠿ Ready ($<call> $<turn> $<session> - M: <m> H: <h> O: <o> - <pct>%)
var (
	reMetrics  = regexp.MustCompile(`\[[0-9]{2}:[0-9]{2}:[0-9]{2}\] \[([^\]]+)\] M: ([0-9]+) H: ([0-9]+) C: ([0-9]+) Th: ([0-9]+)`)
	reReady    = regexp.MustCompile(`╰─⠿ Ready \(\$([0-9.]+) \$([0-9.]+) \$([0-9.]+) - M: ([0-9]+) H: ([0-9]+) O: ([0-9]+) - ([0-9.]+)%\)`)
	rePostAnsi = regexp.MustCompile("\x1b\\[[0-9;]*m")
)

// stripPostTurnANSI removes ANSI SGR sequences so the merged cross-stream capture
// can be searched for plain text (the rendered answer interleaves colour codes).
func stripPostTurnANSI(s string) string { return rePostAnsi.ReplaceAllString(s, "") }

// hasMetricsLine reports whether s carries a round-018 metrics line.
func hasMetricsLine(s string) bool { return reMetrics.MatchString(s) }

// hasReadyLine reports whether s carries a round-018 `╰─⠿ Ready` summary line.
func hasReadyLine(s string) bool { return reReady.MatchString(s) }

// metricsValues returns the provider + M/H/C/Th of the first metrics line in s.
func metricsValues(s string) (provider string, miss, hit, completion, thinking int, ok bool) {
	sub := reMetrics.FindStringSubmatch(s)
	if sub == nil {
		return "", 0, 0, 0, 0, false
	}
	miss, _ = strconv.Atoi(sub[2])
	hit, _ = strconv.Atoi(sub[3])
	completion, _ = strconv.Atoi(sub[4])
	thinking, _ = strconv.Atoi(sub[5])
	return sub[1], miss, hit, completion, thinking, true
}

// readyValues returns the three costs, the session M/H/O, and the hit% of the
// first Ready line in s.
func readyValues(s string) (costs [3]float64, miss, hit, out int, pct float64, ok bool) {
	sub := reReady.FindStringSubmatch(s)
	if sub == nil {
		return costs, 0, 0, 0, 0, false
	}
	for i := 0; i < 3; i++ {
		costs[i], _ = strconv.ParseFloat(sub[1+i], 64)
	}
	miss, _ = strconv.Atoi(sub[4])
	hit, _ = strconv.Atoi(sub[5])
	out, _ = strconv.Atoi(sub[6])
	pct, _ = strconv.ParseFloat(sub[7], 64)
	return costs, miss, hit, out, pct, true
}

// usageFromTable reads the single data row of the
// `| prompt | cached | completion | thinking |` table.
func usageFromTable(t *godog.Table) (prompt, cached, completion, thinking int, err error) {
	if t == nil || len(t.Rows) < 2 {
		return 0, 0, 0, 0, fmt.Errorf("the token-usage table needs a header and one row")
	}
	row := t.Rows[1]
	get := func(i int) (int, error) {
		if i >= len(row.Cells) {
			return 0, fmt.Errorf("the token-usage row is missing column %d", i)
		}
		return strconv.Atoi(strings.TrimSpace(row.Cells[i].Value))
	}
	if prompt, err = get(0); err != nil {
		return
	}
	if cached, err = get(1); err != nil {
		return
	}
	if completion, err = get(2); err != nil {
		return
	}
	if thinking, err = get(3); err != nil {
		return
	}
	return
}

// activeModel returns the MODEL attribute of the default config's selected provider.
func activeModel(sc *scenarioContext) (string, error) {
	data, err := os.ReadFile(sc.homePath("configs/butler.yaml"))
	if err != nil {
		return "", fmt.Errorf("the default configuration must exist: %w", err)
	}
	var cfg map[string]any
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	sel, _ := cfg["SELECTED_PROVIDER"].(string)
	providers, _ := cfg["PROVIDERS"].(map[string]any)
	entry, _ := providers[sel].(map[string]any)
	m, _ := entry["MODEL"].(string)
	return m, nil
}

// addPricing writes a `MODELS` pricing entry for the given model into the default config.
func addPricing(sc *scenarioContext, model string, hit, miss, comp float64) error {
	p := sc.homePath("configs/butler.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("the default configuration must exist before pricing: %w", err)
	}
	var cfg map[string]any
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	models, _ := cfg["MODELS"].(map[string]any)
	if models == nil {
		models = map[string]any{}
	}
	models[model] = map[string]any{"PRICING": map[string]any{"HIT": hit, "MISS": miss, "COMP": comp}}
	cfg["MODELS"] = models
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(p, out, 0o644)
}

// writePriorUsage appends one prior call record to the per-mode usage log
// (the "the session's usage log already records a prior call" Given).
func writePriorUsage(sc *scenarioContext) error {
	rec := `{"timestamp":"2026-09-14T00:00:00Z","provider":"prior","model":"prior","cached_tokens":6,"prompt_tokens":10,"response_tokens":3,"total_tokens":15,"thinking_tokens":2,"cost":0.000100}`
	dir := filepath.Join(sc.home, "output", "butler")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "tokens.log"), []byte(rec+"\n"), 0o644)
}
