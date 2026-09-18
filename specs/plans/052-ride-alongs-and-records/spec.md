# Feature Specification: close [#115](https://github.com/gosharplite/tellme/issues/115) + [#116](https://github.com/gosharplite/tellme/issues/116) — the two `#92` ride-alongs (code) + the three `#92` records (durable disposition) (round 052)

**Feature Branch**: `052-ride-alongs-and-records`

**Created**: 2026-09-18

**Status**: Draft (clarify round 1 **CLOSED** — Q1/Q2 locked by the operator's goal + the round-052 framing exchange) — plan package created by `/axb-specify`.

**Input**: [#115](https://github.com/gosharplite/tellme/issues/115) — the two **ride-alongs** of [#92](https://github.com/gosharplite/tellme/issues/92)'s scope ledger (items **#5**, **#6**), split out at the round-051 closeout. [#116](https://github.com/gosharplite/tellme/issues/116) — the three **records** (items explicitly marked **"not work"**), likewise split out.

**Programme goal (operator-declared)**: **close both [#115](https://github.com/gosharplite/tellme/issues/115) and [#116](https://github.com/gosharplite/tellme/issues/116)** — i.e. retire the last two residual issues of the [#92](https://github.com/gosharplite/tellme/issues/92) lineage, finishing it for good. #115 is a **behaviour-preserving code refactor** (two small ride-alongs); #116 is a **disposition, not a code change** (its body: *"closed-by-intent only when each record is either resolved or relocated"*).

**Behaviour intent**: **MODIFY (behaviour-preserving)** — **no** user-facing change (`stdout`/`stderr` byte-contracts, flags, exit codes, the `[Tool …]` block literals, the DSL vocabulary all unchanged). The two ride-alongs are structural (an owning seam and a construction seam); the records disposition is documentation + governance.

---

## Clarify round 1 — CLOSED (operator-directed)

> Both questions were settled **by the operator's goal statement and the round-052 framing exchange** — the round's close target ("close #115 and #116") and the agreed disposition ("a record that cannot close itself belongs in a durable artifact, not an issue that outlives it — relocate each record, then close"). No question was left open; **no `NEEDS CLARIFICATION` remains**.

| # | Question | Decision |
| --- | --- | --- |
| **Q1** | **Round shape**: does round 052 close **both** issues in **one** delivery, or is #115 (code) separate from #116 (disposition)? | ✅ **One round.** The operator's goal names both issues for 052. #115 is small (tiny → small, behaviour-preserving); #116 is a documentation/governance close. Folding them keeps the #92 lineage retired in one review. |
| **Q2** | **#116 disposition**: how does a record-holder issue close, given "resolve" is foreclosed for all three records (permanent narrowing · declined-round-gated · bounded-then-accepted)? | ✅ **(i) Relocate + close as completed.** Each record is relocated to a **durable, in-tree home** — **ADR 0021** (§Records), the round's own ADR, with its code/ADR cross-references and provenance — and #116 is closed **completed** (its body's own closure path). Rejected: **(ii)** close `not_planned` as a permanent record-holder (leaves the records only in a closed issue's body — the exact "record cosplaying as an issue" the operator rejected); **(iii)** a new `docs/records/**` hierarchy (over-structure; ADRs are already the repo's durable forward-record home — rounds 0020/0021 precedent). |

### Q2 → (i) (LOCKED) — the records relocate to a durable home

The three [#116](https://github.com/gosharplite/tellme/issues/116) records, each with its standing provenance:

1. **Permanent E2E narrowing (round 036 D4)** — the `\n`/`\r` tool-reason row-collision class has **no E2E carrier**; its deterministic carrier is the hostile-fixture unit pin in `internal/ui/toolcall_reason_test.go`. A **permanent narrowing, not a deferral**.
2. **The coordinator models ONE concurrent block (round 040, PR [#86](https://github.com/gosharplite/tellme/pull/86) §5)** — `internal/ui`'s `ToolOutputCoordinator` assumes one open `[Tool Output]` block; a future concurrent-tools round (**declined** — [#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`) would re-scope it.
3. **`End`-while-write-stalled accepted residual** — a wedged `stderr` write cannot be abandoned at `End` (`internal/ui/coordinator.go`); **bounded-then-accepted**, assumed to make progress.

**Relocation home**: **ADR 0021 §Records** — a **non-frozen, indexed, citable** decision record (the same vehicle ADR 0020 used for its RF-51-x forward items). Each record keeps its full text + cross-references (the code pins stay where they are: `toolcall_reason_test.go`, `coordinator.go`). After the relocation lands, [#116](https://github.com/gosharplite/tellme/issues/116) is closed **completed** with a comment pointing at ADR 0021.

> **Assumptions (not escalated — low impact, disclosed):** the ADR number (**0021**); the api/data/dsl-refine **NOOP** set; #115's ride-alongs' exact mechanism (an RD decision — `research.md` D-series); no new Gherkin/DSL row (a behaviour-preserving refactor, the rounds 042–051 precedent).

---

## Grounded in the current system

Measured 2026-09-18 @ `dev` `84a6245` (post-round-051; a static read — the gate re-measures at implementation).

### R-2 (ledger #6) — `BindToolOutput` constructor injection

| Site | Current shape |
| --- | --- |
| `internal/infrastructure/tools/command.go` | `func NewCommandTool() domaintools.Tool { return executeCommand{output: &toolOutputBox{}} }` — the sink is held behind a **pointer box** so the registry's **value copy** shares the binding |
| `internal/infrastructure/tools/command.go` | `func BindToolOutput(reg domaintools.Registry, sink domaintools.OutputSink)` — the **post-construction rebind** seam (round 034 ADR 0005 D7) |
| `internal/app/deps/deps.go` | `BindToolOutput func(reg domaintools.Registry, sink domaintools.OutputSink)` — the injected seam |
| `cmd/tellme/deps.go` | `BindToolOutput: infratools.BindToolOutput`; `agentTools()` appends `infratools.NewCommandTool()` |
| `internal/cli/cli.go:734` | `dp.BindToolOutput(reg, prog.ToolOutput)` — after the registry (`:649`) and the progress object (`:693`) are built |
| `domaintools.OutputSink` | **now an interface** (`Begin()` / `Writer() io.Writer` / `End()` / `Enabled() bool`) the `ui.ToolOutputCoordinator` satisfies directly (round 051 **F-8** / ADR 0020) — so the ctor param is the **interface** |

**Tension to resolve (RD)**: `agentTools()` MUST stay **parameterless + read-free** (round-033 FR-009; the round-031 well-formedness gate iterates it). The live sink exists only on the prompt path. The mechanism is a `research.md` decision (D-series).

### R-1 (ledger #5) — suggestion-selection policy: one owner

| Site | Current shape |
| --- | --- |
| `internal/ui/tui/prompt/suggester.go` | `func (s *suggester) set(items []string)` resets `s.cursor = noChoice` **internally** (a side effect) |
| `noChoice` sentinel | written in **three** places: `newSuggester`, `set`, `cycle` — the reset policy is buried in `set`, not owned by the caller |
| Callers | `internal/ui/tui/prompt/model.go:192` `m.sug.set(nil)`; `:202` `m.sug.set(filtered)` |
| `cycle` / `selected` / `view` | the invariant `cursor ∈ {noChoice} ∪ [0, len(items))`; `selected()`'s `cursor < 0` guard is exactly the no-choice test |

### The #116 records (current homes)

- **Record 1** — plan package `036-tool-reason-sanitize` (**frozen**) + the live pin `internal/ui/toolcall_reason_test.go`.
- **Record 2** — plan package `040-spinner-liveness-and-turn-timer` (**frozen**); PR [#86](https://github.com/gosharplite/tellme/pull/86) §5.
- **Record 3** — `internal/ui/coordinator.go` (a doc comment) + the unit stress `TestCoordinatorEndWhileLineWriteInFlight`.

All three are in **frozen** packages or un-indexed code comments → none is a citable, non-frozen home. ADR 0021 §Records supplies one.

### Invariants that must survive

- **`make verify` green on `dev` at delivery**: RULE-A/B/C stay **0**; RULE-E baseline stays **0** (0 new / 0 stale — the terminal state, **no release valve**); RULE-F consistent; 0 cycles; cross-compile 4/4; lint 0.
- **Zero behavioural change**: identical `stdout`/`stderr` byte-contracts (incl. the `[Tool Output]` block literals + the suggestion list rendering), identical exit codes, unchanged flags, **no** new Gherkin/DSL row.
- **`agentTools()` stays parameterless + read-free** (round-033 FR-009); the round-031 well-formedness gate still iterates it and still asserts `agentTools()` ≡ `newToolRegistry(…)`'s tool set.
- **No cycle**; `internal/domain/**` stays pure (RULE-C); the sanitize/cap/bytes policy stays single-owned in `internal/ui` (ADR 0006/0008/0015).
- **No frozen [#92](https://github.com/gosharplite/tellme/issues/92) decision is re-opened** (composition-root home, the `Dependencies` bag, `agentTools()` relocation, MCP orchestration, the ADR 0014/0015/0017/0018/0019/0020 ports).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - the command tool is born with its `[Tool Output]` sink (Priority: P1)

As a maintainer, I want the `execute_command` tool to receive its `[Tool Output]` sink **at construction**, so the hidden post-construction rebind (`BindToolOutput` + the `toolOutputBox` indirection) disappears and the tool's dependencies are visible in its constructor — while `[Tool Output]` rendering stays byte-identical.

**Why this priority**: it is the larger ride-along (ledger #6) and the one that removes a *registry-rebind* seam; the `F-8` interface (round 051) makes the ctor param possible.

**Independent verification**: `infratools.NewCommandTool(sink)` takes the sink; `grep` finds **0** `BindToolOutput` / `toolOutputBox` references; the round-031 gate + the tool-registry unit pins + the godog E2E are green; the `[Tool Output]` block renders byte-identically.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the sources are inspected, **Then** `NewCommandTool` takes a `domaintools.OutputSink` argument and **no** `BindToolOutput` (nor `toolOutputBox`) exists.
2. **Given** a prompt-bearing turn runs a shell call without `output_file`, **When** the child streams, **Then** the `[Tool Output]` block (`Executing...`, the `[HH:MM:SS] [Tool Output] <line>` lines, the separator, the neutral close) is **byte-unchanged** (godog E2E + unit pins green).
3. **Given** the agent registry is assembled, **When** `agentTools()` is called, **Then** it stays **parameterless and read-free** and the gate's `agentTools()` ≡ registry set assertion still holds.

**Functional Requirements**:

- **FR-001**: `infratools.NewCommandTool` MUST take the sink at construction — `NewCommandTool(sink domaintools.OutputSink) domaintools.Tool` — and the tool MUST hold it directly (no `toolOutputBox` pointer indirection).
- **FR-002**: The `BindToolOutput` function (infrastructure) **and** the `deps.Dependencies.BindToolOutput` seam **and** the `cli.go` rebind call MUST be removed; the registry MUST be built with the sink injected.
- **FR-003**: `agentTools()` MUST remain **parameterless and read-free** (round-033 FR-009); the sink MUST be supplied on the prompt path only, through the injected `Dependencies` seam.
- **FR-004**: The change MUST be **behaviour-preserving** — the `[Tool Output]` block literals, the bound/stop semantics, the teeing, the idle-gap resume, and the neutral close are unchanged; `stdout` stays byte-exact.

---

### User Story 2 - the suggestion-selection policy has one named owner (Priority: P1)

As a maintainer, I want `suggester.set` to take the cursor explicitly (`set(items, cursor)`), so the selection-reset policy is a **caller** decision and the invariant `cursor ∈ {noChoice} ∪ [0, len(items))` has a single, named owner instead of a side effect buried in `set` — with the rendered list unchanged.

**Why this priority**: it is the smaller ride-along (ledger #5), it removes a hidden policy, and it is independently verifiable by a unit pin.

**Independent verification**: `suggester.set(items, cursor)` sets both fields; the callers pass `noChoice`; a unit pin over `set(items, noChoice)` / `set(items, 0)` behaviour; the `-i` Tab/cycle E2E is unchanged.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** `set` is inspected, **Then** it takes `(items []string, cursor int)` and does **not** reset the cursor itself.
2. **Given** a refresh (`computeSuggestions`), **When** a list is set, **Then** the selection is `noChoice` (the caller's explicit choice) and the list renders byte-identically.
3. **Given** the suggester, **When** the invariant is exercised, **Then** `cursor ∈ {noChoice} ∪ [0, len(items))` holds and `selected()`'s guard is exactly the no-choice test.

**Functional Requirements**:

- **FR-005**: `suggester.set` MUST take the cursor explicitly — `set(items []string, cursor int)`; the callers MUST pass `noChoice`; the reset MUST NOT be a `set` side effect.

---

### User Story 3 - the three `#116` records live in a durable home (Priority: P2)

As a maintainer, I want the three standing records relocated from the frozen packages / the closed issue into a **durable, indexed** home (ADR 0021 §Records), so they outlive the plan packages and [#116](https://github.com/gosharplite/tellme/issues/116) can close without dropping them.

**Why this priority**: it is the second half of the operator's goal; it is documentation/governance, not code.

**Acceptance Scenarios**:

1. **Given** the round has landed, **When** the three records are looked up, **Then** each appears in ADR 0021 §Records with its provenance and cross-references.
2. **Given** the relocation, **When** [#116](https://github.com/gosharplite/tellme/issues/116) is closed, **Then** a comment names ADR 0021 as their durable home and the closure reason is **completed**.

**Functional Requirements**:

- **FR-006**: Each of the three `#116` records MUST be relocated into **ADR 0021 §Records** with its provenance (the frozen plan package / the code pin) and its cross-references.

---

### Global Requirements *(cross-story only)*

- **FR-007**: A new **ADR 0021** MUST record the round's decisions (the R-2 construction seam + the R-1 selection-policy owner) **and** host the relocated records, indexed in `docs/decisions/README.md`; `specs/truth/techstack.md` rows updated through `truth-delta.md`.
- **FR-008**: The round MUST introduce **no** new dependency (`go.mod`/`go.sum` unchanged) and **no** new Gherkin/DSL row or topology change.
- **FR-009**: Falsifiability witnesses MUST be reproduced then reverted (ADR 0010): (a) a `_test.go`-level or unit-level assertion that the command tool carries the **injected** sink (a dropped injection fails); (b) a unit pin that `set` no longer resets (a mis-set cursor is observable); (c) the gate stays green at baseline **0** (no layer/cycle regression).
- **FR-010**: The round MUST NOT re-open frozen [#92](https://github.com/gosharplite/tellme/issues/92) decisions (the `Dependencies` bag's existence, the composition-root home, the ADR 0014/0015/0017/0018/0019/0020 ports) nor the rounds 020–051 non-BDD-tooling precedent.

---

## Edge Cases

- **`agentTools()` / registry set drift** — the round-031 gate asserts `agentTools()` ≡ the registry's tool set; a sink-aware registry builder that drops or renames a tool reds it.
- **Nil sink** — `agentTools()` / the offline `--tool-usage` report build the command tool with a **nil** sink; `sinkOpen`/`sinkClose`/`teeSink` already guard `nil`, so the block is a no-op there (unchanged).
- **Seam-signature ripple** — widening the registry seam touches the offline reporting path (`renderToolUsage`) and the test fixtures; each MUST pass a **nil** sink (they never run a tool).
- **Suggester invariant** — a caller passing an out-of-range cursor would break `view()`'s `i == s.cursor`; the documented invariant + the unit pin hold the line; **no** new validation is added (the callers are in-package).
- **Frozen-record immutability** — the relocation **copies** the records into ADR 0021; the frozen `036`/`040` packages are **not** edited; the code pins stay in place.
- **Interface-seam `Validate()`** — the widened `NewToolRegistry` stays **func-typed**, so `Dependencies.Validate()`'s reflect predicate still covers it (no new interface field).

## Key Entities

- **The command-tool construction seam** — `NewCommandTool(sink)` + the sink-aware registry builder; the removal targets `BindToolOutput` / `toolOutputBox`.
- **The `deps.Dependencies` seam** — the injection point (round 044 / ADR 0013); its `NewToolRegistry` field is widened to carry the sink.
- **The `suggester` selection policy** — `set(items, cursor)` + the `noChoice` sentinel.
- **ADR 0021** — the round's decision record + the durable home of the three `#116` records.
- **The layer-discipline ratchet** — `tools/arch/baseline.txt` (stays **0**) + the RULE-F `couplingSurface`.

## Success Criteria

- **SC-001**: `NewCommandTool` takes a `domaintools.OutputSink`; **0** references to `BindToolOutput` / `toolOutputBox` remain in production or test sources.
- **SC-002**: `suggester.set` takes `(items, cursor)`; the callers pass `noChoice`; a unit pin covers both a `noChoice` and a concrete-cursor call.
- **SC-003**: `make verify` green (RULE-A/B/C **0**; RULE-E baseline **0**, 0 new / 0 stale; RULE-F consistent; 0 cycles; lint 0; cross-compile 4/4).
- **SC-004**: `go test -count=1 ./...` green (incl. the godog E2E) with **no** behavioural assertion changed except where a test names a moved/renamed symbol.
- **SC-005**: `go.mod`/`go.sum` unchanged; **no** new Gherkin/DSL row; the topology audit unchanged.
- **SC-006**: ADR 0021 exists + is indexed; ADR 0021 §Records carries all three `#116` records; [#115](https://github.com/gosharplite/tellme/issues/115) and [#116](https://github.com/gosharplite/tellme/issues/116) are closed on delivery.

## Assumptions

- **A1**: The `BindToolOutput` / `toolOutputBox` surface is the measured inventory above (re-measured at implementation).
- **A2**: `/axb-spec-by-example` is **NOOP** — a behaviour-preserving structural refactor has no user-facing journey (the rounds 020/031/036/041–051 non-BDD-tooling precedent).
- **A3**: The R-2 mechanism (how the sink reaches construction while `agentTools()` stays parameterless) is an **RD decision** (`research.md` D-series).
- **A4**: The ADR number is **0021** (next free).
- **A5**: `/axb-api-plan` and `/axb-data-plan` are **NOOP** (no API surface; no persisted/runtime-state change).
- **A6**: `/axb-dsl-refine` is **NOOP** (a `cli`-side refactor is not a tellme CLI-contract change).

---

## Out of scope (recorded forward items)

- Any **user-facing** change: flags, exit codes, stream contracts, formatting, the DSL vocabulary.
- The **concurrent-tools** round that #116 Record 2 anticipates (declined — [#47](https://github.com/gosharplite/tellme/issues/47) `not_planned`).
- Re-scoping #116 Records 1/3 (both remain accepted as recorded — this round only **relocates** them).
- Settled exclusions: no Windows, no security/consent layer, no conversation pruning, no tool-call concurrency.
