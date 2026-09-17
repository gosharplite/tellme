# ADR 0011 — Layer-discipline gate: the pinned layer rule + a fail-on-stale violation baseline

- **Status:** Accepted
- **Date:** 2026-09-17
- **Deciders:** tellme owner
- **Related:** issue [#93](https://github.com/gosharplite/tellme/issues/93) (R1 of the split; the anchor) ·
  umbrella [#92](https://github.com/gosharplite/tellme/issues/92) (the gate-first split) ·
  round 039 review **TD-3** (PR [#81](https://github.com/gosharplite/tellme/pull/81) — the missing layer-discipline gate) ·
  round-020 `verify-cross-compile` (`CROSS_TARGETS`, the host-independent target matrix reused here) ·
  round-032 `verify-mcp-sdk-confinement` (the Makefile-gate precedent) ·
  round 042 (`specs/plans/042-layer-discipline-gate` — this ADR's round) ·
  `docs/decisions/0005-tool-call-log-parity.md` (D1 partitions **rendering**, not **yield** — the R3 obligation) ·
  **ADR-036 parity** + **ADR 0010** (the determinism/verification doctrine the witnesses follow).

## Context

`internal/cli` is both the **application layer** and the **composition root**: `cli.go` (~1236 lines) directly imports five `internal/infrastructure/*` packages and holds eight package-level factory vars; `mcp_discovery.go` imports `internal/infrastructure/di` + `internal/infrastructure/mcp`. Static analysis reports **7 layer violations (0 cycles)**. **No gate in `make verify` checks import direction**, so a *new* illegal cross-layer import sails through every gate today — only a human reviewer would notice.

R1 of [#92](https://github.com/gosharplite/tellme/issues/92) adds the missing gate **as a ratchet**: it must land **today** with the existing violations *baselined* (not fixed — fixing them is R2–R4), so `dev` stays green, a **new** violation fails the standard gate, and the baseline can only **shrink** to 0.

This ADR records the **rule** and the **baseline policy**, because both are decisions other artifacts and future rounds (R2–R4) must be able to **cite** — and because the first draft of this rule read two ways (a one-part "no upward import" statement that could **not** reproduce the 7 `cli → infrastructure` entries it claimed as its baseline). The gate's own **tier table** is the normative machine-readable source of the ranking (see D7); this ADR is the durable narrative + the worked examples.

## Decision

**D1 — The layer ranking.** Governed packages are ranked low → high:

| Tier | Packages |
| --- | --- |
| 0 | `internal/domain/**` (pure) |
| 1 | `internal/config`, `internal/home` (shared utilities) |
| 2 | `internal/app/**` (application) |
| 3 | `internal/infrastructure/**` |
| 4 | `internal/agent` |
| 5 | `internal/ui`, `internal/ui/tui/**` (presentation) |
| 6 | `internal/cli` |
| — | `cmd/**`, `tests/**`, `tools/**` (composition / test-support — **exempt**) |

**D2 — The rule is a two-part predicate** (an earlier one-part "no upward import" statement was **incomplete**: it could not produce the 7 `cli → infrastructure` entries, so the baseline was not derivable from the rule — corrected here):

- **RULE-A (direction):** a governed package MUST NOT import a package of a **strictly higher** tier.
- **RULE-B (application target rule):** the **application tiers** — `internal/app/**` and `internal/cli` — MUST NOT import `internal/infrastructure/**`. *(This is the rule issue [#93](https://github.com/gosharplite/tellme/issues/93) mechanised as "`internal/cli`: domain + stdlib + application utilities only".)*
- **RULE-C (domain purity):** `internal/domain/**` MUST import only `internal/domain/**` + stdlib.
- **RULE-D (default-deny):** an `internal/**` package matching **no** tier is a **violation** ("unranked governed package") — the gate never treats "unknown" as "allowed". Non-`internal` trees are the explicit exemption list (`cmd/**`, `tests/**`, `tools/**`).

**Worked examples — the 8 baselined violations (measured at `dev` `3f8ec08`; re-measured from the gate at implementation):**

```
internal/agent -> internal/ui                        # RULE-A (tier 5 > 4)
internal/cli -> internal/infrastructure/di           # RULE-B
internal/cli -> internal/infrastructure/history      # RULE-B
internal/cli -> internal/infrastructure/llm          # RULE-B
internal/cli -> internal/infrastructure/mcp          # RULE-B
internal/cli -> internal/infrastructure/skills       # RULE-B
internal/cli -> internal/infrastructure/telemetry    # RULE-B
internal/cli -> internal/infrastructure/tools        # RULE-B
```

The 7 `cli → infrastructure` entries are the **downward** imports RULE-B forbids; RULE-A alone would leave them legal. The 8th is the upward `agent → ui` edge. **The 8 are therefore derivable from the two-part rule.** The 7 are **R2**'s to remove and the 8th is **R3/R4**'s, so the count reaches 0 across R2–R4 (how [#92](https://github.com/gosharplite/tellme/issues/92) AC1 is read).

**D3 — The baseline is a committed, sorted file; it is a ratchet.** The baseline lists the known violations, one per line, in the gate's own output form. Behaviour: a violation **not** in the baseline **fails** (a new violation); a baseline entry that no longer violates (**stale**) **fails** with *"remove it from the baseline"*; a violation removed from the baseline while still present **fails**. The baseline can only shrink toward truth.

**D4 — The enumeration is anchored to the module root, and the graph is self-tested.** `go list ./...` is **CWD-relative**, and a Go test binary runs with **CWD = its package directory**, so a guard that shells `go list ./...` from a test would enumerate only its own subtree (⇒ an empty graph ⇒ either all entries stale, or a vacuous green). The guard therefore resolves the module root once and runs `go list` as a child with `cmd.Dir = <moduleRoot>`; its self-test **asserts the enumeration itself** (expected package count + the presence of known governed packages) **before** any ranking assertion, asserts the child's exit status and non-empty output, and passes **no** `-e` (an error must be an error, never an empty graph).

**D5 — Host-independence is evaluated as the union over `CROSS_TARGETS`.** `go list {{.Imports}}` resolves in the **host** build context: files excluded by `GOOS`/`GOARCH` land in `.IgnoredGoFiles` and their imports vanish, so a single-context walk is host-relative. The guard evaluates the **union** over the project's supported targets — `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` (round 020's `CROSS_TARGETS`) — each child `go list` run with the **inherited environment filtered**: a **drop/neutralise set** (`GOOS`/`GOARCH`/`GOARM`/`CGO_ENABLED` set or cleared per target; `GOFLAGS`/`GO111MODULE`/`GOEXPERIMENT`/`GOWORK` neutralised) so an ambient build export (a CI image's `GOFLAGS=-mod=vendor`, `GOWORK=…`, or `GO111MODULE=off`) cannot fail the gate spuriously, **and a preserve set** (`PATH`/`HOME`/`GOPATH`/`GOMODCACHE`/`GOCACHE`) so the warm module cache is still found (an `env -i` child fails with *module cache not found*) — so an OS-gated illegal import cannot hide and the single committed baseline is genuinely host-free.

**D6 — Open (`//go:build` tag-gated) files are out of scope, recorded.** Passing no `-tags` to the child means tag-gated files are invisible to `go list`; and a `go test -tags=…` invocation does **not** propagate its tag to a child `go list` (only `GOFLAGS`/env does). The only custom-tagged file in the module is the **guard itself** (`//go:build arch`), which lives in the exempt `tools/**` tree; the project's other conditional compilation is GOOS-based, which D5 covers. So **custom-tag-gated imports are a recorded out-of-scope residual** — the earlier "respects build tags" claim is **withdrawn**.

**D7 — One normative ranking source + a consistency assertion.** The gate's embedded **tier table** is the single normative source of the ranking; `specs/truth/techstack.md` states the rule as a short **predicate** and **cites** this ADR + the table; a guard self-test asserts the table classifies **every** governed package (RULE-D), so the prose and the machine source cannot silently diverge.

**D8 — Acyclicity is asserted, not assumed.** The gate additionally asserts **0 same-tier/any-tier cycles** (a stdlib SCC/Tarjan pass over the governed graph). Cycles have **no baseline** — the count must be 0 (it is 0 today). This makes issue [#93](https://github.com/gosharplite/tellme/issues/93)'s AC4 ("0 violations, 0 circular references") a real assertion rather than a measurement: a **direction-only** rule can never detect a same-tier cycle, so the separate SCC pass is required.

**D9 — Baseline format is deterministic.** Lines are **sorted in Go** (`sort.Strings`, byte-wise — never via shell `sort`, whose collation is locale-dependent), use **module-relative package paths** (the `github.com/gosharplite/tellme/` prefix stripped, matching the gate's own output), and an **ASCII** delimiter ` -> `. The baseline is **generated from the gate**, never transcribed.

**D10 — What the rule is *not*.** RULE-B binds only the **application tiers** (`internal/app/**` = tier 2, `internal/cli` = tier 6). `internal/agent` (tier 4) and the other non-application tiers may still import `internal/infrastructure/**` (a **downward** import, legal under RULE-A) — RULE-B is **not** a blanket "nothing above infrastructure may reach into infrastructure". A future round (R3/R4) that finds this too loose amends this ADR with a widened RULE-B, rather than assuming it already holds. Likewise RULE-A forbids only imports of a **strictly higher** tier — **same-tier** and **downward** imports are out of scope (the SCC pass in D8 covers cycles separately).

## Alternatives considered

1. **A one-part "no upward import" rule (the first draft)** — rejected: it cannot reproduce the 7 `cli → infrastructure` entries, so `fail-on-stale` would red the gate on the first run and R2's workstream would be unenforced.
2. **Narrow target-rule only** (`cli`/`ui`/`app` ↛ `infrastructure` + domain purity; baseline 7) — rejected: it ignores the `agent → ui` edge the operator chose to flag (clarify Q1/Option 2).
3. **Baseline restated to 1 (upward-only)** — rejected: R2's DoD ceases to be gate-backed; the issue's stated purpose is unmet.
4. **Warn-on-stale baseline** — rejected: the baseline silently rots; "→ 0" loses its teeth.
5. **Makefile `grep` guard** — rejected: cannot resolve package tiers or build-tag context; would mis-rank nested packages.
6. **`golang.org/x/tools/go/packages`** — rejected: a new dependency for no gain (stdlib `go list` suffices).
7. **Ship the absolute "host-independent" claim without the `CROSS_TARGETS` union** — rejected: true for today's tree, false for the mechanism (an OS-gated upward import would go red on one host, green on another, against a single baseline).
8. **Defer the acyclicity assertion** — rejected: AC4 names it; an SCC pass is cheap and keeps the criterion non-vacuous.

## Consequences

- **New:** a `-tags=arch` Go guard under `tools/arch/` + a committed `tools/arch/baseline.txt` (8 entries), invoked by a new `Makefile` target `verify-architecture` that is a member of `make verify` (and named in `make help` + `.PHONY`). `specs/truth/techstack.md` (Build & Tooling) records it, and its **Task runner** row gains the member.
- **No product code**, **no new dependency** (`go.mod`/`go.sum` unchanged); POSIX-only; **hermetic with a warm module cache** (the `verify` aggregate runs it after `vet`/`verify-cross-compile`, which warm the cache; a direct `make verify-architecture` on a cold cache resolves modules like any build — no new module, no network *service*).
- **`-tags=arch` is not compiled by any other gate** (`go vet ./...`, `verify-cross-compile`): a compile error in the guard surfaces only when the gate runs — accepted, recorded so a future "why isn't this vetted?" has an answer.
- **Recorded divergences:** the reference ships a `verify-architecture` Go guard **plus** `modelith-layers`; tellme adopts the Go-guard form and has **no** modelith toolchain (no `modelith-layers` analogue). The reference's guard is also build-tagged `arch` — the form is shared.
- **Scope / forward items:** custom build-tag-gated imports are out of scope (D6); the topology audit's stale-semantics blind spot is a recorded forward item on [#92](https://github.com/gosharplite/tellme/issues/92); the baseline entries are removed by R2 (the 7) and R3/R4 (the 8th).
- **Atomic delivery:** the gate + its baseline + the Makefile wiring land in **one** PR; the baseline is generated from the shipping gate at the merged head (round-040 TD-1 precedent).
- **Witnesses:** a newly added illegal import ⇒ red; a fixed-but-baselined entry ⇒ red (stale) — reproduced, then reverted (ADR 0010 doctrine).
