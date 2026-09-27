# Technical research — round 094 `094-dead-code-hygiene`

**Topic**: adopt a mechanical carrier for the **exported-dead** class (a whole-program
reachability pass) and remove the genuinely-dead exports it finds — without adding a `verify`
gate, a Go dependency, or a `NonFixCatalog` (issue #198; reopens ADR 0042 §D4/§D5).

**Owner**: `axb-technical-research` (owns `specs/truth/techstack.md`; no new technology, no new
dependency). Recorded as **ADR 0064**.

---

## D1 — The tool: vanilla `golang.org/x/tools/cmd/deadcode`

The reference's `cmd/deadcode` is **heavy** (the local PATH binary is `tell-me-go/cmd/deadcode`:
`[DEAD]`/`[PRIVATE]` categories, ports-registry + `verify-exit-query` machinery; it reported
8 `[DEAD]` + 76 `[PRIVATE]`). tellme adopts **vanilla** `golang.org/x/tools/cmd/deadcode`, pinned
**`v0.47.0`** — the version already present in `go.sum` as an indirect entry (so `go.mod`/`go.sum`
stay **unchanged**). It is installed as a **PATH dev-tool binary** (`go install
golang.org/x/tools/cmd/deadcode@v0.47.0`), exactly like `modelith`/`golangci-lint`/`govulncheck`
(ADR 0012). Verified this round: the temp-GOBIN build's `go version -m` shows
`path golang.org/x/tools/cmd/deadcode`, `mod golang.org/x/tools v0.47.0`.

**Provenance hazard (recorded).** The PATH binary name `deadcode` **collides** with the reference's
heavy `tell-me-go/cmd/deadcode`. The target's header comment and ADR 0064 state the required
provenance (vanilla x/tools), so a machine with the wrong binary does not silently change the
advisory's meaning.

## D2 — `-test` is mandatory

`deadcode ./...` (no `-test`) reports **1151** items on the head (`77db4c1`; **1157** on `dev`
`ae300e9`) because test-reachable code reads as dead (tests are not entry points). `deadcode -test
./...` roots the analysis at each package's **test executable too**, so test-reachable symbols stay
live. Measured: on `dev` `ae300e9`, `-test` reports **7** items; without `-test`, every E2E step,
every test helper, the whole test surface — reads as dead. The direction is load-bearing (three
orders of magnitude); the figure is the **measured** one (the former "~330" was a non-reproducing
estimate — F-094-1). The carrier uses **`-test`**.

## D2b — Provenance must be probed, not asserted (F-094-2)

The documented dev host has the reference's **heavy** `tell-me-go/cmd/deadcode` installed at
`$GOPATH/bin/deadcode` (`go version -m $(command -v deadcode)` → `path
github.com/gosharplite/tell-me-go/cmd/deadcode`). Resolving `deadcode` from PATH therefore finds the
**wrong** tool, which prints `[PRIVATE]`/`[DEAD]` noise (284 lines, exit 0) — the ADR 0041
advisory-muse failure. The target therefore **probes the binary's provenance** (`go version -m` →
the `path` line must be `golang.org/x/tools/cmd/deadcode`) and **skips** (exit 0, naming the vanilla
install route) otherwise — a wrong binary is a **no-op**, never a silently re-meant advisory
(ADR 0064 D5 / RF-064-3).

## D3 — Advisory, never fails, not a `verify` member

`make dead-code` mirrors the **`modelith-drift`** shape (ADR 0041): it runs the tool, filters the
recorded FPs, prints findings, and **exits 0 always**. It is **NOT** a `make verify` member and adds
**no** gate. Rationale (the ADR 0041 *advisory muse* lesson): an advisory is an on-demand tool a
human invokes, **not** an automated guard — it proves the tool *works*, not that it *fires*.

**Why not a zero-tolerance gate?** An exported-dead symbol is frequently a **structural-typing FP**
(an interface-conformance method the RTA cannot see) — the same class the heavy reference tool needs
a `[PRIVATE]`/acceptance apparatus to manage. A gate would force per-round exemptions; an advisory
surfaces **new** findings for a human without blocking. (Recorded in ADR 0064 §Forward.)

