# System Analysis Plan — Round 068: surface a Gemini/Vertex round's unpaired tool calls as a diagnostic

**Plan package**: `specs/plans/068-unpaired-call-diagnostic`
**Owner**: `/axb-system-analysis`
**Date**: 2026-09-20
**Inputs**: `spec.md` (US1 · FR-001…FR-007 · I-1…I-7 · Q1 → A · Q2 → A) · `research.md` (D1–D7) · `truth-delta.md` · `specs/truth/**`.

## 1. Interface inventory

| Interface | Kind | Analysis planner | Disposition |
| --- | --- | --- | --- |
| The CLI end (the `tellme` binary) | `cli` | none (a line-oriented CLI) | **Carried forward to its contract owner `/axb-dsl-refine`** at delivery. |

**One** CLI interface. No HTTP/API surface and no persisted-state change, so:

- `/axb-api-plan` → **NOOP** (no `specs/truth/contracts/**`).
- `/axb-data-plan` → **NOOP** (no `specs/truth/data/**`; the diagnostic is transient; `turns.log` is a rendered-text file).
- `/axb-ui-plan` → **skipped** (a plain line-oriented CLI; no TUI surface change).

## 2. Waves

**No dependency waves.** The round is one coherent edit across four tiers (domain account → ui formatter/render port → cli decorator → wiring). Nothing to sequence or fan out.

## 3. Contract owner hand-off — `/axb-dsl-refine`

The CLI end is carried to **`/axb-dsl-refine`**. Expected disposition: **NOOP (recorded narrowing)**.

- The diagnostic is **user-visible in principle** but has **no hermetic producer** (`agentloop.go` appends one result per call ⇒ `M == N` always), so no godog Example can drive it (research D5; spec A5).
- Authoring a Rule with no Example is forbidden (RF-063-10 retired), so `/axb-spec-by-example` authors **no** feature and records the narrowing; `/axb-dsl-refine` records the same. The carriers are the **domain / ui / cli unit pins** (the round-059 documented-narrowing class).
- The topology audit (`axb-gherkin-and-dsl` over `specs/truth/features/cli`) must keep the **same 5 pre-existing errors**, **none new** (no feature/DSL edit).

## 4. Component / tier view (the changed surface)

| Site | Tier | Change |
| --- | --- | --- |
| `internal/domain/llm/unpaired.go` | domain (tier 1) | **NEW** — `UnpairedToolCalls(msgs)` + `roundPairing`; the family-neutral single owner. |
| `internal/infrastructure/llm/gemini/client.go` | infrastructure (tier 4) | `UnpairedCallIDs` **delegates** to `llm.UnpairedToolCalls`; the `dropped`/`unpaired()` machinery is deleted (`buildRound` returns contents only). |
| `internal/domain/render/ports.go` | domain (tier 1) | `Lines` gains `UnpairedCalls(t, ids) string`. |
| `internal/ui/toolcall.go` + `render_ports.go` | ui (tier 6) | `FormatUnpairedCalls` + `Lines.UnpairedCalls` adapter. |
| `internal/cli/unpaired_gateway.go` | cli (tier 7) | **NEW** — the gateway decorator + `unpairedEmitter`. |
| `internal/cli/cli.go` | cli (tier 7) | `runTurn` wraps the gateway with the diagnostic. |

**Gate impact:** none new — no illegal import edge (`cli → domain/llm` + `render` only, no `cli → ui`; ADR 0020). `verify-architecture`'s baseline stays header-only.

## 5. Artifact tree (plan-side + truth-side)

```
specs/plans/068-unpaired-call-diagnostic/
  spec.md · checklists/requirements.md · research.md · plan.md · tasks.md · truth-delta.md
specs/truth/techstack.md                              MODIFY (1 row: Vertex/Gemini adapter)
docs/decisions/0038-unpaired-call-diagnostic.md       ADD (+ README index; ADR 0037 annotated)
```

## 6. Task directives (handed to `/axb-tasks`)

1. **Phase 3 (RED first)** — land the pins before the code: the domain table test; the ui `FormatUnpairedCalls` pin; the cli decorator/emitter pins.
2. **Phase 4 (GREEN)** — the domain `UnpairedToolCalls`; the adapter delegation + dead-code removal; the render port + ui formatter; the cli decorator + `runTurn` wiring.
3. **Regression** — the round-065/066/067 pins + the E2E journeys stay green; `stdout` byte-exact; the OpenAI/Gemini wires unchanged; `make verify` + `go test -count=1 ./...`; topology audit unchanged; `go.mod`/`go.sum` unchanged.
4. **`[BDD-REMOVE]` applicant** — none.
