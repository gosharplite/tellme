# Truth Delta: 057-tool-chrome-colour-and-payload-delta

**Plan Package**: `specs/plans/057-tool-chrome-colour-and-payload-delta`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills. Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **Clarify round 1 CLOSED** — **Q1 → 2** (only the estimated pre-flight line changes) · **Q2 → 1** (the baseline is the last estimate in-process, in memory) · **Q3 → 1** (`turns.log` content-equal but plain). Boundary forms A7 (`+0`) / A8 (signed) / A9 (plain delta) exposed as operator-vetoable assumptions.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Agent tool loop** row | **Round 057 (ADR 0027):** the `[Tool Action]` argument-**value** cap rises **189 → 500** rendered runes (cap set `{500, 200, 200}`; same rune-safe mechanic — the recorded divergence from the reference's 189 **bytes** now differs in magnitude too); the whole `[Tool Action]` line gains the terminal-gated **yellow** accent. | `spec.md` US3 / FR-010…FR-012; `research.md` D3/D4 |
| MODIFY | `specs/truth/techstack.md` — **Turn chrome (operator)** row | **Round 057 (ADR 0027):** the colour element set gains (v) the whole `[Tool Output]` **header** line + **both** separators (**grey** `\033[0;90m`) and (vi) the whole `[Tool Action]` line (**yellow** `\033[0;33m`); the pre-flight **estimated** payload line becomes `+<delta> ~<tokens> tokens - <mode> - <model>` (signed increment over the previous in-process estimate; `/budget` dropped). The **measured** line is unchanged. | `spec.md` US1 / US2 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | tellme exposes a single CLI end; no OpenAPI/HTTP surface changes. | `spec.md` A4 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | No persisted/in-memory state-model change: the payload-increment baseline is **in-memory** and session-scoped (a fresh process has no predecessor). | clarify Q2 → 1; `research.md` D6 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/colouring-the-session-chrome.feature` (+ `chat/dsl.md`) | **Round 057:** a new Rule/Example — at a terminal the `[Tool Output]` header + both separators are **grey** and the `[Tool Action]` line is **yellow**; the no-colour Rule's carrier now covers grey/yellow too. | `spec.md` US1 / FR-001…FR-004 |
| MODIFY | `specs/truth/features/cli/chat/reporting-the-payload-status.feature` (+ `chat/dsl.md`) | **Round 057:** a new Rule/Example for the **increment** (`+0` on the first request; a positive increase on a later request; no `/budget` on the estimated line); the budget Rules re-scope to the **measured** line. | `spec.md` US2 / FR-006…FR-009 |
| MODIFY | `specs/truth/features/cli/chat/watching-the-tool-loop.feature` (+ `chat/dsl.md`) | **Round 057:** the action-value cap Example moves 189 → **500** (with a longer fixture) + `dsl.md` retargets. | `spec.md` US3 / FR-010/FR-012 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0027-tool-chrome-colour-and-payload-delta.md` (+ the `docs/decisions/README.md` index row) | Records the two new colour elements (extending ADR 0023), the `argValueCap` re-parameterisation (ADR 0005 lineage), the signed-increment payload line, and a §Forward (RF-057-x). | `spec.md` A5; `research.md` D8 |
