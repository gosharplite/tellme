# System Analysis Plan — round 029 (`029-agent-write-tools`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/029-agent-write-tools/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/acceptance/*.feature  # /axb-spec-by-example ✓ done (3 journeys: create · edit · offered tool set)
└── tasks.md                       # produced by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY ✓ done
└── features/cli/chat/**           # /axb-dsl-refine — ADD / MODIFY (contract owner)
```

*(No `contracts/**` change (`/axb-api-plan` `NOOP`), no `data/**` change (`/axb-data-plan` `NOOP`),
and no `ui/**` artifact — the write tools are model-facing, not operator-facing.)*

### Repository structure (root)

```text
internal/infrastructure/tools/writer.go   # ADDED   — the two write tools: write_file (create-only + atomic temp+rename + MkdirAll) and replace_text (strict-unique); both on the domain/tools.Tool port
internal/cli/cli.go                        # CHANGED — register the two write tools in the newToolRegistry seam (offer order: readers, write pair, command tool)
internal/infrastructure/tools/writer_test.go # ADDED — unit tests (create-only, atomicity, strict-unique, empty old_text, missing file, parent creation)
tests/e2e/steps/*, suite                    # CHANGED — write-tool steps (fake provider scripts "creates/edits a file before answering"); assert file bytes + refusals
specs/truth/techstack.md                    # MODIFY  — Write filesystem tools row + deferred-bullet rewrite ✓ done
go.mod / go.sum                             # unchanged — stdlib-only (no new dependency)
```

**Structure Decision**: Round 029 adds **two agent tools** *inside* the existing CLI end — it adds no
system boundary. The tools live in `internal/infrastructure/tools` (a new `writer.go`) behind the
**unchanged** `internal/domain/tools.Tool` port, and are registered in the existing `newToolRegistry`
seam in `internal/cli/cli.go`. There is **no** new endpoint, **no** persisted-state change (the write
tools persist nothing of their own), and **no** new dependency — consistent with `research.md`
Decisions 1–8. The only truth change beyond `techstack.md` (`/axb-technical-research`) is the **CLI
interface truth** under `specs/truth/features/cli/chat/**`, owned by the CLI end's contract owner
`/axb-dsl-refine`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface — the **CLI end** (the `chat` capability: the agent
tool surface a prompt run may call). The round **changes the CLI end's observable behaviour**: a prompt
run now offers **two** additional tools (a file-creation tool and a file-editing tool), so the offered
tool set and the tools' observable outcomes change. The interface must therefore be carried to its
contract owner for an interface-truth change.

There is **no** analysis planner for the CLI end (per the CLI-streamlined model): it is not delegated in a
`Wave`; it is carried forward to its **contract owner `/axb-dsl-refine`** at delivery
(`wave-covers-interfaces`'s carried-forward branch).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface; the write tools author no request/response shape).
> - `/axb-data-plan` = **`NOOP`** (the write tools persist no state; the existing persisted shapes — `history.jsonl`, `tokens.log`, the prompt logs — are unchanged).
> - `/axb-dsl-refine` = **contract owner** (ADD a `chat` interface feature for the write pair — create-only / atomic / strict-unique — + `chat/dsl.md` rows; MODIFY the offered-tool-set row/feature so the write tools are part of the offered set).
> - `/axb-ui-plan` = **skipped** (no UX surface change; the tools are model-facing, and the operator `stderr` chrome is unchanged).
> - `/axb-spec-by-example` = **done** (3 acceptance journeys: creating a file · editing a file · the offered tool set).

### Analysis Wave schedule

**No waves.** The CLI end has no analysis planner, so there is no interface to order or delegate within a
wave; the round's contract-owner handoff (`/axb-dsl-refine`) happens at delivery, not inside a wave.

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract owner**: ADD `specs/truth/features/cli/chat/creating-and-editing-files.feature` (Rules: a new file is created with exactly the requested content; a missing folder is created; an existing file is never overwritten; a uniquely-identified block is replaced and nothing else changes; an edit that cannot be uniquely placed is refused) + `chat/dsl.md` rows (the file-creation tool offered and its outcome; the file-editing tool offered and its outcome; the create-only refusal; the strict-unique refusal on 0 / >1 matches); MODIFY the offered-tool-set row/feature so the write tools are part of the offered set. *(Exact file/rule/row names are `/axb-dsl-refine`'s call.)*

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface change).

*Handoff payload (for the next phase)*: plan package `specs/plans/029-agent-write-tools`;
truth root `specs/truth`; truth-delta `specs/plans/029-agent-write-tools/truth-delta.md`;
interfaces: **1 (CLI end / `chat`)** carried to `/axb-dsl-refine`; the round's delivery is the two write
tools in `internal/infrastructure/tools/writer.go`, their registration in `internal/cli/cli.go`, and the
`chat` interface truth.

---

### Gating blockers

*(none — the operator locked the design in-session before `/axb-specify` (scope = the pair; `write_file`
create-only **A**; `write_file` atomic; `replace_text` strict-unique **A**; no security/undo; 30 s
timeout); `research.md` Decisions 1–8 settle the surface, the create-only and atomic write, the
strict-unique replace, the no-security/no-undo stance, the resource-contract integration, the
failure shape, and the dependency stance. No open decision gates the round.)*
