# Truth Delta: 040-spinner-liveness-and-turn-timer

**Plan Package**: `specs/plans/040-spinner-liveness-and-turn-timer`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and must name what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Intended deltas (to be ratified by the owners)**:
> - `/axb-technical-research` — **MODIFY** `specs/truth/techstack.md`: the **Turn progress spinner** row records (a) the **dual elapsed timer** — total since prompt capture + current-turn duration reset per AI-endpoint call (amends round-019 D4's "never reset"); (b) the **idle-gap liveness** during a streaming `[Tool Output]` block (supersedes the round-034 whole-block pause). **ADD** a new ADR (`docs/decisions/0009-…`) that **SUPERSEDES** the round-034 pause semantics recorded in **ADR 0005 D7** (0005's `D7` line is annotated / its `Status` flipped per the ADR immutability rule).
> - `/axb-api-plan` — **NOOP** (no HTTP surface).
> - `/axb-data-plan` — checked **NOOP** (no persisted-state change; the affected text is `stderr`-only presentation).
> - `/axb-dsl-refine` — **MODIFY** `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` (amend the `Rule: The spinner is paused while a command's output streams`; add a `Rule` for the dual elapsed timer) + the matching `specs/truth/features/cli/chat/dsl.md` rows.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Turn progress spinner** row | The elapsed counter becomes a **dual timer**: the total since prompt capture (turn-scoped; never resets) plus the **current AI-endpoint call's** duration (reset per call). While a `[Tool Output]` block streams, the indicator is no longer hidden for the whole block: it resumes after an idle gap and clears on the next output line (single-writer safe). Presentation-only. | `spec.md` FR-001..FR-008; `research.md` D… |
| ADD | `docs/decisions/0009-….md` | Records the dual-timer policy and the idle-gap liveness (with the single-writer constraint); **supersedes ADR 0005 D7**'s whole-block pause and **amends** round-019 D4's "never reset". | `spec.md` FR-001/FR-007; `research.md` D… |
| SUPERSEDE | `docs/decisions/0005-tool-call-log-parity.md` — `D7` | The whole-block pause decision is closed by the idle-gap liveness; 0005 is not edited except its `Status`/`D7` annotation per its own immutability clause. | `research.md` D… |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface. This round changes only the `stderr`-bound spinner presentation. | `contract-authoritative` holds vacuously; `spec.md` FR-009; `plan.md`. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record` and the `~/.tellme/*.jsonl` record shapes | No persisted-state change: the spinner and the `[Tool Output]` block are `stderr`-only presentation and are never persisted. | `spec.md` FR-009; `plan.md`. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` — amend `Rule: The spinner is paused while a command's output streams` | The Rule's Example is amended: the indicator is **not** hidden for the whole block — it resumes during an idle stretch and clears on the next output line. | `spec.md` FR-001..FR-004; `research.md` D… |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` — new `Rule: The spinner shows the total time and the current turn's time` | ADD a Rule with Examples: a single-call turn shows two approximately-equal figures; a multi-call turn resets the second figure per AI-endpoint call while the first grows. | `spec.md` FR-005..FR-008; `research.md` D… |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — new `Given`/`Then` rows + a round-040 note | Rows matching exactly one step each (`dsl-exact-one-match`); `dsl-single-authority` preserved (new rows, no duplication). | `spec.md` FR-001/FR-005; `research.md` D… |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0009-….md` (+ the `docs/decisions/README.md` index row) | Records the dual-timer policy and the idle-gap liveness; supersedes ADR 0005 D7; amends round-019 D4. | The change to two `Accepted` decisions needs a durable, citable home (not a frozen plan package). |
| SUPERSEDE | `docs/decisions/0005-tool-call-log-parity.md` — `D7` (pause) | Closed by WS-A; 0005 stays otherwise immutable. | The ADR immutability rule (0008→0007 precedent). |
