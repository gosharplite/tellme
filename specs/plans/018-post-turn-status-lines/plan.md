# System Analysis Plan — round 018 (`018-post-turn-status-lines`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/018-post-turn-status-lines/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/                # 3 journeys — /axb-spec-by-example
│       ├── reporting-the-metrics-of-the-request.feature
│       ├── reporting-the-cost-and-session-summary.feature
│       └── keeping-the-post-turn-lines-bounded.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY (post-turn lines + MODELS + tokens.log) ✓ done
├── data/data-model.dbml           # /axb-data-plan — ADD (the `usage_record` table)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/chat/…                 # MODIFY — the post-turn status-line contract + DSL rows
```

*(No `contracts/**` change and no `ui/**` artifact this round — a plain-CLI round: `/axb-api-plan` is
`NOOP`, and `/axb-ui-plan` is skipped (the round changes no UX surface — the lines are line-oriented
terminal output, not a TUI).)*

### Source-code structure (repository root)

```text
internal/
├── ui/
│   ├── status.go                  # unchanged — the round-009 payload status formatter
│   ├── metrics.go                 # NEW — the round-018 metric/summary formatters: the metrics line
│   │                              #   `[HH:MM:SS] [<provider>] M: … H: … C: … Th: …` (Th always) and the
│   │                              #   `╰─⠿ Ready ($… $… $… - M: … H: … O: … - …%)` summary (plain text;
│   │                              #   the injected clock supplies the timestamp)
│   └── pricing.go                 # NEW — pure cost arithmetic (miss/hit/completion)×rates + hit-rate
├── config/…                       # CHANGED — the `MODELS` pricing table (HIT/MISS/COMP per model)
├── domain/
│   ├── llm/…                      # CHANGED — widen Usage with CachedTokens + ThinkingTokens
│   └── history/…                  # CHANGED — a Usage record type + store port for the per-mode log
├── infrastructure/
│   ├── llm/openai/…               # CHANGED — parse `prompt_tokens_details` / `completion_tokens_details`
│   └── history/usage.go           # NEW — the per-mode `tokens.log` store (append per call, sum the
│                                  #   session, rotate on `--new`)
├── agent/agentloop.go             # CHANGED — accumulate every provider call's usage on the result
└── cli/cli.go                     # CHANGED — after the answer, emit the metrics line + the Ready line;
                                   #   rotate the usage log on `--new`
```

```text
go.mod / go.sum                     # unchanged — no new module (stdlib + the existing internal packages)
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 018 is a **bounded presentation + local-state change** on the **CLI end**. It
adds the reference's post-turn status — the per-turn **metrics line** (from the API call that just
returned) and the **`╰─⠿ Ready`** session summary (three costs + session token totals + cache-hit rate) —
to the prompt surfaces, backed by a **config-only `MODELS` pricing table** and a **per-mode API-call
usage log** (`tokens.log`). There is **no** new endpoint and **no** new dependency, consistent with
`research.md` Decisions 1–9. The **executable contract** is pinned in `specs/truth/features/cli/**` by
`/axb-dsl-refine`; the formatters live beside the existing status formatter in `internal/ui`; the usage log
is a new per-mode JSON-Lines store beside `history.jsonl`; `specs/truth/contracts/**` and `ui/**` are
untouched.

---

## Analysis Plan

### System interface inventory

This requirement inventories **2** interfaces. Round 018 introduces **no** new system end: it changes the
CLI end's prompt-turn output and adds one piece of persisted local state.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: the **prompt-bearing turn** (positional/piped, the round-012 reader, and the `-i` submit path) — after the answer, report the reference's post-turn status on the **diagnostic stream (`stderr`)**: a **metrics line** `[HH:MM:SS] [<provider>] M: <miss> H: <cached> C: <completion> Th: <thinking>` (from the API call that just returned; `Th:` **always** shown) and a **`╰─⠿ Ready` summary** `($<lastCall> $<turn> $<session> - M: … H: … O: … - <hit%>%)`. `stdout` stays byte-exact; the class-phrase vocabulary is unchanged; both lines are suppressed only when the provider reports no usage; the non-prompt paths are excluded.
   - Requirement evidence: `FR-001`–`FR-014`, `NFR-001`–`NFR-004`; acceptance features `reporting-the-metrics-of-the-request.feature`, `reporting-the-cost-and-session-summary.feature`, `keeping-the-post-turn-lines-bounded.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

2. `Session usage log (per-mode API-call usage)`
   - Endpoint type: `local-state / data endpoint`
   - Primary interface: the per-mode **`tokens.log`** — one JSON record per provider call (`{timestamp, provider, model, cached_tokens, prompt_tokens, response_tokens, total_tokens, thinking_tokens, cost}`), appended after each call under `$TELL_ME_HOME/output/<mode>/`; the `╰─⠿ Ready` summary sums it (session token totals + session cost + cache-hit rate); `--new` archives/rotates it alongside `history.jsonl`.
   - Requirement evidence: `FR-006`, `FR-007`, `FR-010`; edge cases "resumed session" and "`--new` resets".
   - Planner: **`/axb-data-plan`** — entity/field/lifecycle/session responsibilities (a `data/**` ADD).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface of tellme's own; round 018 adds no tellme-owned request/response shape — it only reads the provider's existing `usage`).
> - `/axb-ui-plan` = **skipped** — the round changes no user-facing UX surface; the lines are line-oriented terminal output, not a TUI. No plan-side `ui/**` artifact.
> - **Truth amendment carried to `/axb-dsl-refine`**: **MODIFY** `specs/truth/features/cli/chat/**` — the post-turn status-line contract as executable Rules + DSL rows (the metrics line, the Ready line, the presence/suppression rule, the `stdout`-byte-exact boundary). No new class phrase; the root `cli/dsl.md` vocabulary is unchanged.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
  - `Session usage log (per-mode API-call usage)`
- Analysis focus:
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin the post-turn status-line contract as mechanically assertable interface Rules — the metrics line (M/H/C/Th, `Th` always), the Ready line (three costs + session totals + hit-rate), the presence/suppression rule (emit on every prompt-bearing turn unless the provider reports no usage), and the boundary (`stderr` only, `stdout` byte-exact, no new class phrase) — while `$<call> ≤ $<turn> ≤ $<session>` holds (strict `$<call> < $<turn>` on a tool-using turn).
  - **Session usage log** → **`/axb-data-plan`**: model the per-call usage record, its per-mode append lifecycle, the session-total read, and the `--new` rotation (reusing the existing session-directory layout).
  - **API** → **`NOOP`**; **UI** → **skipped**.
- Scheduling rationale: the CLI contract and the usage-log data model have **no dependency on each other's conclusions** (the log is a self-contained per-mode file; the lines are its reader), so they are analysed together in one wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 (`Session usage log`) → **ADD** the `usage_record` table in `specs/truth/data/data-model.dbml`.
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → the post-turn status-line behaviour in `specs/truth/features/cli/**` (add the metrics + Ready rules; extend `chat/dsl.md`), and confirm `acceptance-coverage` for the three round-018 acceptance features.

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface this round).

*Handoff payload (for the next phase)*: plan package `specs/plans/018-post-turn-status-lines`; truth root `specs/truth`; truth-delta `specs/plans/018-post-turn-status-lines/truth-delta.md`; interfaces `CLI end` + `Session usage log`; analysis focus as above; acceptance features `features/acceptance/**` (3).

---

### Gating blockers

*(none — the operator locked the line shapes, the three-cost semantics, the line-2 source, the `tokens.log` storage, the presence rule, and the config-only pricing source across the interview + clarify round; `research.md` Decisions 1–9 settled the usage widening, the pricing mechanism, the cost formula, the log format, the formatter, the accumulation, the suppression rule, and the verification. **Open (non-blocking):** the exact DSL step vocabulary (`/axb-dsl-refine`) and the exact formatter/accumulation layout (implementation). None gate this round.)*
