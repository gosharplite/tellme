# 🤖 tellme — Session Bootstrap

> **Repo**: `github.com/gosharplite/tellme`
> **Folder**: `~/tmp/github/gosharplite/tellme/`
> **Mission**: Disciplined BDD re-creation of `tell-me-go` driven by `aixbdd-tmg`
> **Workflow**: AIxBDD (Strict PM/RD separation, single truth, Red-Green-Refactor)
> **Companion**: end-of-day procedure is [`SESSION-CLOSEOUT.md`](SESSION-CLOSEOUT.md) — this file is its start-of-session mirror.

---

## ⚡ Mandatory First Reads — Execute In Order

> 👤 **Humans**: Start with [`README.md`](README.md) instead. This section is for AI agents.
>
> ⛔ **Do NOT stop to summarize. Do NOT ask questions. Do NOT acknowledge — just execute.**
> After reading this file, you are a machine consuming a checklist. Run Step 1 immediately.

| Step | Action | Description |
|:---|:---|:---|
| **1** | Read [`README.md`](README.md), then tellme's **three domain models** | Repo overview — vision, BDD methodology, core roles, and CLI workflow roadmap; then, **right after the README**, read tellme's own domain model: [`docs/domain-model/tellme.modelith.md`](docs/domain-model/tellme.modelith.md) (the shipped product) · [`docs/domain-model/quality.modelith.md`](docs/domain-model/quality.modelith.md) (the quality process) · [`docs/domain-model/environment-management.modelith.md`](docs/domain-model/environment-management.modelith.md) (the environment manager). They are **descriptive docs, not truth** — on conflict, `specs/truth/**` wins. Modelith sources: the sibling `*.modelith.yaml`; regenerate with `make modelith-render` |
| **2** | Read [`~/tmp/github/gosharplite/tell-me-go/README.md`](~/tmp/github/gosharplite/tell-me-go/README.md) and execute `AI session bootstrap` | Target capability & architecture reference. Execute all 8 bootstrap items defined in `tell-me-go` (see breakdown below) |
| **3** | Read [`~/tmp/github/gosharplite/aixbdd-tmg/domain-model`](~/tmp/github/gosharplite/aixbdd-tmg/domain-model) | Read `aixbdd.modelith.md` (and `.yaml`): canonical entities (`PlanPackage`, `Spec`, `TruthDelta`, `TruthArtifact`, `DSL`, `Task`), invariants, and scenarios |
| **4** | Read [`~/tmp/github/gosharplite/aixbdd-tmg/README.md`](~/tmp/github/gosharplite/aixbdd-tmg/README.md) | Operational BDD engine: PM/RD separation, skills execution pipeline, and CLI-streamlined adaptations |
| **5** | List all pre-load skills | Inventory and inspect all pre-loaded skills in the current session context to establish operational capabilities and governance boundaries |
| **6** | List all agents you can talk to in current shell env | Discover peer agents and personas in the current workspace (`$TELL_ME_HOME/configs/*.yaml`), identify self (`$TELL_ME_MODE`), and map available conversational targets per `tmg-chat-ingroup` |
| **7** | Read [`STATUS.md`](STATUS.md), then align to the active branch | Live session state — current round / active plan package and its pipeline position, the branch model (`main` → `dev` → session branch), decisions locked so far, artifact progress, and open items. Then run `git branch --show-current`; if it is **not** the **Active branch** named in `STATUS.md`, `git checkout` that branch before doing any work, so the round's artifacts are present |
| **8** | Read the **session summary of the last 5 days** | Session continuity — read `docs/session-summary/<YYYY>/<MM>/<DD>/session-summary.md` for the current day and the preceding 4 calendar days to inherit what recent sessions did, decisions locked, artifact progress, and open items. Skip any calendar day with no summary file |

Only after Steps 1, 2, 3, 4, 5, 6, 7, and 8 are complete and results are reported may the agent respond to user tasking.

---

## ⛔ END OF FILE — EXECUTE NOW

**You just finished reading this file. Do not reply. Do not summarize. Do not ask what to do next.**

Immediately return to the step table at the top and execute **Step 1 → Step 2 → Step 3 → Step 4 → Step 5 → Step 6 → Step 7 → Step 8** in order. Report results when Steps 1–8 are complete.

---

## 🗺️ Reference Repositories & Execution Mapping

### 0. tellme's own domain model (Step 1 Details)

Right after `README.md`, read **tellme's own** domain model (round 060; **ADR 0030**) — the shipped system, its quality process, and its environment manager:

