# ADR 0016 — Application import-ceiling gate (RULE-E): the application tiers' downward-import allow-list

- **Status:** Accepted
- **Date:** 2026-09-18
- **Deciders:** tellme owner
- **Related:** issue [#101](https://github.com/gosharplite/tellme/issues/101) (R5 of the split; the anchor) ·
  umbrella [#92](https://github.com/gosharplite/tellme/issues/92) (**AC2** second clause) ·
  round 047 (`specs/plans/047-application-import-ceiling-gate` — this ADR's round) ·
  **ADR 0011** (the layer-discipline gate: the tier table this rule extends, the baseline ratchet it reuses) ·
  **ADR 0013** (composition-root injection — the pattern that keeps the application tiers thin) ·
  **ADR 0010** (the falsifiability-witness doctrine) · round-040 **TD-1** (never land a stricter rule before its enablers).

## Context

R1 of [#92](https://github.com/gosharplite/tellme/issues/92) (ADR 0011) mechanised the **first** clause of [#92](https://github.com/gosharplite/tellme/issues/92) **AC2** — *"`internal/cli` imports **no** `internal/infrastructure/*`"* — as **RULE-B** (a target rule) and ratcheted its baseline to **0** across rounds 042/044/046. AC2's **second** clause — *"it depends only on `internal/domain/*`, stdlib, and application utilities"* — is a **much stronger** claim that RULE-B cannot express: RULE-B forbids one **target**; it does **not** constrain a *downward* import to any **other** internal package.

So today `make verify` is silent about, e.g., a new `internal/cli → internal/agent` or `internal/cli → internal/ui` import (both packages exist and are **ranked**) — neither violates RULE-A (not upward), RULE-B (not `infrastructure`), RULE-C (not domain), or RULE-D (ranked). Only a human reviewer would notice. *(A brand-new **unranked** package such as a fresh `internal/telemetry` would already be caught by RULE-D — so the gap RULE-E closes is precisely the **ranked-but-unsanctioned** target; the exemplar is chosen accordingly.)* [#101](https://github.com/gosharplite/tellme/issues/101) records this as the R5 ambition and names its precondition: a **new rule** plus a **new ADR**, shipped **with its enablers**.

This ADR records the rule (**RULE-E**) — an application-tier **import ceiling** — and its policy, because future rounds (the R5 de-coupling slices) must be able to **cite** it, and because the rule is a **positive allow-list**, mechanically distinct from ADR 0011's four rules.

## Decision

**D1 — RULE-E (the application import ceiling).** For a **governed application tier** — `internal/app/**` (tier 2) and `internal/cli` (tier 6) — an import of an **`internal/**` package** is a **violation** unless the imported package is in the **sanctioned set**:

| Sanctioned (per application tier) | Why |
| --- | --- |
| `internal/domain/**` | the pure domain (tier 0) |
| `internal/config`, `internal/home` | tier-1 shared utilities (ADR 0011 **D1**) |
| `internal/app/**` | tier-2 application utilities (incl. the injected `deps` seam) |

**`stdlib`** is allowed by construction. Violations are emitted in the gate's existing edge form (`<source> -> <import>`, module-relative, ASCII ` -> `) and are **deduped by edge** with RULE-A/B/C — an edge that breaks several rules is one line.

**D2 — RULE-E binds both application tiers.** `internal/app/**` and `internal/cli` share the **one** sanctioned set. *(Match to RULE-B's target set — the application tiers — per [#101](https://github.com/gosharplite/tellme/issues/101)'s own RULE-E sentence.)* Consequence: `internal/cli → internal/app/**` (the injected `deps`/`suggestions`) and `internal/app/** → internal/config` are **sanctioned**.

> **Live reach (recorded — TD-3).** With today's tier table + sanctioned set, RULE-E has **no reachable firing from `internal/app/**`**: every unsanctioned `internal/**` target an app-tier package could name is either **upward** (RULE-A), **unranked** (RULE-D), or `internal/infrastructure/` (tier 3 > 2, RULE-A). RULE-E's live teeth today are therefore **entirely `internal/cli` → tiers 4/5** (the 3 residuals). "Binds both application tiers" is scope-true; the *reach* is asymmetric, stated here so a later round does not over-credit RULE-E on the app side. (The emitted line is rule-agnostic by design — D1 / dedup — so an app-tier upward edge reads as RULE-A, not RULE-E.)

**D3 — The sanctioned set is normative in the guard's tier table; default-deny.** The set is a normative section of the **embedded tier table** (ADR 0011 **D7** — the single normative machine-readable source). Default-deny: a governed application-tier `internal/**` import matching **no** sanctioned prefix is a violation; the rule never treats "unknown" as "allowed".

**D4 — Fail-on-stale allow-list.** The guard's self-test asserts **every** sanctioned entry is **used** — imported by at least one governed application-tier package. An **unused** sanctioned entry **fails** the gate. This is the deliberate symmetry with ADR 0011 **D3**: a lax allow-list rots exactly as a stale baseline does, so the sanctioned set must also shrink to truth.

> **Two independent nets (stated explicitly).** A **lone widening** of the allow-list (e.g. sanctioning `internal/ui`) is caught **twice**: by the synthetic predicate self-test (the widened entry admits an edge the `want`-list expects to be a violation) **and** by the stale-baseline ratchet (the now-sanctioned edge leaves a baseline line stale). That two-net property — not the coverage assertion alone — is RULE-E's real safety story; the coverage assertion's *own* logic carries a dedicated synthetic witness (`selfTestAllowList`), so gutting it reds the gate (fold **F-2**).

**D5 — The residual edges are baselined; it is the one fail-on-stale ratchet.** RULE-E lands **with** the committed `tools/arch/baseline.txt` extended by the currently-unsanctioned edges — today **3**: `internal/cli -> internal/agent`, `internal/cli -> internal/ui`, `internal/cli -> internal/ui/tui/prompt` — generated from the gate (never transcribed), sorted, module-relative, ASCII ` -> `. The ratchet is unchanged (ADR 0011 **D3**): a violation not listed **fails**; a **stale** entry **fails**; an entry removed while its violation persists **fails**. RULE-A/B/C stay **0**; the R5 de-coupling slices remove the 3 entries one-for-one until the baseline is header-only again.

**D6 — What the rule is *not*.** RULE-E constrains only a governed application tier's imports of **`internal/**`** packages. It does **not** constrain: (i) **stdlib** imports; (ii) **third-party module** imports — **out of RULE-E's scope** and a **live recorded residual**: `internal/cli` already imports `github.com/spf13/pflag` and `golang.org/x/term` (both production, tier 6), so a future *third-party ceiling* for the application tiers (a `go.mod`/module-path-based rule) has **concrete motivation, not a hypothetical one**; (iii) imports made by **non-application** tiers (`internal/agent`, `internal/ui`, `internal/infrastructure/**`, `internal/config`/`home` — governed by RULE-A/B/C); or (iv) `cmd/**`, `tests/**`, `tools/**` (exempt trees). RULE-E is **not** a blanket "nothing may import anything but the domain".

**D7 — RULE-E is a fifth, independent rule; ADR 0011 is extended, not rewritten.** RULE-A (direction) · RULE-B (application target rule) · RULE-C (domain purity) · RULE-D (default-deny coverage) remain in force, unchanged, as recorded in ADR 0011. RULE-E is a distinct kind — a **positive ceiling** on downward imports. No ADR is superseded (ADR 0011 is `Accepted` and immutable except its `Status` line/index). The gate therefore enforces **five** rules; the baseline is one sorted edge list over all of them.

**D8 — Mechanism is inherited, not re-derived.** RULE-E reuses the guard's whole machinery with **no** re-litigation: the **module-root-anchored** `go list` (ADR 0011 **D4**); the **`CROSS_TARGETS` union** with the filtered child env (**D5**); the **merged production+test** graph for the rule and the **production-only** SCC pass for cycles (**D8**); the **deterministic** sorted baseline (**D9**); and the normative tier table (**D7**). RULE-E adds **no** Makefile target — it rides `verify-architecture` (already a member of `make verify`); the `-count=1` invocation remains load-bearing.

**D9 — The residue is temporary by construction.** The 3 baselined edges are the R5 de-coupling workstream ([#101](https://github.com/gosharplite/tellme/issues/101)); at the terminal state (baseline header-only) RULE-E has **no release valve** — a future legitimate ceiling violation must be fixed by refactor (or this ADR amended), the intended terminal ratchet state (ADR 0011, round-046 record).

**Worked example — the RULE-E violations at delivery (measured 2026-09-18 @ `dev` `b42f868`):**

```
internal/cli -> internal/agent                       # RULE-E (unsanctioned)
internal/cli -> internal/ui                          # RULE-E (unsanctioned; also the residual `cli` presentation use)
internal/cli -> internal/ui/tui/prompt               # RULE-E (unsanctioned)
```

`internal/app/**` has **no** RULE-E violation (it imports only `domain` + the sanctioned `config`); `internal/cli → internal/config|home|app/**` and `internal/app/deps → internal/config` are **sanctioned**. ⇒ **RULE-E baseline = 3**, re-measured from the gate at implementation.

## Alternatives considered

1. **Widen RULE-B to the allow-list** — an Accepted ADR is immutable except its `Status` line/index (ADR 0011's own convention); RULE-B's narrow statement remains truthful — rejected; a separate rule + ADR is the established pattern.
2. **A negative rule** ("application tiers must not import anything above the domain bar `X`") — cannot express "domain + stdlib + application utilities only" (AC2), and cannot catch a *sanctioned-tier-but-unsanctioned-package* import — rejected.
3. **A separate committed allow-list manifest** (`tools/arch/allowlist.txt`) — a **second** normative source (ADR 0011 **D7** tension) — rejected; the tier table is the one source.
4. **An edge-level hardcoded allow-list in the test** — no machine-readable home for the coverage assertion (D4) — rejected.
5. **Land RULE-E only when the code is compliant (no baseline)** — forces the entire 6-seam de-coupling into one round; [#101](https://github.com/gosharplite/tellme/issues/101) calls R5 "a multi-round programme"; violates round-040 TD-1 (never a stricter rule before its enablers) — rejected.
6. **Bind `internal/cli` only** (the literal AC2 reading) — leaves `internal/app/**` unconstrained, so RULE-E and RULE-B would govern different sets — rejected.
7. **Extend RULE-E to third-party imports** (a "sanctioned modules" list) — dependency governance is out of R5's scope and needs `go.mod` parsing — rejected; recorded as a residual (D6).
8. **Advisory / warn-only ceiling** — an allow-list that never fails enforces nothing — rejected.

## Consequences

- **New:** RULE-E in `tools/arch` (`//go:build arch`) — a sanctioned-set section in the normative tier table, the rule evaluation, and a coverage self-test (default-deny + fail-on-stale allow-list); the committed `tools/arch/baseline.txt` gains **3** RULE-E edges (header comment updated). No new Makefile target; `verify-architecture` (already in `make verify`) enforces it.
- **No product code**; **no new dependency** (`go.mod`/`go.sum` unchanged); POSIX-only; hermetic with a warm module cache (unchanged from ADR 0011).
- **Truth/governance:** `specs/truth/techstack.md` (Build & Tooling → **Layer-discipline gate** row) is a real **MODIFY** (names RULE-E + the sanctioned set + the residue ratchet); the **Task runner** row is unchanged (`verify-architecture` already listed). This ADR + its `docs/decisions/README.md` index row are added.
- **Recorded residuals:** third-party application-tier imports (**live** — `internal/cli` imports `github.com/spf13/pflag` + `golang.org/x/term`; D6); custom build-tag-gated imports (ADR 0011 **D6**); the 3 edges are the R5 de-coupling workstream; F-4/F-6/F-7/F-8 ([#101](https://github.com/gosharplite/tellme/issues/101)).
- **Governance of the sanctioned set (TD-4).** Staleness is machine-checked (D4); **scope drift is not** — a widening *plus* the removal of the affected baseline line in one commit is indistinguishable from legitimate progress. A **lone** widening is caught twice (D4's two nets), so the residual is narrow. Precedent ADR 0011 **D7** ("a tier-table edit is review-governed"): **a sanctioned-set edit is review-governed — cite ADR 0016 D1, or amend this ADR.**
- **Forward item:** the sanctioned set may be **re-ruled** by a later round (e.g. `config`/`home`/`suggestions` become injected plain data) — a sanctioned-set edit + an ADR 0016 amendment, not this round's. Separately, because RULE-E is a **fifth, independent** rule (D7), a later round MAY **split** RULE-E into its own `specs/truth/techstack.md` row (`Application import ceiling (RULE-E)`) rather than sharing the Layer-discipline gate row — defensible, but deliberately **not** taken this round to keep the truth diff minimal. *(Homed here, on a live surface, per the round-035 **G3** lesson: a deferred option must not live only in a frozen plan package.)*
- **Witnesses:** a new unsanctioned application-tier import ⇒ red; an unused sanctioned entry ⇒ red (fail-on-stale allow-list); a stale baseline line ⇒ red — reproduced then reverted (ADR 0010). The E2E suite is regression, **not** the acceptance carrier (#92 **AC5**).
