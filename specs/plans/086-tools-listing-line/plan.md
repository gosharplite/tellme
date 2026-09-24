# Plan — round 086 `086-tools-listing-line`

## 1. Interfaces identified

| Interface | Kind | Analysis |
| --- | --- | --- |
| CLI (`tellme -l`) | `cli` | the round touches the **existing** `-l`/`--list` interface only — its presentation gains one line |

Exactly **one** CLI end; no backend/frontend. The CLI-streamlined pipeline applies.

## 2. Wave / planner mapping

| Planner | Verdict | Why |
| --- | --- | --- |
| `/axb-api-plan` | **NOOP** | a pure CLI end — no OpenAPI surface (per the CLI-streamlined rule) |
| `/axb-data-plan` | **NOOP** | no persisted-state change — the line reads `history.Entry.Steps` (round 008; schema unchanged) |
| `/axb-ui-plan` | **skipped** | a plain line-oriented CLI (no TUI surface added) |
| `/axb-dsl-refine` | **owner** | the CLI end is carried forward to its contract owner at delivery — the executable `specs/truth/features/cli/history/**` feature + `dsl.md` rows |

## 3. Layering (RULE-E / RULE-B)

- `internal/domain/render` — the `Listing` port + `ListingMessage` gain `ToolCount int` (a domain
  shape; stdlib + domain types only).
- `internal/ui` — the adapter formats the line (the bytes are single-owned here).
- `internal/cli` — computes the figure (`listingMessages`) and passes it through the injected port,
  naming **no** `internal/ui` identifier (ADR 0020).
- No new edge; the layer baseline is unchanged (`verify-architecture` stays at 0 violations).

## 4. Touch points

| Area | File |
| --- | --- |
| Port | `internal/domain/render/ports.go` (`ListingMessage.ToolCount`; the `Listing` doc) |
| Adapter | `internal/ui/listing.go` (emit the line before the `[MODEL]` header) · `internal/ui/colour.go` (a comment note — the yellow is reused, no new constant) |
| Wiring | `internal/cli/cli.go` (`listingMessages` sets `ToolCount: len(e.Steps)`) |
| Unit pins | `internal/ui/listing_test.go` · `internal/cli/list_render_test.go` |
| E2E pins | `tests/e2e/steps/step_r073_listing.go` (parse + Then) · `tests/e2e/steps/scenario_context.go` + `step_t022_history_given_tool_using_exchange.go` (record the arranged tool count) |
| Truth | `specs/truth/features/cli/history/dsl.md` · `specs/truth/features/cli/history/inspecting-the-session-history.feature` · `specs/truth/techstack.md` |
| Records | `docs/decisions/0057-tools-listing-line.md` + `docs/decisions/README.md` · `docs/domain-model/tellme.modelith.{yaml,md}` |

## 5. Domain model

**Modelled** — the `Chrome` entity's `chrome-colour-terminal-gated` invariant names the listing's
accents (round 073, extended by round 082); round 086 extends that sentence with the `[TOOLS]` line
(the ADR 0041 same-PR rule). `modelith-check` must stay green (`make modelith-render`).

## 6. Witness strategy (per upstream ADR 0006)

Every clause of this round is **externally observable** on the `-l` path, so each is carried by an
executable **`[BDD-GREEN]`** (E2E) or a **unit pin** — **no `[WITNESS]` task** and **no
`accepted-unwitnessed` record** is needed. The Claim→Witness ledger (`tasks.md`) records the mapping;
every claim has a discriminating falsifier (a mutation that reddens its carrier).
