# ADR 0064 — Dead-code reachability: an advisory `make dead-code` carrier + the exported-dead cleanup

- **Status:** Accepted
- **Date:** 2026-09-27
- **Deciders:** tellme owner (issue [#198](https://github.com/gosharplite/tellme/issues/198))
- **Related:** **ADR 0042** (§D4 — the declined coverage / reachability-orphan tooling, #144; §D5 — *"no `test-coverage`/`dead-code`"*; §D6 — the `verify-fmt`/PATH-prereq pattern this mirrors — **supersedes §D5's `dead-code` wording**) · **ADR 0041** (the advisory-muse lesson — the reason the carrier is filtered, not a pass-through) · **ADR 0012** (dev-tool binaries resolved from PATH; no `go.mod` change) · **ADR 0030**/`modelith-check` (the fail-on-absent *gate* this deliberately diverges from) · retired **`RF-068-1`** (ADR 0038 — the "reachability orphan answered by reasoning" claim this round's measurement falsifies) · `specs/truth/techstack.md` (*Coverage tooling — DECLINED*) · `docs/domain-model/quality.modelith.yaml` (the `advisory` `GatePolicy` + `modelith-drift` precedent) · the reference's heavy `tell-me-go/cmd/deadcode` (**not** adopted)

## Context

tellme ships **no coverage tooling** (ADR 0042 §D4; #144 closed `not planned`): a coverage
*percentage* misleads because the E2E contract runs the built binary as a subprocess. The one class a
profile uniquely adds — a **reachability orphan** — was recorded as "caught by reasoning" (round 068,
`RF-068-1`). The operator's measurement (issue #198, 2026-09-27) **falsifies that**: a whole-program
reachability pass finds a small set of **genuinely-dead exported symbols** the standard `unused`
linter cannot see (it treats any **exported** symbol as used), and tellme has **no carrier** for this
class at all.

Two facts were **unrecorded**: (i) the exported-dead class is invisible to every shipped gate; (ii)
ADR 0042 §D5 said tellme ships **no** `dead-code`, which this round must **supersede**, not ignore.
Reopening a **declined** decision requires **explicit operator intent** (Bootstrap Agent Rule 11
addendum (b)): the operator filed #198 and directed *"Open a new aixbdd round, the goal is to close
#198."* The **carrier form is deliberately NOT** the reference's heavy `cmd/deadcode`.

## Decision

**D1 — Adopt the *vanilla* `golang.org/x/tools/cmd/deadcode` as an advisory PATH dev-tool.** Pinned
**`v0.47.0`** (the `go.sum`-indirect version ⇒ **no** `go.mod`/`go.sum` change); installed like
`modelith`/`golangci-lint`/`govulncheck` (ADR 0012). **Not** the reference's heavy
`tell-me-go/cmd/deadcode` (no `[DEAD]`/`[PRIVATE]` categories, no ports-registry/`verify-exit-query`
machinery).

**D2 — `-test` is mandatory.** Without it, the whole test surface reads as dead (~330 items); with it,
test-reachable symbols stay live (measured: **7** items on `dev` `ae300e9`).

**D3 — The carrier is advisory, never fails, and is NOT a `make verify` member.** `make dead-code`
mirrors the `modelith-drift` shape: it runs, filters, prints, and **exits 0 always**. It adds **no**
gate. (ADR 0041: an advisory proves the tool *works*, not that it *fires*.)

**D4 — Absent tool ⇒ install hint + exit 0.** A deliberate divergence from `modelith-check`
(fail-on-absent): an advisory that failed on an absent optional tool would be a de-facto gate.

**D5 — The FP policy: filter the *interface-conformance-only* class.** Run with `-test`, then filter
the recorded FP class so the steady-state output on a **clean** tree is **empty** and only **new**
findings appear. **Not** a `NonFixCatalog` — **one documented exclusion predicate** (a class). The
measured member is `ui.sharedSource.Suggest` (an interface-conformance-only method pinned by two
compile-time `var _ I = T{}` assertions that guard the identical method sets of `domaintui.Source`
and `prompt.Source` — structural typing the RTA cannot see); the predicate also covers the
documented **unwrap-class** (`…\.Unwrap$`; with `-test` the RTA currently *sees* `errors.Is/As` and
does not report it, but the class is recorded so a future unwrap-only usage is handled).

**D6 — Remove the genuinely-dead exports (the cleanup).** The issue's 7 grep-verified symbols
(`config.(Config).EffectiveUseTUIPrompt`, `openai.NewWithHTTPClient`, `agent.DefaultToolTimeout`,
`di.TokenResolver`, `harness.RunWithStdin`, `harness.RunBinary`, `mcptest.SchemaWithProperty`) **plus
`cli.noopCallObserver`** — a **production** symbol the issue's inventory missed (its only reference is
its own `var _ agentport.CallObserver = noopCallObserver{}`; never instantiated — `compositeObserver`
is built with a real renderer, `cli.go:875`). A truly-dead type is removed, never filtered (filtering
it would hide it).

**D7 — Start from clean.** Remove the dead set **first**, then add the target, so the advisory begins
from the recorded-FP steady state.

**D8 — This ADR supersedes ADR 0042 §D5's "no `dead-code`" wording.** ADR 0042 §D4 (the
**percentage**-coverage decline) **stands**; only the **reachability-orphan** half is answered. ADR
0042's §D4/D5 body is an Accepted ADR and stays **verbatim**; its **index row** gains a forward
pointer to this ADR.

## Consequences

- The exported-dead class gains a **mechanical surface**: `make dead-code` surfaces **new** findings
  on demand, printing nothing on the clean tree.
- It is **advisory**: no `verify` member, no gate change, no automatic guard (tellme has no CI). The
  carrier proves the tool *works*; a human decides whether a finding is dead or an FP.
- The **provenance hazard** is recorded: the PATH name `deadcode` collides with the reference's heavy
  tool; the header comment + this ADR state the required provenance (vanilla x/tools).
- The tree no longer carries the removed dead exports; no behaviour changes (zero callers).
- **No new dependency** (`x/tools` stays an indirect `go.sum` entry); **no new `make verify` member**;
  `go.mod`/`go.sum` unchanged.

## Alternatives considered

| Alternative | Rejected because |
| --- | --- |
| Widen the `unused` linter config (`exported-is-used: false` + `whole-program: true`) | Tested (golangci-lint v2.12.2); it **never** reports an unused **exported** symbol. |
| Adopt the reference's heavy `cmd/deadcode` | `[DEAD]`/`[PRIVATE]` + ports-registry/exit-query machinery is disproportionate; the vanilla tool is the class carrier. |
| A zero-tolerance `make verify` gate | Interface-conformance FPs would force per-round exemptions; the advisory + FP predicate is the proportionate form. |
| A pure pass-through advisory (no filter) | Always prints the recorded FPs — the ADR 0041 advisory-muse failure. |
| Filter (not remove) `cli.noopCallObserver` | It is a **truly-dead** production type (zero instantiations); filtering would hide it. |
| Adopt a coverage-profile integration | Out of scope; the **percentage** stance (ADR 0042 §D4) is untouched. |

## Forward (non-blocking)

> **⚠ Not open work.** Disclosures / deferred items — not tasking.

- **RF-064-1 (advisory, never a gate).** `make dead-code` is **not** a `make verify` member and never
  fails; a future change to that status is a **new decision** (a superseding ADR), not an edit here.
- **RF-064-2 (the FP predicate is class-based, not a catalog).** The predicate is the
  interface-conformance-only class; adding a member is a **recipe edit in the same PR** as the new FP,
  with a recorded rationale — never a silent widening. tellme keeps **no** `NonFixCatalog`.
- **RF-064-3 (provenance is not machine-checked).** The target resolves `deadcode` from PATH; a
  machine with the reference's heavy binary installed under the same name would silently change the
  advisory's meaning. The provenance is guarded by the header comment + this ADR (**inspection**), not
  by a check.
- **RF-064-4 (host-GOOS only).** The analysis runs on the host (`darwin/arm64` here);
  `verify-cross-compile` covers other targets. No cross-platform coverage is claimed.
- **RF-064-5 (structural-typing FPs are inherent).** The RTA cannot see every interface dispatch;
  the FP class is expected to recur and is handled by the predicate, not by deleting load-bearing
  contract pins.
- **RF-064-6 (`docs/domain-model` same-PR).** The round adds the advisory `dead-code` member to
  `quality.modelith.yaml` (ADR 0041 load-bearing rule); no product entity changes.
