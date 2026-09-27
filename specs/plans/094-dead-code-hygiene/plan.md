# System analysis — round 094 `094-dead-code-hygiene`

**Interface inventory** (axb-system-analysis): the round touches **no system interface**. The object
of work is the **Makefile** (a dev/task surface) plus the source tree's dead-symbol removal.

| End | Kind | Planner / owner | Notes |
| --- | --- | --- | --- |
| The CLI end (the built binary / Gherkin contract) | `cli` | `/axb-dsl-refine` owns the CLI truth tree | **NOOP** — no product-observable behaviour change (the removed symbols have zero callers; a Make target is not a CLI/executable surface) |
| API | — | `/axb-api-plan` | **NOOP** — a standalone CLI; no OpenAPI surface |
| Data | — | `/axb-data-plan` | **NOOP** — no persisted-state shape change |
| UI | — | `/axb-ui-plan` | **NOOP** — a plain line-oriented CLI; no TUI screen change |

**Waves**: none — no interface to delegate. The "interface" of interest is the **`make` task surface**
(the `dead-code` target), which is a dev/tooling concern, not a system interface; it is carried by
`/axb-technical-research` (the `techstack.md` row) and recorded in ADR 0064.

## The change surface

| File | Change | Tier |
| --- | --- | --- |
| `internal/config/config.go` | **remove** `EffectiveUseTUIPrompt` (dead method) | production (config) |
| `internal/infrastructure/llm/openai/client.go` | **remove** `NewWithHTTPClient` (dead ctor) | production (infrastructure) |
| `internal/agent/agentloop.go` (+ `tool_contract.go` comment) | **remove** the `DefaultToolTimeout` alias; fix the doc comment to name `tools.DefaultToolTimeout` | production (agent) |
| `internal/infrastructure/di/mcp_factory.go` | **remove** the `TokenResolver = mcp.TokenSource` alias | production (di) |
| `internal/cli/composite_observer.go` | **remove** `noopCallObserver` (type + 2 methods + its `var _` conformance) | production (cli) |
| `tests/e2e/harness/cmd_helper.go` (+ `step_t014…go` comment) | **remove** `RunWithStdin` + `RunBinary`; fix the doc comment | e2e harness (outside `internal/`) |
| `internal/infrastructure/mcp/mcptest/server.go` | **remove** `SchemaWithProperty` (dead helper; `schemaJSON` stays) | test support (outside `internal/`? no — inside, but a non-`_test.go` helper pkg) |
| `Makefile` | **add** the `dead-code` target + `.PHONY` + `help` line; **no** change to the `verify:` aggregate | task surface |
| `docs/decisions/0064-*.md` + `docs/decisions/README.md` | **ADR 0064** + index row + the ADR 0042 back-pointer | record |
| `specs/truth/techstack.md` | the *Dead-code reachability (advisory)* row + the *Coverage tooling — DECLINED* pointer | truth (owner `/axb-technical-research`) |
| `docs/domain-model/quality.modelith.yaml` (+ re-rendered `.md`) | `dead-code` advisory `QualityGate` member | record / model (ADR 0041) |

## Layering note (ADR 0011/0016)

Every change is a **deletion** of an unreachable symbol or a **Makefile/record** edit. No import edge
is added or moved; the architecture baseline stays at **0**. `internal/cli/composite_observer.go`'s
removal drops an unused type — `agentport` + `llm` imports stay (still used by `compositeObserver`).

## Behaviour identity

The removed symbols have **zero callers** (grep-verified; `spec.md` §2). No production path changes;
the suite passes with **no assertion changed**; E2E counts unchanged (330 scenarios · 2487 steps).
`go.mod`/`go.sum` unchanged; no new `make verify` member.
