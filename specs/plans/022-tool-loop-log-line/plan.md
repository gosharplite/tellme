# System Analysis Plan — round 022 (`022-tool-loop-log-line`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/022-tool-loop-log-line/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/acceptance/*.feature  # /axb-spec-by-example ✓ done (2 journeys: US1–US2)
└── tasks.md                       # produced by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY ✓ done
└── features/cli/chat/**           # /axb-dsl-refine — MODIFY (contract owner)
```

*(No `contracts/**` change (`/axb-api-plan` `NOOP`), no `data/**` change (`/axb-data-plan` `NOOP`),
and no `ui/**` artifact.)*

### Repository structure (root)

```text
internal/agent/agentloop.go          # CHANGED — reshape `logStep` to `[HH:MM:SS] [Tool] <name> - <reason>`
                                     #           (+ a `Now func() time.Time` clock seam; drop arguments/result)
internal/ui/toollog.go               # ADDED   — the pure `FormatToolLog(t, name, reason)` formatter
internal/cli/cli.go                  # CHANGED — set `loop.Now = env.now`; emit one blank line (stderr) before
                                     #           the answer when the turn ran ≥1 tool step (len(result.Steps) > 0)
internal/agent/agentloop_reason_test.go  # CHANGED — pin the new line shape (timestamp + `[Tool] <name>` + reason;
                                     #           no `arguments=`/`result=`; the no-reason form)
internal/ui/toollog_test.go          # ADDED   — the formatter (injected clock; both forms)
tests/e2e/steps/*, suite             # CHANGED — re-pin the tool-loop log Thens + add the blank-line Then
specs/truth/techstack.md             # MODIFY  — Agent tool loop (+ pure-helper unit tests) ✓ done
go.mod / go.sum                      # unchanged — stdlib-only (no new dependency)
```

**Structure Decision**: Round 022 reshapes the **tool-loop diagnostic rendering** *inside* the existing
CLI end — it does not add a system boundary. The loop (`AgentLoop`) stays the single tool-log emitter
(round-008 Decision 7), now printing the operator-chosen single line via a pure `internal/ui` formatter and
an injected clock seam (round-009/017 precedent); the CLI (`runTurn`) owns the answer write and therefore
the one blank line that separates a tool-using turn's log block from the answer (round-017 chrome-gap
precedent). There is **no** new endpoint, **no** persisted-state change, **no** CLI flag/config change, and
**no** new dependency, consistent with `research.md` Decisions 1–8. The one truth change beyond
`techstack.md` (`/axb-technical-research`) is the **CLI interface truth** under
`specs/truth/features/cli/chat/**`, owned by the CLI end's contract owner `/axb-dsl-refine`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface — the **CLI end** (the `chat` capability: the
tool-loop diagnostic it writes to `stderr` during a run). The round **changes the CLI end's observable
behaviour** (the tool-loop log line shape and the blank line before the answer), so the interface must be
carried to its contract owner for an interface-truth change.

There is **no** analysis planner for the CLI end (per the CLI-streamlined model): it is not delegated in a
`Wave`; it is carried forward to its **contract owner `/axb-dsl-refine`** at delivery
(`wave-covers-interfaces`'s carried-forward branch).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface; the log line authors no request/response shape).
> - `/axb-data-plan` = **`NOOP`** (no persisted-state change — the tool-step record `{tool, arguments, result[, signature]}` and the `history.jsonl` shape are unchanged; the log line is operator-facing output, not stored).
> - `/axb-dsl-refine` = **contract owner** (MODIFY the `chat` tool-loop log rows in `chat/dsl.md` and `watching-the-tool-loop.feature`; add the blank-line separation).
> - `/axb-ui-plan` = **skipped** (no operator UX surface change; the log line is a `stderr` diagnostic, not an interactive surface).
> - `/axb-spec-by-example` = **done** (2 acceptance journeys: US1–US2).

### Analysis Wave schedule

**No waves.** The CLI end has no analysis planner, so there is no interface to order or delegate within a
wave; the round's contract-owner handoff (`/axb-dsl-refine`) happens at delivery, not inside a wave.

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract owner**: MODIFY `specs/truth/features/cli/chat/watching-the-tool-loop.feature` + `chat/dsl.md` — the tool-loop log line shape (`[HH:MM:SS] [Tool] <tool name> - <reason>`, no-reason form) and the blank-line separation before the answer.

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface change).

*Handoff payload (for the next phase)*: plan package `specs/plans/022-tool-loop-log-line`; truth root
`specs/truth`; truth-delta `specs/plans/022-tool-loop-log-line/truth-delta.md`; interfaces: **1 (CLI end /
`chat` tool-loop diagnostic)** carried to `/axb-dsl-refine`; the round's delivery is the reshaped
`logStep` (+ the `internal/ui` formatter and the clock seam) in `internal/agent` + `internal/ui` and the
blank line in `internal/cli` + the `chat` interface truth.

---

### Gating blockers

*(none — the operator locked Q1–Q3 before `/axb-specify`; `research.md` Decisions 1–8 settle the line
shape, the timestamp seam, the blank-line emission point, the no-reason form, the streams, the testing, and
the dependency stance. No open decision gates the round.)*
