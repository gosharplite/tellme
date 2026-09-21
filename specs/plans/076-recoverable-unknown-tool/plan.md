# System Analysis — recoverable unknown tool name (round 076)

**Plan Package**: `specs/plans/076-recoverable-unknown-tool`
**Anchor**: [#154](https://github.com/gosharplite/tellme/issues/154)

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The agent tool loop / its failure surface | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the `chat` module gains an unknown-tool-name Rule/Example + a Given/Then pair; `failing-the-tool-loop.feature`'s unavailable-tool Example relocates |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — no persisted-state change (the per-turn counter is runtime-only; an unknown call records nothing) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (no screen change) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The loop behaviour | the implementation | `internal/agent/agentloop.go` — an unknown registry name folds back a `tool`-role result (`unknownToolResult`) + `continue`, bounded per turn by `maxUnknownToolFolds = 3` |
| **W2** | The truth rows | `/axb-technical-research` (done) | `specs/truth/techstack.md` *Agent tool loop* row MODIFY (+ the *Tool-usage accounting* parenthetical); **ADR 0048** + index |
| **W3** | The executable CLI contract | `/axb-dsl-refine` | a new `chat` feature + a Given/Then pair; the relocated Example |
| **W4** | The plan-side acceptance | `/axb-spec-by-example` (done) | `features/acceptance/recovering-from-an-unknown-tool-name.feature` |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is **user-visible** (an unknown tool name no longer aborts the turn; the cap keeps the frozen phrase + exit 7), so `/axb-spec-by-example` is **NOT** NOOP.

## 4. Unchanged surfaces (invariants)

- A real tool's call-time error stays a recoverable `error: …` fold-back; `no tools are registered` is unchanged (`spec.md` I-4, I-1).
- Every requested call gets a paired `tool` result (never a skip) — the Gemini/Vertex pairing (round-065 / #132) is preserved (`spec.md` I-1).
- An unknown call executes nothing and records nothing — no `history.Step`/`Signature`, no tool-usage record (`spec.md` I-2).
- Family-local: provider adapters/wire envelopes, the registry, the reason gate (ADR 0025 D3) and the loop clamp are untouched (zero wire diff) (`spec.md` NFR-002).
- stdlib-only, POSIX-only, hermetic; no new dependency (`spec.md` NFR-004).

## 5. Domain model (ADR 0041)

**Not modelled, and that is recorded here** (the same-PR rule's escape hatch): the round changes an **error-handling classification** in the loop — how the loop responds to a tool name it does not provide. It introduces no modelled entity, attribute, relationship, invariant, or scenario (`ToolCall`'s shape and the `turn-bounded-by-tool-loop` invariant are unchanged; an unknown call records no `outcome`). `docs/domain-model/**` is **unchanged** and `modelith-check` stays green. (`techstack.md`'s *Agent tool loop* row is the truth home for the loop's failure behaviour.)
