# System Analysis Plan — round 027 (`027-ai-call-turn-counter`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/027-ai-call-turn-counter/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/                # 2 journeys — /axb-spec-by-example
│       ├── counting-how-often-the-model-is-asked.feature
│       └── starting-the-count-over-on-a-fresh-session.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY (turn-chrome row + session-history row) ✓ done
├── data/data-model.dbml           # /axb-data-plan — MODIFY (the persisted per-turn AI-call count)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/chat/…                 # MODIFY the turn-header Rule + the `the turn is headed "Turn {number}"…` row
```

*(No `contracts/**` change (`/axb-api-plan` = `NOOP`) and no `ui/**` artifact this round — a plain-CLI
presentation-accounting slice: `/axb-ui-plan` is skipped.)*

### Source-code structure (repository root)

```text
internal/
├── domain/
│   └── history/
│       └── history.go             # CHANGED — add `Calls int` to the persisted `Entry` (the turn's inference-round count)
├── infrastructure/
│   └── history/
│       └── *.go                   # CHANGED — (de)serialize the new `calls` field on the JSONL line (omitempty; a legacy line → 0)
├── agent/
│   └── agentloop.go               # UNCHANGED — `AgentResult.Calls` already carries one entry per inference round (round 018)
└── cli/
    └── cli.go                     # CHANGED — compute the header number as Σ prior-entry `Calls` + 1 (was `len(prior)+1`);
                                   #   persist `Calls: len(result.Calls)` on the appended entry
internal/ui/turn.go                # UNCHANGED — the formatter still renders `╭─⠿ Turn <N> - <mode>`; only `N`'s computation changes
go.mod / go.sum                    # unchanged — no new module (stdlib only)
Makefile                           # unchanged (no new gate)
```

**Structure Decision**: Round 027 is a **presentation-accounting** slice on the **CLI end**. It changes
only how the round-017 turn-header number `<N>` is **computed** — from the completed-**turn** count
(`len(prior)+1`) to the running **AI-endpoint-call** index (**Σ** of the prior turns' inference rounds,
**+ 1**) — and the **persisted state** needed to reconstruct that sum across processes: an integer
**`calls`** field on the session-history entry. The chrome formatter (`internal/ui/turn.go`) is
**untouched**; the loop already exposes the per-turn count as `len(AgentResult.Calls)`, so
`internal/agent` is **unchanged**. The bound is persisted by the `internal/infrastructure/history`
adapter and surfaced by `internal/cli`. There is **no** new endpoint, **no** new dependency, and **no**
change to the chrome cadence/format or any other operator surface. The **executable contract** is
pinned in `specs/truth/features/cli/**` by `/axb-dsl-refine`; `specs/truth/contracts/**` and `ui/**` are
untouched.

---

## Analysis Plan

### System interface inventory

This requirement inventories **2** interfaces.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the prompt turn's **`╭─⠿ Turn <N> - <mode>` header number** now counts the session's **AI-endpoint calls** (inference rounds) — the running index of this prompt's first call (`Σ prior calls + 1`) — advancing by every model request a turn makes (a tool-less turn by one; a tool-using turn by its inference-round count). The chrome **cadence and format are unchanged** (still one header + one `╰─⠿ Ready` per prompt; plain text; no denominator); `stdout` stays byte-exact; `--new` restarts the number at `Turn 1`.
   - Requirement evidence: `FR-001`–`FR-004`, `FR-006`–`FR-007`; acceptance features `counting-how-often-the-model-is-asked.feature`, `starting-the-count-over-on-a-fresh-session.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which MODIFIES the turn-chrome Rule (and its DSL row) under `specs/truth/features/cli/chat/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Session history store (local-state interface)`
   - Endpoint type: `Local-state / file endpoint`
   - Primary interface: the per-turn session record `history_entry` gains an integer **`calls`** field — the turn's **AI-endpoint-call count** (inference rounds, from the loop's per-turn call list) — on the existing append-only JSONL line under `$TELL_ME_HOME/output/<mode>/history.jsonl`; the header number is the **Σ** of the active session's entries' `calls`, **+ 1**. The field is written for every completed turn (≥ 1); a **legacy** line without it counts as **1**; `--new` archives the active file, so the sum restarts.
   - Requirement evidence: `FR-005`, `FR-006`; the key entity *Persisted turn record* in `spec.md`.
   - Planner: **`/axb-data-plan`** — MODIFY the `history_entry` record in `specs/truth/data/data-model.dbml` (add the `calls` attribute).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface of tellme's own).
> - `/axb-data-plan` = **MODIFY** — the session-history record gains the per-turn call count (Decision 2), so the data truth is updated in place.
> - `/axb-ui-plan` = **skipped** — no user-facing UX screen; the header is line-oriented terminal output on `stderr`.
> - **Truth amendment carried to `/axb-dsl-refine`**: **MODIFY** the round-017 turn-chrome contract — the `presenting-the-turn.feature` Rule *"The turn header counts the session's turns"* and the `chat/dsl.md` row `the turn is headed "Turn {number}" for the active mode` — so `{number}` is the **AI-endpoint-call** index (a tool-using history Example must show the advance), and confirm `acceptance-coverage` for both round-027 acceptance features. No new class phrase; the root vocabulary is unchanged.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Session history store (local-state interface)`
- Analysis focus:
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: MODIFY the turn-header contract so `{number}` counts **AI-endpoint calls** (`Σ prior calls + 1`), keeping the chrome cadence/format, the `- <mode>` suffix, the plain text, and `stdout` byte-exact; the existing tool-using-history Example must be reshaped to prove a tool round advances the number by its inference-round count.
  - **Session history store** → handoff to **`/axb-data-plan`**: MODIFY `history_entry` to carry the per-turn **`calls`** attribute (the inference-round count), note the legacy-line default (1) and the `--new`-reset lifecycle.
  - **API** → **`NOOP`**; **UI** → **skipped**.
- Scheduling rationale: the round has two tightly-coupled interfaces on one end (the header number and the persisted state that feeds it); there is no cross-wave dependency, so the whole analysis is one wave, with the CLI end carried forward to its contract owner and the state delegated to the data planner.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`Session history store`) → **MODIFY** `history_entry` in `specs/truth/data/data-model.dbml` (add `calls`).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → MODIFY the turn-header Rule + DSL row, and confirm `acceptance-coverage` for the two round-027 acceptance features.

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface this round).

*Handoff payload (for the next phase)*: plan package `specs/plans/027-ai-call-turn-counter`; truth root `specs/truth`; truth-delta `specs/plans/027-ai-call-turn-counter/truth-delta.md`; interfaces `CLI end`, `Session history store`; analysis focus as above; acceptance features `counting-how-often-the-model-is-asked.feature`, `starting-the-count-over-on-a-fresh-session.feature`.

---

### Gating blockers

*(none — the operator locked the semantics (the number counts AI-endpoint calls; a "call" = an inference round, retries excluded; the chrome stays as-is), and `research.md` Decisions 1–6 settled the unit, the persistence home, the pre-turn emission, the reuse of `AgentResult.Calls`, and the `--new`/legacy handling. **Open (non-blocking):** the exact persisted field name (`calls`) and the DBML prose (`/axb-data-plan`); the DSL row's `{number}` wording (`/axb-dsl-refine`). None gate this round.)*
