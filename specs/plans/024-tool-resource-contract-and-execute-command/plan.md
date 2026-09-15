# System Analysis Plan — round 024 (`024-tool-resource-contract-and-execute-command`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/024-tool-resource-contract-and-execute-command/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/acceptance/*.feature  # /axb-spec-by-example ✓ done (3 journeys: US1–US2 + the offered-tool set)
└── tasks.md                       # produced by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY ✓ done
└── features/cli/chat/**           # /axb-dsl-refine — ADD / MODIFY (contract owner)
```

*(No `contracts/**` change (`/axb-api-plan` `NOOP`), no `data/**` change (`/axb-data-plan` `NOOP`),
and no `ui/**` artifact — the tool surface is model-facing, not operator-facing.)*

### Repository structure (root)

```text
internal/infrastructure/tools/filesystem.go   # CHANGED — retire the fixed 100000-byte / 1 MiB caps; whole-file reads + one aggregate bound + skip marker
internal/infrastructure/tools/get_tree.go      # CHANGED — bound by the shared max_output_tokens
internal/infrastructure/tools/command.go       # ADDED   — the bash-first execute_command tool (os/exec `bash -c`; output_file/append)
internal/domain/tools/tools.go                  # CHANGED — the resolved-bound carrier passed to tools (for early-stop)
internal/agent/agentloop.go                     # CHANGED — the single tool-resource-contract enforcement point (resolve default→param→ceiling; clamp + timeout)
internal/cli/cli.go                             # CHANGED — register execute_command; wire the budget/timeout defaults
internal/**  (tests)                            # CHANGED — unit tests for the contract resolution, execute_command, and the reader bound
tests/e2e/steps/*, suite                        # CHANGED — execute_command steps + bounded-reader steps (fake provider records the request)
specs/truth/techstack.md                        # MODIFY  — Tool resource contract + Agent command tool ✓ done
go.mod / go.sum                                 # unchanged — stdlib-only (no new dependency)
```

**Structure Decision**: Round 024 adds an **execution tool** and a **cross-cutting resource contract**
*inside* the existing CLI end — it adds no system boundary. The new `execute_command` (bash-first
`bash -c`) and the retrofitted readers live in `internal/infrastructure/tools` behind the unchanged
`internal/domain/tools` port; the **agent loop** becomes the single place the contract is resolved and
enforced (it already owns the per-tool timeout, round-008 FR-009). There is **no** new endpoint, **no**
persisted-state change (the tool-step record is unchanged), **no** CLI flag/config change beyond the tool
registration, and **no** new dependency — consistent with `research.md` Decisions 1–8. The one truth
change beyond `techstack.md` (`/axb-technical-research`) is the **CLI interface truth** under
`specs/truth/features/cli/chat/**`, owned by the CLI end's contract owner `/axb-dsl-refine`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface — the **CLI end** (the `chat` capability: the agent
tool surface a prompt run may call, plus the tool resource contract that bounds it). The round **changes
the CLI end's observable behaviour** (a new tool is offered; the readers' bound becomes a parameter; a
repo-wide result-bounding/timeout contract is introduced), so the interface must be carried to its
contract owner for an interface-truth change.

There is **no** analysis planner for the CLI end (per the CLI-streamlined model): it is not delegated in a
`Wave`; it is carried forward to its **contract owner `/axb-dsl-refine`** at delivery
(`wave-covers-interfaces`'s carried-forward branch).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface; the tool surface authors no request/response shape).
> - `/axb-data-plan` = **`NOOP`** (the persisted tool-step shape `{tool, arguments, result[, signature]}` is unchanged; `execute_command` steps store opaquely as today).
> - `/axb-dsl-refine` = **contract owner** (ADD an `execute_command` interface feature + rows; MODIFY the reader features + `chat/dsl.md` for the token-bound/timeout params and the new truncation/skip markers).
> - `/axb-ui-plan` = **skipped** (no UX surface change; the tool surface is model-facing, and the operator `stderr` chrome is unchanged).
> - `/axb-spec-by-example` = **done** (2 acceptance journeys: US1 run a shell command · US2 bounded reader results).

### Analysis Wave schedule

**No waves.** The CLI end has no analysis planner, so there is no interface to order or delegate within a
wave; the round's contract-owner handoff (`/axb-dsl-refine`) happens at delivery, not inside a wave.

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract owner**: ADD `specs/truth/features/cli/chat/running-a-shell-command.feature` (+ `chat/dsl.md` rows: the command tool offered, an exited command, a non-zero exit as a success result, a timed-out command, output captured to a file, and the `reason` echo); MODIFY the reader features (`reading-several-files`, `listing-a-directory`, `surveying-a-folder-tree`) + `chat/dsl.md` for the `max_output_tokens`/`timeout` params, the whole-file read, and the truncation/skip markers.

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface change).

*Handoff payload (for the next phase)*: plan package `specs/plans/024-tool-resource-contract-and-execute-command`;
truth root `specs/truth`; truth-delta `specs/plans/024-tool-resource-contract-and-execute-command/truth-delta.md`;
interfaces: **1 (CLI end / `chat`)** carried to `/axb-dsl-refine`; the round's delivery is the new
`execute_command` tool + the loop-enforced tool resource contract + the retrofitted readers in
`internal/infrastructure/tools` + the contract enforcement in `internal/agent/agentloop.go` + the `chat`
interface truth.

---

### Gating blockers

*(none — the operator locked D1–D7 and clarify Q1–Q3 before `/axb-specify`; `research.md` Decisions 1–8
settle the invocation, the exit semantics, the capture, the contract tiers, the default/ceiling
derivation, the reader retrofit, the markers, and the dependency stance. No open decision gates the
round.)*
