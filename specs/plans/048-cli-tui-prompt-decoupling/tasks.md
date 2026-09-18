# Tasks: de-couple `internal/cli` from the TUI prompt (round 048)

**Plan Package**: `specs/plans/048-cli-tui-prompt-decoupling`
**Anchor**: [#101](https://github.com/gosharplite/tellme/issues/101) — R5.2 of [#92](https://github.com/gosharplite/tellme/issues/92) (the first de-coupling slice)
**Created**: 2026-09-18

> `/axb-implement` execution control plane. Read `spec.md` + `research.md` + `plan.md` + `truth-delta.md` before starting. This is a **non-BDD structural round** (no truth feature/DSL row changes), so Phase 3 carries `[UNIT]` pins and a review gate; there are **no** `[BDD-GREEN]`/`[BDD-REFACTOR]` feature phases — the round-042/043/044/045/046/047 non-BDD precedent.

## Core Inputs

- `spec.md` (Locked **Q1 → B** / **Q2 → A** / **Q3 → A**; US1–US3; FR-001…FR-012)
- `research.md` (D1–D10)
- `plan.md` (0 interfaces; `/axb-api-plan` + `/axb-data-plan` + `/axb-dsl-refine` NOOP)
- `truth-delta.md` (techstack MODIFY ×3 + the Task-runner NOOP; ADR 0017)
- `specs/truth/techstack.md` — the **Layer-discipline gate** row (RULE-E **3 → 2**) + the **Composition root** row + the **Interactive TUI prompt** row
- `docs/decisions/0017-cli-tui-prompt-decoupling.md` (the port + the pattern) · `0016` (RULE-E) · `0011` (tier table/D1, ratchet/D3) · `0013` (the injected seam) · `0015` (the port-in-domain precedent)
- `tools/arch/baseline.txt` (must regenerate to **2** lines — R5.2's DoD)
- `internal/ui/tui/prompt/{run.go,model.go}` (`Run`, `Source`, `DefaultDebounceDuration` — the seam the port mirrors)

## Setup

_(omitted — stdlib-only; no new technology; `go.mod`/`go.sum` unchanged)_

## Phase 2 — Foundational

- [X] **T001** — Create `internal/domain/tui/prompter.go`: the `Source` interface (`Suggest(ctx, query) []string`) + the `Prompter` interface (`Run(ctx, in io.Reader, out io.Writer, src Source, debounce time.Duration) (string, bool, error)` + `DefaultDebounceDuration() time.Duration`) + the route/rationale doc comment. **只做**：the port declaration (imports only `context`/`io`/`time`). **不做**：no `internal/ui` types, no callers.
- [X] **T002** — Land the `internal/ui` adapter: `internal/ui/tuiprompt.go` — `type TUIPrompter struct{}` satisfying `domaintui.Prompter` by delegating to `tuiprompt.Run(...)` + returning `tuiprompt.DefaultDebounceDuration`. Land an empty `internal/ui/tuiprompt_test.go`. **只做**：the adapter + landing file. **不做**：no CLI change yet.

## Phase 3 — Test Alignment & Implementation (unit-only)

> No DSL rows: the round changes no feature step. The "alignment" is the `internal/cli` TUI test seam (a fake `Prompter`); the "RED" is the compile failure (`Options.Prompter` does not exist yet, the adapter does not exist yet).

- [X] **T003 `[UNIT-RED]`** — Re-point the CLI TUI tests (`internal/cli/tui_dispatch_test.go`, `tui_submit_chrome_test.go`) to inject an in-package **fake `domaintui.Prompter`** (canned `(text, ok, err)`) instead of `Options.RunTUIPrompt: func…`; assert the **same behaviour** (the dispatch reaches the seam; `stdout` stays empty; the submit path). Update the `testdeps_test.go` comment. **RED**: `Options.Prompter` does not exist and `cli.go` still imports `tuiprompt`.
  - **DSL 參照**: n/a (no DSL row). **Boundary**: `internal/cli/*_test.go` only.
- [X] **T004 `[UNIT-RED]`** — `internal/ui/tuiprompt_test.go`: pin the adapter contract — `TUIPrompter{}` satisfies `domaintui.Prompter`; `DefaultDebounceDuration()` == `tuiprompt.DefaultDebounceDuration`; `Run(ctx, in, out, fakeSource, 0)` drives the prompt (a canned submit returns the composed text + `ok=true`, abort returns `ok=false`, and `stdout` is untouched). **RED**: `internal/ui.TUIPrompter` does not exist.
  - **Boundary**: `internal/ui/tuiprompt_test.go` only.
- [X] **T005** — Review gate: the port + the fake are non-vacuous; the CLI fake asserts the dispatch contract, and the adapter pin lives at the `ui` tier.

## Phase 4 — Green / Refactor

- [X] **T006 `[GREEN]`** — `internal/cli/cli.go`: drop `import "…/internal/ui/tui/prompt"`; make `cli.Options` = `{ Deps deps.Dependencies; Prompter domaintui.Prompter }`; **delete** the `tuiPromptRunner` func type + the `RunTUIPrompt` field; route `defaultRunTUIPrompt`'s `Run`/debounce through `opts.Prompter` (keep the diagnostic hint + the suggestion-engine wiring in cli); a nil `Prompter` is a clear error (unreachable in production). **做**：the re-wiring. **不做**：no behaviour change (chrome, streams, exit codes).
  - **Boundary**: `internal/cli/cli.go` only.
- [X] **T007 `[GREEN]`** — `cmd/tellme/deps.go` `buildOptions()`: wire `cli.Options{Deps: buildDeps(), Prompter: ui.TUIPrompter{}}` (the composition root injects the adapter, ADR 0013).
  - **Boundary**: `cmd/tellme/deps.go` only.
- [X] **T008 `[CODE-REMOVE]`** — Delete the dead `tuiPromptRunner` type, the `Options.RunTUIPrompt` field, and the in-package nil-default fallback (the `internal/cli → internal/ui/tui/prompt` import is now gone — F-4 closed); sweep the stale doc comments (`cli.go:61`, the seam comments, `testdeps_test.go`).
- [X] **T009 `[REFACTOR]`** — Sweep the in-code citations: the `defaultRunTUIPrompt` doc names the injected port; the `tuiDebounceDuration` fallback goes through `Prompter.DefaultDebounceDuration()`; `internal/ui/tuiprompt.go` states the port home + the RULE-A wrapping rationale.

## Phase 5 — Baseline, regression, witnesses, close-out

- [X] **T010 `[CODE-REMOVE / GATE]`** — Regenerate `tools/arch/baseline.txt` via `make verify-architecture-update`: the `internal/cli -> internal/ui/tui/prompt` line is **gone** (3 → 2); the gate ships with its enabler (round-040 TD-1). **只做**：the baseline regeneration.
- [X] **T011 `[REGRESSION]`** — `make verify` (RULE-E **0 new / 0 stale**, baseline **2**; RULE-A/B/C 0; 0 cycles; lint 0; govulncheck clean; cross-compile 4/4) · `go test -count=1 ./...` green (incl. the godog E2E) · topology audit unchanged · `gofmt`/`go vet` clean. **Falsifiability witnesses** (a) re-introduced edge ⇒ RULE-E red; (b) stale baseline line ⇒ red; (c) missing/false port injection ⇒ the TUI dispatch unit seam (or build) fails loudly — reproduced then reverted (ADR 0010). **Boundary**: verification only.
- [X] **T012** — Land/confirm truth + governance: `specs/truth/techstack.md` (the three MODIFY rows), confirm `docs/decisions/0017-cli-tui-prompt-decoupling.md` + the index row, mark `tasks.md` `[X]`, update `STATUS.md` + the daily summary, and open the PR. Record the remaining edges + F-6/F-7/F-8 on [#101](https://github.com/gosharplite/tellme/issues/101) (F-4 closed).

## Pre-Delivery Orphan Coverage Sweep

| Artifact | Carrier |
| --- | --- |
| `truth-delta` non-NOOP: techstack **Layer-discipline gate** row (RULE-E 2) | T010 + T011 + T012 |
| `truth-delta` non-NOOP: techstack **Composition root** row | T006 + T007 + T012 |
| `truth-delta` non-NOOP: techstack **Interactive TUI prompt** row | T006 + T009 + T012 |
| `research.md` D1–D10 | T001 (D2/D3), T002 (D4), T006 (D1/D5), T007 (D5), T008 (D5/D6), T010 (D1), T011 (D9/D10) |
| `truth-delta` governance: ADR 0017 + index | T012 (recorded — already landed in the plan half) |
| **F-4 closure** (FR-004) | T006 + T008 |
| Baseline reaches **2** (FR-005) | T010 + T011 (gate) |

## Notes

- The round is **behaviour-preserving** — the TUI chrome, `stdout`/`stderr` contracts, flags, and exit codes are unchanged; the E2E suite is **regression**, not the carrier (NFR-004).
- The **port home** is `internal/domain/tui` (Q2-A); the **adapter** is in `internal/ui` (RULE-A: only tiers ≥ 5 may import `internal/ui/**`); the **wiring** is `cmd/tellme` (exempt).
- **No new Makefile target** — the gate rides `verify-architecture` (already a member of `make verify`).

## Outcome (implementation delivered)

- **Product**: `internal/domain/tui/prompter.go` (NEW — the `Source` + `Prompter` port; stdlib-only) · `internal/ui/tuiprompt.go` (NEW — the `TUIPrompter` adapter delegating to `prompt.Run`) · `internal/ui/tuiprompt_test.go` (NEW — the adapter contract pin) · `internal/cli/cli.go` (CHANGED — `Options{Deps; Prompter}`; the `tuiPromptRunner` type + `RunTUIPrompt` field deleted (**F-4 closed**); `runTUIPrompt` routes through the injected port; the nil-default removed) · `cmd/tellme/deps.go` (CHANGED — wires `ui.TUIPrompter{}`) · `internal/cli/{testdeps,tui_dispatch,tui_submit_chrome}_test.go` (CHANGED — the fake port double) · `tools/arch/baseline.txt` (CHANGED — **3 → 2**).
- **Gate**: `make verify` **OK** — `verify-architecture` reports RULE-E **0 new / 0 stale** with the baseline at **2** (`internal/cli -> internal/agent`, `internal/cli -> internal/ui`); RULE-A/B/C **0**; **0** cycles; lint 0 issues; govulncheck clean (0 reachable in code); cross-compile **4/4**.
- **Tests**: `go test -count=1 ./...` green (incl. the godog E2E, ~60 s).
- **Falsifiability witnesses (reproduced then reverted, ADR 0010)**: (a) a re-introduced `internal/cli → internal/ui/tui/prompt` import ⇒ the gate reds and names the edge; (b) a stale `baseline.txt` line ⇒ the gate reds (*"1 stale baseline entr(ies)"*); (c) the nil-port path ⇒ the new `TestTUIDispatchFailsLoudlyWithoutPrompter` pin (EnvironmentError), pinned permanently.
- **No behaviour change**: `stdout`/`stderr` contracts, flags, exit codes, and the TUI chrome are unchanged; `go.mod`/`go.sum` unchanged; the Gherkin/DSL topology audit unchanged.
