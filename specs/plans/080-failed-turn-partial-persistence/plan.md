# System Analysis — Persist a failed turn's completed tool steps (round 080)

**Plan Package**: `specs/plans/080-failed-turn-partial-persistence`
**Anchor**: issue [#161](https://github.com/gosharplite/tellme/issues/161)

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The failed-turn persistence surface (the turn seam + the history record + the failure surface) | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the `chat` module's *remembering the conversation* feature gains failed-turn Rules/Examples; new Given/Then DSL rows script a partial tool turn that then fails (a permanent transport drop / an outright 400) |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — no persisted-state **shape** change (the synthetic answers are the existing `answer` field; the `history_entry`/`history_step` DBML is unchanged) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (one informational `stderr` line; no screen/TUI change) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The two failure synthetic answers | the implementation | `internal/domain/history/history.go` — `ProviderFailedTurnAnswer` / `ToolFailedTurnAnswer` (stored record values; no schema change) |
| **W2** | The failure persistence + the broadened interruption predicate | the implementation | `internal/cli/cli.go` — `failTurn` (extracted from `runTurn`): an operator interruption (`ctx.Err() != nil \|\| errors.Is(err, context.Canceled)`) keeps the round-079 path; a FAILED turn with steps is persisted (best-effort) then reported with the **unchanged** failure surface + one informational line |
| **W3** | The truth rows | `/axb-technical-research` (**done**) | `specs/truth/techstack.md` — the round-079 *Interrupted-turn persistence* row broadened + the *Session history store* clause extended; **ADR 0052** + index; the `docs/domain-model/**` invariant extension |
| **W4** | The executable CLI contract | `/axb-dsl-refine` (**done**) | Rules/Examples on `remembering-the-conversation.feature` + `chat/dsl.md` rows (Givens scripting a partial tool turn then a failure; Thens for the persisted entry/answer/keep-line/empty history) |
| **W5** | The plan-side acceptance | `/axb-spec-by-example` (**done**) | `features/acceptance/preserving-a-failed-turns-completed-work.feature` |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is a **durability/persistence** behaviour on the prompt turn (a failed turn keeps the completed work), so `/axb-spec-by-example` is **NOT** NOOP.

## 4. Unchanged surfaces (invariants)

- The frozen class-phrase vocabulary and the ten-value exit-code set are unchanged; a failed-but-kept turn still exits **6** / **7** with its phrase (`spec.md` I-2/NFR-003).
- Zero completed steps ⇒ nothing written (`spec.md` I-3).
- `history.Entry` / `history.Step` JSON shapes are unchanged (`spec.md` I-4).
- Detection is structural/typed (`ctx.Err()`; the typed error path), never string-matched (`spec.md` I-5/NFR-004).
- `stdout` is byte-unchanged (a failed turn writes no answer); the informational line is `stderr`-only (`spec.md` I-6).
- Stdlib-only; POSIX-only; no new dependency; the E2E needs no pty and no live network (`spec.md` I-7/NFR-002).
- The in-flight child `SIGKILL` is untouched (`spec.md` I-8).

## 5. Domain model (ADR 0041)

**Modelled — amended in this same PR**: the round changes a **modelled behaviour** — the `history-append-after-complete` and `turn-answer-stored-verbatim` invariants (round 079 already amended them for the operator-interrupted case). Both are **MODIFY**-ed (`docs/domain-model/tellme.modelith.yaml` → re-rendered `tellme.modelith.md`; `modelith-check` green) to extend the exception from *operator-interrupted* to *operator-interrupted **or failed***. No new entity/attribute/relationship. The reference **discards** the partial turn in both cases, so this is a **recorded divergence**.

## 6. Behavioural notes locked during implementation (the record)

- **N-1 — the keep line must NOT carry the `tellme: ` prefix.** The failure phrase is `tellme:`-prefixed; the informational keep line is chrome-styled (`[HH:MM:SS] kept N completed tool step(s) in the session history`), consistent with the round-078 retry line and the round-079 interruption line. The round-017 *exactly-one-`tellme:`-line* contract holds.
- **N-2 — the keep line is emitted only when the append succeeded.** A failed-turn `store.Append` is best-effort (D6): an append failure leaves the failure surface unchanged and does NOT print the keep line (never claim a keep that failed).
- **N-3 — the two failure synthetic answers are distinct from the operator-interruption answer**, chosen by the error class (`errors.As(err, &agentport.ErrIncomplete)` → the tool answer; else the provider answer), so a `-l` reader can tell a failure from an operator stop.
- **N-4 — hole #2 is folded by the broadened predicate.** `ctx.Err() != nil || errors.Is(err, context.Canceled)` (both structural) routes a `Ctrl+C` that surfaced as the round-078 retry decorator's `lastErr` to the round-079 interrupted path (exit 0 + the interruption line). The union keeps the round-079 `errors.Is` term (an adapter that wraps the cancellation without the parent ctx being cancelled).
- **N-5 — the E2E failure witness is hermetic.** A new fake-provider per-reply `Drop` / `ErrorStatus` serves a `read_files` tool call on request 1, then drops (retry exhaustion) or answers 400 (non-retryable) on request 2+; `TELL_ME_FORCE_RETRY_DELAY_MS=0` collapses the retry delays (the count is asserted, never wall-clock).
