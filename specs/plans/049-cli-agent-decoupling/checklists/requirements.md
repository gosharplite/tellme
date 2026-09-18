# Specification Quality Checklist: de-couple `internal/cli` from the turn loop — re-cut sub-slice 1: extract the loop's domain-facing contracts (round 049)

**Created**: 2026-09-18

**Feature Directory**: `specs/plans/049-cli-agent-decoupling`

**Spec Path**: `specs/plans/049-cli-agent-decoupling/spec.md`

## Usage

- Check each item against the current `spec.md`.
- On a failure, record the concrete gap and the fix direction under "Issues & corrections".
- If a `NEEDS CLARIFICATION` remains, state explicitly whether it blocks downstream planning.

## Content completeness

- [x] All mandatory sections present
- [x] Feature theme, scope, and main flow are clear (re-cut sub-slice 1 — move the loop's crossing contracts to `internal/domain/**`; the surviving `cli → agent` coupling reduced to the single `AgentLoop` construction; **baseline unchanged**)
- [x] No implementation/framework detail written as a *requirement* (the contract names/home and the alias-vs-reference choice are explicit RD decisions behind FR-002 / RULE-A·C)
- [x] Edge cases cover the main high-risk situations (domain-purity leak; alias-vs-two-names; `errors.As`/`ToolDefs` byte-contracts; partial extraction → second coupling; test imports; cycles; accidental baseline drift; build-tag scope)
- [x] Key entities and success criteria are present

## User stories & requirement attribution

- [x] User stories ordered by value/delivery: US1 (contracts → domain; CLI id count 4 → 1) → US2 (surviving coupling proven edge-sized; baseline unchanged) → US3 (recorded in truth + ADR)
- [x] Every user story is independently verifiable
- [x] Every user story carries acceptance scenarios
- [x] Story-specific FR/NFR attached under the story
- [x] Global requirements hold only cross-story items (FR-010…FR-012, NFR-004, NFR-005)
- [x] No formal requirement duplicated between story sections and global requirements

## Gaps & clarify strategy

- [x] **Q1 → (B) LOCKED** — re-cut the `cli → agent` de-coupling into ordered sub-slices; this round = sub-slice 1 (contracts → domain); the construction inversion is sub-slice 2
- [x] **Q2 → (i) LOCKED** — extract **all three** crossing contracts to `internal/domain/agent` (turn-result + incomplete-turn error + the `ToolDefs` projection); CLI production `→ agent` references drop to exactly one (`AgentLoop`)
- [x] **Q3 → (a) LOCKED** — `internal/agent` **references** the moved contracts directly (no alias, no forwarder); one name per concept
- [x] Questions asked **one at a time** (Q1 → Q2 → Q3); capped at 1–3 per round; **all answered**
- [x] High-impact gap was scoped to a single first question (Q1), with options A/B/C; the operator chose **B** (then Q2 → i, Q3 → a)
- [x] Lower-impact undecided details disclosed as assumptions, not escalated (exact domain type names A3; ADR number A4; NOOP set A5/A6)
- [x] No remaining `NEEDS CLARIFICATION` — clarify round 1 **CLOSED** (Q1/Q2/Q3 locked)

## Verifiability & success criteria

- [x] Acceptance scenarios cover the main success path (green on `dev` with the baseline **unchanged**; the id count 4 → 1; an incomplete extraction → 2+; a flipped contract → compile break)
- [x] Success criteria are measurable, verifiable, technology-neutral (SC-001…SC-006)
- [x] Assumptions express premises/boundaries only, no smuggled requirements
- [x] Requirements, edge cases, key entities, and success criteria are mutually consistent

## Issues & corrections

- **Baseline = 2 today** (re-measured 2026-09-18 @ `dev` `12964d6`): `internal/cli → internal/agent`, `internal/cli → internal/ui`. Under **Q1 → B**, this round does **not** move the baseline (the `AgentLoop` construction keeps the `→ agent` edge); sub-slice 2 drops it **2 → 1**.
- **Sizing correction (important)** — ADR 0017 §Forward classifies the **`→ agent`** edge as *"the deepest slice"* and the **`→ ui`** edge as *"the next natural slice"* (itself **not edge-sized**). An earlier shorthand implied the `→ agent` edge was the *smaller* one — that was **wrong**; the re-cut (Q1 → B) follows from it.
- **Round DoD is NOT a baseline move** — it is "contracts domain-owned + surviving coupling reduced to one construction call site + gate green with the baseline **byte-identical**". Recorded up front so no reader expects `2 → 1` here.
- **RULE-A shapes the contract home** (not just RULE-C): `internal/cli` (tier 6) → `internal/domain/**` (tier 0) is downward/sanctioned; `internal/agent` (tier 4) → `internal/domain/**` is downward/sanctioned.
- **R5.1/R5.2/R5.3** — round 047 = R5.1 (RULE-E gate + baseline); round 048 = R5.2 (`→ ui/tui/prompt`); this round is the **re-cut sub-slice 1** of the `→ agent` de-coupling.
- **Witness discipline**: the witness is the **gate + the identifier-count check + unit seams** (NFR-004), never the E2E suite (#92 AC5); falsifiability witnesses (a)/(b)/(c) reproduced then reverted (FR-011).
- **Atomicity** — extraction + truth/ADR land as one PR (NFR-002). **No new Makefile target**.
- **No baseline edit** — `tools/arch/baseline.txt` is byte-identical at delivery (FR-006).

## Ready determination

- [x] Ready to proceed to downstream planning — **clarify round 1 CLOSED** (Q1 → B · Q2 → i · Q3 → a)
- [ ] A high-impact requirement gap must be closed first — **none remaining**

**Note**: plan package final after the clarify fold. Next pipeline step is `/axb-technical-research` (its precondition is the spec; `/axb-spec-by-example` is **NOOP** — no user-facing journey). `/axb-system-analysis` must record this as a dev-surface structural refactor (0 CLI interfaces; api/data/dsl-refine NOOP).
