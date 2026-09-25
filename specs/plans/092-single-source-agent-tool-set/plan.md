# System analysis — round 092 `092-single-source-agent-tool-set`

**Interface inventory** (axb-system-analysis): the round touches a **CLI end** only.

| End | Kind | Planner / owner | Notes |
| --- | --- | --- | --- |
| The agent tool surface / the offered set | `cli` | `/axb-dsl-refine` owns the CLI truth tree | **NOOP** — no `.feature`/DSL step text changes (the offered set + order are byte-identical); the `chat/dsl.md` `集合` cell is unchanged |
| API | — | `/axb-api-plan` | **NOOP** — a standalone CLI; no OpenAPI surface |
| Data | — | `/axb-data-plan` | **NOOP** — no persisted-state shape change |
| UI | — | `/axb-ui-plan` | **NOOP** — a plain line-oriented CLI; no TUI screen change |

**Waves**: none — the CLI end is carried to its contract owner (`/axb-dsl-refine`), which records a **NOOP**
for this round (no executable-truth change).

## The change surface

| File | Change | Tier |
| --- | --- | --- |
| `internal/infrastructure/tools/agentbase.go` | **NEW** — `NewAgentBaseTools(sink)` (the canonical base-set owner) | production (tools) |
| `cmd/tellme/deps.go` | `assembleAgentTools` delegates its **base** half to `NewAgentBaseTools(spec.Sink)`; the `vision` append + the ceiling stay | production (root; exempt) |
| `tests/e2e/steps/tool_usage.go` | `registeredToolNames()` / `recordableToolNames()` delegate to `NewAgentBaseTools(nil)` (+ the unchanged `read_image` union) | e2e harness (outside `internal/`) |
| `cmd/tellme/deps_test.go` | **+** `TestAgentToolsIsTheCanonicalBaseSet` (binding carrier) | test |
| `tests/e2e/steps/tool_usage_test.go` | **NEW** — `TestRegisteredToolNamesIsTheCanonicalBaseSet` (binding carrier) | test |
| `docs/decisions/0062-*.md` + `docs/decisions/README.md` | **ADR 0062** + index row (the placement decision) | record |
| `specs/truth/techstack.md` | the *Composition root* row gains a one-line placement note | truth (owner `/axb-technical-research`) |

## Layering note (ADR 0011/0016)

`internal/infrastructure/tools` is tier **3** and already imported by the composition root (`cmd/tellme`,
outside `internal/` → exempt) and by `tests/e2e/steps` (outside `internal/`). The canonical owner adds **no
import edge**; the architecture baseline stays at **0**. `internal/cli` is untouched (it consumes the
injected `dp.NewToolRegistry`, never the tools package — RULE-B).

## Behaviour identity

The offered set + order are a pure function of the (unchanged) constructors in the (unchanged) order; the
capability gate and the family-aware ceiling are untouched. The suite passes with **no assertion changed**;
E2E counts unchanged (330 scenarios · 2487 steps).
