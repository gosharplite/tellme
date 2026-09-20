# Truth Delta: 074-cli-help-and-version-shorthands

**Plan Package**: `specs/plans/074-cli-help-and-version-shorthands`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` + `/axb-dsl-refine` RUN (2026-09-21)** — `research.md` D1–D7 + **ADR 0046** + `techstack.md` MODIFY ×1. Clarify resolved at specify time (Q1 → A · Q2 → A). `/axb-api-plan` + `/axb-data-plan` record `NOOP`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *CLI flag parsing* | tellme gains `-h`/`--help` (prints the pflag flag list to `stdout`, exit 0; a successful, offline, prompt-less action, precedence before `--version`) and the `-v` shorthand for `--version` (byte-identical); help emits no `tellme: …` phrase and the unrecognized-flag refusal (phrase on `stderr`, exit 2) is unchanged | `spec.md` US1/US2, FR-001…FR-004; `research.md` D1–D6 |
| ADD | `docs/decisions/0046-cli-help-and-version-shorthands.md` (+ index row) | the decision record for the two shorthands (explicit flags; the pflag flag-list block on stdout; precedence; the stream/exit contract) | `spec.md` SC-001…SC-003; `research.md` D1–D7 |
| NOOP (checked) | `docs/domain-model/**` | the round changes a CLI flag surface, not a modelled entity/invariant (no CLI-flag entity exists); recorded in `plan.md` §5 (ADR 0041's escape hatch) | `plan.md` §5 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state change (flags only). | `spec.md` NFR-002 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/dsl.md` (interface root) | the cross-module `the operator runs tellme with "{flag}"` row (param'd; was the `diagnostics` module's literal `--version` row) | `spec.md` US1/US2 |
| MODIFY | `specs/truth/features/cli/diagnostics/dsl.md` | the literal `--version` When row removed (it lives at the root now) + a pointer note | `plan.md` §1 |
| MODIFY | `specs/truth/features/cli/usage/dsl.md` | two Then rows: `tellme prints its flag list` · `the help is reported as a success` | `spec.md` US1, FR-001…FR-003 |
| ADD | `specs/truth/features/cli/usage/requesting-help.feature` | **NEW** — 2 Rules / 2 Examples for `-h` and `--help` | `spec.md` US1 |
| MODIFY | `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature` | a `-v` Example added to the version Rule | `spec.md` US2, FR-004 |
