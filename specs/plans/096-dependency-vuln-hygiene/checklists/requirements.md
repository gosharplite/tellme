# Requirements checklist — round 096 `096-dependency-vuln-hygiene`

**Spec**: [`spec.md`](../spec.md) · **Anchor**: none (operator-value round)

## Content completeness

- [x] All mandatory sections present (why / change / user stories / requirements / edge cases /
  invariants / success criteria / assumptions).
- [x] Theme, scope, and main flow are clear (a dependency/toolchain vulnerability-hygiene bump).
- [x] Requirements are stated as outcomes, not implementation trivia (the `go get`/`go mod tidy`
  recipe + the truth placement live in `research.md`/`plan.md`).
- [x] Edge cases cover the high-risk shapes (unpatched toolchain, a newer-than-floor version, the
  `go`-directive normalisation).
- [x] Key entities (dependency floor, toolchain floor) and success criteria are complete.

## User stories & requirement ownership

- [x] Stories ordered by value (both P1; Story 1 = the module bump, Story 2 = the recorded floor).
- [x] Each story is independently verifiable (bump = `go.mod` diff + the gate run; floor = a grep of
  the truth clause).
- [x] Each story carries its acceptance shape.
- [x] FRs are attached to a story; no duplication between story and global sections.

## Clarification strategy

- [x] No high-impact gap ⇒ `/axb-clarify` **not** escalated (the directive locks the goal; residual
  choices are RD-owned).
- [x] No `NEEDS CLARIFICATION` remains.

## Verifiability & success criteria

- [x] Acceptance shape verifies the main path (the `vulncheck` gate reaches "No vulnerabilities
  found."; the E2E count is unchanged).
- [x] Success criteria are measurable and tech-neutral (`go.mod` diff, gate exit, E2E figures).
- [x] **Every** normative entry (FR / NFR / SC / EC) carries a Verification Intent.
- [x] Assumptions express premises only (the literal-floor reading, the environment-only scope).

## Consistency

- [x] `spec.md` §1 (the red-gate facts) is consistent with §2 (the surfaces) and §5 (the criteria).
- [x] `spec.md` §3 FR-004/FR-005 (the single truth clause) is consistent with §7 I-2 and §5 SC-004.
- [x] `spec.md` §6 EC-001 (green only on a patched toolchain) is consistent with §1 and §8 A1.

## Ready

- [x] Ready for `/axb-spec-by-example` (**NOOP** — a dependency bump has no product-observable
  behaviour change) and `/axb-technical-research` (the truth owner for the toolchain-floor clause).

**Note**: `/axb-spec-by-example` is **NOOP** (no Gherkin-observable behaviour change). `/axb-dsl-refine`
is **NOOP** (no `.feature` change). The witnesses are the relocated gate run (`govulncheck` red → green),
the `go.mod`/`go.sum` diff, and the truth grep (a build-gate outcome, not a CLI contract).
