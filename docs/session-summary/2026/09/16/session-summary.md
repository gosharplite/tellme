# Session Summary — 2026-09-16

**Project**: `tellme` — a disciplined BDD re-creation of `tell-me-go`
**Repo**: `github.com/gosharplite/tellme`
**Status file**: [`STATUS.md`](../../../../../STATUS.md) *(back-link — the single live-state source)*
**Workspace**: `…/mbp-johndoe-niffler/ait-tellme` (`$TELL_ME_HOME`)
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation)
**Branch**: `028-user-global-prompt-log` (off `dev`) → merged via PR [#59](https://github.com/gosharplite/tellme/pull/59) into `dev` (`9e13f9e`) → propagated `dev → main`.
**Status at end of day**: Round 028 (`028-user-global-prompt-log`) **DELIVERED / FROZEN** — the full AIxBDD pipeline, a plan+truth review loop, an implementation review, a **principal-architect review**, a human merge, propagation, and closeout. `tellme`'s `-i` prompt history is now **per-user** (`~/.tellme/global_prompts.jsonl`) with a first-use seed.

---

## 1. Session at a glance

| Area | Outcome |
| --- | --- |
| Bootstrap | `SESSION-BOOTSTRAP.md` Steps 1–8 at session start (round 027 delivered/frozen; active branch `dev`) |
| Round-028 theme | the `-i` shared prompt log moves from `<TELL_ME_HOME>/output/global_prompts.jsonl` to the **per-user** `~/.tellme/global_prompts.jsonl`, with a one-time `Seed(ctx)` migration |
| Operator directive | verbatim: *"tellme will read from and save to ~/.tellme/global_prompts.jsonl. If ~/.tellme/global_prompts.jsonl does not exist, the existing output/global_prompts.jsonl file will be copy to ~/.tellme/global_prompts.jsonl."* |
| `/axb-specify` | `specs/plans/028-user-global-prompt-log/`; **0** clarify questions (operator-directed) |
| Pipeline | specify ✅ · spec-by-example ✅ · research ✅ · system-analysis ✅ · data-plan ✅ · dsl-refine ✅ · tasks ✅ · implement ✅ · **delivered** |
| Plan+truth reviews | **APPROVED WITH REQUIRED FOLDS** (TD-1 resolver seam) → folds `0eceef2` + `f67f934` → **FINAL ARCHITECTURAL APPROVAL** |
| Impl review | **IMPLEMENTATION CERTIFIED READY TO MERGE** (`d41405d`); forward note `eb682b6` |
| Architect review | **CERTIFIED READY TO MERGE** (TD-2 dormant `wg` · REFACTOR-1 `Seeder` · TD-1) → fold `39821fd` → **RE-CERTIFIED** |
| Merge | PR [#59](https://github.com/gosharplite/tellme/pull/59) **MERGED** into `dev` (`9e13f9e`, by `gosharplite`, 2026-09-15T20:05:06Z); round-028 head frozen at **`39821fd`** (10 commits) |
| Propagation | `028-user-global-prompt-log → dev` (`9e13f9e`) `→ main` — **DONE (no-ff)** |
| Closeout | `make verify` OK · godog **182/182** (1336 steps, 0 undefined) · topology audit PASSED (250 module rows · 1312 steps); `STATUS.md` split (round-027 detail → `docs/archives/status/2026-09-16.md`); `go install ./cmd/tellme` refreshed |
| Session date note | the round's work ran across the 2026-09-15 → **2026-09-16** boundary; `date` at closeout = **2026-09-16**, so this summary is the fresh 2026-09-16 file (§1). |

---

## 2. Round 028 — the full pipeline

1. **Scope.** Operator-request state-location slice: the `-i` shared prompt log moves out of the `TELL_ME_HOME` namespace to the user-global `~/.tellme/` root (reusing the round-026 convention), with a **first-use seed** (copy the env-scoped file verbatim when the new file is absent). The record shape, the read/compaction semantics, and the `-i`-only write rule stay unchanged.
2. **Plan + truth half.** `/axb-specify` → `/axb-spec-by-example` (2 journeys) → `/axb-technical-research` (`research.md` D1–7 + residual risks) → `/axb-system-analysis` (`plan.md`; 2 interfaces; api NOOP; ui skipped) → `/axb-data-plan` (`prompt_log_entry` location + seed lifecycle) → `/axb-dsl-refine` (ADD the carry-over feature; retarget the recording feature + 3 `dsl.md` rows; +4 seed rows).
3. **Tasks.** `/axb-tasks` → T001–T016 (Foundational 3 · Phase 3 = 3 ALIGN + 4 RED + review · 4A MODIFY recording · 4B ADD carry-over · 4C regression); orphan sweep 21/21 PASS.
4. **Implementation.** `/axb-implement` — adapter takes the **injected** `userHome` resolver + a **pure ctor** + an explicit **`Seed(ctx)`** (verbatim copy; `O_EXCL` no-overwrite; blank-home skip; best-effort); CLI wires both call sites and seeds **once** at the composition root; segregated `history.Seeder` capability port.

### Decisions locked (round 028)

| # | Decision |
| --- | --- |
| D1 | The log lives at `~/.tellme/global_prompts.jsonl`, resolved via the **CLI-injected** `userHomeDir` resolver (mirroring `newToolUsageStore`) — never a direct `os.UserHomeDir()` in the adapter. |
| D2 | Both read and write move; the env-scoped file is a **seed source only**. |
| D3 | **Seed-on-absent**: verbatim copy (copy, not move; never overwrite via `O_EXCL`; missing source / blank home → empty; best-effort). |
| D4 | The seed is an **explicit `Seed(ctx) error`** invoked **once** at the composition root (not a constructor side effect — the tracker is built twice per `-i` run). |
| D5 | An unresolvable/unwritable home degrades the log to a no-op. |
| D6 | **Recorded divergence**: tellme no longer shares the prompt log with `tell-me-go` (sharing unit *per-`TELL_ME_HOME`* → *per-user*); **ADR 0004**. |
| D7 | stdlib-only · POSIX-only · hermetic (temp `HOME` + temp runtime home). |
| F1 (review) | TD-1 resolver injection · TD-2 explicit `Seed` · TD-3 contention escalation · RF-1 ADR · RF-2 stale docs · RF-3 nits. |
| F2 (architect) | TD-2 dormant `wg` documented as a reserved drain hook · REFACTOR-1 segregated `history.Seeder` · TD-1 compaction forward note sharpened. |

### Commits (branch `028-user-global-prompt-log`, then merged)

| Commit | Note |
| --- | --- |
| `43000ce` | `docs(028)`: plan package + spec |
| `1d0fe95` | `docs(028)`: acceptance + research + techstack truth |
| `4d24c12` | `docs(028)`: system-analysis plan |
| `65820ad` | `docs(028)`: data + interface truth |
| `b09eec1` | `docs(028)`: tasks.md |
| `0eceef2` | `docs(028)`: plan-review fold (TD-1/TD-2/TD-3/RF-1..3) |
| `f67f934` | `docs(028)`: micro-note fold (blank-home guard) |
| `d41405d` | `feat(028)`: move the shared prompt log to the user-global root + a first-use seed |
| `eb682b6` | `docs(028)`: impl-review forward note (non-atomic seed publish) |
| `39821fd` | `refactor(028)`: architect-review fold (segregated `Seeder`, documented `wg`, compaction note) |
| `9e13f9e` | PR [#59](https://github.com/gosharplite/tellme/pull/59) merge into `dev` |

### Artifacts / truth

- Plan package: `spec.md` · `checklists/requirements.md` · `research.md` (D1–7) · `plan.md` · `features/acceptance/*.feature` ×2 · `tasks.md` (T001–T016) · `truth-delta.md`.
- Truth: `techstack.md` MODIFY (Shared global prompt log) · `data/data-model.dbml` MODIFY (`prompt_log_entry`) · `chat/carrying-over-the-environment-prompt-log.feature` ADD + `chat/recording-the-shared-prompt-log.feature` MODIFY + `chat/dsl.md` MODIFY (3 rows retargeted + 4 seed rows + note) · `contracts/**` NOOP.
- Code: `internal/infrastructure/history/global_prompt_tracker.go` · `internal/domain/history/tracker.go` (+ `Seeder`) · `internal/cli/cli.go`; `docs/decisions/0004-user-global-prompt-log.md` (+ README index).

---

## 3. Verification (2026-09-16, on `dev` @ `9e13f9e`)

- `make verify` **OK** (no test-sleep · offline witness · cross-compile 4/4 · `golangci-lint` 0 issues · `govulncheck` clean).
- `go test -count=1 ./...` green — unit + godog E2E; **182 scenarios (182 passed) · 1336 steps (1336 passed)**, 0 undefined.
- Topology audit **PASSED** — 40 features · 6 modules · 16 root + **250** module rows · **1312** steps.
- `gofmt` clean · `go.mod`/`go.sum` unchanged (stdlib-only).
- **Falsifiability witnesses** (reproduced then reverted): (a) write-target revert → 7 scenario failures · (b) seed source disabled → the carry-over Example fails · (c) never-overwrite guard disabled → the no-carry-over Example fails.

---

## 4. Open items (non-blocking)

- **Round-028 forward items** — (a) the **seed publish is not atomic** (`O_EXCL` create + one `Write`); (b) the **user-global log grows unbounded** and `Recent()` reads the whole file — a future `Compact(maxEntries)` must take an advisory **`flock`** (the escalated no-`flock` item); (c) the **`tell-me-go` divergence** is recorded (ADR 0004).
- Carried: PR #16 **Obs 1** stdout TTY probe OPEN; round-006 **Obs 3**; round-018 gray styling; round-019 macOS CPU leg; round-024 config-gated `CONTEXT_WINDOW`; sequential tools / no pruning / no `flock`; round-011 forward items.
- Future-slice candidate: coverage tooling [#13](https://github.com/gosharplite/tellme/issues/13).

## 5. Next steps

1. Choose the `029-*` theme and start it via `/axb-specify` off `dev`.
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

## 6. PM follow-ups

- None new (spec/acceptance complete; no PM-owned gaps).

## 7. Issue tracker (closeout Step 8)

Reconciled against the delivered state: **#13** open (coverage tooling; future candidate, still accurate); **#47** `not_planned`, **#53**/**#55** completed (earlier sessions). Round 028 was an **operator request** (no anchor issue). **No changes this closeout.**
