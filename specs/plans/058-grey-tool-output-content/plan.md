# System Analysis Plan — round 058 (`058-grey-tool-output-content`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/058-grey-tool-output-content/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md                        # ✓ (/axb-specify) — clarify Q1 opens (A1 non-blocking)
├── research.md                    # ✓ (/axb-technical-research)
├── truth-delta.md                 # ✓ owner rows
├── checklists/requirements.md     # ✓
├── features/acceptance/           # ✓ (/axb-spec-by-example) — colouring-every-output-line.feature
└── tasks.md                       # /axb-tasks

specs/truth/
├── techstack.md                   # MODIFY (Turn chrome + Agent tool loop) ✓
└── features/cli/chat/**           # MODIFY (colouring-the-session-chrome + chat/dsl.md) ✓

docs/decisions/
├── 0028-grey-tool-output-block.md # ADD ✓
└── README.md                      # index row ✓
```

*(No `contracts/**`/`data/**` — `/axb-api-plan` + `/axb-data-plan` **NOOP**. No `ui/**`.)*

### Repository structure (root) — expected changes

```text
internal/ui/tooloutput.go      # CHANGED — formatToolOutputLineColour; WriteWith uses it (w.Colour)
internal/ui/colour_test.go     # CHANGED — the content line is grey (inverts round 057's plain assertion)
tests/e2e/steps/step_r057_payload_and_colour.go  # CHANGED — thenToolOutputGrey also requires content lines grey
specs/truth/features/cli/chat/**  # MODIFY ✓
specs/truth/techstack.md       # MODIFY ✓
docs/decisions/0028-*.md       # NEW ✓
go.mod / go.sum                # unchanged
```

**Structure Decision**: round 058 refines the **existing CLI end** (`cli`) — one formatter on the `[Tool Output]`
block. No new boundary; truth changes are `specs/truth/techstack.md` (two rows) + the CLI interface truth
(`/axb-dsl-refine`) + the governance **ADR 0028**.

---

## Analysis Plan

### System interface inventory

**1** system interface — the **CLI end** (`cli`), the same one rounds 017/034/038/057 already own: the
`[Tool Output]` block on `stderr`. The round **changes existing behaviour** on it (which lines are grey), so it is a
**CLI contract change**, not a new boundary.

> **Scope notes**: `/axb-api-plan` = **NOOP** · `/axb-data-plan` = **NOOP** · `/axb-ui-plan` = **skipped**.

### Analysis Wave schedule

**1 wave**:

| Wave | Interface | Kind | Carried by |
| --- | --- | --- | --- |
| **W1** | The CLI end — the whole `[Tool Output]` block grey | `cli` | **`/axb-dsl-refine`** (the CLI contract owner) |

### Delegation order

1. `/axb-api-plan` — **NOOP**.
2. `/axb-data-plan` — **NOOP**.
3. `/axb-dsl-refine` — **delegated**: extend the existing `the tool output frame is shown in grey` Rule/`dsl.md` row to cover the content lines (in place, semantics unchanged except scope).

*Handoff payload*: plan package `specs/plans/058-grey-tool-output-content`; truth root `specs/truth`;
truth-delta `…/truth-delta.md`; interface: **CLI end** (`chat`); focus: (1) the header + every content line + both
separators are grey; (2) the content text is unchanged (sanitized) and the grey is the line's only escape;
(3) the plain/file legs unchanged. *(Done in this session.)*

### The invariants this round introduces (normative: ADR 0028)

| # | Invariant | Carrier |
| --- | --- | --- |
| **I-1** | Only the content line's **colour** changes; its text is byte-identical | `research.md` D3; ADR 0028 D3 |
| **I-2** | Colour is terminal-gated and never enters `stdout`/`turns.log` | round-054 gate; `research.md` D5/D6; ADR 0028 D3 |
| **I-3** | The wrap is **outside** the sanitizer (the grey pair is the line's only escape) and the neutral close is unchanged | `research.md` D2/D4; ADR 0028 D2 |

### Gating blockers

*(Q1 — the trailing partial line — is **non-blocking**: A1 keeps it dropped.)*
