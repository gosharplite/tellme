# System Analysis — Gemini cached-token usage (round 072)

**Plan Package**: `specs/plans/072-gemini-cached-token-usage`

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The single-prompt reasoning turn + its post-turn metrics | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the metrics Example gains the cached/thinking figures on a Gemini turn |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — the persisted `UsageRecord` **shape** is unchanged (only its values become correct) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (no screen change) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The Gemini adapter usage decode | `/axb-technical-research` (techstack) + the implementation | `internal/infrastructure/llm/gemini/client.go` — add the two struct fields + the two `Usage` mappings, floored at 0 |
| **W2** | The accounting truth row | `/axb-technical-research` | `specs/truth/techstack.md` *Gemini/Vertex adapter* + usage/`UsageRecord` rows |
| **W3** | The executable CLI contract | `/axb-dsl-refine` | a `chat` metrics journey (cached/thinking figures) |
| **W4** | The plan-side acceptance | `/axb-spec-by-example` | `features/acceptance/honest-gemini-token-usage.feature` |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is **user-visible** (the metrics line's `H`/`Th` and the reported cost), so `/axb-spec-by-example` is **NOT** NOOP.

## 4. Unchanged surfaces (invariants)

- The request body, the emitted turn, and the OpenAI-compatible wire are byte-identical (`spec.md` I-1).
- The cost formula (`ComputeCost`, `miss = prompt − cached`) is unchanged (`spec.md` I-2).
- Degenerate responses stay safe: a missing field → 0, no negative miss (`spec.md` I-3).
- No new dependency; stdlib-only; POSIX-only; hermetic.
