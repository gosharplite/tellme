# Plan — Provider-wire projection of MCP-relayed tool schemas (round 061)

**Plan Package**: `specs/plans/061-mcp-schema-provider-projection`
**Truth root**: `specs/truth` · **Interface kind**: `cli` (plain line CLI — no `ui/**`, no TUI surface)

## Interfaces inventoried

| Interface | Kind | Planner | Wave |
| --- | --- | --- | --- |
| The CLI chat end (`specs/truth/features/cli/chat/**`) | `cli` | **carried to its contract owner** `/axb-dsl-refine` (a line CLI has no API/data/UI planner) | 1 |

- `/axb-api-plan` — **NOOP** (no OpenAPI/HTTP surface; single CLI end).
- `/axb-data-plan` — **NOOP** (no persisted/in-memory domain state: the projection is a pure in-flight transform; the floor mutates an already-discovered schema in place).
- `/axb-ui-plan` — **skipped** (plain line CLI; no TUI surface added).

## Waves

**Wave 1 (single, no dependencies)** — the CLI contract owner refines the acceptance journey into executable truth:
- the round-056 declaration row gains the **carve-out** qualification (`chat/dsl.md`);
- a new Rule *A server's argument marks never reach the provider* with three Examples in `chat/using-tools-from-a-remote-mcp-server.feature`;
- five new DSL rows (2 Given + 3 Then) for the new sentences.

## The change surface (RD)

| # | Site | Change |
| --- | --- | --- |
| 1 | `internal/infrastructure/mcp/schema.go` | the **floor**: `isVendorExtension` + recursive `stripVendorExtensions`, applied in `NormalizeMCPSchema` after well-formedness (S-1/S-6) |
| 2 | `internal/infrastructure/llm/gemini/schema.go` (new) | `supportedSchemaKeys` (the named owner) + recursive, fail-closed `projectSchema` (S-2/D2–D5) |
| 3 | `internal/infrastructure/llm/gemini/client.go` | `buildToolDeclarations` projects each declaration's parameters (the wire site) |
| 4 | `internal/infrastructure/mcp/mcptest/server.go` | `AnnotatedSchema()` — the GitHub-shaped fixture (test double) |
| 5 | `tests/e2e/steps/step_r061_mcp_schema.go` (new) + `scenario_context.go` | the five new E2E steps; `writeGeminiConfig` appends the arranged `MCP_SERVERS` block |
| 6 | `internal/infrastructure/mcp/schema_test.go`, `internal/infrastructure/llm/gemini/schema_test.go` (new) | the hermetic pins (S-5) |

## Layer/architecture notes

- The floor lives with the MCP relay; the projection lives with the transport that owns the closed wire. Each seam is **one concern** (S-1).
- `supportedSchemaKeys` is package-local to the gemini adapter and is read by the projection **and** its pin (FR-007's "single named owner") — no new domain type, no cross-layer edge, `verify-architecture` untouched.
- No dependency change (`go.mod`/`go.sum` unchanged); stdlib only; POSIX-only.

## Test strategy

| Layer | Carrier |
| --- | --- |
| Unit (floor) | `TestNormalizeMCPSchema_StripsVendorExtensions` (root + nested + `items`/`anyOf` subtree) |
| Unit (projection + gate) | `TestProjectToolDeclarations_NoUnsupportedKeywordReachesTheWire` (the envelope an MCP tool actually offers; red-capable), `TestProjectSchema_KeepsOnlySupportedKeys`, `TestProjectSchema_FailsClosed` |
| E2E | the truth feature's new Rule (gemini leg + tolerant-family leg + no-mark control) observing the recorded wire bytes |

Witnesses (reproduced then reverted): (a) removing the projection turns the gemini pin RED; (b) removing the floor turns the mcp pin RED.
