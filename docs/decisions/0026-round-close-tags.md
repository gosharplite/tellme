# ADR 0026 — Round-close tags: an annotated `round-NNN` on the propagation merge

- **Status:** Accepted
- **Date:** 2026-09-19
- **Deciders:** tellme owner
- **Related:** [SESSION-CLOSEOUT.md](../SESSION-CLOSEOUT.md) (Step 7 + Rule 15 — where the tag is created), [SESSION-BOOTSTRAP.md](../SESSION-BOOTSTRAP.md) (the frozen-package rule the tag annotates), the aixbdd `plan-package-frozen` / `fresh-package-per-round` invariants; round 056 (`specs/plans/056-mcp-tool-call-reason`, PR [#122](https://github.com/gosharplite/tellme/pull/122)) — the round this convention was adopted with, and its first tag.

## Context

The operator asked, after round 056 closed out: *"we just completed 056 — when and how should a git `tag` '056' be created?"* Grounding showed the repo has **no tag convention at all**: `git tag -l` and `git ls-remote --tags origin` are both empty, neither `SESSION-BOOTSTRAP.md` nor `SESSION-CLOSEOUT.md` mentions tags, and the only "version" in play is the `Makefile`'s `VERSION ?= dev` (`-ldflags -X main.version=…`) — `tellme --version` prints `dev`. So this was a **gap to decide**, not a rule to apply.

Two distinct notions were in play and must not be conflated:

- **A round** is a plan package (`specs/plans/NNN-<slug>/`) — an *internal* iteration container. The domain model defines exactly one lifecycle for it: `active → delivered`, and **`plan-package-frozen`** freezes its artifacts on delivery. Rounds are not versions.
- **A release** is a *product* event — a distributable identified by a version, feeding `--version` and `go install …@vX.Y.Z`. `main` is the repo's "stable/released line", but nothing has ever been released from it as a versioned artifact.

The operator chose the **round-close marker** convention (not a release scheme): mark each delivered round in history.

## Decision

**D1 — Adopt annotated round-close tags, named `round-NNN`.** The name is the plan package's zero-padded three-digit number prefixed with `round-` (so `056`'s tag is **`round-056`**). The `round-` prefix is deliberate: a bare `NNN` reads like a version and invites the version confusion this ADR rules out.

**D2 — The tag points at the `dev → main` propagation merge commit.** That no-ff merge is the first state on `main` that contains the round; tagging it makes `round-NNN` mean "the round, as landed on the stable line". (Not the round's `dev` head — the tag is about the round's *landing*, and `main` is the landing line.)

**D3 — Created at `SESSION-CLOSEOUT.md` Step 7, after propagation is verified.** The prerequisite is the propagation check the closeout already performs — `main^{tree} == dev^{tree}` — so a tag can never mark a half-propagated round. Tagging earlier (at the `dev` merge) is wrong (it would not point at the main-side state); tagging later adds nothing.

**D4 — Operator-approved, like the propagation itself.** Pushing a tag is a **publish** action of the same class as merging/pushing `main`, so it sits behind the same gate as closeout Rule 8: run it only when the round is green, propagated, **and the operator approves**. An agent may prepare and run the command, but it never unilaterally publishes a tag.

**D5 — Annotated, never lightweight.** The tag carries a tagger, a date, and a message naming the round slug, the PR, and the ADR — e.g. `round-056 — 056-mcp-tool-call-reason: MCP tool-call reason envelope + the universal "no reason, no go" gate (ADR 0025); PR #122; folds #121`.

**D6 — Immutable.** A pushed `round-NNN` tag is never moved or deleted. If it is wrong, that is an operator decision recorded with the correction (and, if a round were mis-tagged, a re-point note) — history is relocated/annotated, never silently rewritten. Consistent with the repo's "frozen history" discipline: the tag is the round's permanent marker.

**D7 — Not a version, not SemVer.** `round-NNN` implies **no** release, no artifact, and no version scheme: `tellme --version` stays **`dev`**, `VERSION` stays unset by default, and nothing is published. A real release scheme is a **separate** decision (its own round) and would use `vX.Y.Z` — round tags and version tags are different objects with different purposes.

**D8 — Scope: `main` only, one tag per round.** Exactly one `round-NNN` tag per delivered round, on `main`. The `dev` line is not tagged (it is integration, not history). Earlier rounds are not retro-tagged (forward-only; see Forward items).

## Consequences

- The delivered-rounds index in `STATUS.md` gains a navigable git anchor: `git show round-056` / `git describe` resolve a round to its exact landed state — a durable complement to the frozen plan package and the fold ledger.
- The closeout gains a small, explicit, approval-gated publish step; a round is "done" only after delivery (human merge) **and** propagation, which is exactly when the tag is created.
- The round-vs-version boundary is now written down, so the question does not recur — and it is recorded that `--version` deliberately remains `dev`.
- No code, no dependency, no truth change; `specs/truth/**` is untouched (this is a governance/process decision). The tag itself is created by the operator's closeout, not by the round's implementation.

## Forward items

- **RF-026-1** — a **release scheme** (SemVer tags + `VERSION` wiring + a build/release pipeline) if the project ever ships versioned artifacts; out of scope here by D7.
- **RF-026-2** — **tag signing** (`git tag -s`) if provenance becomes a requirement; the convention is agnostic (annotated is the floor).
- **RF-026-3** — **earlier rounds (001–055) are untagged**; a one-time retro-tag is possible (`round-NNN` on each propagation merge, discoverable from `main`'s merge history) but is deliberately **not** done now — the forward-only record is in the fold ledger + the archives.
- **RF-026-4** — rejected alternative, recorded: tagging the round's **reviewed `dev` head** (the certified fold head) instead of the propagation merge. Rejected because the tag is about the round's *landing on the stable line*, and a `dev`-head tag would not be on `main` and would be ambiguous once `dev` moves on.
