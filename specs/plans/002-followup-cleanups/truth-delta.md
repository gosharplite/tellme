# Truth Delta: 002-followup-cleanups

**Plan Package**: `specs/plans/002-followup-cleanups`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | (i) **Test strategy** row: E2E-only → **E2E for the acceptance path + fast unit tests for pure helpers**; added a **Pure-helper unit tests** row (Go stdlib `testing`, table-driven). (ii) **Build & Tooling**: adopted **`golangci-lint` (with `errcheck`)** as the lint aggregator (+ a committed `.golangci.yml`) and **`govulncheck`** as the dependency-vulnerability gate; removed both from **Not Introduced Yet**; `make` targets now `… lint, vulncheck, verify`. (iii) **CLI flag parsing** row: dropped `--json` from the flag list. (iv) Removed the now-stale **Adopted, Not Yet Instantiated → `godog`** note (instantiating in round 001). | Round-002 spec: NFR-003 (unit tests — finding **F9**), NFR-004 (ignored-error gate), NFR-005 (dependency-vulnerability gate), FR-001 (remove `--json`). **Settles round-001 Decision 7's two deferrals** (the lint aggregator + the vulnerability gate) and **amends round-001 Decision 5** (E2E-only → E2E acceptance path + pure-helper unit tests). |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and no OpenAPI/HTTP surface, so there is no API contract to author or change this round. | `contract-authoritative` is satisfied vacuously (no API surface). The `--json` removal and the message/exit-code freeze are CLI-contract behaviour owned by `/axb-dsl-refine` (`specs/truth/features/cli/**`), not an API contract. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. Round 002 changes only the CLI's operator-facing surface (flag removal, message/exit-code freeze) and its verification tooling; no new persisted or in-memory state is modelled. | Net truth change is zero → recorded as `NOOP` (area checked) per `delta-covers-all-owners`; `data-model-covers-all-state` holds vacuously (no modelled data). |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| DELETE | `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature` | Removed the two `--json` Examples — *"Resolution is reported as structured output"* and *"An unresolved setup is reported as structured output with a dedicated exit code"*. The diagnostics interface now carries only the plain resolved/unresolved Examples. | Round-002 FR-001 (remove `--json`; no machine-readable diagnostic mode) — reverses round-001 FR-013. |
| DELETE | `specs/truth/features/cli/diagnostics/dsl.md` | Removed the two `--json` When rows (`the operator runs tellme's diagnostic with "--json"`) and the two structured-output Then rows (`tellme emits the resolution status as structured output`, `tellme emits the unresolved status as structured output`), plus the pinned JSON key-schema note; added a "single plain-text form; `--json` removed" note; pinned the diagnostic exit code. | Same as above. |
| ADD | `specs/truth/features/cli/usage/unsupported-cli-usage.feature` + `specs/truth/features/cli/usage/dsl.md` | Added two `--json` rejection Examples (alongside `-d`, and on its own) under the existing unrecognized-flag Rule, and the supporting When row `the operator runs tellme's diagnostic with "--json"`. `--json` is now an unrecognized flag → usage error. | Round-002 FR-002 (any `--json` occurrence is an unrecognized-flag usage error, never silently ignored); carries the `single-diagnostic-output` Rule 1. |
| MODIFY | `specs/truth/features/cli/dsl.md` (interface root) + `configuration/dsl.md` + `workspace/dsl.md` + `usage/dsl.md` + `diagnostics/dsl.md` | Pinned the operator-facing failure contract: the root `tellme explains on stderr that "{reason}"` row now asserts the **frozen** wording (`tellme: {reason}` prefix), and the four error-code rows now assert the **pinned numeric values** (usage `2`, configuration `3`, environment `4`, diagnostic `5`; success `0`). | Round-002 FR-004 / FR-005 / FR-006 (freeze the exact stderr wording and the numeric exit codes as contract, asserted verbatim). |
