# ADR 0062 — Single-source the base agent-tool composition (the canonical owner)

- **Status:** Accepted
- **Date:** 2026-09-25
- **Deciders:** tellme owner (issue [#189](https://github.com/gosharplite/tellme/issues/189))
- **Related:** round-090 review finding **R-090-1** · round 091 (`specs/plans/091-record-hygiene-tail/` — deferred issue #188 item (B), homed on #189) · round 090's carrier (`cmd/tellme/deps_offered_set_test.go`) · **ADR 0013** (composition-root extraction — the root lives in `cmd/tellme`) · **ADR 0039** (`ToolSetSpec` — the capability seam; the ceiling stays resolved in the root) · **ADR 0021** (the ctor-injected `[Tool Output]` sink) · **ADR 0011**/**ADR 0016** (the R1 layer table / RULE-A/B/E) · round 092 (`specs/plans/092-single-source-agent-tool-set`)

## Context

The **base agent-tool set** — the tools offered to *every* provider, independent of capability: the
read-only filesystem readers (`list_files`, `read_files`, `get_tree`), the in-file content search
(`search_files`, round 071 / ADR 0043), the write pair (`write_file`, `replace_text`, round 029), the
bash-first command tool (`execute_command`, round 024), and the read-only skills listing
(`list_skills`, round 033) — was composed in **two** live places:

- the **production** assembler — `cmd/tellme/deps.go` `assembleAgentTools(spec)` (and its parameterless
  `agentTools()` facade);
- the **e2e harness** enumerator — `tests/e2e/steps/tool_usage.go` `registeredToolNames()` (and the
  capability-added union `recordableToolNames()`).

Round-090's review recorded this as **R-090-1**: the enumerator **re-implements** the base-set
composition instead of citing the production assembler. It is **not a defect** — the E2E binds the
enumerator to the recorded request **behaviourally** (the offered set the built binary sends is compared
to `registeredToolNames()`) — but it is a **mirror hazard**: a future edit to one composition but not the
other is silent drift. The set is mirrored across ~7 live surfaces; only the `chat/dsl.md` `集合` cell is
**mechanically bound** (round 090's carrier).

Binding the enumerator *by construction* is constrained by the layer table (ADR 0011/0016): the production
assembler is in `cmd/tellme` (**package `main`**, not importable) and the enumerator is in
`tests/e2e/steps`. **No** single Go package can see both, so a "compare the two" carrier cannot be placed;
importing the godog-coupled harness into `cmd/tellme`'s test package would be an **inverted, heavy**
dependency. (This forced round 091's deferral.)

## Decision

**D1 — One canonical owner.** The base set is composed by exactly **one** function:
`internal/infrastructure/tools.NewAgentBaseTools(sink domaintools.OutputSink) []domaintools.Tool` — the
base set in offer order, the command tool's `[Tool Output]` sink injected at construction (ADR 0021). Both
callers **derive** from it:

- `cmd/tellme` `assembleAgentTools(spec)` = `NewAgentBaseTools(spec.Sink)` + the unchanged capability-gated
  `read_image` append;
- `tests/e2e/steps` `registeredToolNames()` = `flattenToolNames(NewAgentBaseTools(nil))`, and
  `recordableToolNames()` = the base **plus** the capability-gated `read_image` (the union authority —
  unchanged).

**D2 — The home is the tools package; this is a *list* relocation, not a composition-root move.**
`NewAgentBaseTools` composes the tools package's **own** constructors and carries **no capability policy**;
it belongs beside `NewFilesystemTools` / `NewWriteTools`, exactly as those return their sub-sets. The
composition **root** stays `cmd/tellme`: it still builds the final registry and **still owns the
capability policy** — the `vision` gate and the family-aware inline-ceiling resolution
(`resolveImageCeiling`) stay there (ADR 0039 D2/D3; `internal/cli` may not name infrastructure). No
capability fact enters the tools layer.

**D3 — No new import edge; the layer baseline stays 0.** `internal/infrastructure/tools` (tier 3) is
already imported by the composition root (`cmd/tellme`, outside `internal/` → exempt) and by
`tests/e2e/steps` (outside `internal/`). The canonical owner adds no edge, so the R1 gate's
`verify-architecture` baseline stays at **0 violations** (ADR 0011/0016).

**D4 — Behaviour is byte-identical.** The offered set and its offer order are a pure function of the
unchanged constructors in the unchanged order; the capability gate and the ceiling are untouched. No new
flag, phrase, exit code, config key, `.feature`/DSL step, or dependency; no new `make verify` member.

**D5 — The binding is carried, and falsifiable.** Two unit carriers pin the delegation (the round-090
equality-of-surfaces shape): `cmd/tellme`'s `TestAgentToolsIsTheCanonicalBaseSet` (the production base set
== the owner) and `tests/e2e/steps`' `TestRegisteredToolNamesIsTheCanonicalBaseSet` (the enumerator == the
owner). Either reddens if a future edit re-inlines a **divergent** base copy; a tool removed from the owner
reddens round-090's `TestOfferedSetDocMatchesTheLiveRegistry` (the owner is **load-bearing**).

## Why an ADR

The placement ("where does the base-set composition live?") is a **project-level rule future rounds
depend on** (`docs/decisions/README.md`): a future round that adds a tool, a capability, or a new consumer
of the offered set must know the **single owner** it cites and the boundary the root keeps (capability
policy). It continues the ADR-0013/0039 composition lineage and records the **partial** relocation of the
`agentTools()` body that round 044 had placed in the root.

## Alternatives considered

| Alternative | Rejected because |
| --- | --- |
| A carrier only (no relocation) — compare the two compositions | **Impossible in one package**: the root is `package main` (not importable) and the enumerator is in `tests/e2e/steps`; no single Go package can see both. This is what forced round 091's deferral. |
| Import the e2e harness (`tests/e2e/steps`, godog) into `cmd/tellme`'s test package | **Inverted, heavy**: the production root would depend on the harness that drives the built binary, pulling godog into the root's test binary. |
| A brittle Go-source-parsing carrier (parse `deps.go` to extract the set) | Rejected (the round-091 reason): semantic source parsing is brittle and disproportionate. The chosen carriers use the round-090 equality-of-surfaces shape instead. |
| Move the whole `agentTools()` assembler out of the composition root | **Reverses ADR 0013** and would put capability policy in the tools layer. Only the base **list** relocates. |
| Decline — record the behavioural E2E binding as sufficient | The operator filed #189 as a round seed (FR-1 prefers binding); the relocation is small, dependency-free, and removes a genuine mirror. |

## Consequences

### Positive

- The base composition has **one owner**: the production assembler and the e2e harness can no longer be
  edited out of step (R-090-1 closed).
- The two binding carriers, plus round-090's doc carrier, form a chain: `chat/dsl.md` ↔ `agentTools()` ↔
  `NewAgentBaseTools` ↔ `registeredToolNames()`.
- **No behaviour change**, **no new import edge**, **no new dependency**; the layer baseline stays 0.

### Negative / Accepted Trade-offs

- The base list now lives in the tools package rather than entirely in the composition root — a **partial
  reversal** of round-044's relocation of the `agentTools()` body. Justified: the function composes the
  package's own constructors and carries no capability policy; the root keeps the policy. (Recorded here so
  a future reader does not re-litigate the placement.)
- The two binding carriers are equal **by construction** today; they have teeth only against a future edit
  that re-inlines a **divergent** copy (the round-090 carrier's shape). Stated honestly.

### Neutral

- No domain entity, persisted record, config key, tool, wire shape, or `.feature`/DSL step changes;
  `docs/domain-model/**` is **not modelled** (ADR 0041 escape hatch — a composition-placement refactor
  changes no modelled behaviour).

## Verification

- **Structural** — the base constructors are composed in `internal/infrastructure/tools/agentbase.go`
  only; both callers delegate. Grep-verifiable (the raw base constructors appear in the tools package and
  the two delegating call sites only).
- **Carriers** — `cmd/tellme` `TestAgentToolsIsTheCanonicalBaseSet`; `tests/e2e/steps`
  `TestRegisteredToolNamesIsTheCanonicalBaseSet` (both pass at head; both redden under a divergent
  re-inline).
- **Behaviour identity** — the full unit suite + the godog E2E stay green with **no assertion changed**;
  E2E counts unchanged (330 scenarios · 2487 steps).
- **Gates** — `make verify` green (`verify-architecture` at its 0-violation baseline; `modelith-check`
  unchanged); `go.mod`/`go.sum` unchanged.
- **Falsifiability** — W-1/W-3 (a divergent re-inline in either caller) redden the carrier; W-2 (a tool
  removed from the owner) reddens round-090's doc carrier; all reproduced then reverted.

## References

- `internal/infrastructure/tools/agentbase.go` (the new canonical owner) · `cmd/tellme/deps.go`
  (`assembleAgentTools` / `agentTools` / `resolveImageCeiling`) · `tests/e2e/steps/tool_usage.go`
  (`registeredToolNames` / `recordableToolNames`) · `cmd/tellme/deps_offered_set_test.go` (round-090
  carrier) · `cmd/tellme/deps_agentbase_test.go` + `tests/e2e/steps/tool_usage_test.go` (the round-092
  carriers).
- [ADR 0013](0013-composition-root-injection.md) · [ADR 0039](0039-toolset-spec-capability-seam.md) ·
  [ADR 0021](0021-tool-output-ctor-injection.md) · [ADR 0011](0011-layer-discipline-gate.md) ·
  [ADR 0016](0016-application-import-ceiling.md).
- Round 090: PR [#187](https://github.com/gosharplite/tellme/pull/187) · round 091:
  `specs/plans/091-record-hygiene-tail/` · round 092: `specs/plans/092-single-source-agent-tool-set/`
  (`spec.md` FR-001…FR-004; `research.md` D1…D6).

## §Forward (deferred, non-blocking)

> **⚠ Not open work.** A decision deferred to a trigger, or a recorded divergence — not tasking.

- **RF-092-1** — the two binding carriers are equal **by construction**; they catch a **divergent**
  re-inline, not the delegation itself. A stronger carrier would need a shared, importable enumerator
  package both callers use (a future option if a third consumer of the base set appears).
- **RF-092-2** — the capability-gated `read_image` append and the union enumerator remain composed
  **separately** in the root and the harness (single tool, single constructor); only the **base** set is
  single-sourced. A future second capability would re-open the question of a shared capability-gated
  builder.
- **RF-092-3** — `newTUIRegistry()` (`cmd/tellme`) composes the 3-reader TUI sub-set directly; it is a
  distinct, narrower set, deliberately not folded into `NewAgentBaseTools` (which owns the agent base set).
