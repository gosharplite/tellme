# System Analysis — the `-l` listing backward turn indices (round 082)

**Plan Package**: `specs/plans/082-listing-backward-turn-indices`

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The offline `-l`/`--list` history listing | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the listing Rule/Example(s) gain the header turn index (`[USER] - N` / `[MODEL] - N`; same index per turn; true distance; whole-label accent) |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — the persisted `history.jsonl` shape is unchanged; the index is derived from the loaded slice (`research.md` D5) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (no screen change; the index is text on `stdout`) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The presentation port + adapter | the implementation | `internal/domain/render/ports.go` (`ListingMessage` gains `TurnIndex int`) + `internal/ui/listing.go` (`header` formats `[USER] - N` / `[MODEL] - N`, whole-label accent, `<= 0` fallback) |
| **W2** | The CLI projection | the implementation | `internal/cli/cli.go` — `listingMessages` computes `len(entries) - i` per entry and stamps it on both messages before truncation |
| **W3** | The truth rows | `/axb-technical-research` (done) | `specs/truth/techstack.md` *Session lifecycle flags* MODIFY + ADR 0054 (amends ADR 0045) |
| **W4** | The executable CLI contract | `/axb-dsl-refine` | a `history` listing header journey (index sequence / same-turn index / true distance / whole-label accent) |
| **W5** | The plan-side acceptance | `/axb-spec-by-example` (done) | `features/acceptance/listing-the-session-with-backward-turn-indices.feature` |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is **user-visible** (the listing's header bytes on `stdout`), so `/axb-spec-by-example` is **NOT** NOOP — its journeys were authored at specify time.

## 4. Unchanged surfaces (invariants)

- The listing stays strictly offline (no provider request, no stdin read); the dispatch precedence (`--help` → `--version` → `-d` → `-l` → `-b` → `-t` → `--tool-usage`) is unchanged (`spec.md` I-5/I-6).
- The `-l N` count/selection semantics are unchanged (last **N messages**); the index is computed on the full entry list before truncation (`spec.md` FR-004, NFR-001).
- The persisted `history.jsonl` / `history.archive.jsonl`, `history.Store`, and `-b`/`--back` are untouched — the index is presentation-only (`spec.md` FR-008).
- The body rendering, the blank separator, and the `-r` behaviour are unchanged; colour is `stdout`-gated and never reaches `stderr` or `turns.log` (`spec.md` I-4/I-7).
- No new dependency (stdlib `fmt`); no network; the exit-code set stays **ten** (`spec.md` NFR-003).

## 5. Domain model

**NOT modelled** (`docs/domain-model/**` unchanged). The listing is a **presentation surface**: the `render.Listing` port and its value types are not domain-model entities, and the model's `Session`/`History` offline-read invariants (the stored history is read, never changed) are unaffected by a header-label change. Recorded here per the ADR 0041 escape hatch (a round that changes modelled behaviour updates the model in the same PR; this round does not change modelled behaviour). `make modelith-check` stays green (no `.yaml`/`.md` change).
