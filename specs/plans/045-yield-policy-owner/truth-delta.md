# Truth Delta: 045-yield-policy-owner

**Plan Package**: `specs/plans/045-yield-policy-owner`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **CLI Application / Agent tool loop** row | The `LoopObserver` hook seam's yield pair is renamed: `Before/AfterToolLog` → **`YieldIndicator`** (clear-only) / **`RestoreIndicator`** (resume) — the split removes the *line-about-to-be-written* conflation so a non-log yield (the round-035 per-call tail) reads as a yield (round 045). | round 045 FR-004/FR-006; `research.md` D3; `spec.md` §Locked decisions C-R3-3. |
| MODIFY | `specs/truth/techstack.md` — **CLI Application / Turn progress spinner** row | The yield behaviour paragraphs (rounds 034/035/040) now name a **single owner** — `ui.YieldController` (`Yield`/`Restore`/`Admit`) — instead of three homes; the mechanism (synchronous joined clear, epoch-preserving resume, row-aware clear) is unchanged. Policy recorded in **ADR 0014**. | round 045 FR-001/FR-002/FR-011; `research.md` D2; `spec.md` §Locked decisions C-R3-1. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**` directory exists**) | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface; `specs/truth/` contains only `data/`, `features/`, and `techstack.md`. This round renames an internal observer hook and consolidates a presentation policy — it authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A4 / `research.md` D6. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry` and the `~/.tellme/*.jsonl` shapes | No persisted/runtime state change: the round is a structural consolidation of an in-memory presentation policy; no record field, file location, or lifecycle changes. | `spec.md` A4 / `research.md` D6. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | No user-facing CLI interface behaviour changes — the round is behaviour-preserving; no feature Rule, Example, step, or `DSLRow` is added or changed (the Gherkin/DSL topology audit is unchanged; the class-phrase vocabulary stays **11**). **Stale-row guard (F-2, measured):** `grep -rn 'BeforeToolLog\|AfterToolLog' specs/truth/` returns **one** hit — `techstack.md`'s own round-045 sentence (this round's MODIFY); `specs/truth/features/**` (incl. `chat/dsl.md`) carries **zero**. So no truth row (and no truth note) is stale. The explanatory notes that *name* the old hooks live in the **frozen plan packages** of the rounds that introduced them (019 / 022 / 034 / 035 / 040), which is expected and out of scope. (Note `chat/dsl.md` *does* carry round-035/036 notes — lines 53/55 — but they describe the yield behaviour **without** naming the hooks, which is why the grep of `specs/truth/**` is zero.) | `spec.md` A1/A5; `research.md` D6 — the round-020/031/041/042/043/044 non-BDD-refactor precedent. |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0014-yield-policy-owner.md` (+ the `docs/decisions/README.md` index row) | Records the **single-owned yield policy** (`ui.YieldController`: `Yield` clear-only / `Restore` resume / `Admit` goroutine-drawn resume; a clear never resumes; restoration is a phase-boundary act) and the **`LoopObserver` hook split** (`YieldIndicator`/`RestoreIndicator` replacing `Before/AfterToolLog`). **Amends ADR 0005 D1 by reference** (0005 partitions *rendering*, not *yield*; its body is not edited — its index row is annotated). No ADR superseded. | round 045 FR-010; `research.md` D5 — a project-level rule future rounds must cite. |
