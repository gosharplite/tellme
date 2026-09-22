# System Analysis — Roll back the last N turns of the session history (`-b`/`--back`) (round 081)

**Plan Package**: `specs/plans/081-back-rollback-turns`
**Anchor**: issue [#163](https://github.com/gosharplite/tellme/issues/163)

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The session-history rollback surface (the `-b`/`--back [N]` flag + the offline action + the durable store capability) | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the `history` module gains rollback Rules/Examples; new Given/When/Then DSL rows arrange K exchanges, roll the last N back, and assert the surviving file + archive + streams |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — no persisted-state **shape** change (a rollback removes whole `history_entry` lines and adds none; the `history_entry`/`history_step` DBML is unchanged) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (one plain `stdout` confirmation; no screen/TUI change) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The durable store capability | the implementation | `internal/domain/history` — `Store.Rollback(n int) (removed int, err error)`; `internal/infrastructure/history/file_store.go` — the temp-file + `fsync` + atomic-rename rewrite (clamp; archive untouched; missing file ⇒ 0) |
| **W2** | The CLI flag + dispatch | the implementation | `internal/cli/cli.go` — register `-b`/`--back` (`NoOptDefVal = "1"`), generalise the optional-int pre-pass (RF-54-4), `renderRollback` (reusing `resolveWorkspace`), run `-b` after `-l` and before the prompt phase; the `-b`×`--new` refusal |
| **W3** | The truth rows | `/axb-technical-research` (**done**) | `specs/truth/techstack.md` — *CLI flag parsing* / *Session history store* / *Session lifecycle flags* MODIFY; **ADR 0053** + index (resolving **RF-54-4**); the `docs/domain-model` `History` rollback action + `Session` offline invariant |
| **W4** | The executable CLI contract | `/axb-dsl-refine` (**done**) | Rules/Examples on the `history` feature + `history/dsl.md` rows (Givens arranging K exchanges; Whens rolling back; Thens for the surviving file/archive/streams + the prompt-bearing form) |
| **W5** | The plan-side acceptance | `/axb-spec-by-example` (**done**) | `features/acceptance/rolling-back-the-session-history.feature` |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is a **new offline session command + a durable store mutation**, so `/axb-spec-by-example` is **NOT** NOOP.

## 4. Unchanged surfaces (invariants)

- The frozen class-phrase vocabulary and the ten-value exit-code set are unchanged; a rollback refuses a bad count with the usage phrase + exit 2 (`spec.md` NFR-003).
- `history.Entry` / `history.Step` JSON shapes are unchanged (`spec.md` FR-010).
- The archive (`history.archive.jsonl`) is never written by a rollback (`spec.md` FR-004).
- Session selection is identical to `-l`/`-t` (`spec.md` I-7).
- Turn numbering follows the trimmed history (`spec.md` NFR-005).
- Stdlib-only; POSIX-only; no new dependency; the E2E needs no pty and no live network (`spec.md` NFR-002).

## 5. Domain model (ADR 0041)

**Modelled — amended in this same PR**: the `History` entity gains the **`history-rollback-removes-complete-turns`** invariant (and a definition clause), and the `Session` entity gains **`session-rollback-stays-offline`** (and a definition clause) — `docs/domain-model/tellme.modelith.yaml` → re-rendered `tellme.modelith.md`; `modelith-check` green. No new entity/attribute/relationship. The reference drops message pairs and treats `backN ≤ 0` as a no-op, so both the line-granularity and the `n ≤ 0` refusal are recorded divergences (**ADR 0053**).

## 6. Behavioural notes locked during implementation (the record)

- **N-1 — a turn is one line; the archive is untouched.** `Rollback` rewrites `history.jsonl` with the surviving entries; `history.archive.jsonl` is never opened.
- **N-2 — durability is atomic.** Write the survivors to `.history.jsonl.tmp` (same dir), `fsync`, `os.Rename` over `history.jsonl`; the live file is never truncated in place.
- **N-3 — the count is clamped; `n ≤ 0` is refused.** `removed = min(n, len(entries))`; the CLI maps `-b 0`/`-b -1` to the usage error (exit 2) before touching the store.
- **N-4 — `-b` runs after `-l` and before `-t`; `-b`×`--new` refuses.** `tellme -l 2 -b` lists then rolls back; `-b`+`--new` (no prompt) is a usage error.
- **N-5 — the confirmation is a plain `stdout` line** (`Rolled back N turns. History now holds M turn(s).`), no class phrase; a failure reuses `emitHistoryError` (exit 4).
- **N-6 — the offline `-b` creates the workspace** (round-053 `EnsureWorkspace`), so a rollback on a never-used mode is a tolerated 0-removed (recorded: RF-081-6).
- **N-7 — the E2E arranges history via the existing Given** and asserts the file bytes (`readSessionEntries`); the archive stays absent; the fake provider records zero requests.
