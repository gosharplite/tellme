# ADR 0064 — Dead-code reachability: an advisory `make dead-code` carrier + the exported-dead cleanup

- **Status:** Accepted
- **Date:** 2026-09-27
- **Deciders:** tellme owner (issue [#198](https://github.com/gosharplite/tellme/issues/198))
- **Related:** **ADR 0042** (§D4 — the declined coverage / reachability-orphan tooling, #144; §D5 — *"no `test-coverage`/`dead-code`"*; §D6 — the `verify-fmt`/PATH-prereq pattern this mirrors — **supersedes §D5's `dead-code` wording**; **§D4's percentage-coverage decline stands and its reachability-orphan half is answered here**) · **ADR 0041** (the advisory-muse lesson — the reason the carrier is filtered, not a pass-through) · **ADR 0012** (dev-tool binaries resolved from PATH; no `go.mod` change) · **ADR 0030**/`modelith-check` (the fail-on-absent *gate* this deliberately diverges from) · the retired **ADR 0038 §Forward `RF-068-1`** — about the *unpaired-call diagnostic's E2E carrier*, **not** reachability; **not** reopened here (it stays retired and inert — this ADR neither fires nor unfires its trigger) · `specs/truth/techstack.md` (*Coverage tooling — DECLINED*; its former "reachability orphan caught by reasoning" clause is **corrected** here) · `docs/domain-model/quality.modelith.yaml` (the `advisory` `GatePolicy` + `modelith-drift` precedent) · the reference's heavy `tell-me-go/cmd/deadcode` (**not** adopted)

## Context

tellme ships **no coverage tooling** (ADR 0042 §D4; #144 closed `not planned`): a coverage
*percentage* misleads because the E2E contract runs the built binary as a subprocess. The one class a
profile uniquely adds — a **reachability orphan** — was recorded in `specs/truth/techstack.md` as
"caught by reasoning" (a round-068 citation). The operator's measurement (issue #198, 2026-09-27)
**falsifies that clause**: a whole-program
reachability pass finds a small set of **genuinely-dead exported symbols** the standard `unused`
linter cannot see (it treats any **exported** symbol as used), and tellme has **no carrier** for this
class at all. (Separately, `RF-068-1` — the diagnostic's *E2E carrier*, ADR 0038 §Forward — is
**not** the reachability claim; it stays retired and inert, unaffected by this round.)

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

**D2 — `-test` is mandatory.** Without it, the whole test surface reads as dead; measured **1151**
items on the head (`77db4c1`; **1157** on `dev` `ae300e9`). With it, test-reachable symbols stay live
(measured: **7** items on `dev` `ae300e9`). The direction is load-bearing (three orders of
magnitude); the figure is the measured one (F-094-1).

**D3 — The carrier is advisory, never fails, and is NOT a `make verify` member.** `make dead-code`
mirrors the `modelith-drift` shape: it runs, filters, prints, and **exits 0 always**. It adds **no**
gate. (ADR 0041: an advisory proves the tool *works*, not that it *fires*.)

**D4 — Absent tool ⇒ install hint + exit 0.** A deliberate divergence from `modelith-check`
(fail-on-absent): an advisory that failed on an absent optional tool would be a de-facto gate.

**D5 — The FP policy: filter the *interface-conformance-only* class.** Run with `-test`, then filter
the recorded FP class so the **findings body** on a **clean** tree is **empty** (the run still prints
its banner + `✓` + `advisory done` lines) and only **new** findings appear. **Not** a `NonFixCatalog`
— **one documented exclusion predicate** (a class). The measured member is
`ui.sharedSource.Suggest` (an interface-conformance-only method pinned by two compile-time
`var _ I = T{}` assertions that guard the identical method sets of `domaintui.Source` and
`prompt.Source` — structural typing the RTA cannot see); the predicate is exactly that member:
`DEADCODE_FP := unreachable func: sharedSource\.Suggest$$`. It is **NAME-KEYED and un-witnessed**
(RF-064-2). The predicate **deliberately does not** carry a name-wide `.*\.Unwrap$` alternative:
measured (F-094-3), that **swallows a genuinely-dead new `Unwrap` method** (a synthetic dead
`W1ProbeError.Unwrap` was reported raw by the tool but hidden by the filter, while its sibling
`.Error` survived) — contradicting FR-4's "only NEW findings appear". The unwrap-class is a
*documented* structural-typing FP the RTA currently does not report; if it ever does, it is added as
a **recorded symbol** then (RF-064-2), never a name-wide alternative.

**D5b — Provenance is probed and a tool failure is distinguished (F-094-2 / F-094-4).** The
documented dev host has the reference's heavy `tell-me-go/cmd/deadcode` at `$GOPATH/bin/deadcode`;
resolving `deadcode` from PATH there finds that wrong tool (284 lines of `[PRIVATE]`/`[DEAD]` noise,
exit 0). So the target **probes the binary's provenance** (`go version -m` → the `path` line must be
`golang.org/x/tools/cmd/deadcode`) and **skips** (exit 0, naming the vanilla install route)
otherwise — a wrong binary is a **no-op**, never a silently re-meant advisory. And it **checks the
tool's exit status**: a non-zero exit (e.g. the tree does not compile) prints `analysis did not run
(tool exit N)` (still exit 0), so a quiet run always means *clean*, never *unknown*.

**D6 — Remove the genuinely-dead exports (the cleanup).** The issue's 7 grep-verified symbols
(`config.(Config).EffectiveUseTUIPrompt`, `openai.NewWithHTTPClient`, `agent.DefaultToolTimeout`,
`di.TokenResolver`, `harness.RunWithStdin`, `harness.RunBinary`, `mcptest.SchemaWithProperty`) **plus
`cli.noopCallObserver`** — a **production** symbol the issue's inventory missed (its only reference is
its own `var _ agentport.CallObserver = noopCallObserver{}`; never instantiated — `compositeObserver`
is built with a real renderer, `cli.go:875`). A truly-dead type is removed, never filtered (filtering
it would hide it).

**D7 — Start from clean.** Remove the dead set **first**, then add the target, so the advisory begins
from the recorded-FP steady state.

**D8 — This ADR supersedes ADR 0042 §D5's "no `dead-code`" wording.** ADR 0042 §D4's
**percentage**-coverage decline **stands**; its **reachability-orphan half is answered** by this
round's carrier. ADR 0042's §D4/D5 body is an Accepted ADR and stays **verbatim**; its **index row**
gains a forward pointer to this ADR.

## Consequences

- The exported-dead class gains a **mechanical surface**: `make dead-code` surfaces **new** findings
  on demand, reporting no findings on the clean tree.
- It is **advisory**: no `verify` member, no gate change, no automatic guard (tellme has no CI). The
  carrier proves the tool *works*; a human decides whether a finding is dead or an FP.
- The **provenance hazard** is **guarded, not merely recorded**: the target probes the binary's
  provenance and skips a wrong one (a PATH `deadcode` that is not the vanilla x/tools binary is a
  no-op, never a silently re-meant advisory — D5b), and the header comment + this ADR state the
  required provenance.
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
- **RF-064-2 (the FP predicate is name-keyed and un-witnessed, not a catalog).** The predicate is the
  recorded interface-conformance-only member
  (`DEADCODE_FP := unreachable func: sharedSource\.Suggest$$`). It is **name-keyed** — a rename of the
  test-local receiver re-noises the clean tree — and **nothing asserts it still matches** its member
  (**TD-094-1**); the class is documented, the member is not pinned. Adding a member is a **recipe
  edit in the same PR** as the new FP, with a recorded rationale — never a silent widening. tellme
  keeps **no** `NonFixCatalog`.
- **RF-064-3 (provenance is probed at run time; the install provenance is un-verified at build time).**
  The target resolves `deadcode` from PATH and **probes** it (`go version -m` → the `path` line) so a
  wrong binary is a **no-op** (D5b). What is *not* machine-checked is that the **vanilla** tool is the
  one the operator is expected to have; that stays a **host fact** (R-094-1) — the `STATUS.md` *Host /
  toolchain* env-note should name the required provenance.
- **RF-064-4 (host-GOOS only).** The analysis runs on the host (`darwin/arm64` here);
  `verify-cross-compile` covers other targets. No cross-platform coverage is claimed.
- **RF-064-5 (structural-typing FPs are inherent).** The RTA cannot see every interface dispatch;
  the FP class is expected to recur and is handled by the predicate, not by deleting load-bearing
  contract pins.
- **RF-064-6 (`docs/domain-model` same-PR).** The round adds the advisory `dead-code` member to
  `quality.modelith.yaml` (ADR 0041 load-bearing rule); no product entity changes.
- **RF-064-7 (the carrier has no invocation occasion — TD-094-2).** `make dead-code` is in no gate,
  no workflow (there is no `.github/workflows`) and no closeout step (`SESSION-CLOSEOUT.md` runs
  `make check`/`check-full`, neither of which reaches it). Per the ADR 0041 lesson ("an advisory
  proves the tool *works*, not that it *fires*"), the exported-dead class will re-accumulate with
  nothing to notice it. Cheap remedy: treat `make dead-code` as an **on-demand closeout / round-open
  step** (human-run, never a gate) — no decision change required.
