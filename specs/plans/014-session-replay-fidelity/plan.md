# System Analysis Plan — round 014 (`014-session-replay-fidelity`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/014-session-replay-fidelity/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── replaying-a-tool-using-conversation.feature
│       └── preserving-existing-conversations.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round (persisted step signature)
├── data/data-model.dbml           # /axb-data-plan — MODIFY this round (`history_step` gains the signature)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/{chat,history}/…       # MODIFY — the resume-with-tools contract + DSL rows
```

*(No `contracts/**` truth artifact in this round — see the `NOOP` note for `/axb-api-plan` below. `data/**` is **in scope**: the round changes the persisted tool-step record, which the `data-model-covers-all-state` invariant reaches (as in rounds 007–008).)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # unchanged — entrypoint

internal/
├── domain/
│   └── history/
│       └── history.go             # CHANGED — `Step` gains an optional, provider-agnostic `Signature`
│                                  #   (`json:"signature,omitempty"`); the record becomes
│                                  #   `{tool, arguments, result, signature?}`
├── agent/
│   ├── agentloop.go               # CHANGED — records `Signature: tc.Signature` when building a `Step`, and
│   │                              #   replays it into the synthesised `llm.ToolCall` in `BuildMessages`
│   │                              #   (the deterministic `call_step_<n>` id is unchanged)
│   ├── agentloop_test.go          # EXTENDED — `BuildMessages` emits the persisted signature (replay unit)
│   └── tool_wire_order_test.go    # unchanged (the active-turn chronology is unaffected)
├── infrastructure/
│   └── history/
│       ├── file_store.go          # unchanged — the append-only JSON-Lines round-trip carries the new field
│       │                          #   for free via the `Step` struct (no store-mechanism change)
│       ├── file_store_test.go     # unchanged
│       └── file_store_widened_test.go  # EXTENDED — the widened step round-trips its signature; a
│                                       #   signature-less step serialises byte-identically (omitempty)
└── infrastructure/llm/gemini/     # unchanged — already captures + echoes the `thoughtSignature` in-turn
                                   #   (round 013); this round only persists what it already carries

tests/e2e/
├── fakeprovider/                  # EXTENDED (only if needed) — the fake already records the request; the
│                                  #   resume witness asserts the replayed `functionCall` carries the signature
├── harness/                       # unchanged — the two-process resume witness reuses the existing runner
└── steps/                          # NEW step files — the resume-with-tools journey (+ the preserved-history
                                   #   regression journeys)

specs/truth/data/data-model.dbml    # MODIFY — `history_step` + signature column; reconcile the record Note
specs/truth/features/cli/**         # MODIFY — the resume-with-tools contract + DSL rows

go.mod / go.sum                     # unchanged — stdlib-only; no new module
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 014 extends the **persisted tool-step record** with one optional, provider-agnostic field and wires it at the loop's two existing mapping points. `internal/domain/history.Step` gains `Signature`; `internal/agent/agentloop.go` records it when a `ToolCall` becomes a `Step` and replays it when a `Step` becomes a `ToolCall` on resume (`BuildMessages`). The store (`internal/infrastructure/history/file_store.go`) needs **no** change — it serializes the `Step` struct, so the field round-trips automatically (and `omitempty` keeps existing lines byte-identical). The Gemini adapter already captures + echoes the `thoughtSignature` in-turn (round 013); this round only **persists** what it already carries, so no adapter change is expected. There is **no** new domain port, config field, CLI flag, or third-party dependency — consistent with `research.md` Decisions 1–8 and Clarify Round 1. The **executable contract** (a resumed tool-using Gemini session replays its step faithfully; a tool-less / OpenAI-family conversation is unaffected) is pinned in `specs/truth/features/cli/**` by `/axb-dsl-refine`, and the **persisted model** (`history_step` + signature) is owned by `/axb-data-plan` in `specs/truth/data/data-model.dbml`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **2** system interfaces. Round 014 does not introduce a new system end: it changes the **CLI end**'s resume behaviour and the **persisted session-history record** that backs it (the round-007/008 store, which reaches the `/axb-data-plan` trigger).

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: on resume, a session that previously used a tool on a **Vertex/Gemini** provider replays the earlier tool step **with** its stored provider token and answers the next prompt — no provider "missing thought_signature" rejection and no `the provider request failed` (exit 6). A conversation that used no tool and a tool-using conversation on the OpenAI-compatible family are **unaffected**; `-l N` still surfaces only prompt/answer; the frozen `tellme: {phrase}` vocabulary and the exit-code table are unchanged.
   - Requirement evidence: `FR-001`–`FR-008`, `NFR-001`, `NFR-003`; acceptance features `replaying-a-tool-using-conversation.feature`, `preserving-existing-conversations.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Session history store (local persisted state)`
   - Endpoint type: `Filesystem / local-persistence endpoint`
   - Primary interface: the persisted **tool step** under the per-mode workspace — the record `{tool, arguments, result, signature?}` stored append-only (one JSON line per completed turn, the ordered steps embedded); the provider token is carried on the step, replayed on resume, and **omitted when empty** so existing `history.jsonl` lines stay byte-identical. The lifecycle stays append-after-complete (an interrupted turn is never written); no per-step id is stored (the adapter synthesises `call_step_<n>` on replay).
   - Requirement evidence: `FR-001`, `FR-002`, `FR-005`, `FR-006`, `FR-007`, `NFR-001`, `NFR-002`; the store backs both acceptance features.
   - Planner: **`/axb-data-plan`** — the persisted record model (fields, lifecycle) is a data responsibility. The owner decides the column shape in `specs/truth/data/data-model.dbml` (`history_step`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own; round 014 changes neither the outbound Vertex nor the OpenAI-compatible request **shape** — it only replays a token the round-013 adapter already re-emits).
> - `/axb-data-plan` = **MODIFY** — `specs/truth/data/data-model.dbml`: the `history_step` table gains the nullable provider-agnostic **signature** field, and the record Note is reconciled (it currently pins the shape and states "no per-step id is stored").
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - The **provider token** is **not** a separate system interface: it is a field of the persisted step (interface 2) that the **CLI end** (interface 1) replays. Its executable contract is carried forward to `/axb-dsl-refine`.
> - **Truth amendment carried to `/axb-dsl-refine`**: **MODIFY** the `chat` interface feature that carries the resume journey (round 007 `chat/remembering-the-conversation.feature`) and/or add a rule pinning the **resume replays the earlier tool step faithfully** journey, plus the `chat`/`history` `dsl.md` rows (a tool step that carried a signature; the replayed request carries the signature); confirm **every round-014 acceptance rule is carried** (`acceptance-coverage`). The root `cli/dsl.md` class-phrase vocabulary stays **10**.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Session history store (local persisted state)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own; the outbound request shape is unchanged (a replayed token, not a new request field).
  - **Data** → handoff to **`/axb-data-plan`**: model the persisted tool-step **signature** — a nullable, provider-agnostic field on `history_step`; reconcile the record-shape Note; keep it optional so existing lines are byte-identical.
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin the **resume-with-tools** contract as mechanically assertable interface Rules — a resumed tool-using Gemini session replays its earlier tool step and is answered; a tool-less conversation and an OpenAI-compatible tool-using conversation are unaffected — while `stdout` stays byte-exact and the class-phrase vocabulary is unchanged.
- Scheduling rationale: there are two interfaces, but the **record shape is already fixed** by `research.md` Decision 1 (per-step signature) and Decision 3 (omit-when-empty), and the CLI end's behaviour by Decisions 2/5 — neither analysis depends on the other's *conclusions*, so they are parallel in a single wave. Per `Wave依賴排序與平行分組判準.md` Rules 1–2, independent interfaces form one wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`Session history store`) → **MODIFY** `specs/truth/data/data-model.dbml` (`history_step` gains the signature; reconcile the Note).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the resume-with-tools behaviour in `specs/truth/features/cli/**` (module feature + DSL rows), and confirm `acceptance-coverage` for both round-014 acceptance features.

*Handoff payload (for the next phase)*: plan package `specs/plans/014-session-replay-fidelity`; truth root `specs/truth`; truth-delta `specs/plans/014-session-replay-fidelity/truth-delta.md`; interfaces `CLI end` + `Session history store`; analysis focus as above; acceptance features `features/acceptance/replaying-a-tool-using-conversation.feature` + `features/acceptance/preserving-existing-conversations.feature`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` is invoked only to record its `NOOP`.

---

### Gating blockers

*(none — Clarify Round 1 settled the persisted-field shape (dedicated `signature`) and the legacy behaviour (unchanged best-effort); `research.md` Decisions 1–8 settled the per-step placement, the loop record/replay points, the store round-trip + `omitempty` back-compat, the provider-neutral treatment, the hermetic two-process resume witness, and the no-dependency posture. **Open (non-blocking):** the exact DBML column name/type and JSON key (`/axb-data-plan`), the DSL step vocabulary (`/axb-dsl-refine`), and the replay-id + signature interaction (implementation, pinned by the unit test). None gate this round.)*
