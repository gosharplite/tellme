# Technical research — round 091 `091-record-hygiene-tail`

**Theme**: record hygiene (tail) — drop a stale tool ordinal (A) · name a stale count's subject and drop
the stale figure (C) · decide the optional e2e-enumerator binding (B). **Truth/record only.**

**Anchor**: [#188](https://github.com/gosharplite/tellme/issues/188). Surveyed on `dev` @ `9f9cd9c`.

---

## D1 — (A) The ordinal is imprecise against the shipped offer order (verified)

`cmd/tellme/deps.go` `assembleAgentTools(spec)` composes the base set in **offer order**:

```go
tools := infratools.NewFilesystemTools()        // list_files, read_files, get_tree
tools = append(tools, infratools.NewSearchTool()...)   // search_files            ← 4th
tools = append(tools, infratools.NewWriteTools()...)   // write_file, replace_text
tools = append(tools, infratools.NewCommandTool(spec.Sink))  // execute_command
tools = append(tools, infratools.NewSkillsTool(nil))         // list_skills
if spec.Vision { tools = append(tools, infratools.NewReadImageTool(...)) }  // read_image (gated)
```

So `search_files` is the **4th** by offer order; the **base set** is **eight** (round 090's reconciled
count); the only **ninth** is the capability-gated `read_image`. The "ninth" is defensible only under a
*chronological add-order* reading nothing states. **Decision (D1.1): drop the ordinal** — identify the tool
by name + round/ADR. Same handling as round 090 (which dropped the hand-count from the `集合` cell).

## D2 — (A) The ADR body is immutable; the index row is live

- `docs/decisions/0043-search-files-tool.md:23` — *"D1 — ADD `search_files` as the **ninth** agent
  tool…"*. `docs/decisions/README.md`'s convention: **an ADR is immutable once Accepted** (supersede, do
  not edit). It is a historical record of the state *at decision time* → **leave verbatim** (issue
  #188 explicitly).
- `docs/decisions/README.md:69` (the ADR 0043 **index row**) — *"… a ninth tool offered to every
  provider"*. The index is a **live, curated** surface (round 085 annotated an index row in place) → in
  scope for FR-2. Remove the bare ordinal.

## D3 — (C) The count's subject is the class-phrase vocabulary — and the figure is stale on both axes

**Live authorities:**

| Fact | Authority | Value |
| --- | --- | --- |
| Frozen **class-phrase vocabulary** | `specs/truth/features/cli/dsl.md` (interface-root row `tellme explains on stderr that "{reason}"`) | **eleven** phrases (enumerated: `the configuration could not be found` · `no configuration could be found` · `the configuration could not be parsed` · `the configuration is invalid` · `the selected provider is not in the registry` · `the provider configuration is invalid` · `the provider request failed` · `the tool request failed` · `the workspace path is not a directory` · `the runtime home is not usable` · `the command-line usage is invalid`) |
| **Exit-code set** | `internal/cli/exitcode.go` | **seven** (`0/2/3/4/5/6/7`) |

Production carries all **eleven** phrases (grep over `internal/`+`cmd/`, non-test) — the interface-root
count is accurate.

**Origin of "ten":** `specs/plans/007-*/research.md` — *"ten phrases, exit codes `0/2/3/4/5/6`"* — a
**frozen-time** figure. Since round 007 the vocabulary grew (round 008 added `the tool request failed`,
+exit 7), so **"ten" is stale on both axes** and drifts with no carrier (RF-089-6).

**Decision (D3.1):** do **not** restate a number (it drifts); **name the subject and drop the count** —
the round-090 approach. Rewrite the three **live** surfaces that restate it:
1. `specs/truth/techstack.md:105` — "the vocabulary stays ten" → "the class-phrase vocabulary is unchanged".
2. `specs/truth/techstack.md:31` — "the ten-value exit-code set" → "the exit-code set".
3. `specs/truth/features/cli/chat/dsl.md` (round-079 note) — "the ten-value exit-code set" → "the exit-code
   set".

**Immutable (left verbatim):** `docs/decisions/0051-*.md:43`, `0052-*.md:23`, `0055-*.md:45`,
`0058-*.md:92` — Accepted ADRs, historical.

## D4 — (C) Carrier decision (honest)

A docs-prose count has **no mechanical carrier** (RF-089-6 / TD-090-1). **Dropping** the figure removes
the claim rather than re-asserting an un-carried number — the round-090 pattern. The witness (W-C) is
therefore a **carried, manual inspection** (the stale figure is gone) — recorded honestly (the F-089-1
lesson: never dress a manual check as a mechanical one).

## D5 — (B) The optional binding — decision: defer, and home it durably

`tests/e2e/steps/tool_usage.go` mirrors the base composition:

```go
func registeredToolNames() []string {
    return flattenToolNames(infratools.NewFilesystemTools(), infratools.NewSearchTool(),
        infratools.NewWriteTools(), []domaintools.Tool{infratools.NewCommandTool(nil)},
        []domaintools.Tool{infratools.NewSkillsTool(nil)})
}
```

Binding it to `agentTools()` **by construction** requires one of:

1. **Moving the base-set composition out of the composition root** (`cmd/tellme`) into an importable
   package both callers share — **reversing round 044's placement** ("the assembler now lives in
   `cmd/tellme`") and re-opening the ADR 0013/0039 composition-root boundary;
2. **Importing the e2e harness** (`tests/e2e/steps`, godog) into `cmd/tellme`'s test package — an
   inverted, heavy test dependency; or
3. a **brittle Go-source-parsing** carrier.

None is proportionate to a **tail** round (NFR-1: no product behaviour change), and there is **no defect**:
`registeredToolNames()` is already bound to the recorded request **behaviourally** by the E2E — round 090's
own framing of **R-090-1** (*"a scoped refactor, not a defect"*). **Decision (D5.1): defer (B)** with this
recorded reason, and **home it on a live issue** — **[#189](https://github.com/gosharplite/tellme/issues/189)**
(a durable home per Bootstrap Agent Rule 11 — not only the
frozen package).

## D6 — Domain model: not modelled (ADR 0041 escape hatch)

Nothing **modelled** changes: no entity/invariant/scenario is touched (the `Tool` entity lists tools by
name, not ordinal; the domain model carries no class-phrase count). Recorded in `plan.md` §5.

## D7 — Realm of the change

- **Truth owners**: `specs/truth/techstack.md` → `/axb-technical-research`;
  `specs/truth/features/cli/chat/dsl.md` → `/axb-dsl-refine`.
- **Records (not truth)**: `docs/decisions/README.md` (the curated index).
- **Immutable**: ADR bodies (0043, 0051, 0052, 0055, 0058); frozen `specs/plans/**`.
- **No** product code, `.feature` step text, `.feature` row semantics, `go.mod`/`go.sum`.
