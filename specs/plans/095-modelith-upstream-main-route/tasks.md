# Tasks — round 095 `095-modelith-upstream-main-route`

**Plan Package**: `specs/plans/095-modelith-upstream-main-route`
**Core Inputs**: `spec.md`, `plan.md`, `research.md`, `truth-delta.md`, `specs/truth/techstack.md`, `docs/decisions/**`, `Makefile`, `docs/domain-model/README.md`
**Type**: toolchain/record round — no product code, no `.feature`, no `go.mod`/`go.sum` change.

## Phase 1: Setup

- [X] **T001** — Created the round branch `095-modelith-upstream-main-route` off `origin/dev`; created the plan package `specs/plans/095-modelith-upstream-main-route/`.

## Phase 2: Foundational

- [X] **T002** — Fail-first evidence (research D1): upstream `main` HEAD resolved to `9008354f19ff13f24273a7c71395c63698c7fbac` (`git ls-remote` + `gh api`); `go install …@main` probed into a temp `GOBIN` (built `v0.5.1-0.20260927062055-9008354f19ff`); installed build's provenance recorded (`go version -m`).

## Phase 3: Test Alignment & Implementation

> A record round: the "alignment" is the doc/route surface. No DSL sentences; `spec-by-example` + `dsl-refine` are NOOP.

- [X] **T003** — `Makefile` — `MODELITH_REF := main` / `MODELITH_INSTALL := go install github.com/stacklok/modelith/cmd/modelith@$(MODELITH_REF)`; `MODELITH_PIN` removed; comment block rewritten. *(FR-002)*
- [X] **T004** — `docs/domain-model/README.md` (*Toolchain → Install*, single source) — the upstream-`@main` route + tracked ref + accepted cost + last-render provenance. *(FR-001)*
- [X] **T005** — `specs/truth/techstack.md` *Domain model* row — reconciled **in place** (fork/pin dropped; upstream `@main`; dev-tool/gate facts + README pointer kept; ADR 0065 cited). *(FR-004)*
- [X] **T006** — `docs/decisions/0065-modelith-install-from-upstream-main.md` (new ADR) — records the decision, supersedes ADR 0030 D2. *(FR-005)*
- [X] **T007** — `docs/decisions/README.md` — ADR 0065 index row appended; forward pointer added on the ADR 0030 row (body verbatim). *(FR-005)*

## Phase 4W: Witness Pins

- [X] **T008** — **[WITNESS] W1** — absent `modelith` fails the gate **naming the new route** (`env PATH=/usr/bin:/bin make modelith-check` ⇒ non-zero + `go install github.com/stacklok/modelith/cmd/modelith@main`). *(CLM-001)*
- [X] **T009** — **[WITNESS] W2** — the documented route resolves & builds (`GOBIN=$(mktemp -d) go install …@main` ⇒ `9008354f19ff`); `make modelith-check` green. *(CLM-002)*
- [X] **T010** — **[WITNESS] W3** — no **live** fork/pin reference remains; new route present (operational surfaces `docs/domain-model/README.md` · `Makefile` · `specs/truth/**` · the bootstrap/closeout/README/STATUS · `scripts/` ⇒ **clean**). *(CLM-003)*
- [X] **T011** — **[WITNESS] W4** — ADR 0065 indexed exactly once (`make verify-adr-index` ⇒ consistent). *(CLM-004)*

## Verification

- [X] **T012** — Frozen history + `go.mod`/`go.sum` + model YAML/MD untouched (`git status` shows none). *(FR-006, FR-007, I-4)* → CLM-005
- [X] **T013** — `make verify` **OK**; `go test -count=1 ./...` **green** (E2E **330 scenarios · 2487 steps**, unchanged); `STATUS.md` + the day log updated.

## Delivery

- [X] **T014** — Commit + push the round branch; open the round PR into `dev` (no Copilot review; only a human merges). Closes [#200](https://github.com/gosharplite/tellme/issues/200) at its merge.

---

## Claim → Witness ledger

| Claim ID | 來源規格 / Truth 錨點 | 宣稱效果 | 見證型態 | 綁定 Task / 決策 | 鑑別性反證 | 狀態 |
| --- | --- | --- | --- | --- | --- | --- |
| CLM-001 | FR-002 / FR-003 · SC-002 | an absent `modelith` fails `modelith-check` **naming the new route** | `[WITNESS]` W1 | T008 | a stale `$(MODELITH_INSTALL)` would print the old fork route (observed: the new route) | verified |
| CLM-002 | FR-001 · research D1 | the documented route resolves & builds upstream `main` HEAD; the gate stays green | `[WITNESS]` W2 | T009 | a fork/`@<branch>` route would not resolve (observed: built `9008354f19ff`) | verified |
| CLM-003 | SC-001 · FR-001/FR-002/FR-004 | no **live** fork/pin reference remains; new route present | `[WITNESS]` W3 | T010 | a live hit on `gosharplite/modelith` / `b4153541` / `feat/self-domain-model` on an operational surface (observed: none) | verified |
| CLM-004 | FR-005 | ADR 0065 indexed exactly once; ADR 0030 body verbatim | `[WITNESS]` W4 + `git diff` | T011 | a missing/duplicate index row, or a 0030 body change (observed: consistent; 0030 body untouched) | verified |
| CLM-005 | FR-006 / FR-007 / I-4 | frozen history + `go.mod`/`go.sum` + model YAML/MD untouched | `[WITNESS]` git-status predicate | T012 | a diff in a frozen path / `go.mod` / `*.modelith.*` (observed: none) | verified |
| CLM-006 | EC-001 | tracking `main` is non-hermetic (accepted) | `accepted-unwitnessed` | ADR 0065 D3 | — (recorded, deliberate) | verified (recorded) |

## Pre-Delivery Orphan Coverage Sweep

- `truth-delta.md` non-NOOP item (the techstack *Domain model* row MODIFY): covered by T005.
- `research.md` Decisions D1–D7: all materialised (D1/D2 → T003/T004/T005; D3 → README/Makefile/ADR; D4 → single source; D5 → ADR 0065 + index; D6 → the not-modelled record in `plan.md` §5; D7 → the NOOP rows).
- No orphan artifacts.
