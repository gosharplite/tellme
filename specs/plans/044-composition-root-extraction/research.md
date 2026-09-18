# Technical Research: composition-root extraction (round 044)

**Plan Package**: `specs/plans/044-composition-root-extraction`
**Anchor**: [#100](https://github.com/gosharplite/tellme/issues/100) — R2 of [#92](https://github.com/gosharplite/tellme/issues/92)
**Created**: 2026-09-18

> RD-side technical decisions. Every decision below is a **technology/structure choice**; the *what* lives in `spec.md`. Decisions are grounded in a static read of `dev` @ `a2fbafc`.

## Context

`internal/cli` is both the application layer and the composition root. It imports 7 `internal/infrastructure/*` packages (the R1 gate's RULE-B baseline) and holds 8 package-level factory vars. The round moves assembly to `cmd/tellme`, injects a domain-typed dependency value, and deletes the globals — **behaviour-preserving**. The R1 layer-discipline gate (`tools/arch`, ADR 0011) is the round's **primary witness**.

## Decisions

### D1 — Composition-root home: `cmd/tellme` (exempt from the tier table)

The R1 gate ranks `internal/app/*` at tier **2** and `internal/infrastructure/*` at tier **3**; **RULE-B** forbids an application-tier package from importing `internal/infrastructure/**`. A composition root in `internal/app/**` would therefore **re-create** the violation R2 removes. `cmd/**` lives outside `internal/`, so the gate never passes it to the tier function — it is **exempt**. `cmd/tellme/main.go` grows from a one-liner to: construct the concrete adapters → build the injected value → `os.Exit(cli.Run(os.Args[1:], version, opts))`. Decided in clarify Q1; settled as **ADR 0013**.

### D2 — Injected value shape: `internal/app/deps.Dependencies` + cli-local `Options`

- **`internal/app/deps`** (new, tier **2**) defines `Dependencies` with **domain-typed fields only** (it may import `internal/domain/**`, `internal/config`, `internal/home`, stdlib). It **cannot** import `internal/ui` (tier 5) or `internal/agent` (tier 4) — either would be a RULE-A upward import.
- Fields (the seams that today cross into `infrastructure/`): `NewGateway`, `NewHistoryStore`, `NewUsageStore`, `NewToolUsageStore` (**signature `func(func() (string, error)) history.ToolUsageStore`** — it gains the argument it previously captured from the `userHomeDir` var), `NewPromptTracker` (**signature `func(home string, userHome func() (string, error)) history.PromptTracker`**), `NewToolRegistry` (the **7-tool agent** registry), `NewTUIRegistry` (the **3-reader** registry the `-i` suggestion source consumes — grill fold fix-1), `BindToolOutput`, `BindSkillsCatalog`, `NewMetricsProvider`, `MCPDiscoverer`, `UserHomeDir`. **Every field has a named consumer** (`dp.NewToolUsageStore(dp.UserHomeDir)`, `dp.NewPromptTracker(res.Home, dp.UserHomeDir)`, `dp.NewMetricsProvider` read **after** the `spinnerGate` short-circuit, `dp.MCPDiscoverer` consumed in `augmentRegistryWithMCP`).
- **`cli.Options`** (defined in `internal/cli`, which may legally import `ui`) carries `deps.Dependencies` + the **one** presentation seam `RunTUIPrompt` (type `func(ctx, resolution, runtimeEnv) (string, bool, error)` — references unexported cli types, so `cmd` cannot populate it; it is **nil-defaulted inside `cli`**). The `newRenderer` var is **deleted** (inline `ui.NewRenderer()`; the renderer's injectable seam stays `runtimeEnv.renderer`).
- Entry point: **`cli.Run(args []string, version string, opts Options) int`** — one injected value. The injected value is **threaded** as a parameter named **`dp`** (`dp deps.Dependencies` — never `deps`, which would shadow the package) through `run`/`runTurn`/`renderTurn`/`dispatchReporting`/`renderToolUsage`/`renderHistoryList`/`renderNewSession`/`newCallRenderer`/`persistTurnUsage`/`runTUIPrompt` and the leaf helpers `newTurnSpinner`/`augmentRegistryWithMCP`. The parsed-flags struct is renamed **`options` → `flags`** (operator G2; kills the `options`/`turnOptions`/`Options` tri-collision).
- Rationale: the **tier table forces the split** (infra-crossing seams → `deps`; the one ui-typed seam → `Options`); a composition bag is **not** a domain concept, so it does **not** go in `internal/domain` (NFR-002). **The wide bag is an accepted smell** — interface segregation is recorded as a *rejected alternative* in ADR 0013 (operator G1).

### D3 — Tool assembly + the two bindings

`internal/cli` touches `infratools` in three roles: **assembly** (`agentTools()`/`newToolRegistry`), **binding** (`BindToolOutput`/`BindSkillsCatalog`), and **data loading** (`infrskills.Load`). All three leave:

- `agentTools()` + `newToolRegistry` **relocate to `cmd/tellme`**.
- `deps` injects: `NewToolRegistry func() domaintools.Registry` (bare), `BindToolOutput func(domaintools.Registry, domaintools.OutputSink)`, `BindSkillsCatalog func(domaintools.Registry, string)` (the string = the resolved skills dir).
- The infra-typed `ToolOutputSink` becomes the neutral domain port **`domaintools.OutputSink`** (`{ Begin func() ; Writer io.Writer ; End func() }` **plus the method `Enabled() bool { return s.Writer != nil }`**), so `internal/cli` names no infra type. `Enabled()` is **required**: `command.go` calls it at four sites and the `{}`-is-disabled contract is load-bearing (`sink()` returns `ToolOutputSink{}`; `BindToolOutput` no-ops safely). The type rename edits `command.go` (five references) and deletes `tooloutput.go`, so `internal/infrastructure/tools/**` is in scope (D11). `internal/cli` still builds the sink from its `ui` coordinator (legal) + computes `SkillsDir` from its own `resolution`, then calls the injected binders — the **same three roles, now via deps**.
- **`agentTools()` stays parameterless + read-free + non-overridable** (round-033 FR-009; round-031 gate, PR #65 ARCH-1). The **property** is preserved; only the **file** moves (A7).
- Rejected: build-the-fully-bound-registry (its neutral `opts` is essentially `OutputSink` anyway) and inject-constructors (would reintroduce a package global — the thing R2 deletes).

### D4 — MCP discovery: orchestration into `internal/infrastructure/mcp`

`mcp_discovery.go` is **orchestration**, not a factory: `discover` (concurrent fan-out over enabled remote servers), `discoverServer` (bounded per-server probe), `serverResult`, and the `mcp.*` wire helpers — all of which already live in (or belong with) `internal/infrastructure/mcp`.

- New `internal/infrastructure/mcp` entry point **`Discover(ctx, servers, bound, newClient, resolveToken)`** returning a domain-typed triple `(tools []domaintools.Tool, warnings []string, close func())`.
- The two `di` constructors (`di.NewRemoteClient`, `di.NewGhTokenResolver`) are **passed in** — `internal/infrastructure/mcp` imports **neither** `di` nor the SDK in any *new* way, so **`verify-mcp-sdk-confinement` stays green** (D8).
- `deps` exposes **`MCPDiscoverer func(ctx, map[string]config.MCPServerConfig) (tools []domaintools.Tool, warnings []string, close func())`** — only types `internal/cli` already imports (`config`, `domaintools`), so no new domain result type.
- `internal/cli` keeps *calling* discovery (it owns the prompt path: `augmentRegistryWithMCP` still merges the returned tools + prints warnings), but names no infra package.
- Rejected: orchestration in `cmd/tellme` (untestable in `main`; MCP-wire knowledge outside the confined package) and keep-in-cli (`di`+`mcp` edges would remain — no DoD).

### D5 — Delete every factory var; migrate the test seams

All 8 package-level vars are **deleted**: `newGateway`, `newHistoryStore`, `newUsageStore`, `newToolUsageStore`, `newToolRegistry`, `newRenderer`, `newTUIPromptRunner`, `userHomeDir` (+ `defaultMCPDiscovery`). Tests construct the injected value instead:

| File | Migration |
| --- | --- |
| `internal/cli/persistence_invariant_test.go` | builds `Options` with a fake usage store + a fake tool registry + a no-op tool-usage store |
| `internal/cli/tool_registry_test.go` | the round-031 assembler gate **relocates to `cmd/tellme`** (it must iterate the relocated, non-overridable assembler — A7) |
| `internal/cli/tui_submit_chrome_test.go` / `tui_dispatch_test.go` | build `Options` with a fake `RunTUIPrompt` + a fake gateway |
| `internal/cli/cli_test.go` | **rewired to inject a failing `ToolUsageStore` double** (the `noopUsageStore` idiom) via `Options`; the ENOTDIR *arrangement* is **retired** (the grill's proposed relocation to `cmd/tellme` is **withdrawn** — `cli.Run` hard-binds `os.Std*` at `cli.go:313-321`, so it cannot observe the stderr/stdout the test asserts; the adapter's read-error path is already pinned at `internal/infrastructure/history/tool_usage_test.go:126`) |
| `internal/cli/turn_test.go` | **8 direct `runTurn` sites** build `Options`/`dp`; `env()` and `factoryReturning` helpers are adapted (the gateway stays an explicit `factory` arg) |
| `internal/cli/prompt_multiline_test.go` | **5 direct `run` sites**; the two archive-path sites (`:183`,`:207`) build `fakeStore`/`capturingUsageStore` doubles for `NewHistoryStore`/`NewUsageStore`; the other three need the threaded value only |
| `internal/cli/mcp_discovery_test.go` | the `mcpDiscoveryConfig` fakes **move to `internal/infrastructure/mcp`** tests (the orchestration's new home) |
| `cmd/tellme` | gains a test: the **relocated assembler gate** + a **deps-construction smoke** — **no stream assertions** (fix-8) |

NFR-003: after the round there is **no global to mutate**; the contract is hermetic by construction (and `t.Parallel()`-safe — the round-029/031 precedent). **Invariant:** no `internal/cli` test file imports `internal/infrastructure/*` (the gate merges `.TestImports`/`.XTestImports`, `arch_test.go:39`) — which is why the migrated tests use in-package doubles.

### D6 — Gate ratchet discipline (self-checking extraction)

The extraction must keep the R1 gate **green at every commit**: a removed violation + its baseline line land **in the same commit** (removing the edge without updating `baseline.txt` fails on **staleness**; leaving a new violation fails on the **count**). Regenerate with `make verify-architecture-update`; verify with `make verify-architecture`.

### D7 — Governance: ADR 0013 (the injected-`Dependencies` pattern)

The pattern is a **project-level rule** future rounds must cite (R3/R4 will touch the same seams; R5/#101 builds on it). ADR **0013** records: the composition-root home (`cmd/tellme`), the domain-typed `deps.Dependencies` + cli-local `Options` split (and *why* the split is forced by the tier table), the `agentTools()` property preservation, and the MCP-discovery re-home. It cites the **ADR-055/060/074** injection lineage; no existing ADR is superseded. Follows the ADR-0011 precedent (a gate/rule change R2–R4 cite).

### D8 — `verify-mcp-sdk-confinement` and the `di` boundary

The gate confines `github.com/modelcontextprotocol/go-sdk` to `internal/infrastructure/mcp/**`. D4's move keeps every SDK/wire call inside that package. `internal/infrastructure/di` (the config-validator + MCP factories) is now imported **only by `cmd/tellme`**; the `tools/arch` gate will confirm `internal/cli` no longer imports it.

### D9 — Interface posture: api/data/dsl NOOP

- `/axb-api-plan` = **NOOP** — no HTTP/OpenAPI surface.
- `/axb-data-plan` = **NOOP** — no persisted/runtime state change (a structural re-home). The baseline file is a **repo artifact**, not runtime state; the gate reads nothing at run time.
- `/axb-dsl-refine` = **NOOP** — no user-facing CLI contract change; the vocabulary stays **11** class phrases; the topology audit is unchanged. Precedent: rounds 020/031/041/042/043 (non-BDD tooling/refactor).
- `/axb-ui-plan` **skipped** (not user-facing).

### D10 — Verification strategy (the witness is the gate, not the suite)

`make verify` (incl. `verify-architecture` with the 7 entries removed + `verify-mcp-sdk-confinement` + 4/4 cross-compile + lint 0 + govulncheck) · `go test -count=1 ./...` green · topology audit unchanged · `stdout` byte-exactness preserved (spot + E2E). **Falsifiability witness**: re-introduce one `internal/cli -> internal/infrastructure/*` import (or restore one removed baseline line) ⇒ the gate goes **red**; revert. This is the round's acceptance carrier (#92: truth/DSL ~NOOP ⇒ a green suite alone is false confidence — the round-009 trap).

### D11 — Scope guard

Touch only: `internal/cli/**`, `internal/app/deps/**` (**new**), `internal/domain/tools/**` (the `OutputSink` port), `internal/infrastructure/mcp/**` (the discovery move), `internal/infrastructure/tools/**` (the `OutputSink` type + binder signature), `cmd/tellme/**`, `tools/arch/baseline.txt`, `docs/decisions/**`, `specs/truth/techstack.md`, and the plan package. **No** adapter behaviour change, **no** domain business-logic change, **no** flag/exit-code/stream change. The strict de-coupling is **out of scope** → [#101](https://github.com/gosharplite/tellme/issues/101).

### D12 — Risks & mitigations

| Risk | Mitigation |
| --- | --- |
| `cmd/**` is untested today | add a `cmd/tellme` test (deps smoke + the relocated assembler gate) — D5 |
| The assembler gate silently weakens when it moves | it must keep iterating the **non-overridable** assembler + the registry-name cross-check (A7) |
| A removed edge without a baseline update reddens `dev` | same-commit baseline regeneration (D6) |
| MCP move leaks an SDK import | `verify-mcp-sdk-confinement` is a member of `make verify` (D8) |
| The `ui`-typed seam accidentally lands in `deps` | `deps` tier-2 ceiling ⇒ RULE-A fails the gate (D2) |
| Behaviour drift in the turn path | `stdout` byte-exactness + full E2E (D10) |
| A `deps` field with no consuming signature (the grill's finding) | every field names its consumer (D2); the threading covers `run`/`runTurn`/`newTurnSpinner`/`augmentRegistryWithMCP` (D2/D5) |
| Hoisting the metrics read before `spinnerGate` | the `dp.NewMetricsProvider` read MUST stay **after** the gate (D5, fix-5) |

### D12b — Grill-round fold (PR #102)

A grill round (architect vs griller) verified the plan against the tree and found the migration surface under-recorded in six places plus a type literal that would not compile; **on the plan as written the 7 → 0 DoD was formally unreachable**. The operator-approved fold (G1–G4 + fixes 1–8 — see `spec.md` *Grill-round fold*) is applied across this file: **fix-1** (D2 — `NewTUIRegistry` + the runner-seam widening + `Options` threading), **fix-2** (D2 — `newRenderer` deleted, one seam), **fix-3** (D3/D11 — `OutputSink.Enabled()` + the tools touch-list), **fix-4/5/6** (D2/D5 — the threading + `newTurnSpinner`/`augmentRegistryWithMCP` + the argument-taking store/prompt-tracker signatures), **fix-7** (D5 — `turn_test.go`/`prompt_multiline_test.go`; the `cmd/tellme` relocation of the read-error test **withdrawn**), **fix-8** (D5 — the `cmd/tellme` test takes no stream assertions; the no-test-infra-import invariant). Operator decisions: **G1** accept the wide bag (ADR 0013 records interface segregation as a *rejected alternative*); **G2** rename `options` → `flags`; **G3** fold into PR #102 (plan half still in flight); **G4** fix the two residuals (below).

## Truth impact (summary)

`specs/truth/techstack.md` **MODIFY** (Build & Tooling carries no change; the CLI-Application / Skills / MCP-Client / Testing rows naming the moved symbols are MODIFYs) — recorded in `truth-delta.md`. `contracts/**` + `data/**` + `features/**` **NOOP**. New **ADR 0013** + index row.
