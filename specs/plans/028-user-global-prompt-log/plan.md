# System Analysis Plan — round 028 (`028-user-global-prompt-log`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/028-user-global-prompt-log/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/                # 2 journeys — /axb-spec-by-example ✓ done
│       ├── sharing-prompts-across-environments.feature
│       └── carrying-over-the-existing-prompt-history.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY (Shared global prompt log row) ✓ done
├── data/data-model.dbml           # /axb-data-plan — MODIFY (prompt_log_entry location + seed lifecycle)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/chat/…                 # MODIFY the shared-prompt-log rows + the module note; ADD a seed Example
```

*(No `contracts/**` change (`/axb-api-plan` = `NOOP`) and no `ui/**` artifact this round — a plain-CLI
state-location slice: `/axb-ui-plan` is skipped.)*

### Source-code structure (repository root)

```text
internal/
├── infrastructure/
│   └── history/
│       └── global_prompt_tracker.go   # CHANGED — resolve the log at ~/.tellme/global_prompts.jsonl (os.UserHomeDir())
│                                      #   instead of <home>/output/global_prompts.jsonl; add the seed-on-absent migration
│                                      #   (verbatim copy of <TELL_ME_HOME>/output/global_prompts.jsonl — copy, not move;
│                                      #   never overwrite an existing destination; missing source → empty); degrade to a
│                                      #   no-op on an unresolvable/unwritable home (never break the prompt)
├── domain/
│   └── history/
│       └── tracker.go                 # CHANGED (doc) — the PromptTracker port doc reflects the user-global home + the seed contract
└── cli/
    └── cli.go                         # CHANGED — construct the tracker with the runtime home (seed source) + the user home (destination);
                                       #   the non-`-i` dispatch paths are untouched (the log stays `-i`-only)
tests/
└── e2e/
    ├── steps/*                        # CHANGED — the shared-prompt-log steps resolve the log at the temp HOME; add seed arrangements/assertions
    └── suite                          # CHANGED — point HOME at a per-scenario temp dir (round-026 precedent) + temp runtime home
go.mod / go.sum                      # unchanged — no new module (stdlib only)
Makefile                             # unchanged (no new gate)
```

**Structure Decision**: Round 028 is a **state-location** slice on the **CLI end**. It changes where the
round-015 interactive prompt log is **read from and written to** — from the environment-scoped
`<TELL_ME_HOME>/output/global_prompts.jsonl` to the per-user `~/.tellme/global_prompts.jsonl` — and adds a
one-time **seed** that copies the existing environment-scoped file into the user-global one when the
latter is absent. The change is confined to the `internal/infrastructure/history` adapter (the path
resolution + the seed) plus the CLI construction seam; the record **shape**, the read **semantics**
(newest-first + deduped), the `-i`-only write rule, and the interactive TUI chrome are **unchanged**, so
`internal/ui` and `internal/agent` are **untouched**. There is **no** new endpoint, **no** new dependency,
and **no** persisted-shape change — only the file's **location and lifecycle**. The CLI end's
**contract owner** is `/axb-dsl-refine` (terminal endpoints have no analysis planner); the log's
**durable home** is delegated to `/axb-data-plan`. `specs/truth/contracts/**` and `ui/**` are untouched.

---

## Analysis Plan

### System interface inventory

This requirement inventories **2** interfaces.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the `-i` interactive **prompt log** now reads (seeds suggestions) and writes (records a submission) a **per-user** file — `~/.tellme/global_prompts.jsonl` — instead of the environment-scoped `output/` file, so a prompt typed in one environment is offered in another; on first use the log is **seeded** from the environment's existing file, so carried-over prompts appear immediately. A non-`-i` run still reads/writes nothing; the interactive chrome and record shape are unchanged.
   - Requirement evidence: `FR-001`–`FR-004` (US1), `FR-005`–`FR-007` (US2, observable through the prompt), `FR-008`; acceptance features `sharing-prompts-across-environments.feature`, `carrying-over-the-existing-prompt-history.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which MODIFIES the shared-prompt-log rows + the `chat` module note under `specs/truth/features/cli/chat/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Shared prompt log store (local-state interface)`
   - Endpoint type: `Local-state / file endpoint`
   - Primary interface: the append-only JSON-Lines `prompt_log_entry` log moves to `~/.tellme/global_prompts.jsonl` (the `output/` root of the **user** home, alongside the round-026 tool-usage log) and gains a **seed-on-absent** lifecycle — when the file is absent it is populated by a **verbatim copy** of `<TELL_ME_HOME>/output/global_prompts.jsonl` (copy, not move; never overwritten; missing source → empty). The record shape (`{"timestamp","prompt"}` per line), the append-only write, the newest-first-deduped read, and the compaction policy are unchanged. **Recorded divergence:** the log is no longer shared with `tell-me-go` (which keeps `<TELL_ME_HOME>/output/global_prompts.jsonl`).
   - Requirement evidence: `FR-005`–`FR-007`; `FR-001`; the key entities *User-global prompt log* and *Environment-scoped prompt log (seed source)* in `spec.md`; `research.md` Decisions 1, 3, 4, 5, 6.
   - Planner: **`/axb-data-plan`** — MODIFY the `prompt_log_entry` record in `specs/truth/data/data-model.dbml` (its `$TELL_ME_HOME/...` location → `~/.tellme/global_prompts.jsonl`, + the seed lifecycle + the `tell-me-go` divergence).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface of tellme's own).
> - `/axb-data-plan` = **MODIFY** — the log's durable home and lifecycle change (Decision 1/3), so the data truth is updated in place.
> - `/axb-ui-plan` = **skipped** — no user-facing UX screen is added or changed; the interactive prompt's terminal surface is unchanged (round-016 parity), and this round's output is a line-oriented log location on the filesystem.
> - **Truth amendment carried to `/axb-dsl-refine`**: **MODIFY** the shared-prompt-log contract — the `specs/truth/features/cli/chat/dsl.md` rows that name `$TELL_ME_HOME/output/global_prompts.jsonl` (the write/read `工作區` references) and the `chat` module note, plus `recording-the-shared-prompt-log.feature`'s header comment — to the user-global path, and **ADD** a seed Example/row (carry-over + never-overwrite). No new class phrase; the root vocabulary is unchanged.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Shared prompt log store (local-state interface)`
- Analysis focus:
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: MODIFY the shared-prompt-log DSL rows + module note so the read/write target is the per-user `~/.tellme/global_prompts.jsonl`; **ADD** a seed Example proving the first-use carry-over (and the never-overwrite negative); keep the `-i`-only rule and the record shape; confirm `acceptance-coverage` for both round-028 acceptance features.
  - **Shared prompt log store** → handoff to **`/axb-data-plan`**: MODIFY `prompt_log_entry` to its new user-global location + the seed-on-absent lifecycle + the tell-me-go divergence; keep the record shape and the compaction policy.
  - **API** → **`NOOP`**; **UI** → **skipped**.
- Scheduling rationale: the round has two tightly-coupled interfaces on one end (the operator-visible `-i` log behaviour and the file that holds it); there is no cross-wave dependency (the store's location/lifecycle is a single, self-contained change), so the whole analysis is one wave — the CLI end carried forward to its contract owner and the store delegated to the data planner, matching the round-015/027 shape.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`Shared prompt log store`) → **MODIFY** `prompt_log_entry` in `specs/truth/data/data-model.dbml` (location → `~/.tellme/global_prompts.jsonl`; seed-on-absent lifecycle; tell-me-go divergence).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → MODIFY the shared-prompt-log rows + module note to the user-global path, ADD a seed Example/row, and confirm `acceptance-coverage` for the two round-028 acceptance features.

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface change this round).

*Handoff payload (for the next phase)*: plan package `specs/plans/028-user-global-prompt-log`; truth root `specs/truth`; truth-delta `specs/plans/028-user-global-prompt-log/truth-delta.md`; interfaces `CLI end`, `Shared prompt log store`; analysis focus as above; acceptance features `sharing-prompts-across-environments.feature`, `carrying-over-the-existing-prompt-history.feature`.

---

### Gating blockers

*(none — the operator directed the target path and the seed rule; `research.md` Decisions 1–7 settle the
location, the read/write move, the seed semantics (verbatim copy, copy-not-move, no-overwrite, empty on a
missing source), the adapter home of the seed, the no-op degrade, and the tell-me-go divergence. **Open
(non-blocking):** the exact DBML prose (`/axb-data-plan`); the DSL row wording + the seed Example
(`/axb-dsl-refine`); the adapter's constructor seam shape (an implementation detail). None gate this
round.)*
