# Technical Research: round 052 — close [#115](https://github.com/gosharplite/tellme/issues/115) (ride-alongs) + [#116](https://github.com/gosharplite/tellme/issues/116) (records)

**Topic**: two behaviour-preserving ride-alongs from the [#92](https://github.com/gosharplite/tellme/issues/92) ledger — **(R-2)** the `execute_command` tool receives its `[Tool Output]` sink **at construction** (removing the `BindToolOutput` registry-rebind + the `toolOutputBox` indirection), and **(R-1)** the `-i` suggester's selection policy gets one named owner (`set(items, cursor)` with the caller passing `noChoice`) — plus the **disposition of the three `#116` records** (relocate → durable home; close the issue). Each decision supports `spec.md` (US1/US2/US3 · FR-001…FR-010 · SC-001…SC-006) and the operator-locked clarifications (**Q1 → one round**; **Q2 → (i) relocate + close completed**).

**Behaviour intent**: **MODIFY (behaviour-preserving)** — no user-facing change. stdlib-only; POSIX-only; no new dependency; no new Gherkin/DSL row.

---

## Terminology

- **Construction seam** — the point where a dependency is handed to an object through its **constructor**, rather than mutated into it afterwards.
- **Registry-rebind step** — the round-034 pattern where the tool registry's stored (value) command tool had its sink *set after* construction via a shared pointer box.
- **Assembler** — `agentTools()` in `cmd/tellme` (round 044): the canonical, non-overridable the tool-set assembler the round-031 well-formedness gate iterates.
- **Record vs work** — a *record* documents a settled standing fact (no code change expected); *work* is a tracked change. `#116` holds records.

---

## Decisions

### D1 — R-2 mechanism: one sink-aware registry builder; `agentTools()` stays parameterless

**Chosen.** `NewCommandTool(sink domaintools.OutputSink) domaintools.Tool` becomes the single constructor, and the agent registry is built through **one** sink-parameterized builder:

```go
// cmd/tellme/deps.go
func agentTools() []domaintools.Tool            { return assembleAgentTools(nil) }   // parameterless + read-free (round-033 FR-009)
func assembleAgentTools(sink domaintools.OutputSink) []domaintools.Tool { … NewCommandTool(sink) … }
func newToolRegistry(sink domaintools.OutputSink) domaintools.Registry { return domaintools.NewRegistry(assembleAgentTools(sink)...) }
```

- `agentTools()` is **unchanged in signature** — still parameterless and read-free — so the round-031 well-formedness gate keeps iterating it verbatim (its only edit is the sibling call `newToolRegistry()` → `newToolRegistry(nil)`).
- The `deps.Dependencies` seam widens: `NewToolRegistry func(sink domaintools.OutputSink) domaintools.Registry`; `Dependencies.Validate()`'s reflect predicate still covers it (it stays **func-typed** — no interface-typed field, so no extra assertion per ADR 0017 §Forward).
- The prompt path builds the progress object **first**, then injects: `prog := dp.NewProgress(…)` → `reg := dp.NewToolRegistry(prog.ToolOutput)`. `prog` reads only `env`, `res.Provider.Model`, `turnStart`, `stderrColumns`, `dp.NewLines()` — it does **not** read `reg`, so the reorder is safe.
- The offline `--tool-usage` path (`renderToolUsage`) takes the same func-typed parameter and calls it with **`nil`** (it never executes a tool).

**Rejected:**
- **(a) A lazy sink seam** (mirror the round-033 `list_skills` catalog seam: an unbound `func() OutputSink` resolved inside `Execute`). *Rejected* — it *keeps* exactly the indirection the issue asks to remove (the tool would still be born without its dependency); it only relocates the mutation from a registry rebind into a lazy field.
- **(b) A second registry seam** (`NewToolRegistryWithOutput`) alongside the existing `NewToolRegistry()`. *Rejected* — two ways to build "the agent registry" is a duplicate surface and a drift hazard (the gate asserts one canonical set); one parameterized builder is fewer concepts.
- **(c) Rebase the assembler into `internal/cli`** so the CLI can see `prog` first. *Rejected* — forbidden: the composition root is `cmd/tellme` (ADR 0013); `internal/cli` must not construct adapters.

**Precedent for widening a `Dependencies` field's signature**: round 051 / ADR 0020 **F-7** reshaped `MCPDiscoverer`'s result into a named `Discovery{Tools, Warnings, Closer}` — a `Dependencies` field signature change inside a behaviour-preserving round. This is the same class of change.

### D2 — the command tool holds the sink directly (no pointer box, no rebind)

`executeCommand` becomes `struct { output domaintools.OutputSink }`; `Execute` passes `c.output` to `runCaptured`. **Deleted**: the `toolOutputBox` type, the `sink()` accessor, `BindToolOutput` (infrastructure), the `deps.Dependencies.BindToolOutput` field, and the `cli.go` rebind call. **Kept**: `sinkOpen` / `sinkClose` / `teeSink` (already nil-guarded), so a **nil** sink (the gate's + the offline assembler's) remains a no-op — the block renders unconditionally on the prompt path exactly as before (rounds 034/038/039/040 unchanged).

### D3 — R-1: `set(items, cursor)` — the caller owns the reset policy

`suggester.set` takes the cursor explicitly and stops resetting it:

```go
func (s *suggester) set(items []string, cursor int) { s.items = items; s.cursor = cursor }
```

The two in-package callers (`model.go` `computeSuggestions`) pass `noChoice`, preserving the round-037 policy (*"a refresh opens unselected"*) as a **caller** decision. `newSuggester` still initialises `cursor: noChoice` (a construction default, not a policy hidden in `set`); `cycle` / `selected` / `view` are untouched. `noChoice` stays package-private (both call sites are in-package; no export). The invariant `cursor ∈ {noChoice} ∪ [0, len(items))` is documented **once**, at `set`.

**Rejected:** exporting `NoChoice` (no external caller; widening the API for no consumer); adding a `reset()` helper (redundant — `set(items, noChoice)` already states it).

### D4 — #116 disposition: relocate the three records into ADR 0021 §Records

ADR 0021 is the round's decision record **and** the durable, indexed, non-frozen home for the three records (D5). Each record is **copied** (never moved) with its provenance:

1. **Permanent E2E narrowing** — the `\n`/`\r` tool-reason row-collision class has no E2E carrier; the deterministic carrier is the hostile-fixture unit pin `internal/ui/toolcall_reason_test.go` (round 036 D4).
2. **One concurrent block** — `internal/ui`'s `ToolOutputCoordinator` assumes one open `[Tool Output]` block; a future (declined, [#47](https://github.com/gosharplite/tellme/issues/47)) concurrent-tools round re-scopes it (round 040 / PR [#86](https://github.com/gosharplite/tellme/pull/86) §5).
3. **`End`-while-write-stalled accepted residual** — a wedged `stderr` write cannot be abandoned at `End` (`internal/ui/coordinator.go`); bounded-then-accepted (the stream is assumed to make progress).

The frozen plan packages `036-*` / `040-*` are **not** edited; the in-code pins stay where they are. After the ADR lands, [#116](https://github.com/gosharplite/tellme/issues/116) closes **completed** (its own closure path).

### D5 — Witness plan (falsifiable, reproduced then reverted)

- **(a) R-2** — a unit pin at the `internal/infrastructure/tools` tier: `NewCommandTool(sink)` carries the sink (a fake `OutputSink` observing `Begin`/`Writer`/`End` receives the block); reverting the injection (a nil sink) makes the pin fail.
- **(b) R-1** — a unit pin at the `internal/ui/tui/prompt` tier over `set(items, noChoice)` and `set(items, 0)`; reverting to the internal reset makes the `set(items, 0)` expectation fail.
- **(c) gate** — `make verify` stays green at RULE-E baseline **0** (no governed import added; 0 cycles); the round-031 gate + registry-set assertion stay green.

### D6 — Truth impact & governance

- **`specs/truth/techstack.md`** — **MODIFY ×3** (three rows):
  - the **Agent command tool (`execute_command`)** row: the sink is now **constructor-injected** (`NewCommandTool(sink)`); the `BindToolOutput` rebind + `toolOutputBox` are removed; the `OutputSink` interface (F-8) is the ctor param.
  - the **Interactive TUI prompt (`-i`)** row **and** the **Prompt suggestion engine** row: the selection policy is single-owned by the **caller** (`set(items, cursor)`; the refresh passes `noChoice`).
- **`specs/truth/techstack.md`** — **NOOP (checked)**: the **Build & Tooling / Task runner** row (no new Makefile target; the gate rides `verify`).
- **ADR 0021** (`docs/decisions/0021-ride-alongs-and-records.md`) + the `docs/decisions/README.md` index row — records D1–D4 and hosts the three relocated records.

### D7 — Scope guard

No user-facing change (flags, exit codes, stream bytes, `[Tool …]` literals, DSL vocabulary); no new dependency (`go.mod`/`go.sum` unchanged); no new Gherkin/DSL row (⇒ no topology change); no frozen [#92](https://github.com/gosharplite/tellme/issues/92)/round-051 decision re-opened. `/axb-spec-by-example`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine` are **NOOP**.

---

## Risks & residual items

- **Seam-signature ripple (R-2)** — widening `NewToolRegistry` touches the offline reporting path and the test fixtures; each passes `nil`. Enumerated in `tasks.md`; the compile step catches a miss.
- **Reorder safety (R-2)** — `prog` is built before `reg`; verified that `NewProgress` reads no registry field (round 051 `ProgressFactory` signature).
- **Two-assembler drift (R-2)** — avoided by construction: `agentTools()` delegates to the same `assembleAgentTools`, so the gate's set-equality assertion cannot drift.
- **Interface-seam `Validate()`** — not triggered (the widened field stays func-typed).
- **No release valve** — at RULE-E baseline **0** (ADR 0011/0016) the round must add **no** governed import; the change is entirely within `internal/infrastructure` + `internal/app/deps` + `cmd/tellme` + `internal/ui`.
