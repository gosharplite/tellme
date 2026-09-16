# System Analysis Plan — round 033 (`033-skills-system`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/033-skills-system/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/acceptance/*.feature  # /axb-spec-by-example — ✅ done (journey: a prompt lists the workspace's pre-loaded skills)
└── tasks.md                       # produced by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY ✓ done (new Skills rows + schema/gate/tool-usage corrections)
└── features/cli/**                # /axb-dsl-refine — ADD (contract owner)
```

*(No `contracts/**` change (`/axb-api-plan` `NOOP`), no `data/**` change (`/axb-data-plan` `NOOP`),
and no `ui/**` artifact — the skills surface is model-facing and its CLI behaviour is carried by the CLI end.)*

> **`/axb-spec-by-example` — done.** Round 033 adds a user-facing CLI journey (on a prompt-bearing turn
> the agent can now list the environment's pre-loaded skills), so a single acceptance journey was warranted
> (unlike the journey-less rounds 020/031). The acceptance feature
> `features/acceptance/discovering-the-available-skills.feature` was authored in `8219150` and consumed by
> `/axb-dsl-refine` (`3b619ca`).

### Repository structure (root)

```text
internal/domain/skills/                 # ADDED   — the Skill value type (Name, Description, Location); no repository/composite/selector framework
internal/infrastructure/skills/loader.go# ADDED   — the file loader: recursive walk of <TELL_ME_HOME>/docs/skills/, frontmatter (name+description) parse, best-effort
internal/infrastructure/tools/skills.go # ADDED   — the read-only `list_skills` agent tool (shared resourceSchema + mandatory `reason`; reader-class 30 s default)
internal/cli/cli.go                     # CHANGED — load the catalog on the prompt-bearing turn path and register `list_skills` in the production `agentTools()` assembler
internal/infrastructure/tools/filesystem_test.go / internal/cli/tool_registry_test.go # ADDED/CHANGED — the loader/tool unit tests + the round-031 gate now sees the 7th tool
tests/e2e/**                            # ADDED   — the E2E `list_skills` scenario (a known `docs/skills` set under the per-scenario TELL_ME_HOME)
specs/truth/techstack.md                # MODIFY  — Skills rows + schema/gate/tool-usage corrections ✓ done
specs/truth/features/cli/**             # ADD     — the executable `list_skills` interface truth (/axb-dsl-refine, contract owner)
go.mod / go.sum                         # unchanged — stdlib-only
```

**Structure Decision**: Round 033 adds a **minimal skills system *inside* the existing CLI end** — it adds
**no** new system boundary. A tiny domain value type (`internal/domain/skills`) plus a file loader
(`internal/infrastructure/skills`) plus **one** read-only agent tool (`list_skills`, in
`internal/infrastructure/tools/`) let a prompt run enumerate the environment's `<TELL_ME_HOME>/docs/skills`
catalog **on demand** (Q1 → on-demand only; Q3 → a `list_skills` tool). The tool flows through the
**unchanged** `internal/domain/tools.Tool` port and the **unchanged** agent tool loop, so there is **no**
change to the sequential execution model and **no** automatic skill injection (the assembled request and the
pre-flight estimate are unchanged — FR-006). The round **changes the CLI end's observable behaviour** (a
prompt can now list the loaded skills) and persists **no** new state (`/axb-data-plan` `NOOP` — the catalog
is read from disk each run; there is no install/remove and no cache). Skill **content is never injected**;
the agent opens a listed skill's file with the existing `read_files`. Consistent with `research.md`
Decisions 1–9. **Paths are pinned** so a `[P]` Phase 3 has no same-file collisions.

### Implementation constraints (review folds — round 033)

- **Catalog → tool wiring seam (pinned; FR-009).** The `docs/skills` load is bound **only on the prompt path** (`runTurn`), where the resolved runtime home is known. `agentTools()` stays **parameterless and performs no filesystem read** — it constructs `list_skills` with an **empty/unbound catalog source**, so the round-031 gate (which iterates `agentTools()`) **and** the offline `--tool-usage` report (which builds the same registry via `newToolRegistry()`) both touch **no** `docs/skills` (they never invoke `Execute`). The `list_skills` tool carries a **lazy `func() ([]skills.Skill, error)` catalog seam** that is set in `runTurn` and resolved **only inside `Execute`**; the `newToolRegistry` DI seam signature is left unchanged. *(Chosen over widening `agentTools()`/`newToolRegistry` to take the catalog, which would churn the round-031 gate call site.)*
- **Single-source the skills path.** `<TELL_ME_HOME>/docs/skills` is derived by a small `internal/home` helper (mirroring `EnsureWorkspace`) rather than re-joined in the loader.
- **Six → seven comment reconciliation (Phase 4/5).** After `list_skills` lands, the code comments that still say "six" (`newToolRegistry`'s doc and the round-031 gate's comment) are reconciled to "seven"; the techstack rows already say seven.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface — the **CLI end** (the agent tool surface a prompt run
offers and may call). The round **changes the CLI end's observable behaviour**: a prompt-bearing turn now
offers a read-only `list_skills` tool that returns the skills loaded from `<TELL_ME_HOME>/docs/skills/`, and
the agent can list them on demand (and open one with the existing reader) — **without** injecting any skill
content into the request.

1. `CLI end — the agent tool surface (incl. the read-only `list_skills` tool)`
   - 端點類型: `CLI 端點`
   - 主要介面: the offered tool set a prompt run presents to the model (now including `list_skills`); the catalog it lists — each skill's **name**, **description**, and **readable location** from `<TELL_ME_HOME>/docs/skills/`; the recursive frontmatter discovery rule; the empty/absent-directory behaviour; the on-demand read path (the agent opens a listed skill with the existing `read_files`); and the **no-injection** property.
   - 需求原文依據: spec US1 (FR-001…004 — load `docs/skills`; expose `list_skills`; recursive discovery that ignores non-skill Markdown; empty catalog reported, not an error) and US2 (FR-005…006 — each entry carries a readable location; **no** skill content is injected into the request).

> **Scope notes**:
> - Skills are **local files** tellme reads from disk — there is **no** outbound dependency and **no** new
>   exposed end. The reference's **skills.sh ecosystem** (`.skills/`, `search/install/remove_skill`) and its
>   **automatic relevance-based injection** are **out of scope** (operator scope, Q1/Q3; recorded in
>   `research.md` Decision 8 and `techstack.md` *Not Introduced Yet*).
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface).
> - `/axb-data-plan` = **`NOOP`** (the round persists no state — the catalog is read from disk each run;
>   there is no install/remove and no cache).
> - `/axb-dsl-refine` = **contract owner** (ADD the executable `list_skills` interface truth under
>   `specs/truth/features/cli/**` + the matching `chat/dsl.md` rows).
> - `/axb-ui-plan` = **skipped** (no UX-surface change; the operator `stderr` chrome is unchanged — the
>   skills system adds no new visible chrome).
> - `/axb-spec-by-example` = **✅ done** (the acceptance feature `discovering-the-available-skills.feature` was delivered; carried by `/axb-dsl-refine`'s interface truth).

### Analysis Wave schedule

**No waves.** The CLI end has no analysis planner, so there is no interface to order or delegate within a
wave; the round's contract-owner handoff (`/axb-dsl-refine`) happens at delivery, not inside a wave.

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract owner**: ADD the `list_skills` interface truth under
   `specs/truth/features/cli/**` — that on a prompt-bearing turn the agent can **list the environment's
   pre-loaded skills** (each skill's **name**, **description**, and **location**, from
   `<TELL_ME_HOME>/docs/skills/`), that discovery is **recursive** and lists only frontmatter skill files
   (a skill's `rules/*.md` is **not** a skill), that an **absent/empty** directory lists **none** (no
   failure), that a listed skill can be **opened** with the existing reader, and that **no** skill content
   is injected into the request — plus the matching `chat/dsl.md` rows. *(Exact file/rule/row names are
   `/axb-dsl-refine`'s call.)*

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX-surface change).

*Handoff payload (for the next phase)*: plan package `specs/plans/033-skills-system`;
truth root `specs/truth`; truth-delta `specs/plans/033-skills-system/truth-delta.md`;
interfaces: **1 (CLI end)** carried to `/axb-dsl-refine`; the round's delivery is the `Skill` value type +
the `docs/skills` file loader + the read-only `list_skills` tool wired into `agentTools()` on the prompt
path, plus the CLI interface truth.

---

### Gating blockers

*(none — the operator locked the design in-session before `/axb-specify` (Q1 → on-demand only; Q2 → moot;
Q3 → a `list_skills` tool); `research.md` Decisions 1–9 settle the source, the discovery/parse rule, the
on-demand surface, the load timing, the minimal shape, the output shape, the tool integration, the
out-of-scope boundary, and the verification posture. No open decision gates the round.)*

*(No open decision gates the round; the acceptance journey is delivered and the catalog→tool wiring seam is pinned above.)*
