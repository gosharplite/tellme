# System Analysis Plan — round 021 (`021-tool-surface-parity`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/021-tool-surface-parity/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/acceptance/*.feature  # /axb-spec-by-example ✓ done (4 journeys: US1–US4)
└── tasks.md                       # produced by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY ✓ done
└── features/cli/chat/**           # /axb-dsl-refine — MODIFY / ADD / DELETE (contract owner)
```

*(No `contracts/**` change (`/axb-api-plan` `NOOP`), no `data/**` change (`/axb-data-plan` `NOOP`),
and no `ui/**` artifact.)*

### Repository structure (root)

```text
internal/infrastructure/tools/filesystem.go   # CHANGED — reshape list_files + multi-file read_files; add get_tree
internal/infrastructure/tools/get_tree.go     # ADDED   — the get_tree tool (+ a stdlib binary probe)
internal/infrastructure/tools/summarize.go    # DELETED — the summarize_history adapter
internal/cli/cli.go                           # CHANGED — newToolRegistry offers the three filesystem tools only
internal/agent/agentloop.go                   # CHANGED — echo the call's `reason` into the `[tool] …` stderr line
internal/**  (tests)                          # CHANGED — unit tests for the three tools + the reason echo
tests/e2e/steps/*, suite                       # CHANGED — remove the summarise steps; add list/read/tree steps
specs/truth/techstack.md                       # MODIFY  — Read-only filesystem tools (+ reason, get_tree) / Agent tool loop ✓ done
go.mod / go.sum                                # unchanged — stdlib-only (no new dependency)
```

**Structure Decision**: Round 021 reshapes the **agent tool surface** *inside* the existing CLI end — it
does not add a system boundary. The reference-aligned contracts (multi-file `read_files`, the
`Contents of …` `list_files`, the new `get_tree`, the required `reason`, and the reference limits) live in
`internal/infrastructure/tools` behind the unchanged `internal/domain/tools` port, and the loop
(`AgentLoop`) is unchanged except for echoing `reason` in its existing `stderr` log line. There is **no**
new endpoint, **no** persisted-state change, **no** CLI flag/config change, and **no** new dependency,
consistent with `research.md` Decisions 1–8. The one truth change beyond `techstack.md`
(`/axb-technical-research`) is the **CLI interface truth** under `specs/truth/features/cli/chat/**`,
owned by the CLI end's contract owner `/axb-dsl-refine`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface — the **CLI end** (the `chat` capability: the agent
tool surface a prompt run may call). The round **changes the CLI end's observable behaviour** (offered
tool set, tool argument schemas, and tool output formats), so the interface must be carried to its
contract owner for an interface-truth change.

There is **no** analysis planner for the CLI end (per the CLI-streamlined model): it is not delegated in a
`Wave`; it is carried forward to its **contract owner `/axb-dsl-refine`** at delivery
(`wave-covers-interfaces`'s carried-forward branch).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface; the tool surface authors no request/response shape).
> - `/axb-data-plan` = **`NOOP`** (the persisted tool-step shape `{tool, arguments, result[, signature]}` is unchanged; only the `read_files` argument *content* now carries `filepaths[]`).
> - `/axb-dsl-refine` = **contract owner** (MODIFY the `chat` `read_files`/`list_files` rows, ADD `get_tree`, DELETE `summarising-the-conversation`, pin the `reason` echo).
> - `/axb-ui-plan` = **skipped** (no UX surface change; the tool surface is model-facing, not operator-facing).
> - `/axb-spec-by-example` = **done** (4 acceptance journeys: US1–US4).

### Analysis Wave schedule

**No waves.** The CLI end has no analysis planner, so there is no interface to order or delegate within
a wave; the round's contract-owner handoff (`/axb-dsl-refine`) happens at delivery, not inside a wave.

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract owner**: MODIFY `specs/truth/features/cli/chat/using-a-tool.feature` (multi-file `read_files`, `reason`, `list_files` format) + `chat/dsl.md`; ADD a `get_tree` interface feature + rows; DELETE `specs/truth/features/cli/chat/summarising-the-conversation.feature` + its rows.

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface change).

*Handoff payload (for the next phase)*: plan package `specs/plans/021-tool-surface-parity`; truth root
`specs/truth`; truth-delta `specs/plans/021-tool-surface-parity/truth-delta.md`; interfaces: **1 (CLI end
/ `chat`)** carried to `/axb-dsl-refine`; the round's delivery is the reshaped/added/removed tools in
`internal/infrastructure/tools` + the `reason` echo in `internal/agent/agentloop.go` + the `chat`
interface truth.

---

### Gating blockers

*(none — the operator locked D1–D5 before `/axb-specify`; `research.md` Decisions 1–8 settle the
contracts, the edge handling, the `reason` echo mechanism, the removal, and the dependency stance. No
open decision gates the round.)*
