# Truth delta — round 096 `096-dependency-vuln-hygiene`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop`
proves the area was checked).

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — the ***Dependency vulnerability scan*** (`govulncheck`) row (the gate's own truth home; `:145`) | Reconcile **in place**: the gate is **toolchain-sensitive** — its *stdlib* half is clean only on a **patched** Go toolchain (**≥ 1.26.9**; the `GO-2026-6617/6613/6605/…` `net/http` fixes land in `go1.26.9`) — so the environment MUST provide one; the shipped *module* half is the `golang.org/x/net ≥ v0.60.0` floor kept in `go.mod`; on an unpatched host the gate reds with **no repo change** (the recorded environmental floor). Current-state wording (`truth-current`); cites ADR 0066. | FR-004: the gate's stdlib half is an *environment* input invisible in `go.mod`, so its floor must be discoverable as truth; the round-095 "a dev-tool/gate fact is a truth clause" precedent. |
| NOOP | `specs/truth/techstack.md` — all other rows (incl. the *Language* row and the `golang.org/x/*` dep rows) | Checked — **no** other technology change. The `Language` row stays `Go 1.26` (true for 1.26.9); `golang.org/x/net` is **indirect** and has no row of its own (its floor is stated on the gate's row + `go.mod`); no new module outside `golang.org/x/*`; no new `make verify` member. | NFR-001 / SC-004 / SC-005 / ADR 0066 **RF-066-4**. |

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/features/cli/**/*.feature` + all `dsl.md` rows | Checked — **no** `.feature`/DSL/step change. A dependency bump has no CLI/executable surface change. | No product-observable behaviour change (behaviour-neutral bump). |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface change (a CLI end). | CLI-streamlined pipeline (`/axb-api-plan` NOOP). |
| NOOP | `specs/truth/data/**` | Checked — no persisted-state change. | The round changes no state. |
| NOOP | `ui/**` | Checked — no user-facing UX surface change. | Plain line-oriented CLI; `/axb-ui-plan` skipped. |

## Records (not truth — noted for completeness)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `go.mod` / `go.sum` | Bump `golang.org/x/net v0.38.0 → v0.60.0` (+ the transitive `golang.org/x/*` rise); `go mod tidy` normalises the `go` directive `1.26 → 1.26.0`. | FR-001 / SC-001: the shipped dependency half of the red gate. |
| TBD | `docs/decisions/00NN-*.md` (+ `docs/decisions/README.md`) | **RD decision (research, A3)**: whether a durable ADR is owed for "the `vulncheck` gate is toolchain-sensitive". If owed, ADR 0042 gets a **forward-pointer index annotation** only (its body stays verbatim). | `adr-index-consistent` + the "advisory/gate is a durable fact" precedent. |
| MODIFY | `STATUS.md` (environment note) | Record the kiosk-environment facts: the user-scoped Go 1.26.9, the `deadcode` install, and the propagated bump. | Session closeout hygiene (not truth). |
