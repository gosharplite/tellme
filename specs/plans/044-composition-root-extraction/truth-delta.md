# Truth Delta: 044-composition-root-extraction

**Plan Package**: `specs/plans/044-composition-root-extraction`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are filled by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked and names what it inspected (round-033 review-fold rule — no unevidenced NOOP).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — **CLI Application** (new **Composition root** row) | Adds a row: the composition root moves to **`cmd/tellme`** (exempt from the R1 tier table), which constructs the concrete adapters and injects them into `internal/cli` as one value — a domain-typed **`internal/app/deps.Dependencies`** plus a cli-local **`cli.Options`** carrying the **one** `internal/ui` presentation seam (`RunTUIPrompt`; the `newRenderer` var is deleted); `internal/cli` constructs **no** `internal/infrastructure/*` adapter and holds no package-level factory var. | round 044 FR-002/FR-003/FR-007; `research.md` D1–D3, D7; grill fold fix-1/fix-2. |
| MODIFY | `specs/truth/techstack.md` — **CLI Application** (**Project layout** row) | Notes that the composition root leaves `internal/cli`: `cmd/tellme` owns assembly and `internal/app/deps` owns the injected port value, so `internal/cli` is a pure application layer. | round 044 FR-001/FR-002; `research.md` D1/D2. |
| MODIFY | `specs/truth/techstack.md` — **Testing & Verification** (**Agent tool-schema gate** row) | The row names "the **non-overridable production assembler `agentTools()`**"; the assembler + its `required ⊆ properties` gate **relocate to `cmd/tellme`** (the *property* — parameterless, read-free, non-overridable — is preserved; the location moves). | round 044 FR-004; `research.md` D5; `spec.md` A7. |
| MODIFY | `specs/truth/techstack.md` — **Skills** (**Skills listing tool (`list_skills`)** row) | The row says "registered in the production `agentTools()` assembler" — the assembler now lives in `cmd/tellme`; the binding contract (`BindSkillsCatalog` on the prompt path only; the offline path performs no `docs/skills` read) is unchanged. | round 044 FR-004/FR-005; `research.md` D3/D5. |
| MODIFY | `specs/truth/techstack.md` — **CLI Application** (**Agent command tool (`execute_command`)** row) | The infra-typed `ToolOutputSink` becomes a **neutral domain port `domaintools.OutputSink`** (`{ Begin func() ; Writer io.Writer ; End func() }` **+ the method `Enabled() bool { return s.Writer != nil }`** — required: `command.go` calls `Enabled()` 4× and the `{}`-is-disabled contract is load-bearing); `BindToolOutput` is injected (a `deps` func), so `internal/cli` no longer names an `infrastructure/tools` type. Behaviour (round-034/038/039/040 block rendering) unchanged. | round 044 FR-005; `research.md` D3; grill fold fix-3. |
| MODIFY | `specs/truth/techstack.md` — **CLI Application** (**Interactive TUI prompt (`-i`)** row) | The `-i` suggestion source's **3-reader** tool set (from `infratools.NewFilesystemTools()`) is now built by the injected **`deps.NewTUIRegistry`** — distinct from the 7-tool agent `NewToolRegistry`; the suggested tool-name set is unchanged (reusing the 7-tool registry would be a behavioural change). The TUI teardown/resume contract is unchanged. | round 044 FR-003/FR-009; `research.md` D2; grill fold fix-1. |
| MODIFY | `specs/truth/techstack.md` — **MCP Client** (**MCP tool discovery (non-stall)** row) | The discovery **orchestration** moves into `internal/infrastructure/mcp` (a cohesive `Discover`), taking its two `di` constructors **as parameters**; `internal/cli` consumes a func-typed `deps.MCPDiscoverer`. The fixed 3 s bound, concurrency, warn+skip, and prompt-path-only behaviour are unchanged. | round 044 FR-006; `research.md` D4. |
| MODIFY | `specs/truth/techstack.md` — **MCP Client** (**MCP client protocol library** row) | `verify-mcp-sdk-confinement` is **preserved** (the SDK/wire stays confined to `internal/infrastructure/mcp/**`); `internal/infrastructure/di` is now imported **only** by the composition root (`cmd/tellme`), never by `internal/cli`. | round 044 FR-006; `research.md` D4/D8. |
| MODIFY | `specs/truth/techstack.md` — **MCP Client** (**MCP credential resolver seam** row) | The bounded `gh`-token-resolver seam stays in `internal/infrastructure/di`; it is now constructed by the composition root and passed into `internal/infrastructure/mcp.Discover` as a parameter (never imported by `internal/cli`). | round 044 FR-006; `research.md` D4. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**` directory exists**) | Inspected: tellme has a single CLI end and **no** OpenAPI/HTTP surface; `specs/truth/` contains only `data/`, `features/`, and `techstack.md` — there is no `contracts/**` tree to change. This round moves a composition root + injects ports and authors no request/response shape. | `contract-authoritative` holds vacuously; `spec.md` A4 / `research.md` D9. *(Grill fold fix: the earlier wording claimed an inspection of `specs/truth/contracts/**`, a path that does not exist — a literally-unevidenced NOOP.)* |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` — inspected `history_entry`/`history_step`/`usage_record`/`prompt_log_entry` and the `~/.tellme/*.jsonl` shapes | No persisted/runtime state change: the extraction is a structural re-home; no record field, file location, or lifecycle changes. | `spec.md` A4 / `research.md` D9. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/features/cli/**` and `specs/truth/features/cli/chat/dsl.md` | No user-facing CLI interface behaviour changes — the round is behaviour-preserving; no feature Rule, Example, step, or `DSLRow` is added or changed (the Gherkin/DSL topology audit is unchanged). The acceptance carrier is the **gate + unit seams**. | `spec.md` A1/A5; `research.md` D9 — the round-020/031/041/042/043 non-BDD-refactor precedent. |

## Governance (ADR)

| Action | Artifact | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0013-composition-root-injection.md` (+ the `docs/decisions/README.md` index row) | Records the injected-`Dependencies` pattern: the composition root lives in `cmd/tellme` (outside the R1 tier table); `internal/cli` receives a domain-typed `internal/app/deps.Dependencies` + a cli-local `Options`, constructs no infrastructure, and holds no package-level factory vars; the `agentTools()` property and the MCP confinement are preserved. Cites the ADR-055/060/074 injection lineage; no existing ADR is superseded. | round 044 FR-011; `research.md` D7 — a project-level rule future rounds must cite. |
