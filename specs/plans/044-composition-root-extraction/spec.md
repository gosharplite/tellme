# Feature Specification: composition-root extraction (round 044)

**Feature Branch**: `044-composition-root-extraction`

**Created**: 2026-09-18

**Status**: Draft — plan half only (specify → spec-by-example → technical-research → system-analysis). Anchor issue [#100](https://github.com/gosharplite/tellme/issues/100) — **R2** of [#92](https://github.com/gosharplite/tellme/issues/92). Clarify round 1 locked **Q1–Q7** (below); the PR #102 **grill round** then drove a **fold** (operator decisions G1–G4 + fixes 1–8) recorded in the *Grill-round fold* section — the plan's migration surface was under-recorded and the 7 → 0 DoD was formally unreachable until this fold.

**Input**: Issue [#100](https://github.com/gosharplite/tellme/issues/100) (R2 of the #92 gate-first split), resolved via the AIxBDD process. #92: *"`internal/cli` is both the application layer and the composition root: `cli.go` (1236 lines) directly imports five `internal/infrastructure/*` packages and holds package-level factory vars … Static analysis reports 7 layer violations (0 cycles)."* R1 ([#93](https://github.com/gosharplite/tellme/issues/93), round 042 / PR [#95](https://github.com/gosharplite/tellme/pull/95), ADR 0011) shipped the `verify-architecture` layer-discipline gate + a committed baseline of those 7 violations; R2 makes the baseline **shrink to 0**.

**Behaviour intent**: **MODIFY (structural — behaviour-preserving)**. Move the composition root out of `internal/cli` (to **`cmd/tellme`**), inject the concrete construction through a domain-typed **`Dependencies`** value so `internal/cli` imports **no** `internal/infrastructure/*` package, and delete the package-level factory vars — with **zero** change to any flag, exit code, stream contract, formatting, or the `tellme` CLI contract. The round's DoD is **falsifiable**: the R1 gate reported `0 new, 0 stale` and the 7 `internal/cli -> internal/infrastructure/*` baseline entries **removed** (7 → 0). This is a **refactor slice** in the ADR-055/060/074 injection lineage.

---

## Locked decisions (clarify round 1 — Q1–Q7)

| # | Decision |
| --- | --- |
| **Q1** | **`cmd/tellme` is the composition-root home (option A).** It is **outside `internal/`**, hence **exempt** from the R1 gate's tier table (which ranks `internal/app/*` at **2** and `internal/infrastructure/*` at **3**; RULE-B bans an application-tier package from importing `internal/infrastructure/**`). Placing the root in `internal/app/**` would **re-create** the exact violation R2 removes. `cmd/**` currently has no test file — the round adds a `cmd/tellme` test. Rejected **(B)** extending `internal/infrastructure/di` (viable but conceptually surprising: build-time assembly under `infrastructure/`). |
| **Q2** | **A new `internal/app/deps` package holds the injected `Dependencies` struct (option β), carrying domain-typed fields only.** `internal/app/deps` is tier **2**, so it may import **domain (0)** and **config/home (1)** but **not** `ui (5)` / `agent (4)` (either would be an upward import — RULE-A). It therefore holds **only the seams that today cross into `infrastructure/`**. `cmd/tellme` constructs the concrete adapters (exempt) and assigns them into the struct. Rejected **(α)** the struct inside `internal/cli` (would enable the stricter AC2 that Q5 declines) and **(γ)** ~10 func-args to `Run`. |
| **Q3** | **Relocate the assembler; inject `NewToolRegistry` + `BindToolOutput` + `BindSkillsCatalog`; move `ToolOutputSink` → `domaintools.OutputSink` (option 1).** `agentTools()` + `newToolRegistry` move to `cmd/tellme`; `deps` gains a bare-registry build func and the two binder funcs; the infra-typed `ToolOutputSink` becomes a neutral domain port (`domaintools.OutputSink{ Begin() ; Writer io.Writer ; End() ; Enabled() }` — `Enabled()` is a **method**, see the grill fold fix-3). Most faithful three-roles→three-funcs translation; strengthens the port story. Rejected **(2)** build-the-fully-bound-registry (its neutral `opts` is essentially the domain `OutputSink` anyway) and **(3)** inject constructors (would reintroduce a mutable package global — the very thing R2 deletes). |
| **Q4** | **Move MCP discovery orchestration into `internal/infrastructure/mcp`; inject a func-typed `deps.MCPDiscoverer` (option a).** `mcp_discovery.go`'s `discover`/`discoverServer`/`serverResult` + the `mcp.*` wire helpers move next to the adapter; `di.NewRemoteClient` + `di.NewGhTokenResolver` are **passed in**, so `internal/infrastructure/mcp` imports neither `di` nor the SDK in any new way — preserving `verify-mcp-sdk-confinement`. `deps` exposes `MCPDiscoverer func(ctx, map[string]config.MCPServerConfig) (tools []domaintools.Tool, warnings []string, close func())` — types `internal/cli` already imports; no new domain result type. Rejected **(b)** orchestration in `cmd/tellme` (untestable; MCP-wire knowledge outside the confined package) and **(c)** keep-in-cli (the `mcp` baseline edge would remain — does not reach the DoD). |
| **Q5** | **R2's DoD is the 7 RULE-B edges → 0 (falsifiable by the gate).** The stricter #92 AC2 wording ("`internal/cli` depends only on `internal/domain/*`, stdlib, and application utilities") is **not** R2's scope: RULE-B is an *upward*-import rule and does **not** constrain a downward import, so `cli → ui/agent/config/home/app` is legal and unbaselined; the strict form would require inverting nearly every wiring call (a multi-round programme, and the gate cannot even express it without a new rule). Recorded as follow-up issue **[#101](https://github.com/gosharplite/tellme/issues/101)** (R5 — strict de-coupling). |
| **Q6** | **One injected value: `cli`'s entry takes `Options{Deps deps.Dependencies; RunTUIPrompt tuiPromptRunner}`; delete every factory var (option T1, as amended by the grill fold).** Because `internal/app/deps` (tier 2) **cannot** import `ui`, the **one** presentation seam whose type is unexported-cli (`tuiPromptRunner` over `resolution`/`runtimeEnv`) lives in a **`cli.Options`** struct defined **in `internal/cli`** and is **nil-defaulted inside `cli`** (a `cmd` package cannot name the unexported types). The `newRenderer` var is **deleted** (inlined `ui.NewRenderer()`; the renderer's injectable seam stays `runtimeEnv.renderer`). The entry point takes **one** value: `cli.Run(args, version, opts)`. All 8 package-level vars (`newGateway`, `newHistoryStore`, `newUsageStore`, `newToolUsageStore`, `newToolRegistry`, `newRenderer`, `newTUIPromptRunner`, `userHomeDir`) are **deleted**; tests build values, making the contract hermetic by construction. The round-031 **`agentTools()` well-formedness gate** moves to `cmd/tellme`; `mcp_discovery_test.go`'s `mcpDiscoveryConfig` fakes move into `internal/infrastructure/mcp` tests. Rejected **(T2)** keep any var (leaves a global) and **(T3)** neutral app-layer ui ports (invents ports for pure presentation). |
| **Q7** | **The strict-scope follow-up lives as a new child of [#92](https://github.com/gosharplite/tellme/issues/92): [#101](https://github.com/gosharplite/tellme/issues/101)** (R5). Repo doctrine: a forward item belongs on a **live issue body**, not a frozen plan package or a review comment (§19-c; cf. #60→#91, #69→#92). |

---

## Grill-round fold (PR #102 — operator decisions G1–G4 + verified fixes 1–8)

A grill round (architect vs griller; transcript: https://gist.github.com/gosharplite/b3e8f0bc4d328187399cff800a738829) verified the plan against the tree and found the migration surface **under-recorded in six places** plus a type literal that would not compile. **On the plan as written the 7 → 0 DoD was formally unreachable.** The fold below (operator-approved) supersedes the corresponding clauses of the *Locked decisions* table and the FRs.

**Operator decisions:**

| # | Decision |
| --- | --- |
| **G1** | **Accept the wide `deps.Dependencies` bag** (now ~12 fields); record **interface segregation** as an explicit *rejected alternative* in ADR 0013. The bag is a known design smell, deliberately deferred. |
| **G2** | **Rename the parsed-flags struct `options` → `flags`** (kills the `options`/`turnOptions`/`Options` tri-collision). The injected value's parameter is named **`dp`** (`dp deps.Dependencies`), never `deps` (which would shadow the package). |
| **G3** | **Fold fixes 1–8 into PR #102** (the plan half is still in flight — the package is *not* delivered/frozen, so `plan-package-frozen` does not bite). |
| **G4** | **Fix both non-blocking residuals** (see *Residuals* below). |

**Verified fixes (folded into FRs / `research.md` D2/D3/D5/D11 / ADR 0013):**

1. **TUI assembly (fix-1).** The `-i` suggestion source consumes a **3-reader** tool set (`cli.go:229`), **not** the 7-tool `NewToolRegistry`; the `tuiPromptRunner` seam's type (`cli.go:206`, unexported `resolution`/`runtimeEnv`) cannot receive deps. → add **`deps.NewTUIRegistry func() domaintools.Registry`**, **widen the runner seam** to receive `Options`, and **thread `Options`** through `Run` → `run` → `runTUIPrompt`. Without this, the `infratools`/`infrhistory` edges survive.
2. **Renderer (fix-2).** `newRenderer` is a **production-only** default (0 test hits; its doc comment is false) and its `answerRenderer` type is **unexported** → **delete `newRenderer`**, inline `ui.NewRenderer()` at its single site; `cli.Options = { Deps deps.Dependencies ; RunTUIPrompt tuiPromptRunner }` — **one** presentation seam (nil-defaulted *inside* `cli`; `cmd/tellme` cannot populate unexported-typed fields). The renderer's injectable seam stays `runtimeEnv.renderer`.
3. **`domaintools.OutputSink` (fix-3).** The port MUST carry **`Enabled()`** (a *method*, not a field) — `command.go` calls it **4×** and the `{}`-is-disabled contract is load-bearing. → `domaintools.OutputSink{ Begin func() ; Writer io.Writer ; End func() }` + `func (s OutputSink) Enabled() bool { return s.Writer != nil }`; correct the pinned literal everywhere, and **reconcile FR-013's touch-list with `research.md` D11** (include `internal/infrastructure/tools/**`).
4. **`runTurn` threading (fix-4).** `runTurn` widens with `dp deps.Dependencies` (the gateway factory stays an explicit `factory` argument); `research.md` D5 MUST add **`turn_test.go`** (8 direct `runTurn` sites).
5. **Orphan `deps` fields (fix-5).** **`newTurnSpinner`** (`cli.go:865`, telemetry) and **`augmentRegistryWithMCP`** (`cli.go:1134`, MCP) own `deps.NewMetricsProvider` / `deps.MCPDiscoverer`; both widen with `dp`, and the metrics-provider read MUST stay **after** the `spinnerGate` short-circuit (hoisting it would construct the provider on a gated-off turn).
6. **`UserHomeDir` orphan (fix-6).** `deps.NewToolUsageStore` and `NewPromptTracker` MUST carry the **argument-taking signatures** they have today (`func(func() (string, error)) …` / `func(home string, userHome func() (string, error)) …`), so `internal/cli` calls `dp.NewToolUsageStore(dp.UserHomeDir)` / `dp.NewPromptTracker(res.Home, dp.UserHomeDir)` — no field without a consumer.
7. **`run` callers (fix-7).** `run` is called directly from **8 test sites in 3 files**; `research.md` D5 MUST add them — including **`prompt_multiline_test.go`** (5 sites; the two archive-path sites build `fakeStore`/`capturingUsageStore` doubles). And the grill's proposed **`cmd/tellme` relocation of `TestRenderToolUsageDiagnosesReadError` is WITHDRAWN**: `cli.Run` hard-binds `os.Stdin/Stdout/Stderr` (`cli.go:313-321`; no stream seam, no `os.Std*` swap in the repo), so that test stays in `internal/cli` with a **failing `ToolUsageStore` double** (only the ENOTDIR *arrangement* is retired; the adapter's read-error path is pinned at `tool_usage_test.go:126`).
8. **Record the constraints (fix-8).** The `cmd/tellme` test is **assembler-gate + deps-smoke only — no stream assertions**; and NFR-003 gains the invariant *no `internal/cli` test file imports `internal/infrastructure/*`* (the gate merges `.TestImports`/`.XTestImports`, `arch_test.go:39`).

**Residuals (G4):**

- The `truth-delta.md` `/axb-api-plan` row said it inspected `specs/truth/contracts/**`, a directory that **does not exist** → reworded to name what was actually inspected.
- The `STATUS.md` line "`dev` = last delivered round 043 (`a2fbafc`)" was loose beside the branch base → tightened to name both heads unambiguously.

---

## Grounded in the current system

Static read 2026-09-18 @ `dev` `a2fbafc`.

**The 7 baselined RULE-B violations** (`tools/arch/baseline.txt`, ADR 0011):

```
internal/cli -> internal/infrastructure/di
internal/cli -> internal/infrastructure/history
internal/cli -> internal/infrastructure/llm
internal/cli -> internal/infrastructure/mcp
internal/cli -> internal/infrastructure/skills
internal/cli -> internal/infrastructure/telemetry
internal/cli -> internal/infrastructure/tools
```

(The gate also baselines **1** RULE-A `internal/agent -> internal/ui`; that is **R3/R4** territory, **not** R2 — see *Out of scope*.)

**Exact construction sites in `internal/cli` production code** (must all leave):

| infra package | symbol(s) | site(s) |
| --- | --- | --- |
| `infrastructure/history` | `NewFileStore`, `NewUsageStore`, `NewToolUsageStore`, `NewGlobalPromptTracker` | `cli.go:168,179,198,219,300` |
| `infrastructure/llm` | `NewGateway` | `cli.go:619` (`var newGateway`) |
| `infrastructure/tools` | `NewFilesystemTools`, `NewWriteTools`, `NewCommandTool`, `NewSkillsTool`, `NewRegistry`, `BindToolOutput`, `BindSkillsCatalog` | `cli.go:229,756,1153-1155,1161,1170` |
| `infrastructure/skills` | `Load` | `cli.go:1171` |
| `infrastructure/telemetry` | `NewSystemMetricsProvider` | `cli.go:869` |
| `infrastructure/di` + `infrastructure/mcp` | `di.NewRemoteClient`, `di.NewGhTokenResolver`; `mcp.TokenSource`, `ResolveAuthorization`, `ResolveMCPTimeout`, `NamespacedName`, `ValidToolName`, `NormalizeMCPSchema`, `NewTool`, `SwitchToCallTimeout`, the `*Warning` builders | `mcp_discovery.go:28,45,46,130-164` |

**Package-level factory vars** (the current DI seam — tests override them): `newGateway`, `newHistoryStore`, `newUsageStore`, `newToolUsageStore`, `newToolRegistry`, `newRenderer`, `newTUIPromptRunner`, `userHomeDir` (`cli.go:155-619,1124`), plus `defaultMCPDiscovery` (`mcp_discovery.go:43`).

**The gate's tier table** (`tools/arch/arch_test.go`): `config`/`home` = 1 · `internal/app/*` = 2 · `infrastructure` = 3 · `agent` = 4 · `ui` = 5 · `internal/cli` = 6 · `domain` = 0. **RULE-B** = an application-tier (`internal/cli` **or** `internal/app/*`) package MUST NOT import `internal/infrastructure/**`. Packages outside `internal/` (i.e. `cmd/**`) are **exempt** — never passed to the tier function.

**The round-031 recurrence gate** `TestAgentToolSchemasAreWellFormed` (`internal/cli/tool_registry_test.go`) iterates the **non-overridable** `agentTools()` (PR #65 ARCH-1) and cross-checks `newToolRegistry().Tools()` names — wherever the assembler goes, this gate **moves with it**.

**`mcp_discovery.go` is orchestration, not a factory**: `discover` (fan-out over enabled remote servers), `discoverServer` (bounded per-server probe), `serverResult` (aggregation), the `mcpRun` result, and the `mcp.*` wire calls. Its result is already domain-typed (`[]domaintools.Tool`, `[]string`, `close func()`); its only infra couplings are the two `di` constructors.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - `internal/cli` stops being the composition root (Priority: P1)

As an architect/maintainer, I want `internal/cli` to import **no** `internal/infrastructure/*` package, so the application layer no longer constructs its own adapters and the R1 layer-discipline gate reports the 7 `internal/cli -> internal/infrastructure/*` violations **gone** (0 violations, 0 cycles).

**Why this priority**: it is the round's entire reason to exist; it is the one clause of #92's AC2 that is **falsifiable by the gate**.

**Independent verification**: run the R1 gate (`make verify-architecture` / the tagged test) — before the round it reports 7 baselined `internal/cli -> infrastructure` edges; after, those baseline lines are **removed** and the gate is green (0 new, 0 stale). Property: the graph's `internal/cli` out-edges contain **no** `internal/infrastructure/**` node.

**Acceptance Scenarios**:

1. **Given** the R1 layer-discipline gate on the round head, **When** it evaluates the module graph, **Then** the `internal/cli -> internal/infrastructure/*` baseline entries are **absent** and the gate reports **0 new, 0 stale** violations.
2. **Given** the extracted `internal/cli`, **When** its imports are enumerated, **Then** no `internal/infrastructure/**` path appears (verified by the gate, **not** by inspection).
3. **Given** the gate as a **ratchet**, **When** any non-baselined `cli -> infrastructure` edge reappears, **Then** the gate **fails** (the falsifiability witness, FR-012).

**Functional Requirements**:

- **FR-001**: `internal/cli` (production **and** test files) MUST import **no** `internal/infrastructure/*` package after the round.
- **FR-002**: The concrete construction of infrastructure adapters, the LLM gateway, the MCP discovery client, and the tool registry MUST move to a composition root located at **`cmd/tellme`** (exempt from the tier table); `cmd/tellme` MUST build the adapters/pipelines and pass them into `internal/cli` via a single entry-point value.
- **FR-008**: The extraction MUST remove the 7 `internal/cli -> internal/infrastructure/*` lines from `tools/arch/baseline.txt` **in the same commit** as the edge removal (removing a violation without updating the baseline fails on **staleness**; a new violation fails on the **count**), so `dev` is green at **every** commit.

**Non-Functional Requirements**:

- **NFR-001**: The extraction MUST be **deterministic** (no wall-clock/timing dependence; the seam is a value, not a global).

---

### User Story 2 - Wiring is injected through a domain-typed seam (Priority: P2)

As a maintainer, I want the CLI's dependencies injected through a domain-typed `Dependencies` value (built by the composition root), so `internal/cli` depends only on domain ports and the package-level factory globals are gone.

**Why this priority**: it is the *mechanism* that retires US1's edges, and it removes the mutable-global test seam (#92 AC5).

**Independent verification**: the round head defines `internal/app/deps.Dependencies` (domain-typed fields only) + a `cli.Options` (ui seams); no package-level factory var remains in `internal/cli`; the tests that mutated globals construct values instead.

**Acceptance Scenarios**:

1. **Given** the composition root, **When** `cmd/tellme` runs, **Then** it constructs the concrete adapters and calls `cli.Run(args, version, opts)` with **one** injected value.
2. **Given** `internal/app/deps`, **When** it is compiled, **Then** it imports **only** `internal/domain/**`, `internal/config`, `internal/home`, and stdlib (no `internal/infrastructure/**`, no `internal/ui`, no `internal/agent`).
3. **Given** the `internal/cli` package, **When** it is compiled, **Then** it declares **no** package-level factory var (`newGateway`/`newHistoryStore`/`newUsageStore`/`newToolUsageStore`/`newToolRegistry`/`newRenderer`/`newTUIPromptRunner`/`userHomeDir`).

**Functional Requirements**:

- **FR-003**: A new `internal/app/deps` package MUST define the injected `Dependencies` struct with **domain-typed fields only** (the seams that cross into `infrastructure/`): the gateway factory, the history/usage/tool-usage store factories, the prompt-tracker factory, the **agent** tool-registry build func (`NewToolRegistry`), the **TUI** tool-registry build func (`NewTUIRegistry` — the **3-reader** set the `-i` suggestion source consumes, *distinct* from the 7-tool agent registry), the two tool binders, the metrics-provider factory, the MCP discoverer, and `UserHomeDir`. The store/prompt-tracker fields MUST carry the **argument-taking signatures they have today** (`NewToolUsageStore func(func() (string, error)) history.ToolUsageStore`; `NewPromptTracker func(home string, userHome func() (string, error)) history.PromptTracker`) so `UserHomeDir` has real consumers (`dp.NewToolUsageStore(dp.UserHomeDir)`, `dp.NewPromptTracker(res.Home, dp.UserHomeDir)`) — **no field without a consumer**.
- **FR-004**: `agentTools()` MUST stay **parameterless + read-free** and the **non-overridable** production assembler (round-033 FR-009; round-031 gate, PR #65 ARCH-1). It relocates to `cmd/tellme`; the round-031 **well-formedness gate** (`TestAgentToolSchemasAreWellFormed`) moves with it and keeps iterating the assembler (property preserved, location moved).
- **FR-005**: The tool bindings MUST be injected (`NewToolRegistry`, `BindToolOutput`, `BindSkillsCatalog`) and the infra-typed `ToolOutputSink` MUST become a neutral domain port **`domaintools.OutputSink`** (`{ Begin func() ; Writer io.Writer ; End func() }` **plus a method `Enabled() bool { return s.Writer != nil }`** — the `command.go` block calls `Enabled()` 4× and the `{}`-is-disabled contract is load-bearing, so the port mirrors today's type member-for-member) so `internal/cli` names no infra type.
- **FR-006**: The MCP discovery orchestration MUST move into `internal/infrastructure/mcp` (as a cohesive `Discover`), taking its two `di` constructors **as parameters**; `deps` MUST expose a func-typed `MCPDiscoverer`. `verify-mcp-sdk-confinement` MUST stay green (the SDK/wire stays confined to `internal/infrastructure/mcp/**`).
- **FR-007**: Every package-level factory var in `internal/cli` MUST be **deleted**; the **one** presentation seam whose type references unexported cli types (`newTUIPromptRunner`, type `func(ctx, resolution, runtimeEnv) (string, bool, error)`) MUST live in a **`cli.Options`** struct defined in `internal/cli` (nil-defaulted internally; `cmd/tellme` populates only `Options.Deps`); the `newRenderer` var is **deleted** (inline `ui.NewRenderer()`; the renderer seam stays `runtimeEnv.renderer`). The injected value MUST be **threaded** through `Run`→`run`→`runTurn`/`renderTurn`/`dispatchReporting`/`renderToolUsage`/`renderHistoryList`/`renderNewSession`/`newCallRenderer`/`persistTurnUsage`/`runTUIPrompt` **and** the leaf helpers `newTurnSpinner`/`augmentRegistryWithMCP`, as a parameter named **`dp`** (`dp deps.Dependencies`); the parsed-flags struct is renamed **`options` → `flags`** (G2). All test files that mutated the globals MUST be migrated to construct the injected value — **including `turn_test.go` (8 `runTurn` sites) and `prompt_multiline_test.go` (5 `run` sites; the two archive-path sites build `fakeStore`/`capturingUsageStore` doubles)** — and no `internal/cli` test file may import `internal/infrastructure/*`.

**Non-Functional Requirements**:

- **NFR-002**: `internal/domain/**` MUST stay **100% pure** (RULE-C); the `Dependencies` type MUST NOT be placed in `internal/domain` (a composition bag is not a domain concept), and no domain package may carry concrete factory types.
- **NFR-003**: Tests MUST stay **hermetic** — no *required* reliance on mutable package globals; a test constructs its own injected value. **Invariant:** **no `internal/cli` test file imports `internal/infrastructure/*`** — the R1 gate merges `.TestImports`/`.XTestImports` (`tools/arch/arch_test.go:39`), so a test import would keep a baseline edge alive and make 7 → 0 unreachable (this is why the migrated tests use in-package doubles, not real adapters).

---

### User Story 3 - Zero behavioural change (Priority: P3)

As a maintainer, I want the round to be a **pure structural refactor**: under the same inputs every flag, exit code, stream byte, and formatting outcome is identical, and the full regression is green.

**Why this priority**: it bounds the round to structure and protects the "no user-facing change" #92 constraint.

**Independent verification**: `make verify` green; godog E2E green; `stdout` byte-exactness unchanged; the DSL vocabulary (11 class phrases) unchanged; the topology audit unchanged.

**Acceptance Scenarios**:

1. **Given** a supported prompt path, **When** the round head runs it, **Then** `stdout` is byte-identical to the pre-round head and every exit code is unchanged.
2. **Given** the full `make verify` + E2E, **When** it runs on the round head, **Then** it is green (no behavioural/stream-contract regression).
3. **Given** the offline `--tool-usage` report and the `-i` TUI surface, **When** they run, **Then** their existing contracts are unchanged (the offline path still builds its registry without reading `docs/skills`; the TUI still tears down and resumes the standard surface).

**Functional Requirements**:

- **FR-009**: The round MUST NOT change any user-facing CLI behaviour: no flag, no exit code, no stream contract, no formatting, no DSL row (vocabulary stays **11**), no stored record.
- **FR-010**: `internal/domain/**` MUST stay pure and tests MUST stay hermetic (NFR-002/NFR-003 restated as a global gate).
- **FR-012**: The round's witness MUST be **the gate + unit seams**, NOT the E2E suite (#92 constraint: truth/DSL are ~NOOP, so a green suite alone is false confidence). The round MUST include a **falsifiability witness**: re-introduce one `internal/cli -> internal/infrastructure/*` import (or one removed baseline line), confirm the gate goes **red**, then revert.

**Non-Functional Requirements**:

- **NFR-004**: `make verify` (incl. `verify-architecture`, `verify-mcp-sdk-confinement`, 4/4 cross-compile, lint, govulncheck), `go test -count=1 ./...`, and the Gherkin/DSL topology audit MUST be green; the round introduces **no** new Gherkin/DSL rows (`/axb-dsl-refine` **NOOP**).

---

## Requirements *(mandatory)*

### Global Requirements

#### Functional Requirements

- **FR-011**: The round MUST record the injection pattern in a new **ADR 0013** (`docs/decisions/0013-*.md` + the `docs/decisions/README.md` index row) citing the ADR-055/060/074 lineage, and MUST update `specs/truth/techstack.md` for every **named/relocated symbol** (see *Truth obligations* below) — these are **MODIFY**s, **not** NOOPs (`truth-current`).
- **FR-013**: Scope guard — the round MUST touch only: `internal/cli/**`, `internal/app/deps/**` (**new**), `internal/domain/tools/**` (the `OutputSink` port), `internal/infrastructure/tools/**` (the `ToolOutputSink` type removal + the `command.go`/`tooloutput.go` type rename — reconciles this list with `research.md` D11), `internal/infrastructure/mcp/**` (the discovery move), `cmd/tellme/**`, `tools/arch/baseline.txt`, `docs/decisions/**`, `specs/truth/techstack.md`, and the plan package. It MUST NOT change infrastructure adapter behaviour, domain business logic, flags, exit codes, or stream contracts.

#### Non-Functional Requirements

- **NFR-005**: **stdlib-only** — the round introduces **no** new module dependency; `go.mod`/`go.sum` unchanged.
- **NFR-006**: **POSIX-only** (the repo is bash/POSIX-only); no Windows variant.

### Key Entities

- **Composition root**: the assembly site (`cmd/tellme`) that constructs concrete adapters and injects them into `internal/cli`.
- **`deps.Dependencies`**: the domain-typed value carrying the infra-crossing seams (gateway, stores, prompt tracker, the **agent** tool-registry build func, the **TUI 3-reader** tool-registry build func, the two tool binders, metrics provider, MCP discoverer, `UserHomeDir`).
- **`cli.Options`**: the cli-local value wrapping `deps.Dependencies` + the **one** `internal/ui`-typed presentation seam (`RunTUIPrompt`; the `newRenderer` var is deleted).
- **`domaintools.OutputSink`**: the neutral domain port replacing the infra-typed `ToolOutputSink` (`Begin`/`Writer`/`End` + the `Enabled()` method).
- **`MCPDiscoverer`**: a func-typed port (`tools []domaintools.Tool`, `warnings []string`, `close func()`) implemented by `internal/infrastructure/mcp`.
- **`flags`** (renamed from `options`): the parsed-flags struct in `internal/cli` — renamed to avoid the `options`/`turnOptions`/`Options` tri-collision (G2).

### Truth obligations (explicit — `truth-current`)

`specs/truth/techstack.md` **names refactored symbols**; each is a real **MODIFY** with a `truth-delta.md` row:

- **Agent tool-schema gate** row — "the **non-overridable production assembler `agentTools()`**" (its home moves to `cmd/tellme`).
- **Skills listing tool** row — "registered in the production `agentTools()` assembler".
- **MCP client protocol library** row — `verify-mcp-sdk-confinement` (the confinement boundary is preserved; the orchestration home changes).
- **MCP tool-schema normalization** row and **Local fake MCP server** row — the discovery/normalization logic's home (`internal/infrastructure/mcp`).
- **Stronger tool-schema recurrence guards** — the round-031 gate's home.
- `ToolOutputSink` / `BindToolOutput` — renamed/moved (`domaintools.OutputSink` + injected binder).
- *(grill fold)* The **`-i` suggestion tool set** (the 3-reader registry from `NewFilesystemTools()`) is now built by `deps.NewTUIRegistry`; the **`flags`** rename and the `cli.Options` reduction to one seam are plan-side mechanism (no techstack symbol).

`/axb-api-plan` = **NOOP** (no HTTP surface); `/axb-data-plan` = **NOOP** (no persisted/runtime state change).

## Success Criteria *(mandatory)*

- **SC-001**: The R1 layer-discipline gate is **green** with the `internal/cli -> internal/infrastructure/*` baseline entries **removed** (7 → 0), reporting **0 new, 0 stale** violations and **0 cycles**; `internal/cli` imports no `internal/infrastructure/**` (verified by the gate). (covers FR-001, FR-002, FR-008)
- **SC-002**: `internal/app/deps.Dependencies` exists with **domain-typed fields only**; `internal/cli` declares **no** package-level factory var; tests construct injected values (no *required* mutable global). (covers FR-003, FR-007, NFR-003)
- **SC-003**: `agentTools()` remains **parameterless + read-free + non-overridable**, and the round-031 well-formedness gate iterates it at its new home. (covers FR-004)
- **SC-004**: Full regression green with **zero** behavioural / stream-contract change — `make verify`, godog E2E, exit codes, `stdout` byte-exactness, DSL vocabulary **11**, topology audit unchanged. (covers FR-009, NFR-004)
- **SC-005**: `verify-mcp-sdk-confinement` is green; `internal/infrastructure/mcp` imports neither `di` nor the SDK in a new way; `internal/domain/**` stays pure. (covers FR-006, FR-010, NFR-002)
- **SC-006**: `specs/truth/techstack.md` carries MODIFY rows for every named/relocated symbol; **ADR 0013** + its index row are present; `go.mod`/`go.sum` unchanged. (covers FR-011, NFR-005)
- **SC-007**: **Falsifiability witness** — re-introducing one `internal/cli -> internal/infrastructure/*` import (or restoring one removed baseline line) makes the gate **red**; reverted clean. (covers FR-012)

## Edge Cases

- **`internal/app/deps` tier-2 ceiling** → it may import domain/config/home only; the **one** `ui`-typed presentation seam (`RunTUIPrompt`) therefore lives in `cli.Options` (Q6), not in `deps`. A future attempt to add a `ui` field to `deps` MUST fail the gate (RULE-A).
- **`cmd/**` is untested today** → the extraction adds a `cmd/tellme` test covering **(a)** the relocated `TestAgentToolSchemasAreWellFormed` assembler gate and **(b)** a deps-construction smoke. It takes **NO stream assertions**: `cli.Run` hard-binds `os.Stdin/Stdout/Stderr` (`cli.go:313-321`) and the round adds no stream seam and no in-process `os.Std*` swap (neither pattern exists in the repo). Any stream-observing behaviour test stays in `internal/cli` behind `runtimeEnv`.
- **The `-i` TUI path** → the **one** presentation seam (`RunTUIPrompt`) is injected via `cli.Options`; the `-i` suggestion source's **3-reader** tool set comes from `deps.NewTUIRegistry` (not the 7-tool `NewToolRegistry`; reusing the latter would change the suggested tool names — a behavioural change FR-009 forbids); the existing `-i` teardown/resume contract is unchanged.
- **The round-031 assembler gate relocation** → it must keep iterating the **non-overridable** assembler and cross-checking the registry name set; moving it to `cmd/tellme` preserves the *property* (`agentTools()` parameterless + read-free), not the location (Assumption A7).
- **The offline `--tool-usage` path** → it builds a registry **without reading `docs/skills`** (round-033 FR-009); the extraction MUST preserve that (the injected tool-registry build func must not read the catalog; the catalog binder is called only on the prompt path).
- **A test that mutates a global** → after the round there is no global to mutate; the test constructs the injected value (the `persistence_invariant_test.go` / `tui_*_test.go` / `cli_test.go` (**failing-double, not a relocation**) / `tool_registry_test.go` (**relocates to `cmd/tellme`**) / `mcp_discovery_test.go` (**fakes move to `internal/infrastructure/mcp`**), plus **`turn_test.go` (8 `runTurn` sites)** and **`prompt_multiline_test.go` (5 `run` sites)** migrations).
- **`verify-mcp-sdk-confinement`** → the discovery move must not leak an SDK import outside `internal/infrastructure/mcp/**`.
- **One RULE tier per seam** → the gateway/stores/tool binders cross into `infrastructure` (tier 3) and so are `deps` fields; the **only** `ui`-typed seam is the TUI runner (tier 5), so it is the **one** `cli.Options` field — the split is *forced* by the tier table, not a choice.

## Assumptions

- **A1 (weak E2E carrier)** — the round changes no `tellme` CLI behaviour, so the E2E suite is a **weak acceptance carrier**; the witness is the **gate** + unit seams (FR-012). Per this, **`/axb-spec-by-example` is expected to be thin or NOOP** (no new user-facing business journey) — precedent: rounds 020/031/036/041/042/043.
- **A2 (mechanism is RD)** — the exact `Dependencies` field shapes, the `MCPDiscoverer` signature, the `cmd/tellme` construction layout, and the baseline-regeneration steps are **RD** decisions (`research.md` D-x), not FRs.
- **A3 (ADR required)** — **ADR 0013** records the injected-`Dependencies` pattern (a project-level rule future rounds may cite).
- **A4 (no other interfaces)** — `/axb-api-plan` = **NOOP** (no HTTP surface); `/axb-data-plan` = **NOOP** (no persisted/runtime state); `/axb-ui-plan` **skipped** (not user-facing).
- **A5 (`/axb-dsl-refine` NOOP)** — no feature Rule, Example, step, or `DSLRow` changes (the topology audit stays unchanged); the acceptance carrier is the gate + unit seams.
- **A6 (scope guard)** — see FR-013.
- **A7 (property, not location)** — #100's constraint is the `agentTools()` *property* (parameterless, read-free, non-overridable), not its *file*; relocation to `cmd/tellme` satisfies it, and the round-031 gate moves with the assembler.
- **A8 (ratchet discipline)** — the gate stays green at **every commit** (baseline regeneration in the same commit as each removed edge).

## Out of scope (recorded)

- **The strict de-coupling** ("`internal/cli` imports only `internal/domain/*` + stdlib + app utilities") — follow-up **[#101](https://github.com/gosharplite/tellme/issues/101)** (R5); requires a new gate rule + ADR (Q5).
- **R3/R4 presentation-policy ownership** — the spinner-yield policy (single owner), the `LoopObserver` hook split (`YieldIndicator`/`RestoreIndicator`), the blank-reason predicate / loop→`ui` predicate, and the `internal/agent -> internal/ui` RULE-A edge — all **out of R2**.
- **Ride-alongs** — suggestion-selection `set(items, cursor)` and `NewCommandTool(sink)` ctor injection — only if a touched file makes one free; otherwise parked.
- Changing user-facing CLI behaviour, flags, exit codes, or formatting.
- Rewriting infrastructure adapters or domain business logic.
- Windows · a security/consent layer · conversation pruning · tool-call concurrency (settled exclusions).

## Provenance

- [#92](https://github.com/gosharplite/tellme/issues/92) workstream **#1** (large) — composition-root extraction; AC2/AC5 wording.
- [#100](https://github.com/gosharplite/tellme/issues/100) — the round's anchor (grounded seam inventory + the 6 pre-scoped clarify decisions).
- Round **044** clarify **Q1–Q7** (locked above); the strict-scope follow-up filed as **[#101](https://github.com/gosharplite/tellme/issues/101)**.
- [#93](https://github.com/gosharplite/tellme/issues/93) → round 042 / PR [#95](https://github.com/gosharplite/tellme/pull/95) (ADR 0011) — the gate + baseline that makes this round falsifiable.

## References

- [`tools/arch/baseline.txt`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/baseline.txt) · [`tools/arch/arch_test.go`](https://github.com/gosharplite/tellme/blob/dev/tools/arch/arch_test.go) — the gate + tier table.
- [`docs/decisions/0011-layer-discipline-gate.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0011-layer-discipline-gate.md) — RULE-A/B/C/D + the ratchet/baseline policy.
- [`docs/decisions/0005-tool-call-log-parity.md`](https://github.com/gosharplite/tellme/blob/dev/docs/decisions/0005-tool-call-log-parity.md) — the presentation-policy partition R3/R4 own.
- `specs/truth/techstack.md` (Build & Tooling) — the rows naming the refactored symbols.
