# System Analysis Plan — round 031 (`031-tool-schema-wellformedness`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/031-tool-schema-wellformedness/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
└── tasks.md                       # produced by /axb-tasks

specs/truth/
└── techstack.md                   # /axb-technical-research — MODIFY ✓ done (Agent tool schemas + Agent tool-schema gate rows)
```

*(No `features/acceptance/**` (no new user-facing CLI journey — `/axb-spec-by-example` **skipped**), no `features/cli/**` change
(`/axb-dsl-refine` **`NOOP`**), no `contracts/**` change (`/axb-api-plan` **`NOOP`**), no `data/**` change
(`/axb-data-plan` **`NOOP`**), and no `ui/**` artifact.)*

### Repository structure (root)

```text
internal/infrastructure/tools/filesystem.go    # CHANGED — the shared schema builder (resourceSchema) declares the mandatory `reason` property, so every tool it backs satisfies required ⊆ properties
internal/infrastructure/tools/filesystem_test.go # CHANGED — strengthen/replace the blind-spot TestToolSchemasRequireReason (it asserted only that `required` contains `reason`)
internal/cli/tool_registry_test.go             # ADDED   — the well-formedness gate over newToolRegistry(): for EVERY registered tool, every name in `required` is declared under `properties` and the schema parses
internal/infrastructure/tools/command.go       # unchanged — execute_command builds its schema inline and already declares `reason` (the lone compliant tool)
internal/infrastructure/tools/{writer,get_tree}.go # unchanged callers — they already pass `"reason"` as required; the builder now declares its property
specs/truth/techstack.md                        # MODIFY — Agent tool schemas row + Agent tool-schema gate row ✓ done
go.mod / go.sum                                 # unchanged — stdlib-only (no new dependency)
```

**Structure Decision**: Round 031 is an **internal tool-schema correctness + verification-gate** change
*inside* the existing CLI end — it adds **no** system boundary. The fix lives in the **shared schema
builder** (`internal/infrastructure/tools/filesystem.go`), which builds the argument schema for every
agent tool behind the **unchanged** `internal/domain/tools.Tool` port; the correction makes each tool's
advertised schema satisfy `required ⊆ properties` (the invariant a strict provider such as Vertex/Gemini
enforces). The regression gate is a **hermetic unit test over the production assembler `agentTools()`**
(`internal/cli/tool_registry_test.go`, `newToolRegistry()`), not a runtime behaviour. There is **no** new
endpoint, **no** persisted-state change, **no** change to the CLI's observable behaviour or the offered
tool set, and **no** new dependency — consistent with `research.md` Decisions 1–6. The **only** truth
change is `specs/truth/techstack.md`, owned by `/axb-technical-research` (already applied).

---

## Analysis Plan

### System interface inventory

This requirement inventories **0** system interfaces. The round corrects an **internal advertised-schema
detail** and adds a **hermetic verification gate**: it changes **no** backend, frontend, or CLI interface
truth — the offered tool set, the CLI flags/stdout/stderr/exit-code contract, and the CLI's user-visible
behaviour are all unchanged (the existing Vertex/Gemini journeys simply stop failing). There is therefore
no interface to delegate or carry forward.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface; the schema fix authors no request/response shape).
> - `/axb-data-plan` = **`NOOP`** (no persisted or in-memory state change; the round writes no file format).
> - `/axb-dsl-refine` = **`NOOP`** (no CLI interface truth change — the tool-schema well-formedness is asserted by a **production-assembler (`agentTools()`) gate**, not a new/changed Gherkin rule or DSL row; the offered-tool-set rows are unchanged — round-020 precedent).
> - **Recorded divergence (ARCH-5)** — the round's acceptance (US1/US2) is carried by the **production-assembler (`agentTools()`) gate + a manual live Vertex/Gemini check**, **not** by executable Gherkin (`/axb-spec-by-example` skipped, `/axb-dsl-refine` NOOP): the tool schemas are not observable through the built binary hermetically. Recorded so `acceptance-coverage`'s intent — *PM-defined acceptance is executable* — is not silently re-interpreted (round-020 precedent).
> - `/axb-ui-plan` = **skipped** (no UX surface; nothing operator-facing changes).
> - `/axb-spec-by-example` = **skipped** (no new user-facing CLI journey — the defect fix restores an existing journey rather than adding one).

### Analysis Wave schedule

**No waves.** There is no interface to order or delegate; the round's only truth change (`techstack.md`)
is RD-side and owned by `/axb-technical-research` (already applied in the plan/truth half).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — `NOOP` (no CLI interface truth change).

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface).

*Handoff payload (for the next phase)*: plan package `specs/plans/031-tool-schema-wellformedness`;
truth root `specs/truth`; truth-delta `specs/plans/031-tool-schema-wellformedness/truth-delta.md`;
interfaces: **none**; the round's delivery is the shared-builder fix
(`internal/infrastructure/tools/filesystem.go`) + the registry well-formedness gate
(`internal/cli/tool_registry_test.go`) + the `techstack.md` truth rows.

---

### Gating blockers

*(none — the two high-impact gaps were resolved in Clarify Round 1 (Q1 → hermetic gate + manual live
check; Q2 → minimal fix + gate), and `research.md` Decisions 1–6 settle the fix site, the gate's location,
the `execute_command` stance, the verification posture, the dependency stance, and the deferred recurrence
guards. No open decision gates the round.)*
