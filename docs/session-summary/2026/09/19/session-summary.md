# Session Summary — 2026-09-19

**Project**: `tellme` — a disciplined BDD re-creation of `tell-me-go`
**Repo**: `github.com/gosharplite/tellme`
**Status file**: [`STATUS.md`](../../../../../STATUS.md) *(back-link — the single live-state source)*
**Workspace**: `…/mbp-johndoe-niffler/ait-tellme` (`$TELL_ME_HOME`); darwin/arm64 host (Go 1.26.6).
**Session mode**: `butler`.
**Branch**: `053-offline-session-config-and-turns-flag` (off `dev`) → **PR [#118](https://github.com/gosharplite/tellme/pull/118)**.
**Status at end of day**: round **053** `053-offline-session-config-and-turns-flag` **DELIVERED (PR #118)** — the offline session commands honour `-c`; a new `-t` turn-log flag + a per-session `turns.log` writer; **closes [#103](https://github.com/gosharplite/tellme/issues/103)**; **ADR 0022**.

---

## 1. Session at a glance

| Area | Outcome |
| --- | --- |
| Bootstrap | `SESSION-BOOTSTRAP.md` Steps 1–8 (round 052 delivered/frozen; active branch `dev` → new round-053 branch) |
| Theme | [#103](https://github.com/gosharplite/tellme/issues/103) — `-l` ignored `-c` (silently reading the default session) + no `-t` |
| Clarify (one at a time) | **Q1 → (C1)** tellme writes its **own** `turns.log` (its rendered chrome); `-t` prints it · **Q2 → (A)** an **explicit** unreadable `-c` **fails** |
| Pipeline | specify ✅ · clarify ✅ (Q1/Q2) · spec-by-example ✅ · technical-research ✅ (**ADR 0022** + `techstack.md`) · system-analysis ✅ (1 CLI interface; api NOOP) · dsl-refine ✅ · tasks ✅ (T001–T015) · implement ✅ (all `[X]`) |
| Review (PR #118) | architectural review `5736721300` — **APPROVE WITH REQUIRED FOLDS** (F-53-1…F-53-5 + RF-53-x + nits) → **folded** |
| Verification | `gofmt` clean · `go vet` clean · `make verify` **OK** (arch gate 0 new/0 stale; lint 0; vulncheck clean; cross-compile 4/4) · `go test -count=1 ./...` **green** (incl. the godog E2E, **Strict**) · witnesses (a)/(b)/(c) reproduced + reverted |

## 2. Round 053 — what landed

1. **US1 (`-c` selects the session)** — `historyMode(homeDir, configPath)`: mode = `TELL_ME_MODE` → else the **`-c` config's `MODE`** → else the default → else `butler`. Mode-only read (`config.Load` pure parse) ⇒ stays **offline**. One seam `resolveWorkspace(homeDir, configPath)` covers `-l`, `-t`, prompt-less `--new`. **Q2 → (A)**: an explicit unreadable `-c` fails (existing phrase, exit 3).
2. **US2 (`-t` + the writer)** — `-t`/`--turns` prints `output/<mode>/turns.log` (order `-d` → `-l` → `-t` → `--tool-usage`); a new domain port `history.TurnsLogStore` + infra adapter; the prompt path tees the per-call renderer's chrome (turn rule/header, payload status, reasons, metrics, `Ready`) via `runtimeEnv.turnsLog`/`diag()`; best-effort; `--new` archives it (`turns.archive.log`).
3. **Truth/governance** — **ADR 0022** + index · `techstack.md` (CLI flags, Session lifecycle, **Turn log** row, Prompt-input precedence) · `data-model.dbml` `turns_log_line` · CLI truth (`inspecting-the-session-history.feature` + `reviewing-the-turn-log.feature` + `history/dsl.md`).

## 3. The fold (PR #118 review `5736721300`)

| # | Fold |
| --- | --- |
| **F-53-1** | The writer had **no permanent carrier** → added the CLI-tier pin `TestRunTurnTeesChromeIntoTurnsLog` (positive: chrome reaches the injected store; negative: empty workspace tees nothing). |
| **F-53-2** | Prompt-less `--new` now **refuses before archiving** on an explicit bad `-c` — the round-012 ordering is **superseded**: recorded in ADR 0022 §Consequences + the `cli.go` comment rewritten. |
| **F-53-3** | Chose **(i) narrow the sink**: `turns.log` records **the per-call renderer's chrome only**; the input-capture ack, the `-i` prompt echo, error phrases, the spinner, and the `[Tool …]` block stay stderr-only. Every surface (ADR D5, data Note, techstack row, acceptance) re-aligned. |
| **F-53-4** | Gherkin: added the `--new -c` interface Rule (`starting-a-fresh-session.feature`), the write-side turn-log Rule + split the non-atomic missing-file Rule (`reviewing-the-turn-log.feature`), an env-wins `-l` Example, and the `-t` precedence pin (`TestDispatchReportingPrecedence`); new `dsl.md` rows + stepdefs. |
| **F-53-5** | `techstack.md` **Prompt input** row aligned to the `-t` precedence + never-read-stdin list. |
| RF | **RF-53-2** done — the port is streaming (`Writer()`/`Reader()`, no whole-file materialisation). **RF-53-3** done — renamed `runtimeEnv.turnsLog`. RF-53-1/5/6/7 recorded in ADR 0022 §Forward. |
| Nits | `thenPrintsNothing` now asserts `sc.stdout == ""`; `-t` help wording; data-model `archived` token removed; STATUS header + daily-summary link fixed. |

## 4. Open items

- **PR [#118](https://github.com/gosharplite/tellme/pull/118)** awaits **human merge** into `dev`; then propagate `dev → main`, refresh the installed binary, close [#103](https://github.com/gosharplite/tellme/issues/103).
- **ADR 0022 forward items** RF-53-1…RF-53-7 (chrome-follow, archive-failure drift, `-l`-default-1 out of scope, self-diagnosing retrieve, the positional-factory seam, `-l`/`-t` create the workspace, a failed turn's chrome persisted).
- Carried: PR #16 **Obs 1**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; [#91](https://github.com/gosharplite/tellme/issues/91) / [#13](https://github.com/gosharplite/tellme/issues/13).

## 5. Next steps

1. Human merges PR [#118](https://github.com/gosharplite/tellme/pull/118) → propagate `dev → main` → close [#103](https://github.com/gosharplite/tellme/issues/103) → `SESSION-CLOSEOUT.md`.
2. Re-read `SESSION-BOOTSTRAP.md` next session.

## 6. PM follow-ups

- None new (spec/acceptance complete; no PM-owned gaps).
