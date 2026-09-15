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
internal/infrastructure/tools/filesystem.go   # CHANGED — retire the fixed 100000-byte / 1 MiB caps; whole-file reads + one aggregate BYTE bound + skip marker
internal/infrastructure/tools/get_tree.go      # CHANGED — bound by the shared bound (byte budget)
internal/infrastructure/tools/command.go       # ADDED   — the bash-first execute_command tool (os/exec `bash -c`; bounded StdoutPipe/StderrPipe; output_file/append; process-group timeout)
internal/domain/tools/tools.go                  # CHANGED — widen the Tool port TWO-WAY: an upward per-tool contract descriptor (default timeout) + the downward resolved BYTE budget on Execute (grill Q2); each tool's JSON-schema `timeout`/`max_output_tokens` description MUST derive from the same constant as its `DefaultTimeout()` (review R5)
internal/agent/agentloop.go                     # CHANGED — the single tool-resource-contract enforcement point (applies the timeout; delegates bound resolution/conversion/clamp to the pure resolver below)
internal/agent/tool_contract.go                 # ADDED   — the PURE contract resolver (`resolveBound(param, ceiling, default)` + `clampBytes(result, byteBudget)`); unit-tested, kept out of `Run` (review R4)
internal/config/config.go                       # CHANGED — add `ModelPricing.ContextWindow` (yaml `CONTEXT_WINDOW`) + a `ContextWindowFor(model)` accessor (mirrors `PricingFor`) (grill Q7)
internal/cli/cli.go                             # CHANGED — register execute_command; extend `resolution` with the run-static effective budget and compute it (min of MAX_HISTORY_TOKENS and the window) in `resolve()`; thread it into the AgentLoop literal (≈L638)
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
persisted-state change (the tool-step record is unchanged), and **no** new dependency — consistent with `research.md` Decisions 1–8. It **does** add a config surface: an optional per-model `MODELS.<model>.CONTEXT_WINDOW` read by `internal/config/config.go` and resolved into the run-static effective budget in `internal/cli/cli.go`'s `resolve()` (grill Q7 — replacing an earlier "no config change" claim). The one truth
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
> - `/axb-dsl-refine` = **contract owner** (ADD an `execute_command` interface feature + rows, incl. a **process-tree stop** witness; MODIFY `reading-several-files` + `chat/dsl.md` for the token-bound/timeout params — the bound stated in **bytes** — and the new truncation/skip markers; `listing-a-directory`/`surveying-a-folder-tree` keep their shape but gain an **ADD** bound witness).
> - `/axb-ui-plan` = **skipped** (no UX surface change; the tool surface is model-facing, and the operator `stderr` chrome is unchanged).
> - `/axb-spec-by-example` = **done** (3 acceptance journeys: US1 run a shell command · US2 bounded reader results · the offered-tool set).

### Analysis Wave schedule

**No waves.** The CLI end has no analysis planner, so there is no interface to order or delegate within a
wave; the round's contract-owner handoff (`/axb-dsl-refine`) happens at delivery, not inside a wave.

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract owner**: ADD `specs/truth/features/cli/chat/running-a-shell-command.feature` (+ `chat/dsl.md` rows: the command tool offered, an exited command, a non-zero exit as a success result, a timed-out command — a **nil-error result**, not an `error: ` render — output captured to a file, a **process-tree stop** witness, and the `reason` echo); MODIFY `reading-several-files` + `chat/dsl.md` for the `max_output_tokens`/`timeout` params, the whole-file read, and the truncation/skip markers (the bound stated in **bytes**). `listing-a-directory` / `surveying-a-folder-tree` keep their **output shape** (NOOP) but gain an **ADD** bound witness (a scripted small `max_output_tokens` trips the truncation marker) — grill Q5, since FR-011 must be falsifiable.

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
