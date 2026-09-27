# Truth delta — round 094 `094-dead-code-hygiene`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop`
proves the area was checked).

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` (a new *Dead-code reachability (advisory)* row, next to the *Domain model* / `modelith-drift` row) | Record the **advisory** `make dead-code` target: vanilla `golang.org/x/tools/cmd/deadcode` (pinned `v0.47.0`, a PATH dev-tool binary — **no** `go.mod`/`go.sum` change) run with `-test`, filtering the recorded **interface-conformance-only** FP class so a clean tree prints nothing; **never fails**, **not** a `make verify` member; **supersedes ADR 0042 §D5's "no `dead-code`"** wording (ADR 0064). | The round adds a new dev-tool carrier; `techstack.md` is its truth home (owner `/axb-technical-research`). |
| MODIFY | `specs/truth/techstack.md` (*Coverage tooling — DECLINED* row) | Add a **one-line pointer** that the **reachability-orphan** half of the declined #144 decision is now answered by the advisory `make dead-code` carrier (ADR 0064), while the **percentage** half stays declined. | `truth-current`: the old row said the reachability orphan class was "caught by reasoning (round 068)" — this round's measurement falsifies that; keep the record accurate without reversing the percentage decline. |
| NOOP | `specs/truth/techstack.md` (all other rows) | Checked — no other technology change (no new dependency, no new `make verify` member). | NFR-1. |

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/features/cli/**/*.feature` + all `dsl.md` rows | Checked — **no** `.feature`/DSL/step change. A Make target is not a CLI/executable surface (godog drives the built binary). | No product-observable behaviour change (removed symbols have zero callers). |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface change (a CLI end). | CLI-streamlined pipeline (`/axb-api-plan` NOOP). |
| NOOP | `specs/truth/data/**` | Checked — no persisted-state change. | The round changes no state. |
| NOOP | `ui/**` | Checked — no user-facing UX surface change. | Plain line-oriented CLI; `/axb-ui-plan` skipped. |

## Records (not truth — noted for completeness)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0064-dead-code-advisory-carrier.md` | **ADR 0064** — the advisory `make dead-code` carrier: tool pin, advisory-not-a-gate, the FP policy (interface-conformance-only class), not-a-catalog, not the reference's `cmd/deadcode`, the PATH provenance hazard, and the explicit statement that it **supersedes ADR 0042 §D5's "no `dead-code`"** wording. | FR-5: a durable decision needs its authoritative home. |
| MODIFY | `docs/decisions/README.md` (ADR 0042 index row + a new 0064 index row) | Annotate the ADR 0042 index row with a forward pointer to 0064; add the 0064 row. The **ADR 0042 body is left verbatim** (immutable Accepted ADR — supersede, never edit). | `adr-index-consistent` + the recurring "no back-pointer" review fold. |
| MODIFY | `docs/domain-model/quality.modelith.yaml` (+ re-rendered `quality.modelith.md`) | Add `dead-code` as an **advisory** `QualityGate` member alongside `modelith-drift` in the `QualityPipeline` description; **`make modelith-render`**; `modelith-check` no drift. | The domain model is **load-bearing** (ADR 0041): the round adds a modelled quality carrier → same-PR update. |
| MODIFY | `Makefile` | Add the `dead-code` target + `.PHONY` + `help` line; **no** change to the `verify:` aggregate. | FR-2/FR-3. |
| MODIFY (delete) | `internal/config/config.go` · `internal/infrastructure/llm/openai/client.go` · `internal/agent/agentloop.go` (+ `tool_contract.go` comment) · `internal/infrastructure/di/mcp_factory.go` · `tests/e2e/harness/cmd_helper.go` (+ `step_t014…go` comment) · `internal/infrastructure/mcp/mcptest/server.go` · `internal/cli/composite_observer.go` | Remove the **8** genuinely-dead entries of `spec.md` §2 (10 symbols incl. the two `noopCallObserver` methods); correct the two **comment** references. | FR-1: the cleanup (behaviour-neutral — zero callers). |
