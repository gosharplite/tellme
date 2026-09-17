# Truth Delta: 043-hermetic-make-go-env

**Plan Package**: `specs/plans/043-hermetic-make-go-env`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **Build & Tooling** table | Adds a **Hermetic toolchain invocation** row: the single `Makefile` `export`/`unexport` boundary; the **neutralise set** (`GOENV=off`, `GOWORK=off`, `GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT` unset, ambient `GOOS`/`GOARCH`/`GOARM` unset), the **preserve set** (warm-cache `PATH`/`HOME`/`GOPATH`/`GOMODCACHE`/`GOCACHE` + network/checksum `GOPROXY`/`GOSUMDB`/`GOPRIVATE`/`GONOSUMDB`/`GOINSECURE`), the **host-default `CGO_ENABLED`** stance, the **`GOENV=off`** rationale (the env-file fallback defeats an unset/empty value — round-042 F-2 / ADR 0011 D5), the unconditional **all-targets** scope, the defence-in-depth relationship to `tools/arch` `childEnv`, and the two recorded residuals (bare-`go`-outside-`make`; the `GOENV=off` env-file escape hatch); cites **ADR 0012** as the normative host. | round 043 FR-007/FR-011; `research.md` D1–D8. |
| MODIFY | `specs/truth/techstack.md` — **Task runner** row | Notes that `make` runs the Go toolchain **hermetically** (the invariant invocation boundary; `GOENV=off` + the neutralise/preserve sets), so an ambient/persisted Go env cannot redden a target; the `verify` aggregate list is unchanged. | round 043 FR-007; `research.md` D3/D6. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/contracts/**` | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface. This round changes the build pipeline's toolchain invocation; it authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A5; `research.md` D9. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry`/the `~/.tellme/*.jsonl` shapes | No persisted/runtime state change: the round changes how `make` launches the Go toolchain; it adds no table, no log, no file shape. | `spec.md` A5; `research.md` D9. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | No user-facing CLI interface behaviour changes — the change is a **dev surface** (`make`), not the `tellme` binary's CLI contract; no feature Rule, Example, step, or `DSLRow` is added or changed (the Gherkin/DSL topology audit is unchanged). The acceptance carrier is the scripted hostile-env `Makefile` invocation + the two positive controls. | `spec.md` A2/A5; `research.md` D9 — the round-020/031/041/042 non-BDD-tooling precedent. |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0012-hermetic-make-go-env.md` (+ the `docs/decisions/README.md` index row) | Records the **hermetic `make` invocation rule**: the single top-of-`Makefile` boundary; the neutralise set + preserve set; the **`GOENV=off`** rationale (the env-file fallback defeats an unset/empty value); the host-default `CGO_ENABLED`; the unconditional all-targets scope; the defence-in-depth relationship to round-042's `childEnv` (ADR 0011 D5); and the two recorded residuals. Generalises round-042 D5 (gate child) + round-020 TD1 (`CGO_ENABLED=0`, one target). No existing ADR is superseded. | round 043 FR-007/FR-011; `research.md` D6 — `docs/decisions/README.md` names "a project-level rule … other artifacts (or future rounds) depend on and must be able to cite". |
