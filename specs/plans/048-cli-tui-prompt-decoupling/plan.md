# System Analysis Plan — round 048 (`048-cli-tui-prompt-decoupling`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/048-cli-tui-prompt-decoupling/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
└── tasks.md                       # produced later by /axb-tasks (NOT this branch)

specs/truth/
└── techstack.md                   # /axb-technical-research — MODIFY (Build & Tooling / Layer-discipline gate;
                                   #   CLI Application / Composition root; Interactive TUI prompt) ✓ done

docs/decisions/
├── 0017-cli-tui-prompt-decoupling.md   # governance — ADD ✓ done
└── README.md                      # index row ✓ done
```

*(No `features/acceptance/**` — `/axb-spec-by-example` **NOOP** (no user-facing business journey).
No `features/cli/**` change — `/axb-dsl-refine` **NOOP** (a `cli`-reader refactor is not the `tellme`
CLI contract). No `contracts/**` change — `/axb-api-plan` **NOOP**. No `data/**` change —
`/axb-data-plan` **NOOP**. No `ui/**` artifact.)*

### Repository structure (root) — expected changes (implementation, a later phase)

```text
internal/domain/tui/                # NEW — the port: `Source` + `Prompter` (stdlib-only; RULE-C-pure)
internal/ui/tuiprompt.go            # NEW — the adapter satisfying domaintui.Prompter (delegates to tuiprompt.Run)
internal/cli/cli.go                 # CHANGED — drop the tuiprompt import; Options{Deps; Prompter}; delete tuiPromptRunner + RunTUIPrompt; route through the port
internal/cli/*_test.go              # CHANGED — re-point the dispatch/submit seam to a fake Prompter (test-only)
cmd/tellme/deps.go                  # CHANGED — wire the adapter into cli.Options
tools/arch/baseline.txt             # CHANGED — drop the `internal/cli -> internal/ui/tui/prompt` line (3 -> 2); generated, never transcribed
docs/decisions/0017-cli-tui-prompt-decoupling.md  # NEW — the de-coupling; index row ✓ done
specs/truth/techstack.md            # MODIFY — the gate row (RULE-E 2) + the composition-root / TUI-prompt rows ✓ done
Makefile                            # unchanged — the gate rides the existing `verify-architecture` target
go.mod / go.sum                     # unchanged — no new dependency
internal/agent , internal/infrastructure , tests/e2e/**   # unchanged — no behaviour change
```

**Structure Decision**: Round 048 is a **behaviour-preserving structural refactor** of the application
tier's wiring — not a runtime-interface change. It inverts the **one** residual edge
`internal/cli → internal/ui/tui/prompt` into an **injected domain port** (`internal/domain/tui.Prompter`,
satisfied by an `internal/ui` adapter, wired at `cmd/tellme`), so `internal/cli` imports only sanctioned
packages and the RULE-E baseline ratchets **3 → 2**. There is **no** new endpoint, **no** persisted
state, **no** CLI behaviour change, **no** new Makefile target, and **no** new dependency. The truth
changes are `specs/truth/techstack.md` + the governance **ADR 0017**.

---

## Analysis Plan

### System interface inventory

This requirement inventories **0** system interfaces. The round re-wires an existing subsystem behind an
inverted port; the `tellme` CLI end's observable behaviour (flags, streams, exit codes, TUI chrome) is
**unchanged**, so there is no interface to delegate or carry forward.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface).
> - `/axb-data-plan` = **`NOOP`** (no persisted/in-runtime state; the refactor moves wiring, not data).
> - `/axb-dsl-refine` = **`NOOP`** (no CLI interface truth change — the refactor is not the `tellme` binary's CLI contract).
> - `/axb-ui-plan` = **skipped** (no new UX surface; the TUI chrome is unchanged).
> - `/axb-spec-by-example` = **NOOP/skipped** (no user-facing business journey — an internal de-coupling is not a business journey).

### Analysis Wave schedule

**No waves.** There is no interface to order or delegate; the round's truth changes (`techstack.md` +
ADR 0017) are RD-side and owned by `/axb-technical-research` (already applied).

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted state).
3. **`/axb-dsl-refine`** — `NOOP` (no CLI interface truth change).

Not delegated:
- `/axb-ui-plan` — **skipped** (no new UX surface).

*Handoff payload (for the next phase)*: plan package `specs/plans/048-cli-tui-prompt-decoupling`;
truth root `specs/truth`; truth-delta `specs/plans/048-cli-tui-prompt-decoupling/truth-delta.md`;
interfaces: **none**; the round's delivery is the **domain `tui.Prompter` port** + the **`internal/ui`
adapter** + the **CLI re-wiring** + the **baseline 3 → 2** + **ADR 0017**.

---

### The port the round introduces (normative: ADR 0017 + the tier table, ADR 0011 D1)

The tier ranking (ADR 0011 **D1**): 0 `domain/**` · 1 `config`/`home` · 2 `app/**` · 3 `infrastructure/**`
· 4 `agent` · 5 `ui`/`ui/tui/**` · 6 `cli`; `cmd/**`, `tests/**`, `tools/**` **exempt**. **RULE-A** forbids
an **upward** import, so only tiers ≥ 5 may import `internal/ui/**` — the adapter must live in a tier-5
package (or the exempt composition root).

| Element | Home | Tier | Rule check |
| --- | --- | --- | --- |
| **Port** `domaintui.Source` + `domaintui.Prompter` | `internal/domain/tui` (NEW) | 0 | RULE-C-pure (stdlib `context`/`io`/`time` only) |
| **Adapter** (satisfies `Prompter`) | `internal/ui` (NEW file) | 5 | imports `internal/ui/tui/prompt` — same tier (RULE-A clean) |
| **Consumer** (wires the suggestion engine; calls the port) | `internal/cli` | 6 | imports `internal/domain/tui` (sanctioned) — the `ui/tui/prompt` import is **gone** |
| **Composition root** (injects the adapter) | `cmd/tellme` | exempt | may import anything |

**RULE-E (ADR 0016).** For a governed application tier (`internal/app/**`, `internal/cli`) an import of an
`internal/**` package is a violation unless it is in the sanctioned set (`internal/domain/**`,
`internal/config`, `internal/home`, `internal/app/**`); **default-deny**; **fail-on-stale allow-list**;
violations **deduped by edge**.

**Baseline: 3 → 2.** The removed edge is `internal/cli -> internal/ui/tui/prompt`; the remaining baselined
residuals are `internal/cli -> internal/agent` and `internal/cli -> internal/ui` (later slices,
[#101](https://github.com/gosharplite/tellme/issues/101)). RULE-A/B/C stay **0**; **0** cycles; **0**
unranked. The baseline is regenerated from the gate (`make verify-architecture-update`), never transcribed.

**F-4 closure.** `cli.Options` = `{ Deps deps.Dependencies; Prompter domaintui.Prompter }` — both fields
of **exported** types; the `tuiPromptRunner` func type + `RunTUIPrompt` field are **deleted**; the in-package
nil-default (which imported the TUI package) is **removed**; `cmd/tellme` wires the adapter.

**Enumeration / mechanism (inherited, not re-derived).** The gate runs `go list` anchored to the **module
root**, evaluates the **`CROSS_TARGETS` union**, governs the **merged production+test** graph, asserts
**0 cycles** on the **production-only** graph, and its self-test asserts the enumeration before any ranking
assertion (ADR 0011 **D4/D5/D8**). **No** new Makefile target — the gate rides `verify-architecture`
(already a member of `make verify`); the `-count=1` invocation stays load-bearing.

---

### Gating blockers

*(none — the operator locked the theme (**R5.2** of [#101](https://github.com/gosharplite/tellme/issues/101),
the TUI-prompt de-coupling) and answered clarify **Q1 → B** (slice), **Q2 → A** (the port lives in
`internal/domain/**`), **Q3 → A** (F-4 folded in) one at a time. No open decision gates the round.)*
