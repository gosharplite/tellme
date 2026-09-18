# Truth Delta: 053-offline-session-config-and-turns-flag

**Plan Package**: `specs/plans/053-offline-session-config-and-turns-flag`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify round 1 CLOSED** — **Q1 → (C1)** (`tellme` writes its own `turns.log` — its rendered chrome — and `-t` prints it); **Q2 → (A)** (an explicit `-c` that cannot be honoured fails; an absent default stays tolerant). No truth has been written yet.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **CLI flag parsing** row | **Round 053 (closes [#103](https://github.com/gosharplite/tellme/issues/103); ADR 0022)**: the flag set gains `-t`/`--turns`. | `spec.md` US2 / FR-005 |
| MODIFY | `specs/truth/techstack.md` — **Session lifecycle flags** row | **Round 053 (ADR 0022)**: the offline session commands (`-l`, prompt-less `--new`, `-t`) resolve the mode as `TELL_ME_MODE` → else the **`-c` config's `MODE`** → else the default → else `"butler"` (previously ignored `-c`); the `-c` read is mode-only (stays offline); an **explicit** `-c` that cannot be read **fails**; `-t` prints `turns.log`. | `spec.md` US1 / FR-001…FR-003, FR-009; `research.md` D1–D4 |
| ADD | `specs/truth/techstack.md` — **Turn log (`turns.log`)** row | **Round 053 (Q1 → (C1); ADR 0022)**: a per-session plain-text `output/<mode>/turns.log` holding tellme's rendered turn chrome; written by an injected port (best-effort); read by `-t`; archived by `--new`. | `spec.md` US2 / FR-004, FR-006, FR-007; `research.md` D5 |
| NOOP (checked) | `specs/truth/techstack.md` — **Build & Tooling / Task runner** row | Inspected: the round adds **no** Makefile target — the gates ride the existing `verify` aggregate. | `spec.md` SC-004 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — expected NOOP)_ | `specs/truth/` (**no `contracts/**`**) | tellme has a single CLI end and no OpenAPI/HTTP surface. | `spec.md` A5 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — expected MODIFY)_ | `specs/truth/data/data-model.dbml` | ADD a `turns_log` artifact (the per-session rendered turn chrome, `output/<mode>/turns.log`; archived on `--new`). | `spec.md` A4; Q1 → (C1) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(pending — expected MODIFY)_ | `specs/truth/features/cli/**` and `specs/truth/features/cli/**/dsl.md` | A user-facing session-selection correction + a new `-t` flag likely add/align a CLI Example + `DSLRow`. | `spec.md` A5 |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0022-offline-session-config-and-turns-log.md` (+ the `docs/decisions/README.md` index row) | Records D1–D5 (the mode-resolution precedence + explicit-`-c` failure; the widened `resolveWorkspace`; the `-t` flag; the `turns.log` artifact/writer) + a §Forward (RF-53-1…4). | `spec.md` FR-010; `research.md` D6/D7 |
