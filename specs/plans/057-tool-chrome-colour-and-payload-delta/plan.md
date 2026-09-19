# System Analysis Plan — round 057 (`057-tool-chrome-colour-and-payload-delta`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/057-tool-chrome-colour-and-payload-delta/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md                        # ✓ draft (/axb-specify) + clarify CLOSED
├── research.md                    # ✓ done (/axb-technical-research)
├── truth-delta.md                 # ✓ skeleton + owner rows
├── checklists/
│   └── requirements.md            # ✓ ready
├── features/
│   └── acceptance/                # ✓ done (/axb-spec-by-example) — 3 journeys:
│       ├── colouring-the-tool-chrome.feature
│       ├── showing-the-payload-increment.feature
│       └── capping-the-action-argument.feature
└── tasks.md                       # produced by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY (Agent tool loop + Turn chrome) ✓ done
└── features/cli/chat/**           # /axb-dsl-refine — MODIFY ✓ done

docs/decisions/
├── 0027-tool-chrome-colour-and-payload-delta.md   # ADD ✓ done
└── README.md                                      # index row ✓ done
```

*(No `contracts/**` — `/axb-api-plan` **NOOP**. No `data/**` — `/axb-data-plan` **NOOP** (the delta
baseline is in-memory). No `ui/**` — no new screen; existing `stderr` chrome only.)*

### Repository structure (root) — expected changes (implementation)

```text
internal/ui/colour.go          # CHANGED — colorGray/colorYellow + grey()/yellow()/wrap()
internal/ui/toolcall.go        # CHANGED — argValueCap 189 -> 500; formatToolActionColour
internal/ui/toolrenderer.go    # CHANGED — ActionLine applies the yellow accent
internal/ui/tooloutput.go      # CHANGED — ToolOutputWriter.Colour; grey header + separators
internal/ui/coordinator.go     # CHANGED — NewToolOutputCoordinator(+colour)
internal/ui/status.go          # CHANGED — FormatPayloadEstimate + colour variant
internal/ui/render_ports.go    # CHANGED — Lines.PayloadEstimate; NewTurnProgress(+colour)
internal/domain/render/ports.go# CHANGED — Lines.PayloadEstimate; ProgressFactory(+colour)
internal/cli/call_renderer.go  # CHANGED — the in-memory prev-estimate tracker; emit PayloadEstimate
internal/cli/cli.go            # CHANGED — pass colourOn to NewProgress
cmd/tellme/deps.go             # CHANGED — the NewProgress binding signature
tests/e2e/steps/*              # CHANGED/NEW — payload helpers (delta), the 500 cap, the colour steps
specs/truth/features/cli/chat/**  # MODIFY ✓ done
specs/truth/techstack.md       # MODIFY ✓ done
docs/decisions/0027-*.md       # NEW ✓ done
go.mod / go.sum                # unchanged — no dependency change
```

**Structure Decision**: round 057 refines the **existing CLI end** (`cli`) — the `stderr` chrome and
one rendered-value constant. No new boundary; the truth changes are `specs/truth/techstack.md` (two
rows) + the CLI interface truth (`/axb-dsl-refine`) + the governance **ADR 0027**.

---

## Analysis Plan

### System interface inventory

**1** system interface — the **CLI end** (`cli`) that rounds 017/034/039/054 already own: the turn
chrome (`stderr`). The round **changes existing behaviour** on it (two colour accents, one cap, one
line shape), so it is a **CLI contract change**, not a new boundary.

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no API surface).
> - `/axb-data-plan` = **`NOOP`** (the increment baseline is in-memory; no persisted/in-runtime model change).
> - `/axb-dsl-refine` = **MODIFY** (contract owner): the colour feature gains a grey/yellow Rule; the payload-status feature gains the increment Rule and re-scopes the budget Rules to the measured line; the tool-loop feature's cap Example moves to 500.
> - `/axb-ui-plan` = **skipped** (no new screen / keybinding / state transition).

### Analysis Wave schedule

**1 wave** (cli-only):

| Wave | Interface | Kind | Carried by |
| --- | --- | --- | --- |
| **W1** | The CLI end — chrome colour + payload increment + the 500-rune cap | `cli` | **`/axb-dsl-refine`** (the CLI contract owner) |

### Delegation order

1. `/axb-api-plan` — **NOOP**.
2. `/axb-data-plan` — **NOOP**.
3. `/axb-dsl-refine` — **delegated**: write/retarget the executable Rules for (i) the grey `[Tool Output]` frame, (ii) the yellow `[Tool Action]` line, (iii) the estimated-payload increment, and (iv) the 500-rune action cap; update `chat/dsl.md`.

*Handoff payload (for `/axb-dsl-refine`)*: plan package
`specs/plans/057-tool-chrome-colour-and-payload-delta`; truth root `specs/truth`; truth-delta
`…/truth-delta.md`; interface: **CLI end** (`chat`); focus: (1) the header + both separators grey;
(2) the action line yellow; (3) `Payload: +<delta> ~<n>` and no allowance on the estimated line;
(4) the 500-rune argument-value cap. *(Done in this session.)*

### The invariants this round introduces (normative: ADR 0027)

| # | Invariant | Carrier |
| --- | --- | --- |
| **I-1** | Only the operator-stated attribute changes per element; everything else on the line is unchanged | `research.md` D1–D5; ADR 0027 |
| **I-2** | Colour is terminal-gated (`stderr` terminal + `-r` off) and never enters `stdout`/`turns.log` | round-054 gate; `research.md` D1/D7; ADR 0027 D1/D6 |
| **I-3** | The delta is display-only (no persistence; no accounting/budget/exit-code change) | `research.md` D6; ADR 0027 D5/D7 |
| **I-4** | The measured line, the round-024 budget display, and the round-056 gate are unchanged | `research.md` D5/D9; ADR 0027 D7 |

### Gating blockers

*(none — the three clarify questions are LOCKED and the boundary forms are exposed as vetoable
assumptions A7/A8.)*
