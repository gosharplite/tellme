# System Analysis Plan — round 007 (`007-session-history-persistence`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/007-session-history-persistence/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── remembering-the-conversation.feature
│       ├── starting-a-fresh-conversation.feature
│       └── inspecting-the-session-history.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round (history store, --new/-l, conversation context)
├── data/**                        # /axb-data-plan — the session-history model (ADD this round; see interface 2)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/<module>/{*.feature, dsl.md}
```

*(No `contracts/**` truth artifact in this round — see the `NOOP` note for `/axb-api-plan` below. `data/**` is **in scope** this round: the round introduces persisted system state, unlike round-001's config/workspace.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # unchanged — entrypoint; only calls cli.Run(os.Args[1:], version)

internal/
├── cli/                           # dispatch (--version → -d → -l → (--new) prompt turn → boot)
│   ├── cli.go                     #   CHANGED — adds `--new` (bool) and `-l N` (int); wires the history store;
│   │                              #   resumes prior turns into the request on a prompt run; persists the completed
│   │                              #   turn; archives on `--new`; prints the last N on `-l`. Map a history I/O
│   │                              #   failure to the environment class phrase + code 4
│   └── exitcode.go                # unchanged this round (0/2/3/4/5/6)
├── domain/
│   ├── llm/gateway.go             #   CHANGED — `llm.Request` carries the resumed prior messages (role/content);
│   │                              #   empty history keeps the round-004 single-user-message request byte-identical
│   └── history/                   #   NEW — the history domain port (`Store` interface + `Entry` value type:
│                                  #   a completed turn's prompt + answer). Network-free, no I/O
├── infrastructure/
│   ├── llm/openai/client.go       #   CHANGED — sends the `messages` array (resumed history + current prompt)
│   └── history/                   #   NEW — the JSON-Lines adapter behind `internal/domain/history`:
│                                  #   append-after-complete, read-all on resume, `--new` archive, deterministic
│                                  #   `encoding/json` (no timestamp/ID). Bounded by the workspace
├── config/                        # unchanged this round
└── home/                          # unchanged this round (the per-mode workspace already exists)

tests/e2e/                         # godog suite driving the built binary (interface Gherkin)
├── harness/                       # CHANGED — filesystem assertions over history.jsonl / history.archive.jsonl;
│                                  #   the fake provider records the request `messages` array (resume witness)
├── steps/                         # NEW step files for remembering / --new / -l
└── network_guard_test.go          # CHANGED — the offline set now includes `-l` and a prompt-less `--new`
go.mod / go.sum                    # unchanged — the round is stdlib-only (encoding/json, os)
Makefile                           # unchanged (no new gate)
```

**Structure Decision**: The change stays **inside the existing CLI surface** and adds **one domain port + one adapter** for the history store, mirroring the round-004 pattern (a network-free `internal/domain` port with an `internal/infrastructure` adapter, injected so tests can substitute a fake). The store is an append-only JSON-Lines file under the already-existing per-mode workspace; no new directory, no database, no new dependency. The provider port gains a **prior-message field** on `llm.Request` (a value-type widening, not a transport change); its adapter sends the widened `messages` array. `internal/config` and `internal/home` are untouched. The E2E harness gains filesystem assertions and the fake provider records the sent conversation. The resume/persist decision is already settled by `research.md` Decisions 1–2, so this round adds no new architecture dimension.

---

## Analysis Plan

### System interface inventory

This requirement inventories **2** system interfaces. Round 007 does introduce **persisted system state** (the session history), which — unlike round-001's config (**input**) and workspace (**interface behaviour**) — is genuine state and therefore reaches the `/axb-data-plan` trigger ("local storage (JSON)").

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: **standard output / standard error / exit code** for the operator — the new `--new` (start a fresh session) and `-l N` (list the last N messages) flags; the auto-resume behaviour on a prompt run (prior turns sent as context); the frozen `tellme: {phrase}` class phrases and the exit-code table (`0`/`2`/`3`/`4`/`5`/`6`), with a history I/O failure reusing the environment phrase + code `4`. The stdin read stays on the prompt path (round 005) — `-l` and a prompt-less `--new` never read it.
   - Requirement evidence: `FR-001`, `FR-002`, `FR-004`, `FR-005`, `FR-007`–`FR-012`, `NFR-001`–`NFR-003`; acceptance features `remembering-the-conversation.feature`, `starting-a-fresh-conversation.feature`, `inspecting-the-session-history.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Session history store (local persisted state)`
   - Endpoint type: `Filesystem / local-persistence endpoint`
   - Primary interface: the persisted session model under the per-mode workspace — a **turn** record carrying the operator's `prompt` and the provider's `answer`, stored append-only (one JSON line per completed turn) and read wholesale on resume; the archive on `--new`; the deterministic byte content (no timestamp/ID). The lifecycle is append-after-complete (an interrupted turn is never written).
   - Requirement evidence: `FR-001`, `FR-003`, `FR-006`, `NFR-002`; the store backs every acceptance feature.
   - Planner: **`/axb-data-plan`** — the persisted state model (entities, fields, lifecycle) is a data responsibility. The owner decides the artifact granularity for a minimal append-only log.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own; round 007 does not change the outbound provider request *shape* beyond carrying prior messages in the existing `messages` array).
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - **Distinction from round 001**: round 001's data truth was argued out (`/axb-data-plan` = NOOP) because the config file is **input** (a CLI input contract owned by the CLI end) and the workspace lifecycle is **interface behaviour**. Round 007's session history is neither — it is **persisted state with a record model** (a turn's prompt + answer), which the `data-model-covers-all-state` invariant reaches. Hence interface 2 is delegated rather than NOOP'd.
> - The provider remains an **external dependency reached outbound by the CLI**; only the client → provider request gains the prior-message context (no new external contract).
> - **Truth amendment carried to `/axb-dsl-refine`**: the round-004 `chat/answering-a-single-prompt` interface feature (and its `dsl.md`) gain a **MODIFY** — a prompt run now **also persists** the completed turn and **carries prior turns** as context when history is present; a fresh workspace is unchanged.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Session history store (local persisted state)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own; the outbound provider request carries prior messages in the existing array (no new contract).
  - **Data** → handoff to **`/axb-data-plan`**: model the persisted session history — the turn record (prompt/answer), the append-only file lifecycle, resume-read, and the `--new` archive — as the minimal data truth for the session store.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: add the session-history behaviour to `specs/truth/features/cli/**` — remembering the conversation (persist + auto-resume), starting a fresh session (`--new`), and inspecting the last N messages (`-l N`) — and **MODIFY** the round-004 `chat/answering-a-single-prompt` feature + DSL rows so a prompt run is understood to persist and resume (a fresh workspace is unchanged).
- Scheduling rationale: there are two interfaces, but the history store's **record shape is already fixed** by `research.md` Decision 1 (a turn = prompt + answer) and the CLI end's behaviour by Decisions 2–8 — neither analysis depends on the other's *conclusions*, so they are parallel in a single wave. Per `Wave依賴排序與平行分組判準.md` Rules 1–2, independent interfaces form one wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`Session history store`) → author the persisted session-history data truth (`specs/truth/data/**`), or record a justified `NOOP` if the owner judges the append-only turn log below the data-model bar.
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the session-history behaviour (`remembering-the-conversation`, `starting-a-fresh-conversation`, `inspecting-the-session-history`) plus the **round-004 `chat/answering-a-single-prompt` MODIFY**, in `specs/truth/features/cli/**` (module feature + DSL rows; interface-root `dsl.md` if any row becomes cross-module).

*Handoff payload (for the next phase)*: plan package `specs/plans/007-session-history-persistence`; truth root `specs/truth`; truth-delta `specs/plans/007-session-history-persistence/truth-delta.md`; interfaces `CLI end` + `Session history store`; analysis focus as above; acceptance features `features/acceptance/remembering-the-conversation.feature` + `features/acceptance/starting-a-fresh-conversation.feature` + `features/acceptance/inspecting-the-session-history.feature`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` is invoked only to record its `NOOP`.

---

### Gating blockers

*(none blocking — Clarify Round 1 settled resume semantics, the `--new`/`-l` scope, and the failure contract; `research.md` Decisions 1–8 settled the store format, the request widening, the archive mechanism, the `-l` contract, and the dispatch order. The only judgement call — whether the session history reaches `data/**` — is resolved by the CLI-streamlined `/axb-data-plan` trigger ("local storage (JSON)") and the `data-model-covers-all-state` invariant; it is delegated to the owner rather than guessed. **Open (non-blocking, named pin):** PR #16 Final-Review **Obs 1** — the stdout TTY probe — remains open; tellme ships no own presentation chrome, so the probe is not wired.)*
