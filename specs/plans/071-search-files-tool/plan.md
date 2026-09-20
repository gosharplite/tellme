# System Analysis — `search_files` (round 071)

**Plan Package**: `specs/plans/071-search-files-tool`

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The single-prompt reasoning turn + its agent tool surface | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY/ADD** — a new executable journey (`chat/searching-file-contents.feature`) + `chat/dsl.md` rows |
| API surface | — | `/axb-api-plan` | **NOOP** — no OpenAPI/HTTP surface (a single CLI end) |
| Data surface | — | `/axb-data-plan` | **NOOP** — no persisted-record change (the tool result is in-flight only) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (no TUI screen change) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The tool adapter + the assembler + the schemas | `/axb-technical-research` (techstack) + `/axb-dsl-refine` (CLI truth) | one new package `internal/infrastructure/tools/search.go`; the assembler (`cmd/tellme/deps.go`) offers it; the shared `resourceSchema` backs it |
| **W2** | The recordable union + the offline report | `/axb-technical-research` (techstack, *Tool-usage accounting*) | `recordableToolNames()` gains `search_files` |
| **W3** | The executable CLI contract | `/axb-dsl-refine` | `specs/truth/features/cli/chat/searching-file-contents.feature` (ADD) + `chat/dsl.md` rows |
| **W4** | The plan-side acceptance | `/axb-spec-by-example` | `features/acceptance/searching-file-contents.feature` |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is **no** API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The new behaviour is user-visible (the model is offered a new tool and its result is fed back), so `/axb-spec-by-example` is **NOT** NOOP this round (unlike the 042/043/047/049/069/070 structural rounds).

## 4. Unchanged surfaces (invariants)

- The reason gate (ADR 0025), the round-024 resource contract, the per-call timeout, and the observer hooks are unchanged (`spec.md` I-1/I-2).
- No config key, no CLI flag, no chrome/`stdout` shape change (`spec.md` I-8).
- No new dependency; stdlib-only; POSIX-only; hermetic.
