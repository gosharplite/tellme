# System Analysis Plan — round 005 (`005-stdin-piping`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/005-stdin-piping/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── piping-a-prompt.feature
│       └── piping-the-answer-out.feature
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
└── main.go                        # entrypoint; supplies os.Stdin/Stdout/Stderr + the real TTY detector; single `-X main.version` target

internal/
├── cli/                           # dispatch (--version → -d → prompt turn → boot)
│   ├── cli.go                     #   CHANGED — injects the stdin/stdout/stderr + TTY-detection seam; combines
│   │                              #   positional args + piped stdin; gates presentation on isTTY(stdout); reads stdin
│   │                              #   only on the prompt-turn path
│   └── exitcode.go                # unchanged this round (0/2/3/4/5/6)
├── config/                        # unchanged this round
├── home/                          # unchanged this round
├── domain/llm/                    # unchanged this round (no provider request-shape change)
└── infrastructure/llm/openai/     # unchanged this round

tests/e2e/                         # godog suite driving the built binary (interface Gherkin)
├── harness/cmd_helper.go          # CHANGED — the subprocess runner injects a scripted stdin (a pipe) for piped scenarios
├── steps/                         # NEW step files for piped stdin + the output contract
└── network_guard_test.go          # unchanged
Makefile / go.mod / go.sum         # unchanged — no new dependency (stdlib char-device detection)
```

**Structure Decision**: The change stays **inside the existing CLI packages** — no new package, no new dependency, no new layered seam. The I/O coupling is inverted the same way round 003 inverted the environment lookup: `internal/cli` receives `stdin`/`stdout`/`stderr` and a TTY-detection function, so the prompt-combination and input/output-mode selection are pure and unit-testable, while `cmd/tellme/main.go` supplies the real `os.*` handles and detector. The provider path (`internal/domain/llm` + `internal/infrastructure/llm/openai`) and `internal/{config,home}` are untouched — round 005 changes only how the prompt is ingested and how the output is contractually framed. The E2E harness gains a scripted-stdin path (today `cmd.Stdin` is unset → the child reads `/dev/null`).

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface (round 005 changes prompt ingestion and the output contract on the existing end — it introduces **no** new end and **no** new outbound provider contract):

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: **standard input** (a piped/redirected prompt, combined with the positional instruction as `args (joined by spaces)` + `"\n"` + piped content, then trimmed); **standard output** (the plain answer, exactly one trailing newline, presentation suppressed when stdout is not a terminal); **standard error** (the frozen class phrase); and the unchanged exit-code table (`0`/`2`/`3`/`4`/`5`/`6`).
   - Requirement evidence: `FR-001`–`FR-008`, `FR-010`–`FR-012`, `NFR-001`/`NFR-004`; acceptance features `piping-a-prompt.feature`, `piping-the-answer-out.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own; round 005 does not change the provider request shape, so there is no new API contract either).
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - `/axb-data-plan` = **`NOOP`** (round 005 introduces no persisted or in-memory state — stdin is read into an ephemeral prompt string; no `history.jsonl`, no session state).
> - The provider remains an **external dependency reached outbound by the CLI**; its request/response contract is unchanged from round 004 and is not re-analysed here. The system still has exactly one end (the CLI).

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own; no provider request-shape change.
  - **Data** → **`NOOP`**: no persisted or in-memory state introduced.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: extend `specs/truth/features/cli/**` with the piped-prompt behaviour (read stdin on the prompt-turn path; combine `args + "\n"` + piped content; empty → boot) and the output contract (answer on stdout only, one trailing newline, presentation suppressed when stdout is not a terminal), adding the necessary Given/When/Then DSL rows in the right module(s) and interface-root `dsl.md`.
- Scheduling rationale: there is a single interface, so no dependency ordering is needed — it settles entirely from the upstream sources (`spec.md` §US1/US2, `research.md` Decisions 1–6). Per `Wave依賴排序與平行分組判準.md` Rules 1–2, a lone interface forms a single wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no persisted state).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the piped-prompt behaviour + the TTY-aware output contract in `specs/truth/features/cli/**` (module feature + DSL rows; interface-root `dsl.md` if any row becomes cross-module).

*Handoff payload (for the next phase)*: plan package `specs/plans/005-stdin-piping`; truth root `specs/truth`; truth-delta `specs/plans/005-stdin-piping/truth-delta.md`; interface `CLI end`; analysis focus as above; acceptance features `features/acceptance/piping-a-prompt.feature` + `features/acceptance/piping-the-answer-out.feature`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` and `/axb-data-plan` are invoked only to record their `NOOP`.

---

### Gating blockers

*(none — Clarify Round 1 settled the piping scope, the combine rule, and the output contract; the TTY-detection mechanism and the stdin cap are settled by `/axb-technical-research` Decisions 1–2. No cross-repo or architectural blocker remains.)*
