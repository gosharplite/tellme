# Tasks — round 096 `096-dependency-vuln-hygiene`

**Plan Package**: `specs/plans/096-dependency-vuln-hygiene`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `go.mod`/`go.sum`, `docs/decisions/**`
**Type**: dependency / toolchain hygiene round — no product code, no `.feature`, no new module outside the `golang.org/x/*` family.

## Phase 1: Setup

- [X] **T001** — Created the round branch `096-dependency-vuln-hygiene` off `dev` `99f1114`; created the plan package `specs/plans/096-dependency-vuln-hygiene/` (`spec.md`, `checklists/requirements.md`, `truth-delta.md`).

## Phase 2: Foundational

- [X] **T002** — Fail-first baseline (research §Environment): on the kiosk container, `make check-full` stops at `vulncheck` (Go **1.26.6**, `x/net@v0.38.0`); captured the 9 flagged stdlib advisories + `GO-2026-6617/6612/6611/6610/6603` (`x/net`, fixed `v0.60.0`).

## Phase 3: Test Alignment & Implementation

> A hygiene round: the "alignment" is the dependency graph + the truth/gate surface. No DSL sentences; `spec-by-example` + `dsl-refine` are NOOP.

- [X] **T003** — `go.mod` / `go.sum` — `go get golang.org/x/net@v0.60.0` + `go mod tidy`; confirm the transitive `golang.org/x/*` rise and the inert `go 1.26 → 1.26.0` normalisation. *(FR-001, research D1/D4)*
- [X] **T004** — `specs/truth/techstack.md` *Dependency vulnerability scan* (`govulncheck`) row — reconcile in place: the gate is toolchain-sensitive; the toolchain floor (**≥ 1.26.9**) + the module floor (`golang.org/x/net ≥ v0.60.0`); DOI ADR 0066. *(FR-004)*
- [X] **T005** — `docs/decisions/0066-dependency-vuln-hygiene.md` (new ADR) — records the decision (the gate's toolchain-sensitivity + the two floors + the standing clearing method); cites ADR 0012 R1 (toolchain outside the make-hermetic boundary); does not amend any ADR body. *(FR-004, research D3)*
- [X] **T006** — `docs/decisions/README.md` — append the ADR 0066 index row. *(FR-004)*

## Phase 4W: Witness Pins

- [X] **T007** — **[WITNESS] W1** — `govulncheck` **red → green**: on the pre-fix tree with Go < 1.26.9, `govulncheck ./...` reports the stdlib advisories; after the fix (Go ≥ 1.26.9 + `x/net v0.60.0`), it reports **"No vulnerabilities found."** *(CLM-001; SC-002)*
- [X] **T008** — **[WITNESS] W2** — the module floor holds: `grep 'golang.org/x/net v0.60.0' go.mod` ⇒ present; `go mod tidy` ⇒ **no diff** (consistent). *(CLM-002; SC-001)*
- [X] **T009** — **[WITNESS] W3** — the truth clause holds: `grep -c '1.26.9' specs/truth/techstack.md` ⇒ ≥ 1, and `git diff --stat specs/truth` touches **only** the `govulncheck` row. *(CLM-003; SC-004)*
- [X] **T010** — **[WITNESS] W4** — `make verify-adr-index` consistent (ADR 0066 indexed once). *(CLM-004)*

## Verification

- [X] **T011** — Behaviour-neutrality: `make verify` **OK**; `go test -count=1 ./...` **green** with E2E **330 scenarios · 2487 steps** (unchanged); `make test-race` no races; `go.mod` diff introduces **no** non-`golang.org/x/*` module; `specs/truth` diff touches one row; no `*.modelith.*` change. *(NFR-001, FR-003, FR-005, I-1/I-2, SC-003/SC-004/SC-005)*
- [X] **T012** — Record: `truth-delta.md` filled (research rows); `STATUS.md` environment note + day log updated.

## Delivery

- [X] **T013** — Commit + push the round branch; open the round PR into `dev` (no Copilot review; only a human merges).

---

## Claim → Witness ledger

| Claim ID | Source spec / Truth anchor | Claimed effect | Witness type | Bound Task / Decision | Discriminating falsifier | Status |
| --- | --- | --- | --- | --- | --- | --- |
| CLM-001 | FR-002 · SC-002 · research D1/D2 | on Go ≥ 1.26.9, the `vulncheck` gate reports **no called vulnerabilities** | `[WITNESS]` W1 | T007 | the pre-fix host (Go 1.26.6 / `x/net v0.38.0`) reports 9 stdlib + the `x/net` advisories (observed: red before, green after) | verified |
| CLM-002 | FR-001 · SC-001 · research D1 | `go.mod` carries `golang.org/x/net v0.60.0`; `go.sum` consistent | `[WITNESS]` W2 | T008 | a stale `v0.38.0` (or a `go mod tidy` diff) would redden the claim (observed: v0.60.0, no diff) | verified |
| CLM-003 | FR-004 · SC-004 · research D2 | the truth records the toolchain floor; only the `govulncheck` row changes | `[WITNESS]` W3 | T009 | a *Language*-row edit or a second truth-row change would redden (observed: one row) | verified |
| CLM-004 | FR-004 · research D3 | ADR 0066 is indexed exactly once | `[WITNESS]` W4 | T010 | a missing/duplicate index row (observed: consistent) | verified |
| CLM-005 | FR-003 · NFR-001 · I-1/I-2 | behaviour-neutral; no new module; no model change | `[WITNESS]` `git diff` + gates | T011 | a new non-`golang.org/x/*` module / an E2E count change / a `*.modelith.*` diff (observed: none) | verified |
| CLM-006 | EC-001 · ADR 0066 §Forward | green only on a **≥ 1.26.9** toolchain; the floor is **recorded, not machine-checked** | `accepted-unwitnessed` | ADR 0066 | — (recorded; the environment floor is a ConfigHub/image concern) | verified (recorded) |

## Pre-Delivery Orphan Coverage Sweep

- `truth-delta.md` non-NOOP item (the `govulncheck` row MODIFY): covered by T004.
- `research.md` Decisions D1–D6: all materialised (D1 → T003; D2 → T004; D3 → T005/T006; D4 → T003's inert-normalisation check; D5 → the not-modelled record in `plan.md` §5; D6 → the NOOP rows).
- `specs/truth/techstack.md` changed section (the `govulncheck` row): covered by T004.
- No orphan artifacts.
