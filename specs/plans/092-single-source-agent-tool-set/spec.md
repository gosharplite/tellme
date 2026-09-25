# Round 092 — `092-single-source-agent-tool-set`

**Theme**: **single-source the base agent-tool composition** — the base set (the readers + `search_files` +
the write pair + `execute_command` + `list_skills`) is today composed in **two** places that can drift: the
production assembler `agentTools()` / `assembleAgentTools()` (`cmd/tellme/deps.go`) and the e2e harness
enumerator `registeredToolNames()` / `recordableToolNames()` (`tests/e2e/steps/tool_usage.go`). The round
introduces **one canonical owner** (`internal/infrastructure/tools.NewAgentBaseTools`) both derive from, so
the two can no longer be edited out of step.

**Anchor issue**: [#189](https://github.com/gosharplite/tellme/issues/189). **DoD = close it.**

---

## 1. Why this round

Round-090 reviewer finding **R-090-1** (*recorded, not required*) observed that the offered-tool set is
mirrored across ~7 live surfaces, and that only the `chat/dsl.md` `集合` cell is **mechanically bound** (by
round 090's carrier, `cmd/tellme/deps_offered_set_test.go`). The **e2e enumerator**
`tests/e2e/steps/tool_usage.go` → `registeredToolNames()` **re-implements the base-set composition**
instead of citing the production assembler.

It is **not a defect**: the E2E binds `registeredToolNames()` to the recorded request **behaviourally**
(run 091, issue [#188](https://github.com/gosharplite/tellme/issues/188) item (B), deferred to this live
issue). It is a **scoped refactor** — a **mirror hazard** (a future edit to one composition but not the
other) that the round removes by giving the base composition a **single owner**.

Round 091 deferred item (B) because binding the enumerator *by construction* looked to require either
moving the composition out of the composition root, importing the e2e harness into `cmd/tellme`'s test
package (an inverted, heavy dependency), or a brittle source-parsing carrier. This round owns that design
question (see `research.md`).

## 2. The change

1. **A canonical owner** — `internal/infrastructure/tools.NewAgentBaseTools(sink domaintools.OutputSink)
   []domaintools.Tool`: the base set in offer order (readers · `search_files` · write pair ·
   `execute_command` · `list_skills`), with the command tool's `[Tool Output]` sink injected at
   construction. It composes the constructors **the tools package already owns** — no new import edge, no
   new dependency.
2. **The composition root cites it** — `cmd/tellme` `assembleAgentTools(spec)` =
   `NewAgentBaseTools(spec.Sink)` + the unchanged capability-gated `read_image` append (the family-aware
   ceiling stays resolved in the root — ADR 0039 D2/D3). `agentTools()` is unchanged (`assembleAgentTools(deps.ToolSetSpec{})`).
3. **The e2e enumerator cites it** — `registeredToolNames()` = `flattenToolNames(NewAgentBaseTools(nil))`;
   `recordableToolNames()` = the same base **plus** the capability-gated `read_image` (the union authority,
   unchanged semantics).
4. **Binding carriers** — a `cmd/tellme` unit test binds `agentTools()` to the canonical owner, and an
   e2e-package unit test binds `registeredToolNames()` to it, so a future edit that **re-inlines a
   divergent** base composition in either place reddens (the round-090 carrier shape).

No product behaviour change; the offered set + order are byte-identical; `go.mod`/`go.sum` unchanged.

## 3. Requirements

- **FR-001** the base-set composition has **one** canonical owner
  (`internal/infrastructure/tools.NewAgentBaseTools`); both the production assembler
  (`cmd/tellme.agentTools()`) and the e2e enumerator (`tests/e2e/steps.registeredToolNames()`) **derive**
  from it rather than re-list the constructors.
- **FR-002** the e2e enumerator's base set is **bound to** the canonical owner by a carrier that reddens if
  the delegation is replaced by a divergent inline composition (W-1).
- **FR-003** the production base set (`agentTools()`) is **bound to** the canonical owner by a carrier of
  the same shape.
- **FR-004** the capability gate is **unchanged**: `read_image` is appended only for a vision-enabled
  provider (the root), and the e2e **union** enumerator (`recordableToolNames()`) still includes it.
- **NFR-001** **no observable behaviour change** — the offered set + order are byte-identical; no new flag,
  phrase, or exit code; no `.feature`/DSL step change; `go.mod`/`go.sum` unchanged; stdlib-only.
- **NFR-002** `make verify` green (incl. `verify-architecture` at its **0-violation** baseline and
  `modelith-check`); `go test -count=1 ./...` green — E2E counts **unchanged** (330 scenarios · 2487 steps).

## 4. Invariants

- **I-1 — One composition owner.** The base tool constructors are composed in exactly one production
  function (`NewAgentBaseTools`); no second production site re-lists them. (Grep-verifiable: the raw base
  constructors appear in the tools package and the two delegating call sites only.)
- **I-2 — Behaviour identity.** The set, the offer order, and every observable surface are unchanged; the
  full suite passes with **no assertion changed** (the ADR-0039 behaviour-identity witness).
- **I-3 — Layer baseline 0.** No new import edge: `internal/infrastructure/tools` is already imported by
  `cmd/tellme` (exempt) and `tests/e2e/steps` (outside `internal/`); the canonical owner adds none.
  `verify-architecture` stays at its 0-violation baseline.
- **I-4 — Capability policy stays in the root.** The `vision` gate and the family-aware ceiling resolution
  (`resolveImageCeiling`) remain in `cmd/tellme` (ADR 0039 D3); the canonical owner carries the **base**
  set only.
- **I-5 — Frozen history untouched.** `specs/plans/NNN-*/**` is never edited.

## 5. Scope

**In**: the canonical owner + the two delegating call sites + the two binding carriers + the placement ADR
(**ADR 0062**) + the `techstack.md` placement note.
**Out**: any product behaviour change; changing the offered set/order; moving the capability gate or the
ceiling resolution; adding a `make verify` member; the docs-prose gate question (ADR 0060 RF-089-6);
frozen plan packages.

## 6. Success criteria

- **SC-001** `NewAgentBaseTools` is the single base-set composition; both callers delegate to it (FR-001).
- **SC-002** the binding carriers pass at head and **redden** under the W-1/W-3 mutations (a divergent
  inline composition in either caller).
- **SC-003** the offered set + order are byte-identical; E2E counts unchanged (330 · 2487).
- **SC-004** `make verify-architecture` stays at 0 violations; `make verify` green; `go.mod`/`go.sum`
  unchanged.
- **SC-005** **ADR 0062** records the placement decision (indexed once; `verify-adr-index` green).

## 7. Assumptions

- **A1** The operator's round instruction grants the intent: bind the enumerator to the production
  assembler (issue #189 FR-1, option "single-source the base-set composition"). No `/axb-clarify` needed
  (0 questions).
- **A2** The **base** set is the right single-source granularity; the capability gate (`read_image`) stays
  in the root (A + the union enumerator keeps it).
- **A3** No `NEEDS CLARIFICATION`.
