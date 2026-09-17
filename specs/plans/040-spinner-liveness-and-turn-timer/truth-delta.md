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
| MODIFY | `specs/truth/techstack.md` — **Turn progress spinner** row | The elapsed display becomes a **dual timer** `({total}s {turn}s)`: the **total** since prompt capture (turn-scoped; **never reset** — round-019 D4 preserved) **plus** the **current turn's** duration, **reset at each AI-endpoint call** (stamped in `OnInferenceStart`; an **amendment** to round-019 D4). And, during a `[Tool Output]` block, the spinner is **no longer hidden for the whole block**: after an **idle gap** (default 3 s, a hermetic env seam) the indicator **resumes**, and the next complete output line is preceded by a **synchronous clear** (single-writer safe) — a **supersession of round-034's whole-block pause** (ADR 0005 D7). Presentation-only. | `spec.md` FR-001..FR-008; `research.md` D1–D5. |
| ADD | `docs/decisions/0009-spinner-dual-timer-and-streaming-liveness.md` (+ the `docs/decisions/README.md` index row) | Records the dual-timer policy (per-AI-call reset; the round-019-D4 amendment) and the idle-gap liveness (single-writer; the presenter's goroutine-joined clear); **supersedes ADR 0005 D7** and carries the two recorded reference divergences. | `spec.md` FR-001/FR-007; `research.md` D1–D5/D8. |
| SUPERSEDE (partial) | `docs/decisions/0005-tool-call-log-parity.md` — **`D7` only** | The whole-block pause decision is narrowed to an idle-gap pause. **D7 only** — the rest of ADR 0005 (D1–D6, D8…) stands, so 0005's overall `Status` stays `Accepted` and its **body is not edited**; the supersession is named in ADR 0009 and in the decisions index (research D8). | `research.md` D8. |

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
| ADD | `docs/decisions/0009-spinner-dual-timer-and-streaming-liveness.md` (+ the `docs/decisions/README.md` index row) | Records the dual-timer policy (per-AI-call reset; the round-019-D4 amendment) and the idle-gap liveness (single-writer; goroutine-joined clear); supersedes ADR 0005 D7; carries the two recorded reference divergences. | The change to two `Accepted` decisions needs a durable, citable home (not a frozen plan package). |
| SUPERSEDE (partial) | `docs/decisions/0005-tool-call-log-parity.md` — **`D7` only** | The whole-block pause is narrowed to an idle-gap pause; 0005 stays otherwise immutable and its body is not edited (the supersession is named in ADR 0009 + the decisions index). | The ADR immutability rule (0008→0007 precedent), applied to a single decision. |
