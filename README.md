# tellme

A disciplined re-creation of **tell-me-go** driven by a strict **Behavior-Driven Development (BDD)** workflow.

---

## 🎯 Project Vision & Intention

`tell-me-go` is a high-performance, multi-provider reasoning agent CLI designed for terminal developer workflows. While the original `tell-me-go` project was built in a rapid, vibe-coding style, **tellme** aims to re-create the exact same capabilities through an engineering-grade, test-first **BDD methodology**.

This is a long-term journey. Rather than rushing code implementation, `tellme` will be developed incrementally through structured plan packages, explicit requirements verification, and continuous executable tests:

- **PM Responsibility**: Clear requirements in business language (`spec.md`), acceptance journeys in Gherkin (`features/acceptance/*.feature`), and acceptance criteria verification.
- **RD Responsibility**: Decision-driven technical research (`research.md`), system architecture, single-source-of-truth contracts (`specs/truth/**`), executable interface Gherkin and DSL definitions (`specs/truth/features/**`), and strict Red-Green-Refactor implementation via TDD/BDD.
- **CLI-Streamlined Workflow**: Adapted specifically for a terminal CLI (omitting web UI mockups, focusing on terminal commands, flags, arguments, exit codes, stdin/stdout streams, and local state persistence).

---

## 📚 Essential References

The development of `tellme` is anchored against two primary reference repositories:

### 1. `tell-me-go` (Capability & Architecture Reference)
- **Local Path**: `/home/pos/tmp/github/gosharplite/tell-me-go`
- **Upstream**: [github.com/gosharplite/tell-me-go](https://github.com/gosharplite/tell-me-go)
- **Role**: The functional benchmark and architecture reference.
  - Multi-provider reasoning support (Google Gemini, OpenAI, DeepSeek, Anthropic Claude, Moonshot Kimi, Z.ai GLM).
  - Provider-agnostic `Thought` reasoning model.
  - Built-in agentic tools: FileSystem, Git, AST Go analysis, system commands, MCP client, and enterprise integrations.
  - Context window control: token budgeting, automatic turn summarization, and turn pinning.
  - Session durability and persistence (`history.jsonl`, O(1) archive navigation, SQLite task/state storage).
  - Security and safety guardrails: `SafePath` authorization, run-away loop detection, and cost auditing.
  - Environment management: Niffler group/persona templates and provider hot-swapping.

### 2. `aixbdd-tmg` (BDD Execution, Skills & Domain Model Reference)
- **Local Path**: `/home/pos/tmp/github/gosharplite/aixbdd-tmg`
- **Upstream**: [github.com/gosharplite/aixbdd-tmg](https://github.com/gosharplite/aixbdd-tmg)
- **Role**: The operational engine and domain model for the BDD workflow.
  - **Canonical Domain Model** (`domain-model/aixbdd.modelith.md`): Defines core entities (`PlanPackage`, `Spec`, `AcceptanceFeature`, `TruthArtifact`, `TruthDelta`, `DSL`, `Task`), truth single-ownership, and step-to-DSL matching invariants.
  - **15+ Specialized Skills** (`skills/axb-*`): Directs each phase from specification (`axb-specify`), acceptance journeys (`axb-spec-by-example`), technical research (`axb-technical-research`), system analysis (`axb-system-analysis`), executable DSL refinement (`axb-dsl-refine`), task generation (`axb-tasks`), to TDD implementation (`axb-implement`, `axb-bdd`).
  - **Dedicated Roles** (`roles/pm.yaml` and `roles/rd.yaml`): Enforces clean boundaries between PM requirement definition and RD engineering implementation.
  - **CLI-Streamlined Workflow**: Standards for CLI BDD (skipping `/axb-ui-plan`, marking `/axb-api-plan` as NOOP, conditional `/axb-data-plan`, and treating interface Gherkin/DSL as executable CLI contract).

---

## 🛤️ Workflow Roadmap

When development begins, iterations will proceed through the standard AIxBDD lifecycle:

```text
[Requirement]
      │
      ▼
/axb-specify ─────────► Create specs/plans/NNN-<slug>/ (spec.md, checklist, truth-delta.md)
      │
      ├─────────────────► (Optional) /axb-clarify-over-specs
      │
      ▼
/axb-spec-by-example ──► Formalize PM acceptance criteria into features/acceptance/*.feature
      │
      ▼
/axb-technical-research ► Select tech stack & record decisions (research.md, specs/truth/techstack.md)
      │
      ▼
/axb-system-analysis ──► Map CLI interfaces & data models (plan.md, /axb-data-plan)
      │
      ▼
/axb-dsl-refine ───────► Decompose acceptance rules into executable Gherkin & DSL (specs/truth/features/**)
      │
      ▼
/axb-tasks ────────────► Break down work into test-aligned execution tasks (tasks.md)
      │
      ▼
/axb-implement ────────► Execute TDD cycle (Red → Green → Refactor via /axb-bdd) to completion
```
