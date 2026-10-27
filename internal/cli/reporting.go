package cli

import (
	"fmt"
	"io"

	"github.com/gosharplite/tellme/internal/app/deps"
	"github.com/gosharplite/tellme/internal/domain/history"
	"github.com/gosharplite/tellme/internal/domain/render"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// dispatchReporting handles the offline reporting commands in precedence order
// — `-d` → `-l` → `-b` (standalone) → `-t` → `--tool-usage` — before any prompt
// or stdin access. It returns the exit code and whether a reporting command
// handled the run. (`--version` is handled by the caller, ahead of the reporting
// batch.) Extracted from `run` so its cyclomatic complexity stays under the
// cyclop gate (round 026). Round 081 (ADR 0053): `-l` composes with `-b` (list
// then roll back) and the standalone rollback is an offline action; hasPrompt
// distinguishes the standalone rollback from the prompt-bearing form (handled by
// the prompt path).
func dispatchReporting(f *flags, target sessionRef, env runtimeEnv, hasPrompt bool, in cliDeps) (int, bool) {
	// -d is the reporting path: it always produces a report, and it takes
	// precedence over a prompt or piped input (round-004 Decision 7).
	if f.diagnostic {
		return renderDiagnostic(target, env.stdout), true
	}
	// Round 081 (ADR 0053): `-b`/`--back` validation, beside its action and below
	// the `-d` tier (symmetric with `-l`). A non-positive count is a usage error
	// (symmetric with `-l`), and a rollback combined with `--new` is refused (the
	// two session actions are contradictory).
	if f.backSet && (f.back <= 0 || f.newSession) {
		return emitUsageError(env.stderr), true
	}
	// -l lists the last N messages and exits, strictly offline (round-007). A
	// non-positive N is a usage error, evaluated before any network or stdin.
	if f.listSet {
		if f.list <= 0 {
			return emitUsageError(env.stderr), true
		}
		if code := renderHistoryList(target, f.list, f.raw, env, in); code != Success {
			return code, true
		}
		// Round 081 (ADR 0053): compose with `-b` (list then roll back), matching
		// the reference's order. With `-b` absent, `-l` is terminal (unchanged).
		if !f.backSet {
			return Success, true
		}
	}
	// Round 081 (ADR 0053): the standalone rollback (no positional prompt) is an
	// offline action, ordered after `-l` and before `-t`. A prompt-bearing `-b` is
	// handled by the prompt path (the rollback runs just before the turn).
	if f.backSet && !hasPrompt {
		return renderRollback(target, env, f.back, in), true
	}
	// -t prints the resolved session's turn log and exits, strictly offline
	// (round 053; ADR 0022). Ordered after -l (which keeps its round-007
	// precedence).
	if f.turns {
		return renderTurnsLog(target, env, in), true
	}
	// --tool-usage is the offline tool-usage report: like --version it needs no
	// configuration, no TELL_ME_HOME, and no workspace (round-026 FR-006).
	if f.toolUsage {
		return renderToolUsage(env, in), true
	}
	return 0, false
}

// renderToolUsage prints the offline per-tool roll-up (round-026 FR-006..FR-009):
// it enumerates the LIVE registry, streams the user-global tool-usage log into
// per-tool counts, and writes plain text to stdout. It is `--version`-class — it
// requires neither a configuration, nor TELL_ME_HOME, nor a session workspace.
//
// A GENUINE read failure of the log is surfaced as a one-line diagnostic on the
// diagnostic stream (the report path is offline, so stderr is free), so an
// unreadable log is distinguishable from "no tool ever used"; the all-zero report
// still prints and the command succeeds.
func renderToolUsage(env runtimeEnv, in cliDeps) int {
	// The offline report never executes a tool; it lists every tool that can be
	// RECORDED — the UNION of the base set and the capability-gated set (round
	// 062; PR #129 fold F-062-1) — so a capability-gated tool (`read_image`) is
	// shown (with zero) rather than invisible while its records accumulate. One
	// authority for "a tool that can be recorded": the same two registries the
	// prompt path builds. The provider TYPE argument (round 063) is unused here
	// (the report is capability-blind by design — ADR 0032 RF-062-13), so it is
	// passed empty for both variants.
	names := unionToolNames(in.dp.NewToolRegistry(deps.ToolSetSpec{}), in.dp.NewToolRegistry(deps.ToolSetSpec{Vision: true}))
	counts, err := in.dp.NewToolUsageStore(in.dp.UserHomeDir).Aggregate()
	if err != nil {
		_, _ = fmt.Fprintf(env.stderr, "[tool-usage] could not read the usage log: %v\n", err)
		counts = nil
	}
	rows := make([]history.ToolUsageRow, 0, len(names))
	for _, n := range names {
		rows = append(rows, history.ToolUsageRow{Tool: n, Counts: counts[n]})
	}
	_, _ = fmt.Fprint(env.stdout, in.dp.NewLines(false).ToolUsage(rows))
	return Success
}

// unionToolNames enumerates the tool names across the given registries, in
// first-seen order, de-duplicated by name (round 062; PR #129 fold F-062-1).
func unionToolNames(regs ...domaintools.Registry) []string {
	seen := make(map[string]bool)
	var names []string
	for _, reg := range regs {
		if reg == nil {
			continue
		}
		for _, t := range reg.Tools() {
			if seen[t.Name()] {
				continue
			}
			seen[t.Name()] = true
			names = append(names, t.Name())
		}
	}
	return names
}

// renderHistoryList lists the last N persisted messages (round-007 FR-007..FR-009)
// and exits — strictly offline, no provider request. It resolves only the
// workspace (no configuration/provider requirement), so listing works even when
// the configuration is absent. Round 053 (ADR 0022): the session is selected by
// the `-c` configuration's MODE (when the env override is unset).
func renderHistoryList(target sessionRef, n int, raw bool, env runtimeEnv, in cliDeps) int {
	res, rerr := resolveWorkspace(target)
	if rerr != nil {
		return emitBootError(env.stderr, res, rerr)
	}
	entries, err := in.dp.NewHistoryStore(res.Workspace).Load()
	if err != nil {
		return emitHistoryError(env.stderr, err)
	}
	// Round 073 (ADR 0045): the listing is presented by the injected render.Listing
	// port (the bytes stay single-owned in internal/ui). The count/selection
	// semantics are unchanged (the last N MESSAGES). The header accent is gated on
	// the STDOUT probe (not the stderr one) and `-r`; the width is best-effort.
	in.dp.NewListing().Render(env.stdout, listingMessages(entries, n), render.ListingSpec{
		Colour: env.stdoutIsTerminal(),
		Raw:    raw,
		Width:  res.WrapWidth, // resolved once in resolveWorkspace (fold RF-8)
		Warn:   env.stderr,
	})
	return Success
}

// listingMessages projects the persisted exchanges into the listing's message
// sequence (round 073): the operator prompt then the model answer per exchange,
// truncated to the last n messages. It is the successor of the former toMessages
// — the widened tool activity's CONTENT is never projected (FR-017 / clarify
// Q2 -> A). Round 086 (ADR 0057): the model message also carries the turn's
// TOOL-call count (len(e.Steps)) — the `[TOOLS] - M (N calls)` line the adapter
// prints. The count is the persisted turn's executed tool steps, NOT e.Calls
// (the AI-endpoint inference-round count), and only the count crosses the seam.
func listingMessages(entries []history.Entry, n int) []render.ListingMessage {
	msgs := make([]render.ListingMessage, 0, len(entries)*2)
	for i, e := range entries {
		// Round 082 (ADR 0054): the backward turn index — the distance of this
		// entry's turn from the MOST RECENT one (1 = newest). It is the TRUE
		// distance from the end of the loaded history, computed BEFORE the
		// message-count truncation below, so a lone leading message of an odd
		// `-l N` keeps its true distance (never renumbered by its position in the
		// printed slice). Both messages of the entry carry the SAME index.
		turnIndex := len(entries) - i
		msgs = append(msgs,
			render.ListingMessage{Role: render.ListingOperator, Body: e.Prompt, TurnIndex: turnIndex},
			render.ListingMessage{Role: render.ListingModel, Body: e.Answer, TurnIndex: turnIndex, ToolCount: len(e.Steps)})
	}
	if len(msgs) > n {
		msgs = msgs[len(msgs)-n:]
	}
	return msgs
}

// renderTurnsLog prints the resolved session's turn log to stdout and exits —
// strictly offline (round 053; ADR 0022). A missing/empty log is success (it
// mirrors `-l` on an empty session). The session is selected by the `-c`
// configuration's MODE (when the env override is unset).
func renderTurnsLog(target sessionRef, env runtimeEnv, in cliDeps) int {
	res, rerr := resolveWorkspace(target)
	if rerr != nil {
		return emitBootError(env.stderr, res, rerr)
	}
	// Round-053 fold RF-53-2: stream the log to stdout (never materialise it), a
	// missing file reading as empty (the store's Reader tolerance). A read
	// failure reuses the environment class phrase via emitHistoryError (the
	// frozen vocabulary), though the wording is history-flavoured for a
	// turn-log file.
	rc, err := in.dp.NewTurnsLogStore(res.Workspace).Reader()
	if err != nil {
		return emitHistoryError(env.stderr, err)
	}
	defer func() { _ = rc.Close() }()
	_, _ = io.Copy(env.stdout, rc)
	return Success
}

// renderNewSession starts a fresh session without a prompt: it archives the
// active history (retaining it) and returns success (round-007 FR-005/FR-006).
// Round 053 (ADR 0022): the session is selected by the `-c` configuration's MODE
// (when the env override is unset), and the turn log is archived alongside the
// history and usage logs.
func renderNewSession(target sessionRef, env runtimeEnv, in cliDeps) int {
	res, rerr := resolveWorkspace(target)
	if rerr != nil {
		return emitBootError(env.stderr, res, rerr)
	}
	if err := in.dp.NewHistoryStore(res.Workspace).Archive(); err != nil {
		return emitHistoryError(env.stderr, err)
	}
	if err := in.dp.NewUsageStore(res.Workspace).Archive(); err != nil {
		return emitHistoryError(env.stderr, err)
	}
	// Best-effort: the turn log is a trace, not session state.
	_ = in.dp.NewTurnsLogStore(res.Workspace).Archive()
	return Success
}

// renderRollback rolls back the last n complete turns of the resolved session
// and prints a plain confirmation to stdout (round 081 / ADR 0053). It is an
// OFFLINE action: it resolves only the runtime home + effective mode + session
// workspace (no provider), reusing the same selection as `-l`/`-t`. A resolve
// failure reuses the boot error surface; a store failure reuses the environment
// class phrase (exit 4). A standalone rollback never reads stdin.
func renderRollback(target sessionRef, env runtimeEnv, n int, in cliDeps) int {
	res, rerr := resolveWorkspace(target)
	if rerr != nil {
		return emitBootError(env.stderr, res, rerr)
	}
	store := in.dp.NewHistoryStore(res.Workspace)
	removed, err := store.Rollback(n)
	if err != nil {
		return emitHistoryError(env.stderr, err)
	}
	remaining, err := store.Load()
	if err != nil {
		return emitHistoryError(env.stderr, err)
	}
	_, _ = fmt.Fprint(env.stdout, rollbackConfirmation(removed, len(remaining)))
	return Success
}

// rollbackConfirmation renders the rollback result line (round 081 / ADR 0053):
// plain text on stdout, no `tellme: ` class prefix. The effect (the file bytes)
// is the contract; the wording is not frozen.
func rollbackConfirmation(removed, remaining int) string {
	return fmt.Sprintf("Rolled back %d turn%s. History now holds %d turn%s.\n",
		removed, pluralS(removed), remaining, pluralS(remaining))
}

// pluralS returns "" for 1 and "s" otherwise.
func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
