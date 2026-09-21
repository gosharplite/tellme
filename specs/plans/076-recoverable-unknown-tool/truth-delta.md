# Truth Delta: 076-recoverable-unknown-tool

**Plan Package**: `specs/plans/076-recoverable-unknown-tool`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` + `/axb-dsl-refine` RUN (2026-09-22)** — `research.md` D1–D8 + **ADR 0048** + `techstack.md` MODIFY ×2. Clarify resolved at specify time (**0 questions** — not escalated). `/axb-api-plan` + `/axb-data-plan` record `NOOP`. **Architect review 1 fold (`APPROVE WITH REQUIRED FOLDS`, F-1…F-6 + TDs):** the falsifiability folds added pins (pairing/no-usage/mixed-round, the cap discriminating Then + value pin, the ordering pin, the action-line chrome) and corrected the false `clampBytes` claim (F-4) + synced the stale `offering-the-agent-tools.feature` Example (F-6) — no truth *semantic* change beyond the D2 wording correction.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Agent tool loop* | an **unknown tool name** is a **recoverable fold-back** (`tool`-role result `error: no tool named "<name>"; available tools: …`, then `continue`) instead of terminal; **bounded per turn** by `maxUnknownToolFolds = 3`; the unknown-name check stays before the reason gate; round-008 **FR-010 narrowed** (phrase + exit 7 for bound reached / cap exhausted / no tools registered); a tellme-side robustness improvement (reference not asserted) | `spec.md` US1/US2, FR-001…FR-006; `research.md` D1–D8 |
| MODIFY | `specs/truth/techstack.md` — *Tool-usage accounting* | the parenthetical "an unregistered tool aborts the loop" is corrected: an unregistered tool **records nothing and no longer aborts** (folds back + continues, bounded per turn) | `spec.md` I-2, FR-001 |
| ADD | `docs/decisions/0048-recoverable-unknown-tool-name.md` (+ index row) | the decision record: the recoverable fold-back, the message, the per-turn cap (and why), the ordering, the narrowed FR-010 contract, the not-modelled call | `spec.md` SC-003; `research.md` D1–D8 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state change (the per-turn counter is runtime-only; an unknown call records nothing). | `spec.md` 關鍵實體; `plan.md` §5 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/using-an-unknown-tool-name.feature` | **NEW** feature: Rule 1 (an unknown name is fed back and the turn continues) + Rule 2 (a provider that keeps asking is stopped at the per-turn cap) | `spec.md` US1/US2, FR-001…FR-006 |
| MODIFY | `specs/truth/features/cli/chat/failing-the-tool-loop.feature` | the `A request for a tool that is not available is reported` Example is **relocated** to the new feature (an unknown name is now recoverable; the always-unknown fixture becomes the cap case) | `spec.md` US2, FR-004 |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | +2 rows — the `Given` that scripts an unknown-then-read-then-answer exchange and the `Then` that reads the folded-back unavailable-tool result | `spec.md` US1/US2 |
