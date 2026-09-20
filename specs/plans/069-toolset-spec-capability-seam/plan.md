# System Analysis Plan — the `ToolSetSpec` capability seam (round 069)

**Plan Package**: `specs/plans/069-toolset-spec-capability-seam`
**Inputs**: `spec.md` · `research.md` (D1–D7) · `truth-delta.md` · `specs/truth/**`
**Anchor**: issue [#140](https://github.com/gosharplite/tellme/issues/140)

---

## 1. Interface inventory

| Interface | Kind | Where exercised | Planner |
| --- | --- | --- | --- |
| The **prompt turn**'s tool-registry construction | `cli` | `internal/cli` (`runTurn` path) → `deps.NewToolRegistry` | carried forward to its contract owner **`/axb-dsl-refine`** (no API/data/UI planner applies) |
| The **offline `--tool-usage`** report | `cli` | `internal/cli` (`renderToolUsage`) | same |

Both are **one CLI end**: a construction seam internal to the binary. No HTTP/OpenAPI surface, no persisted state, no user-facing UX surface. The change is **not user-visible**, so no new acceptance journey is required (the `042/043/047/049` structural-round precedent).

## 2. Waves

| Wave | Delegates | Outcome |
| --- | --- | --- |
| **W1** | `/axb-api-plan` | **NOOP** — a single CLI end; no `specs/truth/contracts/**`. |
| **W2** | `/axb-data-plan` | **NOOP** — no persisted state; the spec is in-memory construction state (`spec.md` I-3). |
| **W3** | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI; no TUI surface change. |
| **W4** | `/axb-dsl-refine` (the CLI contract owner) | **NOOP** — no user-visible behaviour change; no interface feature / DSL row changes (the existing `offering-the-agent-tools` / `reading-a-local-image` / `accounting-for-the-tool-use` journeys stay green **unchanged**). |

`analysis-plan-never-writes-truth`: this plan orchestrates only — the truth edits are recorded by their owner (`/axb-technical-research` → `techstack.md`).

## 3. Affected files (expected change set)

```text
internal/app/deps/deps.go          CHANGED  + ToolSetSpec; NewToolRegistry takes it
cmd/tellme/deps.go                 CHANGED  assembleAgentTools/newToolRegistry take the spec
internal/cli/cli.go                CHANGED  the prompt-path call site + renderToolUsage (named signature)
cmd/tellme/deps_test.go            CHANGED  the three call sites → named-field constructions
internal/cli/testdeps_test.go      CHANGED  the double's signature
internal/cli/cli_test.go           CHANGED  the injected registry func + the renderToolUsage call
docs/decisions/0039-*.md (+ index) ADD      the structural ADR
specs/truth/techstack.md           MODIFY   *Composition root* + *Image filesystem tool* rows
specs/plans/069-toolset-spec-capability-seam/**  ADD  the plan package
```

**No** `internal/infrastructure/**` production change; **no** config schema change; **no** new dependency.

## 4. Contract / layering notes

- `internal/cli` already imports `internal/app/deps` (it consumes `deps.Dependencies`), so naming `deps.ToolSetSpec` at the call site adds **no** import edge.
- `internal/app/deps` already imports `internal/domain/tools`, so the `ToolSetSpec.Sink domaintools.OutputSink` field adds **no** import edge.
- The ceiling resolution (`infrallm.Family` + `infratools.ImageCeilingForFamily`) **stays in `cmd/tellme`** — `internal/cli` must not name infrastructure (the RULE-B 0-violation baseline; `research.md` D2).
- `agentTools()` keeps its parameterless shape (`assembleAgentTools(deps.ToolSetSpec{})`).
