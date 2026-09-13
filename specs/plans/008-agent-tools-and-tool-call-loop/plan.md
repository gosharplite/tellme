# System Analysis Plan — round 008 (`008-agent-tools-and-tool-call-loop`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/008-agent-tools-and-tool-call-loop/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── answering-with-a-declared-tool.feature
│       ├── watching-the-tool-loop.feature
│       ├── bounding-and-failing-the-tool-loop.feature
│       └── summarising-the-conversation.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round (tools, loop, widened history)
├── data/**                        # /axb-data-plan — the session-history model (MODIFY this round; see interface 2)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/<module>/{*.feature, dsl.md}
```

*(No `contracts/**` truth artifact in this round — see the `NOOP` note for `/axb-api-plan` below. `data/**` is **in scope**: the round **widens** the persisted turn to embed tool activity.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # unchanged — entrypoint

internal/
├── agent/                         #   NEW — the tool-loop orchestrator (`AgentLoop`): the bounded
│   │                              #   think→act→observe cycle; consumes `llm.Gateway` + `tools.Registry` +
│   │                              #   `history.Store`; emits the `stderr` tool-loop logs; derives a per-tool
│   │                              #   timeout context; maps an incomplete loop to
│   │                              #   `the tool request failed` + code 7 (RF-1 — keeps the loop out of `cli`)
├── cli/
│   ├── cli.go                     #   CHANGED — wires `AgentLoop`; stays presentation/dispatch glue
│   │                              #   (flag parsing, resolve, stream routing, exit-code mapping)
│   └── exitcode.go                #   CHANGED — add `ToolError = 7` (0/2/3/4/5/6/7)
├── domain/
│   ├── llm/gateway.go             #   CHANGED — `Request` gains tool definitions; `Response` gains the
│   │                              #   model's structured tool-call requests (id/name/arguments)
│   ├── tools/                     #   NEW — the tool domain port: `Tool` (name/description/parameters) +
│   │                              #   `Registry` dispatch. Network-free, no I/O
│   └── history/history.go         #   CHANGED — `Entry` widened to embed the turn's tool steps
│                                  #   (`{prompt, answer, steps:[{tool, arguments, result}]}`)
├── infrastructure/
│   ├── llm/openai/client.go       #   CHANGED — sends the `tools` array, assistant `tool_calls`, and `tool`-role
│   │                              #   result messages; parses `choices[0].message.tool_calls`
│   ├── tools/                     #   NEW — the read-only filesystem tools (`list_files`, `read_files`) behind
│   │                              #   the tool port (`read_files` bounded by a fixed 1 MiB cap; no path boundary),
│   │                              #   plus the LLM-backed `summarize_history` tool (injected `history.Store` + `llm.Gateway`)
│   └── history/file_store.go      #   CHANGED — serialize/parse the widened JSON line (fixed field order,
│                                  #   no timestamp/id); replay steps into the conversation on load
├── config/                        #   CHANGED — resolve `MAX_TOOL_LOOP` (env/config, default 1000)
└── home/                          # unchanged this round

tests/e2e/                         # godog suite driving the built binary (interface Gherkin)
├── harness/                       # CHANGED — the fake provider serves scripted tool-call responses and records
│                                  #   the sent tool definitions; stderr is captured for the tool-loop-log assertions
├── steps/                         # NEW step files for the tool loop / tool logs / failure / summarisation
└── network_guard_test.go          # unchanged — the offline set (`--version`, `-d`, `-l`, prompt-less boot) is unchanged
go.mod / go.sum                    # unchanged — the round is stdlib-only (os, path/filepath, io, encoding/json)
Makefile                           # unchanged (no new gate)
```

**Structure Decision**: The change stays **inside the existing CLI surface**, following the round-004/007 seam pattern: it adds **one new domain port + one adapter** for the tools (`internal/domain/tools` ↔ `internal/infrastructure/tools`), and **widens two existing seams** — the provider port (`llm.Request`/`llm.Response`) and the history record (`history.Entry`). The tool loop is **CLI orchestration** (a counted loop in `runTurn`), not a new layer. `MAX_TOOL_LOOP` resolution joins `internal/config` beside the existing effective-value helpers. No new directory beyond `domain/tools` + `infrastructure/tools`; no database, no new dependency. The E2E harness extends the fake provider to serve tool calls; the offline set is unchanged (the tool-using turn is a chat path). The loop mechanics are already settled by `research.md` Decisions 1–9, so this round adds no new architecture dimension.

---

## Analysis Plan

### System interface inventory

This requirement inventories **2** system interfaces. Round 008 does not introduce a new system end; it extends the **CLI end**'s behaviour (the tool loop + the two read-only filesystem tools) and **widens** the already-existing **persisted session state** (a completed turn now also carries its tool activity).

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: **standard output / standard error / exit code** for the operator — the agent tool loop (a prompt run may call the read-only tools and iterate to a final answer); the **live tool-loop log** on `stderr`; the loop bound `MAX_TOOL_LOOP` (default 1000, env/config); the per-tool timeout; the frozen `tellme: {phrase}` class phrases and the exit-code table (`0`/`2`/`3`/`4`/`5`/`6`/**`7`**), with an incomplete loop reported as the new `tellme: the tool request failed` + code `7`; and the **three registered tools** — the read-only filesystem tools (`list_files`, `read_files`) with the 1 MiB read cap, plus the LLM-backed `summarize_history` tool (Story 4). The `-l` contract is unchanged (prompt/answer only).
   - Requirement evidence: `FR-001`–`FR-011`, `FR-013`, `FR-015`, `FR-017`, `FR-018`, `NFR-001`, `NFR-002`, `NFR-005`–`NFR-007`; acceptance features `answering-with-a-declared-tool.feature`, `watching-the-tool-loop.feature`, `bounding-and-failing-the-tool-loop.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Session history store (local persisted state)`
   - Endpoint type: `Filesystem / local-persistence endpoint`
   - Primary interface: the persisted session model under the per-mode workspace — a **completed-turn** record that now carries the operator's `prompt`, the provider's `answer`, **and the turn's tool steps** (`{tool, arguments, result}`); stored append-only (one JSON line per completed turn, fixed field order, no timestamp/id) and read wholesale on resume, replaying the tool steps into the conversation. Lifecycle stays append-after-complete (an interrupted turn is never written).
   - Requirement evidence: `FR-016` (persist the widened turn), `FR-014` (stories 1–4 persistence), `NFR-006`; the store backs every acceptance feature.
   - Planner: **`/axb-data-plan`** — the persisted record model is a data responsibility; this round **widens** the round-007 `history_entry` (a MODIFY, owned by `/axb-data-plan`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own). The **outbound** provider request gains the `tools` array + `tool` messages, but that is the provider's OpenAI-compatible wire (an external dependency reached outbound), not a contract tellme authors.
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - The **read-only filesystem tools** are **not** a separate system interface: they have no schema/contract artifact and no persisted state to model — their operator-facing behaviour (the tool loop, the log, the size cap, the failure contract) belongs to the **CLI end**, and there is no analysis planner for a local read. They are carried to `/axb-dsl-refine` with the CLI end.
> - The **provider** remains an **external dependency reached outbound by the CLI**; only the client → provider request gains tool definitions/results (no new tellme-authored contract).
> - **Truth amendments carried to `/axb-dsl-refine`**: the round-004 `chat/*` features and the round-007 `chat/remembering-the-conversation` + `history/*` features are **MODIFY**-able — a prompt run may now run a tool loop and **persist + replay** tool activity; the root `cli/dsl.md` vocabulary gains the 11th class phrase (`the tool request failed`), and the exit-code table extends to `7`. A no-tool run stays byte-identical to rounds 004–007.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Session history store (local persisted state)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own; the outbound provider request carries tool definitions/results on the existing Chat Completions wire (no new contract).
  - **Data** → handoff to **`/axb-data-plan`**: **widen** the persisted turn model to carry tool steps alongside `prompt`/`answer`, preserving the append-only lifecycle and byte-determinism.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: add the tool-loop behaviour to `specs/truth/features/cli/**` — answering with a declared tool, watching the loop, bounding/failing the loop, and summarising via an agent tool — and **MODIFY** the round-004/007 `chat`/`history` features + DSL rows for the now tool-capable, widened turn (with the root `dsl.md` class-phrase vocabulary 10→11 and the exit-code table extended to `7`).
- Scheduling rationale: there are two interfaces, but the widened record shape is **already fixed** by `research.md` Decision 5 and the CLI-end behaviour by Decisions 3–7 — neither analysis depends on the other's *conclusions*, so they are parallel in a single wave. Per `Wave依賴排序與平行分組判準.md` Rules 1–2, independent interfaces form one wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`Session history store`) → **MODIFY** the persisted session model (`specs/truth/data/**`) to embed the completed turn's tool steps.
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the tool-loop behaviour (`answering-with-a-declared-tool`, `watching-the-tool-loop`, `bounding-and-failing-the-tool-loop`, `summarising-the-conversation`) plus the **round-004/007 `chat`/`history` MODIFY** and the root `dsl.md`/exit-code updates, in `specs/truth/features/cli/**`.

*Handoff payload (for the next phase)*: plan package `specs/plans/008-agent-tools-and-tool-call-loop`; truth root `specs/truth`; truth-delta `specs/plans/008-agent-tools-and-tool-call-loop/truth-delta.md`; interfaces `CLI end` + `Session history store`; analysis focus as above; acceptance features `features/acceptance/answering-with-a-declared-tool.feature` + `features/acceptance/watching-the-tool-loop.feature` + `features/acceptance/bounding-and-failing-the-tool-loop.feature` + `features/acceptance/summarising-the-conversation.feature`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` is invoked only to record its `NOOP`.

---

### Gating blockers

*(none — Clarify Rounds 1–3 settled the tool set, the turn persistence, the no-boundary safety posture, the failure contract, the loop bound, and the log destination; `research.md` Decisions 1–9 settled the tool port, the provider widening, the loop, the tools + size cap, the widened record, the failure contract, the log stream, and the testing strategy. The only judgement call — whether the read-only tools are a separate interface — is resolved by `系統介面盤點與端點歸類判準.md` Rule 3 (they have no independent analysis responsibility or planner) and carried with the CLI end to `/axb-dsl-refine`. **Open (non-blocking):** tool-call concurrency stays sequential, and the exact read-cap size (1 MiB proposed) is a `/axb-technical-research` determination — neither gates this round.)*
