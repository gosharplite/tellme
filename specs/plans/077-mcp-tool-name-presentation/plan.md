# System Analysis — MCP tool-name presentation (round 077)

**Plan Package**: `specs/plans/077-mcp-tool-name-presentation`
**Anchor**: [#155](https://github.com/gosharplite/tellme/issues/155)

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The offered MCP tool declaration (its description) | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the `chat` module gains an MCP offered-declaration Rule/Example + a Then reading the description; observed on the fake-provider wire |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — no persisted-state change (the offered description is computed, never stored) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (no screen change) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The offered-description behaviour | the implementation | `internal/infrastructure/mcp/tool.go` — the call-name note (a field computed once in `NewTool`) + the fallback naming the callable name; `internal/domain/agent/result.go` is untouched |
| **W2** | The truth rows | `/axb-technical-research` (done) | `specs/truth/techstack.md` — a new *MCP tool-name presentation* row (ADD); **ADR 0049** + index |
| **W3** | The executable CLI contract | `/axb-dsl-refine` | a Rule/Example on `using-tools-from-a-remote-mcp-server.feature` + a `dsl.md` Then |
| **W4** | The plan-side acceptance | `/axb-spec-by-example` | `features/acceptance/…` (a presentation journey) — a small, non-NOOP change |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is **model-visible** (the offered MCP declaration's description now states the callable wire name), so `/axb-spec-by-example` is **NOT** NOOP (a short presentation journey).

## 4. Unchanged surfaces (invariants)

- The namespaced wire name stays `mcp_<server>_<tool>`; the offered declaration's **`name`** is unchanged; the bare name is never offered (`spec.md` I-2).
- The server's own definition (name/description/input schema) is never mutated; any tellme note is added **outside** the server text (ADR 0025 D1) (`spec.md` I-1).
- The offered **schema** bytes are unchanged; the vendor-extension floor (round 061) and the closed-wire projection (ADR 0031) are untouched (`spec.md` I-3 / NFR-001).
- No system-prompt / persona change; the ask stays declaration-carried (ADR 0025 D4) (`spec.md` I-4).
- Family-local to the MCP adapter's description; the naming rule, the schema normalizer, the reason envelope, and the `ToolDefs` projection are untouched (`spec.md` NFR-002).
- stdlib-only, POSIX-only, hermetic; no new dependency; no live network in the gate (`spec.md` NFR-003).

## 5. Domain model (ADR 0041)

**Not modelled, and that is recorded here** (the same-PR rule's escape hatch): the round changes a **presentation string** on the already-modelled `MCPTool`/`Tool` — the description the offered declaration carries. It introduces no modelled entity, attribute, relationship, invariant, or scenario (the `Tool` entity's `name`/`gate`/`requiresReason` and the `mcp-*` invariants are unchanged; a description is not a modelled attribute). `docs/domain-model/**` is **unchanged** and `modelith-check` stays green. (`techstack.md`'s new *MCP tool-name presentation* row is the truth home for the offered-description behaviour.)
