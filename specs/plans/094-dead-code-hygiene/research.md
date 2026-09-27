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

`deadcode ./...` (no `-test`) reports **~330** items because test-reachable code reads as dead
(tests are not entry points). `deadcode -test ./...` roots the analysis at each package's **test
executable too**, so test-reachable symbols stay live. Measured: on `dev` `ae300e9`, `-test` reports
**7** items; without `-test`, `harness.RunInWithSyncedStdin`, every E2E step, every test helper,
etc. — the whole test surface — reads as dead. The carrier uses **`-test`**.

## D3 — Advisory, never fails, not a `verify` member

`make dead-code` mirrors the **`modelith-drift`** shape (ADR 0041): it runs the tool, filters the
recorded FPs, prints findings, and **exits 0 always**. It is **NOT** a `make verify` member and adds
**no** gate. Rationale (the ADR 0041 *advisory muse* lesson): an advisory is an on-demand tool a
human invokes, **not** an automated guard — it proves the tool *works*, not that it *fires*.

**Why not a zero-tolerance gate?** An exported-dead symbol is frequently a **structural-typing FP**
(an interface-conformance method the RTA cannot see) — the same class the heavy reference tool needs
a `[PRIVATE]`/acceptance apparatus to manage. A gate would force per-round exemptions; an advisory
surfaces **new** findings for a human without blocking. (Recorded in ADR 0064 §Forward.)

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

The filter predicate therefore = **"an interface-conformance-only method"**, implemented as a small
documented symbol regex covering the measured member plus the documented **unwrap-class**
(`…\.Unwrap$`) so a future unwrap-only usage (if the RTA ever stops seeing it) is handled without a
recipe edit. The **predicate** is the durable artifact; the members are its current instances.

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
