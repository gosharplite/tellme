# System Analysis Plan — round 004 (`004-first-reasoning-turn`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/004-first-reasoning-turn/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── answering-a-single-prompt.feature
│       └── reporting-a-failed-provider-request.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/<module>/{*.feature, dsl.md}
```

*(No `contracts/**` or `data/**` truth artifacts in this round — see the `NOOP` notes for `/axb-api-plan` and `/axb-data-plan` below.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # entrypoint; `var version` is the single `-X main.version` target

internal/
├── cli/                           # dispatch (--version → -d → prompt turn → boot); wires the provider port;
│   ├── cli.go                     #   maps a provider failure → frozen class phrase + exit 6
│   └── exitcode.go                # exit-code values (0/2/3/4/5/6)
├── config/                        # unchanged this round — supplies the resolved Provider
├── home/                          # unchanged this round
├── domain/
│   └── llm/                       # NEW — the provider gateway port (network-free)
│       └── gateway.go             #   Gateway interface + Request/Response/ProviderError value types
└── infrastructure/
    └── llm/
        └── openai/                # NEW — the OpenAI-compatible adapter (stdlib net/http)

tests/e2e/                         # godog suite driving the built binary (interface Gherkin)
├── fakeprovider/                  # NEW — in-process httptest OpenAI-compatible fake
├── steps/                         # NEW chat/turn step files
└── network_guard_test.go          # re-scoped to the offline-path (no-dial canary) witness
Makefile                           # fmt, tidy, build, test, lint, vulncheck, verify (verify-no-network re-scoped)
go.mod / go.sum
```

**Structure Decision**: Adds the project's **first layered seam** — a network-free domain port (`internal/domain/llm`) plus one infrastructure adapter (`internal/infrastructure/llm/openai`) — while keeping the existing `internal/{cli,config,home}` packages (round 004 research Decision 1). `internal/cli` becomes the composition root: it wires the port to the adapter, dispatches the prompt turn, and maps a provider failure to the frozen class phrase + exit code. No new external dependency is introduced — the transport is stdlib `net/http` (research Decision 2). `internal/config` and `internal/home` are unchanged; the resolved `Provider` carried by round 003 is consumed as-is (research Decision 3).

---

## Analysis Plan

### System interface inventory

This requirement inventories **2** system interfaces (the issue's *"the CLI end (+ possibly a provider-gateway interface)"*):

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the **prompt argument** (`tellme "<prompt>"`); stdout (the printed provider answer); stderr (the frozen class phrase `tellme: the provider request failed`); and the exit-code extension (success `0`, provider error `6`, distinct from `2`/`3`/`4`/`5`).
   - Requirement evidence: `FR-001`, `FR-005`, `FR-006`, `FR-008`, `FR-009`; acceptance features `answering-a-single-prompt.feature`, `reporting-a-failed-provider-request.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Provider gateway interface`
   - Endpoint type: `third-party / external provider endpoint` (outbound)
   - Primary interface: the outbound OpenAI-compatible request to the resolved provider — endpoint (`<URL>/chat/completions`), auth (`API_KEY` → `Authorization: Bearer`), `MODEL`, `MAX_TOKENS`, merged `HEADERS`, and `reasoning_effort`; plus response normalization (`choices[0].message.content`) and the failure taxonomy (transport / non-2xx / uninterpretable body).
   - Requirement evidence: `FR-002`, `FR-003`, `FR-004`, `FR-006`, `FR-007`; `research.md` Decisions 3–5.
   - Planner: **none** — no analysis planner takes a third-party provider contract (`/axb-api-plan` covers tellme's *own* API surface, of which there is none). Its contract is exercised **through the CLI end** (the E2E local fake provider) and is carried forward to its contract owner **`/axb-dsl-refine`** at delivery (the CLI `chat`/`turn` module).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own).
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - `/axb-data-plan` = **`NOOP`** (a single in-memory turn — per Clarify Q1 no `history.jsonl`, no session state; no persisted model or state store is introduced).
> - The provider is an **external dependency reached outbound by the CLI**, not a new *system end of tellme*; the system still has exactly one end (the CLI).

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Provider gateway interface`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own.
  - **Data** → **`NOOP`**: a single in-memory turn; no persisted state this round.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: extend `specs/truth/features/cli/**` with a new `chat`/`turn` module (prompt → one provider request → printed answer; provider-failure class + code `6`), add the frozen class phrase `the provider request failed` to the interface-root `specs/truth/features/cli/dsl.md`, and add the provider-error exit-code row.
  - **Provider gateway** → analysed and carried **together with the CLI end** to `/axb-dsl-refine`: its request/response contract and failure taxonomy are exercised through the CLI `chat`/`turn` module (the E2E local fake provider), so it shares the same contract-owner handoff and the same main artifact.
- Scheduling rationale: the two interfaces are **information-independent** — both settle from the same upstream sources (`spec.md` §US1/US2 and `research.md` Decisions 3–5; the provider error taxonomy is already fixed, so the CLI failure mapping does not wait on the provider analysis). Per `Wave依賴排序與平行分組判準.md` Rule 2 they therefore share a **single wave**.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`, `Provider gateway`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`CLI end`, `Provider gateway`) → **`NOOP`** (no persisted state; single in-memory turn).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for **both** interfaces (merged into a single handoff — they share the main artifact `specs/truth/features/cli/**`, per `分析介面委派與planner對應判準.md` Rule 3) → new `chat`/`turn` module + interface-root `dsl.md` class-phrase vocabulary + provider-error exit-code row.

*Handoff payload (for the next phase)*: plan package `specs/plans/004-first-reasoning-turn`; truth root `specs/truth`; truth-delta `specs/plans/004-first-reasoning-turn/truth-delta.md`; interfaces `CLI end` + `Provider gateway`; analysis focus as above.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` and `/axb-data-plan` are invoked only to record their `NOOP`.

---

### Gating blockers

*(none — Clarify Round 1 settled the turn scope, provider family, and failure contract; the transport choice and the no-network guard amendment are settled by `/axb-technical-research` Decisions 2 & 6. No cross-repo or architectural blocker remains.)*
