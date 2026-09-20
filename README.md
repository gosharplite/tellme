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

## 🧭 Design Intent & Direction (operator-declared)

`tellme` is not a feature-for-feature port of `tell-me-go`; it is a deliberate **re-specification** of the same capability, carried out with an engineering-grade BDD/SDD/TDD process. Three operator-declared directions define the target shape of `tellme`:

1. **No security layer.** `tell-me-go` shipped `SafePath` authorization, consent prompts, command whitelists, and forbidden-character rules. In real usage the operator **always bypassed them** — the rules never held in practice, yet they caused the AI to *repeatedly fail* tool calls for no protection gained. A guardrail that is always disengaged is pure overhead, so `tellme` removes it entirely: tools read, write, and execute whatever they are given. The resulting risk (destructive commands, out-of-tree writes) is an **explicitly accepted operator decision**, not an oversight.
2. **No Windows.** Dropping Windows removes the entire cross-platform tax — path translation, `cmd`/PowerShell shell wrappers, Windows built-in probing, and separator handling. `tellme` targets **bash on POSIX** only.
3. **Bash-first execution.** `execute_command` (run through `bash -c`) is a **first-class primitive**, not a gated escape hatch. Because bash already provides piping and redirection, a separate `pipe_commands` tool is unnecessary and is not offered.

**Consequence — a deliberately small tool surface.** With no security layer and no Windows, a dedicated agent tool only earns its place if it beats bash on a real axis:

- **Context boundedness** — bash `cat`/`grep` are unbounded and can blow the model's context window; a bounded, own-contract tool protects it.
- **Determinism / testable contract** — in BDD a tool's output is *executable truth*; a fixed result shape is testable where a shell one-liner's is not.
- **Reliability** — exact-content writes and exact-block replacements are easy for a schema'd tool and error-prone in shell (`sed`/quoting).

The surface is therefore kept intentionally minimal: `execute_command` as the universal primitive, plus only the tools that clearly clear that bar — the reader family (`list_files`/`read_files`/`get_tree`), the bounded in-file content search `search_files` (round 071; ADR 0043), `write_file`, and `replace_text`. Tools that merely duplicate a trivial shell command are omitted rather than carried for parity.

*This direction is consistent with the project's settled decisions (round 008 Clarify Q3 and round 021 D4 record "no security/consent layer" as a settled exclusion; rounds 012/015 declare POSIX-only). Direction changes are recorded here and in [`STATUS.md`](STATUS.md).*

---

## 📚 Essential References

The development of `tellme` is anchored against two primary reference repositories:

### 1. `tell-me-go` (Capability & Architecture Reference)
- **Local Path**: `~/tmp/github/gosharplite/tell-me-go`
- **Upstream**: [github.com/gosharplite/tell-me-go](https://github.com/gosharplite/tell-me-go)
- **Role**: The functional benchmark and architecture reference.
  - Multi-provider reasoning support (Google Gemini, OpenAI, DeepSeek, Anthropic Claude, Moonshot Kimi, Z.ai GLM).
  - Provider-agnostic `Thought` reasoning model.
  - Built-in agentic tools: FileSystem, Git, AST Go analysis, system commands, MCP client, and enterprise integrations.
  - Context window control: token budgeting, automatic turn summarization, and turn pinning.
  - Session durability and persistence (`history.jsonl`, O(1) archive navigation, SQLite task/state storage).
  - Security and safety guardrails: `SafePath` authorization, run-away loop detection, and cost auditing. *(The `SafePath`/consent rules are deliberately **not** re-created in `tellme` — see [Design Intent & Direction](#-design-intent--direction-operator-declared).)*
  - Environment management: Niffler group/persona templates and provider hot-swapping.

### 2. `aixbdd-tmg` (BDD Execution, Skills & Domain Model Reference)
- **Local Path**: `~/tmp/github/gosharplite/aixbdd-tmg`
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
/axb-system-analysis ──► Map the CLI interfaces (plan.md); for this CLI: /axb-api-plan = NOOP, /axb-data-plan conditional / NOOP, /axb-ui-plan skipped
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
