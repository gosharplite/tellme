# System Analysis Plan — Round 067: Gemini/Vertex tool-call id follow-ups

**Plan package**: `specs/plans/067-toolcall-id-followups`
**Owner**: `/axb-system-analysis`
**Date**: 2026-09-20
**Inputs**: `spec.md` (US1/US2 · FR-001…FR-010 · SC-001…SC-007 · S-1…S-7 · I-1…I-8 · A1…A7) · `research.md` (D1–D8) · `truth-delta.md` · `specs/truth/**`.

## 1. Interface inventory

| Interface | Kind | Analysis planner | Disposition |
| --- | --- | --- | --- |
| The CLI end (the `tellme` binary) | `cli` | none (a line-oriented CLI) | **Carried forward to its contract owner `/axb-dsl-refine`** at delivery (the round-065/066 precedent). |

**One** CLI interface. There is **no** HTTP/API surface and **no** new persisted state, so:

- `/axb-api-plan` → **NOOP** (no `specs/truth/contracts/**`). `contract-authoritative` holds vacuously.
- `/axb-data-plan` → **NOOP** (no `specs/truth/data/**` change). The round adds no persisted field — the wire `id` is derived at build time from `parseResponse`, so `data-model-covers-all-state` holds.
- `/axb-ui-plan` → **skipped** (a plain line-oriented CLI; no TUI surface changes).

## 2. Waves

**No dependency waves.** The round is **adapter-local** (`internal/infrastructure/llm/gemini`): a pure response-decode branch (id provenance) + a round-builder accessor (unpaired calls) + unit pins. There is no shared interface to sequence and nothing to fan out; the work is one coherent edit.

## 3. Contract owner hand-off — `/axb-dsl-refine`

The CLI end is carried to **`/axb-dsl-refine`**. Expected disposition: **NOOP**.

- The round changes no **user-visible** behaviour: the unpaired-call observability is a **code accessor** (`UnpairedCallIDs`), not a `[Tool …]`/`stderr` diagnostic (research S-3/D1), and the provider-id preference has no shipped user surface.
- The round-065/066 interface journeys (`specs/truth/features/cli/chat/calling-several-tools-in-one-round.feature`) are **unchanged**.
- The topology audit (`axb-gherkin-and-dsl` over `specs/truth/features/cli`) must keep the **same 5 pre-existing errors**, **none new** (no feature/DSL edit).

## 4. Component / tier view (the changed surface)

| Site | Tier | Change |
| --- | --- | --- |
| `internal/infrastructure/llm/gemini/client.go` `parseResponse` | infrastructure (tier 4) | read `candidates[0].content.parts[].functionCall.id`; prefer it when non-empty, else the deterministic `call_<n>` (`callID` helper). |
| `internal/infrastructure/llm/gemini/client.go` `roundBuilder` | infrastructure (tier 4) | add `unpaired() []string` (the current round's unpaired call ids, call order); record them in `flush()` (`b.dropped`) at the one drop site; factor the message switch into `consume()`; add the package-level `UnpairedCallIDs(prior)` accessor. `buildContents` keeps its signature + output. |
| `internal/agent/agentloop.go` | agent (tier 5) | **no change** (family-blind; the id it carries is produced by the adapter). |
| `internal/infrastructure/llm/openai/**` | infrastructure (tier 4) | **no change** (I-1: a session is single-family; the OpenAI-compatible wire is untouched). |

**Gate impact:** none — no new import edge (only `encoding/json`/`fmt`, already imported), no `internal/cli` change, so `verify-architecture`'s baseline stays header-only. No `make verify` member is added.

## 5. Artifact tree (plan-side + truth-side)

```
specs/plans/067-toolcall-id-followups/
  spec.md            (axb-specify)
  checklists/requirements.md
  research.md        (axb-technical-research)
  plan.md            (this file — axb-system-analysis)
  tasks.md           (axb-tasks — next)
  truth-delta.md
specs/truth/techstack.md                                 MODIFY (1 row: Vertex/Gemini adapter)
docs/decisions/0037-gemini-toolcall-id-provenance.md     ADD (+ README index; ADR 0036 annotated)
```

## 6. Task directives (handed to `/axb-tasks`)

1. **Phase 3 (RED first)** — land the unit pins *before* the change: (a) a response carrying `functionCall.id` ⇒ the emitted parts carry that id; (b) a response without a provider id ⇒ the deterministic `call_<n>` fallback; (c) an `M < N` round ⇒ `UnpairedCallIDs` returns exactly the unpaired id(s); (d) `M = N` ⇒ `UnpairedCallIDs` empty.
2. **Phase 4 (GREEN)** — implement the `callID` provenance branch in `parseResponse`; implement `unpaired()`/`dropped`/`consume()`/`UnpairedCallIDs`; keep the round-065 batched shape, the FIFO fallback, and the empty/foreign-id omission.
3. **Regression** — the round-065/066 pins + the E2E journeys stay green unchanged; `make verify` + `go test -count=1 ./...`; topology audit unchanged (no feature/DSL edit); `go.mod`/`go.sum` unchanged.
4. **`[BDD-REMOVE]` applicant** — none (no dead stepdef).