| File | Focus & Key Insights |
|---|---|
| [`docs/domain-model/tellme.modelith.md`](docs/domain-model/tellme.modelith.md) | **Product** model — entities (`Session`, `Turn`, `Provider`, `Tool`, `ToolCall`, `Context`, `History`, `Skill`, `MCP` server/tool, `Config`, `Persona`, `Chrome`, `PromptInput`, …), their invariants, and execution scenarios |
| [`docs/domain-model/quality.modelith.md`](docs/domain-model/quality.modelith.md) | **Quality process** model — the `QualityPipeline` gates (`make verify`), the E2E contract, the topology audit, ADR governance, and the triage loop. Records tellme's **no-`NonFixCatalog`** divergence |
| [`docs/domain-model/environment-management.modelith.md`](docs/domain-model/environment-management.modelith.md) | **Environment management** model — environments/groups/personas/provisioning/hot-swap; models the **external** Niffler manager (`tellme.sh`) |

- **Descriptive docs, not truth**: on any conflict with [`specs/truth/**`](specs/truth), the **truth wins** and the model is corrected (ADR 0030 §D4).
- **Source vs. rendered**: edit the `*.modelith.yaml`; the `*.modelith.md` is **generated** — never hand-edit. Regenerate with `make modelith-render`; `make modelith-check` (a `make verify` member) fails on drift.

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

### 6. Recent Session Summaries (Step 8 Details)

Read the per-day session summaries for the **last 5 days** to inherit recent session context:
1. Locate summaries under `docs/session-summary/<YYYY>/<MM>/<DD>/session-summary.md` (e.g. `docs/session-summary/2026/09/10/session-summary.md`).
2. Read the summary for the **current day and the preceding 4 calendar days** — newest first is fine.
3. Extract, per day: what was done, decisions locked, artifacts produced, commits, and open items; reconcile against `STATUS.md` (they should agree).
4. **Skip any calendar day with no `session-summary.md` file** — do not treat a missing day as an error.
5. Keep it current: maintain the daily summary (`docs/session-summary/<YYYY>/<MM>/<DD>/session-summary.md`) alongside `STATUS.md` so the next session inherits accurate state.

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
9. **Session History Continuity**: Read the **session summaries of the last 5 days** at bootstrap (Step 8) — `docs/session-summary/<YYYY>/<MM>/<DD>/session-summary.md` — to inherit recent context, decisions, artifact progress, and open items before tasking; reconcile them with `STATUS.md` and skip missing days.
10. **Domain Model is Descriptive Docs**: tellme's domain model (`docs/domain-model/**`, Step 1) is **subordinate** to truth — on any conflict, `specs/truth/**` wins and the model is corrected. Edit the `*.modelith.yaml` source only; the `*.modelith.md` is generated (`make modelith-render`), and `make modelith-check` (a `make verify` member) fails on drift (**ADR 0030**). The model is **load-bearing** (**ADR 0041**): a round that changes **modelled behaviour** updates `docs/domain-model/**` **in the same PR** (or records, in its plan package, why the change is not modelled), and the **advisory** `make modelith-drift` (never fails; **not** a `verify` member) surfaces a modeled entity whose concept has left the code.
11. **Open Items Are Disclosures, Not a Work Queue**: [`STATUS.md`](STATUS.md)'s **Open items (non-blocking)** section is an **index of disclosures** — a *forward item* is a decision *deferred to a trigger*, **not** tasking. Do **not** re-raise, re-litigate, or propose a forward item as a round theme unless its **trigger has fired**; an item marked **⚠ trigger-gated** is inert (a pointer to its `ADR 00NN §Forward`, which is the authority). Pick a round theme from the **roadmap / a live issue**, **never** from this index. **Two addenda (session 52):** (a) **a forward-item cluster must never generate a round theme** — a theme comes from operator value or a live issue, never from "this would clear the backlog"; (b) **a settled / declined decision is never reopened — even to fire a trigger — without explicit operator intent.** **Third addendum (session 55):** when answering "what is next?" / listing candidates, name **only** (i) **live issues** and (ii) **operator-value themes** — **never** an `ADR §Forward` item, **not even labelled "disclosure, not a candidate"**. Putting a disclosure in a candidates list *is* surfacing it (it invites exactly the "what is this?" round-trip the curation rule exists to prevent). If asked about a forward item, answer **from its ADR** and **do not add it to any next-steps list**. *(The `RF-063-10`/`RF-068-1`/`RF-069-2..4` lesson: an un-curated deferral sitting on a bootstrap-read surface becomes a permanent muse — the recurrence, not the item, is the bug.)*
