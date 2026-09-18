# Truth Delta: 048-cli-tui-prompt-decoupling

**Plan Package**: `specs/plans/048-cli-tui-prompt-decoupling`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).
>
> **Status**: `/axb-technical-research` (D1–D10 + **ADR 0017** + the `techstack.md` MODIFY rows below), `/axb-system-analysis` (`plan.md`: 0 interfaces + no waves) and `/axb-dsl-refine` (NOOP) have **landed**; the api/data/dsl-refine **NOOP** rows below are ratified (their owners inspected the surfaces). Clarify locked: **Q1 → B** (the TUI prompt edge), **Q2 → A** (the port lives in `internal/domain/**`), **Q3 → A** (F-4 folded in). `/axb-tasks` `tasks.md` (T001–T012) + the implementation are on this branch (see `tasks.md` §Outcome).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Build & Tooling / Layer-discipline gate** row | The RULE-E baseline figure changes **3 → 2**: the `internal/cli → internal/ui/tui/prompt` edge is removed (round 048, R5.2), leaving `internal/cli → internal/{agent, ui}` baselined. The row cites **ADR 0017** and records the mechanism (the `internal/domain/tui.Prompter` port + the `internal/ui` adapter + composition-root injection). The rule + policy (ADR 0011/0016) are otherwise unchanged. | `truth-current`; round 048 FR-008; `research.md` D1/D3/D4/D5/D7. |
| MODIFY | `specs/truth/techstack.md` — **CLI Application / Composition root (dependency injection)** row | The `cli.Options` presentation-seam clause is restated: it now holds the **domain** `internal/domain/tui.Prompter` (an exported interface type) instead of the unexported-typed `RunTUIPrompt` field; the adapter is an `internal/ui` type injected from `cmd/tellme`; the nil-default is removed. **F-4** (exported field of an unexported type) is closed. | round 048 FR-004/FR-008; `research.md` D5/D7. |
| MODIFY | `specs/truth/techstack.md` — **CLI Application / Interactive TUI prompt (`-i`)** row | Adds the round-048 note: the prompt is reached through the injected domain port (satisfied by an `internal/ui` adapter, wired at `cmd/tellme`); `internal/cli` no longer imports `internal/ui/tui/prompt`. No chrome/behaviour change. | round 048 FR-001/FR-008; `research.md` D1/D4. |
| NOOP (checked) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | Inspected: the `verify` aggregate already names `verify-architecture`. The round adds no Makefile target — RULE-E rides that target — so the row is unchanged. | round 048 FR-008; `research.md` D8. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface (`specs/truth/` = `data/`, `features/`, `techstack.md`). This round is an internal wiring refactor; it authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5; `research.md` D8 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | Inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry` and the `~/.tellme/*.jsonl` shapes: **no** persisted/runtime-state change — the refactor moves wiring (a port + an adapter), not data. | `spec.md` A5; `research.md` D8 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | Inspected: a `cli`-reader refactor is **not** a `tellme` CLI-contract change — no feature Rule, Example, step, or `DSLRow` changes (the Gherkin/DSL topology audit is unchanged). The acceptance carrier is the gate + the adapted unit seams. | `spec.md` A2/A6; `research.md` D8 — the rounds 020/031/036/041–047 non-BDD-tooling precedent |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0017-cli-tui-prompt-decoupling.md` (+ the `docs/decisions/README.md` index row) | Records the de-coupling: the **domain** `tui.Prompter` port, its home (`internal/domain/tui`), the `internal/ui` adapter, the composition-root injection, the behaviour-preservation claim, the baseline movement **3 → 2**, the **F-4 closure**, **what the change is *not***, and its relation to ADR 0011/0016/0013/0015. | round 048 FR-007; `research.md` D7 — a structural change the later R5 slices must cite. |
