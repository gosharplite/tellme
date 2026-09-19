package steps

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

// T010 [BDD-RED] — Then: the progress spinner reports the machine's resource usage.
func init() {
	registrars = append(registrars, func(ctx *godog.ScenarioContext) {
		ctx.Then(`^the progress spinner reports the machine's resource usage$`, thenSpinnerReportsResources)
	})
}

// reT010Resource matches a tool-execution spinner line carrying the CPU/MEM
// segment (one decimal each).
var reT010Resource = regexp.MustCompile(`[⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏] Executing[^\r\n]*CPU: [0-9]+\.[0-9]%[^\r\n]*MEM: [0-9]+\.[0-9]%`)

// reT010Mem captures a MEM percentage value so the round-059 strengthening can
// require a real (non-zero) figure.
var reT010Mem = regexp.MustCompile(`MEM: ([0-9]+)\.([0-9])%`)

// thenSpinnerReportsResources (必查 / 呈現結果): the captured stderr carries a
// tool-execution spinner line whose status also reports the machine's CPU and
// memory percentages (one decimal each), and — round 059 — the memory figure is
// a REAL non-zero value (the round-019 macOS arm reported 0.0% for every host).
// 不該發生: the resource segment must not appear on the model-phase spinner.
func thenSpinnerReportsResources(ctx context.Context) error {
	sc := scenarioFrom(ctx)
	if !reT010Resource.MatchString(sc.stderr) {
		return fmt.Errorf("the tool-execution spinner carries no CPU/MEM segment: stderr=%q", sc.stderr)
	}
	nonZeroMem := false
	for _, m := range reT010Mem.FindAllStringSubmatch(sc.stderr, -1) {
		if m[1] != "0" || m[2] != "0" {
			nonZeroMem = true
			break
		}
	}
	if !nonZeroMem {
		return fmt.Errorf("the resource segment's MEM figure is 0.0%% on every frame (expected a real value): stderr=%q", sc.stderr)
	}
	for _, seg := range strings.FieldsFunc(sc.stderr, func(r rune) bool { return r == '\r' || r == '\n' }) {
		if strings.Contains(seg, "Thinking") && strings.Contains(seg, "CPU:") {
			return fmt.Errorf("the model-phase spinner carried the resource segment: %q", seg)
		}
	}
	return nil
}