**A quiet run must mean *clean*, not *unknown* (F-094-4).** The advisory's one promise is that no
findings = clean. A **tool failure** (non-zero exit — e.g. the tree does not compile) yields empty
stdout; the target therefore **checks the tool's exit status** and, if non-zero, prints `analysis
did not run (tool exit N)` (and still exits 0) rather than the `✓ no unreachable functions found`
line. "Never fails" stays the policy; "reports success for an analysis that never ran" does not.

## D4 — Absent tool ⇒ install hint + exit 0 (a deliberate divergence)

`modelith-check` **fails** when its binary is absent (it is a zero-tolerance gate member). The
advisory `dead-code` instead prints the install hint **and exits 0** — deliberate: an advisory that
fails on an absent optional tool would be a de-facto gate.

## D5 — The FP policy: filter the interface-conformance-only class

**D1 (operator-locked).** Run with `-test`, then **filter the recorded FP symbols** so the
steady-state output on a **clean** tree is **empty**, and only **new** findings appear. This is
**not** a `NonFixCatalog` — it is **one documented exclusion predicate** (a class), recorded in
ADR 0064. Rationale: a pure pass-through that always prints noise is exactly the ADR 0041 advisory
muse failure.

**The measured clean-tree FP set** (re-verified this round — the issue's assumed "Unwrap-class" was
miscalibrated: with `-test` the RTA **sees** the `errors.Is/As` unwrap calls and does **not** report
them):

- `ui.sharedSource.Suggest` (`internal/ui/tuiprompt_test.go`) — an **interface-conformance-only**
  method: it exists so `sharedSource` satisfies `domaintui.Source` **and** `prompt.Source` in two
  compile-time assertions (`var _ domaintui.Source = sharedSource{}`, `var _ prompt.Source =
  sharedSource{}`) that pin the two interfaces' **identical method sets** — the load-bearing
  assignability the TUI adapter relies on. Structural typing the analyzer cannot see; the method is
  never *called*. **Filtered, never deleted** (deleting it deletes a contract pin).

The filter predicate is therefore **exactly this measured member** —
`DEADCODE_FP := unreachable func: sharedSource\.Suggest$$` — an **interface-conformance-only class**
the predicate documents. It is **NAME-KEYED and un-witnessed** (RF-064-2): a rename of the receiver
re-noises the clean tree; nothing asserts the predicate still matches its member.

**Not** included: a name-wide `.*\.Unwrap$` alternative. Measured (F-094-3), that **swallows a
genuinely-dead new** `Unwrap` method (a synthetic dead `W1ProbeError.Unwrap` was reported raw by the
tool but hidden by the filter, while its sibling `.Error` survived) — directly contradicting FR-4's
"only NEW findings appear". The unwrap-class is a *documented* structural-typing FP that the RTA
currently does **not** report (so it needs no term); if it ever does, it is added as a **recorded
symbol** then, never a name-wide alternative.

**The cleanup set** (grep-verified zero callers; §2 of `spec.md`): the issue's 7 plus
`cli.noopCallObserver` — a **production** symbol the issue's inventory missed, whose **only**
reference is its own `var _ agentport.CallObserver = noopCallObserver{}` (never instantiated;
`compositeObserver` is built with a real renderer at `cli.go:875`). Removing it is the round's
`remove-the-genuinely-dead` mandate; it is **not** a filter (a filter would hide a truly-dead type).
Recorded as a divergence from the issue's inventory.

## D6 — Starting from clean

Remove the dead set **first**, then add the target — so the advisory begins from the recorded-FP
steady state (empty output on a clean tree), not from real hits.

## D7 — Records

**ADR 0064** records the decision; the ADR 0042 **index row** gains a forward pointer (its body — an
Accepted ADR — stays **verbatim**); `docs/decisions/README.md` gains the 0064 row;
`specs/truth/techstack.md` gains the *Dead-code reachability (advisory)* row + a pointer on the
*Coverage tooling — DECLINED* row; `docs/domain-model/quality.modelith.yaml` adds `dead-code` as an
**advisory** `QualityGate` member (re-render + `modelith-check`).

## D8 — Alternatives considered (and why not)

1. **Widen the `unused` linter config** (`exported-is-used: false` + `whole-program: true`) —
   rejected: tested; it **never** reports an unused **exported** symbol.
2. **Adopt the reference's heavy `cmd/deadcode`** — rejected: the `[DEAD]`/`[PRIVATE]` +
   ports-registry/exit-query machinery is disproportionate; the light vanilla tool is the class
   carrier.
3. **A zero-tolerance `verify` gate** — rejected (D3): FPs would force per-round exemptions; the
   advisory + FP predicate is the proportionate form.
4. **A coverage-profile integration** — out of scope; the percentage stance (ADR 0042 §D4) is
   untouched; this round adds **only** the reachability-orphan carrier.

## D9 — Scope guard

No product behaviour, no `.feature`/DSL text, no `go.mod`/`go.sum`, no new dependency, **no new
`verify` member**, no `//nolint`, no `NonFixCatalog`. Frozen `specs/plans/NNN-*/**` untouched. The
**percentage**-coverage decline (ADR 0042 §D4) stands; only the **reachability-orphan** half is
answered.
