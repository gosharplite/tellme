# 🤖 tellme — Session Bootstrap

> **Repo**: `github.com/gosharplite/tellme`
> **Folder**: `~/tmp/github/gosharplite/tellme/`
> **Mission**: Disciplined BDD re-creation of `tell-me-go` driven by `aixbdd-tmg`
> **Workflow**: AIxBDD (Strict PM/RD separation, single truth, Red-Green-Refactor)

---

## ⚡ Mandatory First Reads — Execute In Order

> 👤 **Humans**: Start with [`README.md`](README.md) instead. This section is for AI agents.
>
> ⛔ **Do NOT stop to summarize. Do NOT ask questions. Do NOT acknowledge — just execute.**
> After reading this file, you are a machine consuming a checklist. Run Step 1 immediately.

| Step | Action | Description |
|:---|:---|:---|
| **1** | Read [`README.md`](README.md) | Repo overview — vision, BDD methodology, core roles, and CLI workflow roadmap |
| **2** | Read [`~/tmp/github/gosharplite/tell-me-go/README.md`](~/tmp/github/gosharplite/tell-me-go/README.md) and execute `AI session bootstrap` | Target capability & architecture reference. Execute all 8 bootstrap items defined in `tell-me-go` (see breakdown below) |
| **3** | Read [`~/tmp/github/gosharplite/aixbdd-tmg/domain-model`](~/tmp/github/gosharplite/aixbdd-tmg/domain-model) | Read `aixbdd.modelith.md` (and `.yaml`): canonical entities (`PlanPackage`, `Spec`, `TruthDelta`, `TruthArtifact`, `DSL`, `Task`), invariants, and scenarios |
| **4** | Read [`~/tmp/github/gosharplite/aixbdd-tmg/README.md`](~/tmp/github/gosharplite/aixbdd-tmg/README.md) | Operational BDD engine: PM/RD separation, skills execution pipeline, and CLI-streamlined adaptations |
| **5** | List all pre-load skills | Inventory and inspect all pre-loaded skills in the current session context to establish operational capabilities and governance boundaries |
| **6** | List all agents you can talk to in current shell env | Discover peer agents and personas in the current workspace (`$TELL_ME_HOME/configs/*.yaml`), identify self (`$TELL_ME_MODE`), and map available conversational targets per `tmg-chat-ingroup` |
| **7** | Read [`STATUS.md`](STATUS.md), then align to the active branch | Live session state — current round / active plan package and its pipeline position, the branch model (`main` → `dev` → session branch), decisions locked so far, artifact progress, and open items. Then run `git branch --show-current`; if it is **not** the **Active branch** named in `STATUS.md`, `git checkout` that branch before doing any work, so the round's artifacts are present |

Only after Steps 1, 2, 3, 4, 5, 6, and 7 are complete and results are reported may the agent respond to user tasking.

---

## ⛔ END OF FILE — EXECUTE NOW

**You just finished reading this file. Do not reply. Do not summarize. Do not ask what to do next.**

Immediately return to the step table at the top and execute **Step 1 → Step 2 → Step 3 → Step 4 → Step 5 → Step 6 → Step 7** in order. Report results when Steps 1–7 are complete.

---

## 🗺️ Reference Repositories & Execution Mapping

### 1. `tell-me-go` AI Session Bootstrap (Step 2 Details)

When executing Step 2, consult the reference implementation at `~/tmp/github/gosharplite/tell-me-go/` across its 8 mandatory bootstrap targets:

| # | File / Command | Focus & Key Insights |
|---|---|---|
| **2.1** | `~/tmp/github/gosharplite/tell-me-go/README.md` | Core capabilities: multi-provider reasoning (`Thought` model), agentic tools, MCP client, context self-healing, durability |
| **2.2** | `~/tmp/github/gosharplite/tell-me-go/Makefile` | Build/test conventions, coverage exclusions (`*test/`, `testing/`), architectural guards (`verify-architecture`, etc.) |
| **2.3** | `~/tmp/github/gosharplite/tell-me-go/docs/domain-model/tell-me-go.modelith.md` | Core domain model: `Session`, `Turn`, `Provider`, `Tool`, `ContextWindow`, invariants, and execution scenarios |
| **2.4** | `~/tmp/github/gosharplite/tell-me-go/docs/domain-model/quality.modelith.md` | Quality process model: `QualityPipeline` gates, complexity triage, `NonFixCatalog` curation rules |
| **2.5** | `~/tmp/github/gosharplite/tell-me-go/docs/architect/environments/` | Environment management evolution (Toby, Dobby, Porter, Sprawl, Niffler) |
| **2.6** | `~/tmp/github/gosharplite/tell-me-go/docs/architect/environments/domain-model/environment-management.modelith.md` | Domain model for environment isolation, persona templates, and dynamic provider switching |
| **2.7** | `~/tmp/github/gosharplite/tell-me-go/docs/architect/INTENTIONAL_NON_FIXES.md` | Authoritative catalog of known patterns deliberately left as-is to avoid circular refactoring |
| **2.8** | Execute `list_skills` | Discover available skills loaded in the current agent session |

