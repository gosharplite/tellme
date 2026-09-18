# System Analysis Plan — round 049 (`049-cli-agent-decoupling`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/049-cli-agent-decoupling/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
└── techstack.md                   # /axb-technical-research — MODIFY ×2 (Build & Tooling / Layer-discipline
                                   #   gate: re-cut, baseline unchanged 2; CLI Application / Agent tool loop) ✓ done

docs/decisions/
├── 0018-cli-agent-contracts-extraction.md   # governance — ADD ✓ done
└── README.md                      # index row ✓ done
```

*(No `features/acceptance/**` — `/axb-spec-by-example` **NOOP** (no user-facing business journey).
No `features/cli/**` change — `/axb-dsl-refine` **NOOP** (a `cli`-reader refactor is not the `tellme`
CLI contract). No `contracts/**` change — `/axb-api-plan` **NOOP**. No `data/**` change —
`/axb-data-plan` **NOOP**. No `ui/**` artifact.)*

### Repository structure (root) — expected changes (implementation, a later phase)

```text
internal/domain/agent/              # CHANGED — the extracted contracts: the turn-result type (Result),
                                    #   the incomplete-turn error (ErrIncomplete), the ToolDefs projection
                                    #   (peer of LoopObserver/CallObserver/ToolLineRenderer); stdlib+domain only
internal/agent/agentloop.go         # CHANGED — delete the local AgentResult/ErrIncomplete/ToolDefs;
                                    #   reference the domain contracts directly (no alias)
internal/agent/*_test.go            # CHANGED — repoint to the domain contracts (test-only adaptation)
internal/cli/call_renderer.go       # CHANGED — read the result/ToolDefs from internal/domain/agent
internal/cli/cli.go                 # CHANGED — read the result/ErrIncomplete from internal/domain/agent;
                                    #   the only remaining `agent.` reference is the AgentLoop construction
tools/arch/baseline.txt             # UNCHANGED — byte-identical (still 2); this round moves no edge
docs/decisions/0018-cli-agent-contracts-extraction.md   # NEW — the re-cut sub-slice 1; index row ✓ done
specs/truth/techstack.md            # MODIFY ×2 — the gate row (re-cut; baseline unchanged 2) + the Agent tool loop row ✓ done
Makefile                            # unchanged — the gate rides the existing `verify-architecture` target
go.mod / go.sum                     # unchanged — no new dependency
cmd/tellme , internal/infrastructure , tests/e2e/**   # unchanged — no behaviour change
```

**Structure Decision**: Round 049 is a **behaviour-preserving structural refactor** of an existing
intra-tier contract boundary — not a runtime-interface change. It is the **re-cut sub-slice 1** of the
`internal/cli → internal/agent` de-coupling: it **extracts the loop's crossing contracts** (the turn-result
type, the incomplete-turn error, and the `ToolDefs` projection) into **`internal/domain/agent`**, and makes
`internal/agent` **reference** them directly (no alias). After the round `internal/cli`'s `internal/agent`
references drop **4 → 1** (`agent.AgentLoop` only). **The RULE-E baseline does NOT move** — the `→ agent`
edge persists via the surviving construction; the baseline-moving round is **sub-slice 2** (**2 → 1**).
There is **no** new endpoint, **no** persisted state, **no** CLI behaviour change, **no** new Makefile target,
and **no** new dependency. The truth changes are `specs/truth/techstack.md` + the governance **ADR 0018**.

---

## Analysis Plan

### System interface inventory

This requirement inventories **0** system interfaces. The round relocates an existing intra-tier contract;
the `tellme` CLI end's observable behaviour (flags, streams, exit codes, turn chrome) is **unchanged**, so
there is no interface to delegate or carry forward.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface).
> - `/axb-data-plan` = **`NOOP`** (no persisted/in-runtime state; the refactor moves a contract, not data).
> - `/axb-dsl-refine` = **`NOOP`** (no CLI interface truth change — the refactor is not the `tellme` binary's CLI contract).
> - `/axb-ui-plan` = **skipped** (no UX surface at all).
> - `/axb-spec-by-example` = **NOOP/skipped** (no user-facing business journey — an internal contract relocation is not a business journey).

### Analysis Wave schedule

**No waves.** There is no interface to order or delegate; the round's truth changes (`techstack.md` +
ADR 0018) are RD-side and owned by `/axb-technical-research` (already applied).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — `NOOP` (no CLI interface truth change).

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface).

*Handoff payload (for the next phase)*: plan package `specs/plans/049-cli-agent-decoupling`;
truth root `specs/truth`; truth-delta `specs/plans/049-cli-agent-decoupling/truth-delta.md`;
interfaces: **none**; the round's delivery is the **extracted `internal/domain/agent` contracts** + the
**`internal/agent` reference-and-delete** + the **CLI re-read** (CLI `→ agent` references **4 → 1**) +
**ADR 0018** + the **baseline unchanged (2)** statement.

---

### The extracted contracts (normative: ADR 0018 + the tier table, ADR 0011 D1)

The tier ranking (ADR 0011 **D1**): 0 `domain/**` · 1 `config`/`home` · 2 `app/**` · 3 `infrastructure/**`
· 4 `agent` · 5 `ui`/`ui/tui/**` · 6 `cli`; `cmd/**`, `tests/**`, `tools/**` **exempt**. **RULE-A** forbids
an **upward** import.

| Element | Home | Tier | Rule check |
| --- | --- | --- | --- |
| **Extracted contracts**: the turn-result type (`Result`), the incomplete-turn error (`ErrIncomplete`), the `ToolDefs` projection | `internal/domain/agent` (CHANGED — peer of the existing ports) | 0 | RULE-C-pure (stdlib + `domain/history` + `domain/llm` + `domain/tools` only) |
| **Loop** (deletes its local declarations; references the domain contracts directly — no alias) | `internal/agent` (`AgentLoop`, `Run`) | 4 | imports `internal/domain/agent` — **downward** (RULE-A clean) |
| **Consumer** (reads the contracts from `internal/domain/agent`) | `internal/cli` | 6 | imports `internal/domain/agent` (sanctioned per RULE-E); its only remaining `internal/agent` reference is the `AgentLoop` construction |

**RULE-E (ADR 0016).** For a governed application tier (`internal/app/**`, `internal/cli`) an import of an
`internal/**` package is a violation unless it is in the sanctioned set (`internal/domain/**`,
`internal/config`, `internal/home`, `internal/app/**`); **default-deny**; **fail-on-stale allow-list**;
violations **deduped by edge**. The CLI's new `internal/domain/agent` import is **sanctioned**; the surviving
`internal/cli -> internal/agent` construction import stays **baselined**.

**Baseline: UNCHANGED (still 2).** This round **removes no edge** — the `internal/cli -> internal/agent`
edge persists via the `AgentLoop` construction, so `tools/arch/baseline.txt` is **byte-identical** (0 new /
0 stale). The remaining baselined residuals are `internal/cli -> internal/agent` and
`internal/cli -> internal/ui`; the `→ agent` line is removed by **sub-slice 2** (**2 → 1**,
[#101](https://github.com/gosharplite/tellme/issues/101)).

**Edge-sizing evidence for sub-slice 2.** After the round, `internal/cli` production references **exactly
one** `internal/agent` identifier (`agent.AgentLoop`) — the 9-field construction plus the `Run` call whose
result/error types are already domain-owned. This is the machine-checkable surface sub-slice 2 must invert
(an **identifier count**, not a baseline move).

**Enumeration / mechanism (inherited, not re-derived).** The gate runs `go list` anchored to the **module
root**, evaluates the **`CROSS_TARGETS` union**, governs the **merged production+test** graph, asserts
**0 cycles** on the **production-only** graph, and its self-test asserts the enumeration before any ranking
assertion (ADR 0011 **D4/D5/D8**). **No** new Makefile target — the gate rides `verify-architecture`
(already a member of `make verify`); the `-count=1` invocation stays load-bearing.

---

### Gating blockers

*(none — the operator locked the theme (the re-cut of **R5.3** of [#101](https://github.com/gosharplite/tellme/issues/101),
the `cli → agent` de-coupling) and answered clarify **Q1 → B** (re-cut into sub-slices), **Q2 → (i)**
(extract all three crossing contracts), **Q3 → (a)** (reference + delete, no alias) one at a time. No open
decision gates the round.)*
