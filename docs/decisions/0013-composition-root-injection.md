# ADR 0013: Composition-root extraction — an injected, domain-typed `Dependencies` seam

**Status**: Accepted

**Date**: 2026-09-18

**Round**: 044 (`044-composition-root-extraction`) — R2 of [#92](https://github.com/gosharplite/tellme/issues/92); anchor [#100](https://github.com/gosharplite/tellme/issues/100).

**Context**: `internal/cli` was both the application layer **and** the composition root: `cli.go` imported five `internal/infrastructure/*` packages and held eight package-level factory vars; `mcp_discovery.go` imported `internal/infrastructure/di` + `internal/infrastructure/mcp`. The R1 layer-discipline gate (ADR 0011) baselined those **7** RULE-B violations. R2 removes them: the composition root moves out of the application layer and dependencies are injected.

## Decision

**D1 — The composition root lives in `cmd/tellme`.** The R1 tier table ranks `internal/app/*` at **2** and `internal/infrastructure/*` at **3**, and RULE-B forbids an application-tier package from importing `internal/infrastructure/**`. `cmd/**` is outside `internal/`, so the gate never ranks it — it is **exempt**. Choosing `cmd/tellme` keeps `internal/app/**` rule-clean; a root under `internal/app` would **re-create** the violation R2 removes.

**D2 — `internal/cli` receives one injected value: a domain-typed `internal/app/deps.Dependencies` plus a cli-local `Options`.** `internal/app/deps` (tier 2) holds the seams that cross into `infrastructure/` (`NewGateway`, `NewHistoryStore`, `NewUsageStore`, `NewToolUsageStore`, `NewPromptTracker`, `NewToolRegistry` — the **7-tool agent** registry, `NewTUIRegistry` — the **3-reader** registry the `-i` suggestion source consumes, `BindToolOutput`, `BindSkillsCatalog`, `NewMetricsProvider`, `MCPDiscoverer`, `UserHomeDir`) using **domain types only**; the store/prompt-tracker fields carry the **argument-taking signatures they have today** so `UserHomeDir` has real consumers. Because tier 2 **cannot** import `internal/ui` (tier 5), the **one** presentation seam whose type references unexported cli types (`RunTUIPrompt`, over `resolution`/`runtimeEnv`) lives in a **`cli.Options`** struct defined in `internal/cli` and is **nil-defaulted internally** (a `cmd` package cannot name the unexported types, so it populates only `Options.Deps`). The entry point is **`cli.Run(args, version, opts)`**; the value is threaded as a parameter named **`dp`** (`dp deps.Dependencies`). The `newRenderer` var is **deleted** (inline `ui.NewRenderer()`; the renderer's injectable seam stays `runtimeEnv.renderer`). The split is **forced by the tier table**, not a preference. The composition bag is **not** placed in `internal/domain` (a composition concept is not a domain concept).

**D3 — `internal/cli` constructs no infrastructure, and holds no package-level factory var.** All eight vars are deleted; the `agentTools()` assembler + `newToolRegistry` relocate to `cmd/tellme`; the infra-typed `ToolOutputSink` becomes the neutral domain port **`domaintools.OutputSink`** (`{ Begin func() ; Writer io.Writer ; End func() }` **plus the method `Enabled() bool { return s.Writer != nil }`** — `command.go` calls it 4× and the `{}`-is-disabled contract is load-bearing); the two tool binders are injected. The **`agentTools()` property** — parameterless, read-free, non-overridable (round-033 FR-009; round-031 gate) — is preserved; only its **file** moves, and the round-031 `required ⊆ properties` gate moves with it.

**D4 — MCP discovery orchestration moves into `internal/infrastructure/mcp`.** The concurrent, bounded, warn+skip discovery – and the `mcp.*` wire helpers it uses – live beside the adapter as a cohesive `Discover`, taking its two `di` constructors **as parameters**. `internal/cli` consumes a func-typed `deps.MCPDiscoverer`. The MCP Go SDK stays confined to `internal/infrastructure/mcp/**` (`verify-mcp-sdk-confinement` preserved); `internal/infrastructure/di` is imported **only** by the composition root.

**D5 — The round is behaviour-preserving and fails falsifiably.** No flag, exit code, stream contract, formatting, or DSL vocabulary changes. The DoD is the **gate**: the 7 `internal/cli -> internal/infrastructure/*` baseline entries are removed and the gate reports **0 new, 0 stale**; a re-introduced import (or a restored baseline line) turns it **red**.

## Why an ADR

The decision is a **project-level rule future rounds depend on** (`docs/decisions/README.md`): R3/R4 touch the same seams (yield policy, observer hooks, blank-reason predicate, the `agent → ui` RULE-A edge), and R5 ([#101](https://github.com/gosharplite/tellme/issues/101)) builds directly on this pattern. It continues the **ADR-055 / ADR-060 / ADR-074** injection lineage (define a domain port, construct the adapter once at the composition root, inject it). No existing ADR is superseded.

## Alternatives considered

- **(α) Put the `Dependencies` struct inside `internal/cli`** — rejected (Q2): it would leave composition semantics in the application layer and enable the stricter de-coupling that R2 declines.
- **(γ) ~10 func args to `Run`** — rejected: unreadable and every new seam widens the signature (the round-033 FR-009 anti-pattern).
- **(β) Interface-segregated ports** (several small purpose-built ports inside `internal/app/deps` instead of one wide bag) — **rejected for R2, recorded**: the 12-field `Dependencies` is a wide structural bag that couples `internal/cli` to every seam and will only grow, so segregation is the more scalable shape. It is deferred because it would expand a behaviour-preserving refactor round into a design round with no gate to falsify the extra work; the operator accepted the bag (G1). Re-open when a future round adds a seam or when a strict-de-coupling round ([#101](https://github.com/gosharplite/tellme/issues/101)) is scheduled.
- **(B) Composition root under `internal/infrastructure/di`** — rejected (Q1): viable but conceptually surprising (build-time assembly under `infrastructure/`).
- **(T2) keep any presentation var / (T3) neutral app-layer ui ports** — rejected (Q6): (T2) leaves a global; (T3) invents ports for pure presentation.

## Consequences

- `internal/cli` imports no `internal/infrastructure/*` package (RULE-B clean); the R1 gate's `internal/cli` RULE-B baseline is now **empty**; `verify_architecture` reports **0** RULE-A/B/C violations and **0** cycles once R3/R4 remove the single RULE-A `internal/agent -> internal/ui` entry.
- Tests become **hermetic by construction** — there is no package global to mutate (#92 AC5); a test builds its own `Options`.
- `cmd/tellme` gains a test; the round-031 assembler gate relocates there.
- The composition root is now the single assembly site; adding a new adapter is a change in `cmd/tellme` + a new `deps` field, never a change inside `internal/cli`.
- **Records / residuals**: this ADR does **not** de-couple the *downward* imports `internal/cli` legitimately keeps (`agent`/`ui`/`config`/`home`/`app`) — the stricter #92 AC2 clause is parked on **#101** (it needs a new gate rule + ADR). `internal/domain/**` purity is unchanged.
