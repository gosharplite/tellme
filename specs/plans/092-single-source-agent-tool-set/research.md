# Technical research — round 092 `092-single-source-agent-tool-set`

**Topic**: where the base agent-tool composition should live, so both the production assembler and the e2e
enumerator derive from **one** owner instead of re-implementing it (round-090 **R-090-1**; issue #189).

**Owner**: `axb-technical-research` (owns the `techstack.md` placement note; no new technology, no new
dependency).

---

## D1 — The defect (a mirror, not a bug)

The base set (readers · `search_files` · write pair · `execute_command` · `list_skills`) is composed in
**two** live places:

- the **production** assembler — `cmd/tellme/deps.go` `assembleAgentTools(spec)` (and its parameterless
  `agentTools()` facade);
- the **e2e harness** enumerator — `tests/e2e/steps/tool_usage.go` `registeredToolNames()` (and the union
  `recordableToolNames()`).

A future edit to one but not the other is silent drift. Round-090 finding **R-090-1** recorded the
enumerator re-implements the composition; the issue records it is **not a defect** (the E2E binds the
enumerator to the recorded request **behaviourally**) — it is a **maintenance hazard**, i.e. a **scoped
refactor**.

## D2 — Decision: single-source the base composition in the tools package

Introduce **one canonical owner**:
`internal/infrastructure/tools.NewAgentBaseTools(sink domaintools.OutputSink) []domaintools.Tool` — the
base set in offer order, the command tool's `[Tool Output]` sink injected at construction (ADR 0021). Both
callers **derive** from it:

- `cmd/tellme` `assembleAgentTools(spec)` = `NewAgentBaseTools(spec.Sink)` + the unchanged capability-gated
  `read_image` append;
- `tests/e2e/steps` `registeredToolNames()` = `flattenToolNames(NewAgentBaseTools(nil))`, and
  `recordableToolNames()` = the same base **plus** `NewReadImageTool(0)` (the union authority, unchanged).

**Why the tools package is the right home (the placement decision).** The function is a composition of the
package's **own** constructors and carries **no capability policy**; it belongs beside them, exactly as
`NewFilesystemTools` / `NewWriteTools` return their sub-sets. It adds **no import edge**
(`internal/infrastructure/tools` is already imported by `cmd/tellme` — exempt — and by
`tests/e2e/steps` — outside `internal/`), so the R1 layer gate's architecture baseline stays **0**
(ADR 0011/0016). This is a *relocation of the base **list***, not a move of the **composition root**: the
root (`cmd/tellme`) still assembles the final registry and **still owns the capability policy** — the
`vision` gate and the family-aware ceiling resolution stay there (ADR 0039 D2/D3; `internal/cli` may not
name infrastructure). Recorded as **ADR 0062**.

## D3 — Alternatives considered (and why not)

1. **A carrier only, no relocation** — impossible in one package: the production assembler is in
   `cmd/tellme` (**package `main`**, not importable) and the enumerator is in `tests/e2e/steps`; **no**
   single Go package can see both, so a "compare them" carrier cannot be placed. (This is what forced
   round 091's deferral.)
2. **Import the e2e harness (`tests/e2e/steps`, godog) into `cmd/tellme`'s test package** — rejected: an
   **inverted, heavy** test dependency (the production root depending on the harness that drives the
   built binary), pulling godog into the root's test binary.
3. **A brittle Go-source-parsing carrier** (parse `cmd/tellme/deps.go` to extract the set) — rejected: the
   round-091 deferral reason stands (semantic source parsing is brittle and disproportionate); the
   round's binding carriers are the round-090 **equality-of-surfaces** shape instead.
4. **Move the whole `agentTools()` assembler back out of the composition root** — rejected: that reverses
   round-044's placement (ADR 0013) and would put capability policy in the tools layer; only the base
   **list** relocates.
5. **Decline (record the behavioural E2E binding as sufficient)** — rejected: the operator filed the issue
   as a round seed (FR-1 prefers binding); the relocation is small, dependency-free, and removes a genuine
   mirror.

## D4 — The witness (falsifiability)

- **W-1** — re-inline the base constructors in `registeredToolNames()` (omitting one) ⇒
  `TestRegisteredToolNamesIsTheCanonicalBaseSet` **reddens** (the enumerator no longer equals the owner).
- **W-2** — remove a tool from `NewAgentBaseTools` ⇒ round-090's `TestOfferedSetDocMatchesTheLiveRegistry`
  **reddens** (`doc (8) vs live (7)`): the canonical owner is **load-bearing** and the doc carrier binds to
  it (transitively, to the e2e enumerator now too).
- **W-3** — re-inline a divergent base composition in `assembleAgentTools()` ⇒
  `TestAgentToolsIsTheCanonicalBaseSet` **reddens**.
- **W-4 (behaviour identity)** — the full unit suite + the godog E2E stay **green with no assertion
  changed**; E2E counts unchanged (330 · 2487). This is the ADR-0039 "structural change, byte-identical
  behaviour" witness.

## D5 — No product behaviour; records

The round changes **no** observable behaviour: the base set + order are byte-identical; the capability
gate is untouched; no `.feature`/DSL step text changes; `go.mod`/`go.sum` unchanged; stdlib-only. The only
new production file is `internal/infrastructure/tools/agentbase.go`; the rest are one-line delegations and
`_test.go` carriers.
`docs/domain-model/**` is **not modelled** (a composition-placement refactor changes no entity, invariant,
or scenario — ADR 0041 escape hatch; the modelled tool surface is unchanged). The `techstack.md`
*Composition root* row gains a one-line placement note (owner `/axb-technical-research`).

## D6 — Scope guard

No product behaviour, no `.feature`/DSL text, no `go.mod`/`go.sum`, no new dependency, no `make verify`
change, no capability-gate move. Frozen `specs/plans/NNN-*/**` untouched. The docs-prose *gate* question
(a general carrier for such claims) stays out of scope (ADR 0060 RF-089-6).