### 2. `aixbdd-tmg` Domain Model & Workflow (Steps 3 & 4 Details)

The operational engine at `~/tmp/github/gosharplite/aixbdd-tmg/` provides the formal BDD methodology:

- **Canonical Domain Model** (`~/tmp/github/gosharplite/aixbdd-tmg/domain-model/aixbdd.modelith.md`):
  - **Entities**: `PlanPackage`, `Spec`, `AcceptanceFeature`, `TruthArtifact` (`Contract`, `DataModel`, `DSL`, `InterfaceFeature`, `Techstack`), `TruthDelta`, `DeltaEntry`, `Task`.
  - **Truth Governance**: Each `TruthArtifact` under `specs/truth/**` has exactly one authoritative owner skill (`truth-single-owner`).
  - **Package Immutability**: A delivered `PlanPackage` (`specs/plans/NNN-<slug>/`) is frozen history (`plan-package-frozen`); new work creates a fresh package (`fresh-package-per-round`).
  - **Executable Alignment**: Every acceptance rule is carried by at least one `InterfaceFeature` (`acceptance-coverage`), and every step must match exactly one `DSLRow` (`dsl-exact-one-match`).
- **Operational Workflow** (`~/tmp/github/gosharplite/aixbdd-tmg/README.md`):
  - **Phase Pipeline**: `/axb-specify` → `/axb-spec-by-example` / `/axb-technical-research` → `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → `/axb-implement` (Red → Green → Refactor via `/axb-bdd`).
  - **Streamlined CLI Rules**:
    1. Skip `/axb-ui-plan` (no HTML mockups).
    2. Lean `/axb-system-analysis` (`/axb-api-plan` marked as `NOOP` in `truth-delta.md`; `/axb-data-plan` conditional on local state persistence).
    3. Executable CLI contract via `/axb-dsl-refine` (`specs/truth/features/**` and `dsl.md`).

### 3. Pre-loaded Skills Verification (Step 5 Details)

Identify and enumerate all pre-loaded skills injected into the session context.

### 4. In-Group Agent Discovery (Step 6 Details)

Inspect the current shell environment and discover available peer agents using the `tmg-chat-ingroup` protocol:
1. Identify the current workspace root via `$TELL_ME_HOME`.
2. Enumerate all agent configurations in `$TELL_ME_HOME/configs/*.yaml` and extract their corresponding `MODE` names and personas (`PERSON`).
3. Identify the active agent identity via `$TELL_ME_MODE`.
4. Report all peer agents that can be reached (every mode except self; never message your own mode to avoid session self-pollution).

### 5. Session Status Verification (Step 7 Details)

Read the repo-root `STATUS.md` to establish live session state, then align the working branch:
1. Current round / active plan package and its position in the phase pipeline.
2. The **Active branch** and the branch model (`main` → `dev` → session branch).
3. **Align the working tree**: run `git branch --show-current`; if it is not the **Active branch** named in `STATUS.md`, check that branch out before doing any work (otherwise the round's artifacts are absent).
4. Decisions locked so far, plus any open non-blocking items.
5. Keep it current: update `STATUS.md` at each pipeline phase gate and whenever a decision is locked.

---

## ⚠️ Agent Rules

1. **No Vibe Coding**: Never write speculative code directly. All product code in `tellme` must be driven by an active `PlanPackage` with executable tests.
2. **Strict BDD Pipeline**: Always follow the sequence: Requirements (`spec.md`) → Acceptance Gherkin (`features/acceptance/`) → Truth Delta (`truth-delta.md`) → Executable DSL (`specs/truth/features/**`) → Tasks (`tasks.md`) → TDD Implementation.
3. **Reference, Never Copy Blindly**: `tell-me-go` is the benchmark for capability, behavior, and architecture; `aixbdd-tmg` is the benchmark for development discipline. Clean architecture, testability, and determinism take precedence over legacy shortcuts.
4. **Frozen History**: Never modify delivered `specs/plans/NNN-<slug>/` directories. Always create a new package for new iterations or modifications.
5. **Truth Integrity**: Keep `specs/truth/**` as the single source of truth for current system behavior. Record all modifications through `truth-delta.md`.
6. **Skill Awareness**: Verify pre-loaded skills before taking action; follow the specific SOP and invariants defined in each active skill.
7. **In-Group Protocol**: Respect peer agent boundaries and messaging rules defined in `tmg-chat-ingroup` (clear `TELL_ME_MODE`, sequential dispatch, and never message self).
8. **Session Status Discipline**: Read `STATUS.md` at bootstrap (Step 7) and keep it current — update it at every pipeline phase gate and whenever a decision is locked, so the next session inherits accurate state.
