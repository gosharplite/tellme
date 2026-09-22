# System Analysis — Preserve completed tool steps on an interrupted turn (round 079)

**Plan Package**: `specs/plans/079-interrupted-turn-partial-persistence`
**Anchor**: issue [#159](https://github.com/gosharplite/tellme/issues/159)

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The interrupted-turn persistence surface (the turn seam + the history record + the exit surface) | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the `chat` module's *remembering the conversation* feature gains an interrupted-turn Rule/Examples; new Given/Then DSL rows script an `execute_command` that signals the turn process (`kill -INT $PPID`) and observe the persisted entry + the exit |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — no persisted-state **shape** change (the synthetic answer is the existing `answer` field; the `history_entry`/`history_step` DBML is unchanged) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (one informational `stderr` line; no screen/TUI change) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The synthetic-answer constant | the implementation | `internal/domain/history/history.go` — `InterruptedTurnAnswer` (a stored record value; no schema change) |
| **W2** | The interruption detection + persist + line | the implementation | `internal/cli/cli.go` (`runTurn`) — after `loop.Run`, `errors.Is(err, context.Canceled) && len(result.Steps) > 0` ⇒ append the synthetic entry, emit the informational `stderr` line, return `Success` |
| **W3** | The truth rows | `/axb-technical-research` (**done**) | `specs/truth/techstack.md` — *Session history store* MODIFY + *Interrupted-turn persistence* ADD; **ADR 0051** + index; the `docs/domain-model/**` invariant amendment |
| **W4** | The executable CLI contract | `/axb-dsl-refine` | a Rule/Examples on `remembering-the-conversation.feature` + `chat/dsl.md` rows (a Given scripting an `execute_command` that signals the turn process; Thens for the persisted entry/`calls`, the informational line, the exit) |
| **W5** | The plan-side acceptance | `/axb-spec-by-example` (**done**) | `features/acceptance/preserving-an-interrupted-turns-completed-work.feature` |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is a **durability/persistence** behaviour on the prompt turn, so `/axb-spec-by-example` is **NOT** NOOP (an interrupted-turn journey).

## 4. Unchanged surfaces (invariants)

- The frozen class-phrase vocabulary and the ten-value exit-code set are unchanged; a new phrase and a new code are refused (`spec.md` NFR-004; the zero-step path keeps today's phrase + exit 6).
- A zero-step interruption writes **nothing** (`spec.md` I-2/US2).
- The in-flight `execute_command` child kill (process-group `SIGKILL`) is untouched (`spec.md` I-3).
- `history.Entry` / `history.Step` JSON shapes are unchanged (`spec.md` FR-007).
- Interrupt detection is **typed** (`errors.Is(err, context.Canceled)`), never string-matched (`spec.md` I-6/FR-004).
- `stdout` is byte-unchanged (the interrupted path writes none); any diagnostic is `stderr`-only (`spec.md` NFR-003).
- Stdlib-only; POSIX-only; no new dependency; the E2E needs no pty and no live network (`spec.md` NFR-002/I-8).

## 5. Domain model (ADR 0041)

**Modelled — and amended in this same PR** (the same-PR rule): the round changes a **modelled behaviour** — the `history-append-after-complete` and `turn-answer-stored-verbatim` invariants. Both are **MODIFY**-ed (`docs/domain-model/tellme.modelith.yaml` → re-rendered `tellme.modelith.md`; `modelith-check` green) to scope the operator-interrupted exception. The change introduces no new entity/attribute/relationship. The reference **discards** the partial turn, so this is a **recorded divergence**.

## 6. Behavioural notes locked during implementation (the record)

- **N-1 — the informational line must NOT carry the `tellme: ` prefix.** The round-017 *failed provider request* carrier asserts **exactly one** `tellme: `-prefixed line (the class phrase). The interrupted-with-work path emits **no** class phrase (it succeeds), so its informational line is **chrome-styled** (`[HH:MM:SS] interrupted by operator; …`), consistent with the round-078 retry line and the turn chrome. Pinned by the unit test + the E2E.
- **N-2 — the synthetic answer is NOT printed to `stdout`.** The interrupted path's `stdout` is empty (the answer is a stored record value, not a rendered answer); the operator sees the informational `stderr` line. Keeps `stdout` byte-exact.
- **N-3 — the interrupted path emits no deferred tail, but it DOES leave the aborted call's opening frame (corrected at fold F-079-4).** The cancelled `Complete` returns before `notifyCallEnd`, so the renderer's deferred final tail is nil and `runTurn` returns without `EmitFinalTail()` — no dangling *deferred* tail. But `OnCallBegin` fires **before** `Complete`, so the aborted call's `╭─⠿ Turn N` header + its estimated-payload line are already on `stderr` (and, via the round-053 sink, in `turns.log`) with **no** measured/metrics/`Ready` tail — an unavoidable dangling frame (round-040 `OnCallBegin` ordering). Recorded as **RF-079-B**; suppressing the frame when the ctx is already cancelled is a candidate, not done.
- **N-4 — the informational line is `stderr`-only, not `turns.log`.** Consistent with the ADR 0050/0038 precedent (a diagnostic that is not turn chrome).
