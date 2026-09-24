# Requirements checklist — round 086 `086-tools-listing-line`

## Readiness

- [x] A fresh `PlanPackage` (`specs/plans/086-tools-listing-line/`) — `fresh-package-per-round`.
- [x] The `Spec` is PM-authored / operator-locked (the notation, placement, and the three sub-decisions).
- [x] Anchor: **operator request** (no anchor issue — the rounds 073/074/075/078 precedent).
- [x] Clarify **not escalated** — the operator locked every high-impact choice in-session (A3); no
      requirement gap remains that would change the story split or the acceptance logic.
- [x] Every normative clause (FR / NFR / EC / SC) carries a **Verification Intent** (observable /
      unobservable tier) — the downstream Claim→Witness sweep sees a complete set.

## Requirements → judgeable carriers

| # | Requirement | Judgeable by |
| --- | --- | --- |
| FR-001 | `[TOOLS] - M (N calls)` between `[USER]` and `[MODEL]` | E2E `the listing reports each turn's tool activity` + the unit byte pin |
| FR-002 | printed for every turn incl. `(0 calls)` | E2E (plain exchange ⇒ `(0 calls)`) + the unit byte pin |
| FR-003 | its own block (blank line each side) | the unit byte pin + the separator Then |
| FR-004 | whole label yellow on a terminal stdout | E2E `the listing accents the tool line in yellow` + the unit pin |
| FR-005 | plain under `-r` / redirected | E2E (no-accents Then) + `TestListingRawSuppressesColour` |
| FR-006 | `-l N` = last N messages (rider) | the existing `tellme lists the last {count} messages` Then (unchanged) |
| FR-007 | no tool step content surfaced | the modified `tellme lists the tools' activity but not their contents` row |
| FR-008 | not in `-t`/turns.log nor the chrome | the existing `-t` Thens + a `turns.log` grep pin |
| FR-009 | count = `len(Steps)` (not `Calls`) | the CLI wiring pin `TestRenderHistoryListProjectsToolCount` |
| NFR-001 | offline | the existing `tellme sends no request to any provider` Then |
| NFR-002 | no new dependency | `git diff go.mod go.sum` empty + `make lint` |
| EC-001 | bare `[TOOLS] (N calls)` at M ≤ 0 | the unit pin (non-history caller) |
| EC-002 | partial listing keeps the line | E2E `A partial listing keeps the answer's tool line` |
| EC-003 | legacy line with no `steps` ⇒ `(0 calls)` | E2E (plain exchange) + the unit pin |
| SC-001 | counts correct | the E2E tool-activity Example |
| SC-002 | message shape unchanged | the existing count/selection Thens (unchanged) |
| SC-003 | gates green; no dep change | `make check` + `git diff go.mod go.sum` |

## Boundaries

- **In**: the `-l` listing's `[TOOLS]` line (adapter + `ListingMessage.ToolCount` + the CLI
  projection); the listing truth rows/feature; ADR 0057; the `techstack.md` / `docs/domain-model/**`
  records; the E2E + unit carriers.
- **Out**: `-b`/`--back`, the stores, `--new`, `-t`, the turn chrome, the answer path, `-l N`
  semantics, the reference's per-call `[Tool Call]`/`[Tool Response]` lines (**RF-073-4**),
  `go.mod`/`go.sum`.

## Verdict

**Ready** — no `NEEDS CLARIFICATION`; the falsifiable witnesses exist on day one (the byte pin on the
line and its placement; the yellow-accent pin; the count-from-`Steps` wiring pin; the no-content
negative).
