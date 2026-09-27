# Requirements checklist — round 094 `094-dead-code-hygiene`

**Spec**: [`spec.md`](../spec.md) · **Anchor**: [#198](https://github.com/gosharplite/tellme/issues/198)

## Content completeness

- [x] All mandatory sections present (why / change / user stories / requirements / edge cases /
  invariants / success criteria / assumptions).
- [x] Theme, scope, and main flow are clear (dead-code cleanup + an advisory carrier).
- [x] Requirements are stated as outcomes, not implementation trivia (the tool pin + Makefile shape
  live in `research.md`/`plan.md`).
- [x] Edge cases cover the high-risk shapes (absent tool, new finding, provenance collision).
- [x] Key entities (`QualityGate` advisory member) and success criteria are complete.

## User stories & requirement ownership

- [x] Stories ordered by value (both P1; US-1 cleanup, US-2 carrier).
- [x] Each story is independently verifiable (cleanup = grep + gates; carrier = the target's run).
- [x] Each story carries its acceptance shape.
- [x] FRs are attached to a story; no duplication between story and global sections.

## Clarification strategy

- [x] No high-impact gap ⇒ `/axb-clarify` **not** escalated (the two decisions are operator-locked).
- [x] No `NEEDS CLARIFICATION` remains.

## Verifiability & success criteria

- [x] Acceptance shape verifies the main path (clean tree prints nothing; injected export prints).
- [x] Success criteria are measurable and tech-neutral (grep, exit codes, `modelith-check`).
- [x] **Every** normative entry (FR / NFR / SC / EC) carries a Verification Intent.
- [x] Assumptions express premises only (the Rule-11 intent + the two locked decisions).

## Consistency

- [x] `spec.md` §2/§3 (the re-verified inventory) are consistent with the issue's §inventory +
  the recorded divergence (`cli.noopCallObserver`).
- [x] `spec.md` §5 (FR-2/FR-3) is consistent with §4 (US-2) and the issue's carrier spec.
- [x] `spec.md` §7 (I-4: not-a-catalog) is consistent with the quality model's `no-nonfix-catalog`.

## Ready

- [x] Ready for `/axb-spec-by-example` (**NOOP** — a Make target is not godog-drivable) and
  `/axb-technical-research`.

**Note**: `/axb-spec-by-example` is **NOOP** (no product-observable behaviour change; a Make target
is not an executable Gherkin surface). `/axb-dsl-refine` is **NOOP** (no `.feature` change). The
witnesses are the target's own run + the gate suite (a Make target, not a CLI contract).
