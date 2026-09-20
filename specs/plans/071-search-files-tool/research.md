# Research — `search_files` (round 071)

**Plan Package**: `specs/plans/071-search-files-tool`
**Anchor**: issue [#147](https://github.com/gosharplite/tellme/issues/147) (the candidate inventory) — `search_files` is the one Filesystem candidate.
**Method**: compare the reference (`tell-me-go/internal/tools/workspace/{registration.go,search.go}` + `internal/pkg/concurrentsearch/concurrentsearch.go`, `8ca180f4`) against tellme's reader family (`internal/infrastructure/tools/**`) and the round-024 resource contract, then decide the shape. Clarify Q1–Q3 were answered (see `spec.md` §Clarify).

## Decisions

| # | Decision | Rationale |
| --- | --- | --- |
| **D1** | **ADD a ninth agent tool `search_files`**, offered to every provider (gate `none` — not vision-gated). | It is the missing half of the reader trio; the design-intent bar (context-boundedness + determinism) is met. `spec.md` S-1/I-5. |
| **D2** | **Own tool adapter** (`internal/infrastructure/tools/search.go`), a sibling of the reader family; a dedicated `NewSearchTool()` group appended in `assembleAgentTools` **after** `NewFilesystemTools()` (so `newTUIRegistry`'s reader trip let is untouched). | Keeps `NewFilesystemTools()` = the trio (its documented contract) and the `-i` suggestion registry unchanged. |
| **D3** | **Search mode = literal by default, `is_regex` opt-in** (clarify Q1 → A). Literal uses `strings.Contains`; regex compiles `regexp` (RE2). An invalid pattern is a **stable tool error naming it** (never the loop's terminal request failure — the tool-error path). | Predictable default; reference parity; the tool's schema is the acceptance contract. |
| **D4** | **Bound = the round-024 byte budget (primary, every return path) + a degenerate 100-match cap (first 100 by path of the first 1000 walk-order candidates) + a 500-byte per-line trim** (clarify Q3 → A). Matches sorted **path asc, then line asc** (the ordering divergence from the reference's worker-dependent order). | Reuses the `read_files` "cap *and* budget" precedent; determinism is the testability axis. `spec.md` I-2/I-4. |
| **D5** | **Walk = recursive (`os.ReadDir`, sorted by name), binary files skipped, a 10 MB max-line token, NO directory ignore list, unreadable entries skipped best-effort; a file path is searched directly** (clarify Q2 → A). | tellme owns no `WorkspacePolicy` (that is a secret-scan concern); the caller scopes with `path`. `spec.md` S-7. |
| **D6** | **No `SafePath`/consent gate** — the settled no-security-layer direction; the tool reads the path it is given. | `spec.md` I-3; README design intent. |
| **D7** | **Recorded in a new ADR (0043)** with three deliberate divergences (no security gate · no workspace policy · deterministic order) **plus three further recorded differences** (per-line marker · >1 MiB skip · probe magnitude) and the design-intent justification. | Durable decision record; ADR governance (guard `verify-adr-index`). |

## Options considered (the shape choices)

- **Where does the tool live?** (a) the `Tool` port widened for a search mode — **rejected** (no port change needed; `Execute(ctx, args, budget)` already carries everything). (b) a new sibling adapter + a dedicated constructor — **taken** (D2).
- **Walk library?** stdlib `os.ReadDir` recursion — **taken** (no new dependency; the round-024 readers are stdlib-only). A `filepath.WalkDir` alternative was considered; explicit recursion gives the same deterministic order and keeps the ctx/scan-cap checks visible.
- **Ordering?** the reference's worker-pool order — **rejected** (non-deterministic); `sort.SliceStable` by (path, line) — **taken** (D4).

## Risks / mitigations

- **Unbounded walk cost** — a very large tree. Mitigation: the **scan cap** (`searchScanCap = 1000` candidates) stops collection early; the per-call `timeout` (reader default 30 s) yields a nil-error timeout result; the byte budget bounds the result.
- **Pathological line** — a 60 MB minified line. Mitigation: the 10 MB scanner token bound ends that file's scan best-effort (Q2).
- **A literal query containing regex metacharacters** — the literal default keeps its meaning (D3); the regex opt-in is explicit.

## Verification intent (witnesses)

- **Boundedness (SC-004)**: a fixture with far more matches than a small injected budget ⇒ the result stays within the budget and ends with `... (truncated)`.
- **Determinism (I-4)**: a fixture whose matches span a sub-folder and a file ⇒ the result is path-ascending (unit pin + E2E).
- **Mode (Q1)**: the same `.` query matches one line literally but every line as a regex (unit pin).
- **Failure shape (FR-009)**: an empty query and an invalid pattern are **stable tool errors**, not crashes (unit pin + E2E).

## Truth / decisions

- **ADR 0043** (`docs/decisions/0043-search-files-tool.md` + index) — the tool, the design-intent justification, the three divergences, and the Q1–Q3 locked choices.
- `specs/truth/techstack.md` — MODIFY the reader/schema/tool-usage rows; ADD a *file-content search tool* row.
- `specs/truth/features/cli/chat/searching-file-contents.feature` (ADD) + `chat/dsl.md` (rows).
- No `contracts/**`, no `data/**` (single CLI end; no persisted record).
