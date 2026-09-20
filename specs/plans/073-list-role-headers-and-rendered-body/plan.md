# System Analysis — the `-l` listing presentation (round 073)

**Plan Package**: `specs/plans/073-list-role-headers-and-rendered-body`

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The offline `-l`/`--list` history listing | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the listing Example(s) gain the role header, the rendered model body, the verbatim operator body, the blank separator, the terminal-gated accents, and the `-r` raw form |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — the persisted `history.jsonl` shape is unchanged (only its *presentation* changes; I-2) |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (no screen change; the listing is text on `stdout`) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The presentation port + adapter | the implementation | `internal/domain/render/ports.go` (`Listing`/`ListingMessage`/`ListingRole`/`ListingSpec`) + `internal/ui/listing.go` (`NewListing`) + the two colour codes in `internal/ui/colour.go` |
| **W2** | The CLI wiring | the implementation | `internal/cli/cli.go` — `renderHistoryList` assembles `[]render.ListingMessage`, resolves the width best-effort, and calls the port; `runtimeEnv` gains the `stdout` probe (+ `stdoutTerminalDetector`, `TELL_ME_FORCE_STDOUT_TTY`); the `-r` flag reaches the listing; `cmd/tellme/deps.go` binds `NewListing` |
| **W3** | The truth rows | `/axb-technical-research` (done) | `specs/truth/techstack.md` ×4 MODIFY + `docs/domain-model` MODIFY (`research.md` D1–D9) |
| **W4** | The executable CLI contract | `/axb-dsl-refine` | a `history` listing journey (header / rendered body / verbatim prompt / separator / accents / raw) |
| **W5** | The plan-side acceptance | `/axb-spec-by-example` (done) | `features/acceptance/listing-the-session-as-a-conversation.feature` |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is **user-visible** (the listing's bytes on `stdout`), so `/axb-spec-by-example` is **NOT** NOOP — its journeys were authored at specify time.

## 4. Unchanged surfaces (invariants)

- The listing stays strictly offline (no provider request) and terminal; the dispatch precedence (`-d` → `-l` → `-t` → `--tool-usage`) is unchanged (`spec.md` I-1).
- The persisted `history.jsonl` / `history.archive.jsonl` are untouched — the role naming (`[MODEL]`) and the rendering are presentation-only (`spec.md` I-2).
- Colour is `stdout`-gated and never reaches `stderr` or `turns.log` (`spec.md` I-3).
- No new dependency (glamour is already direct); no network; the `-l N` count semantics are unchanged (`spec.md` I-4/I-5).
- Tool activity stays omitted (clarify Q2 → A); the round-007 operator-only contract is preserved.
