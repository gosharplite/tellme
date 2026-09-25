# System analysis — round 087 `087-mcp-tool-cache`

**Owner**: `axb-system-analysis` (produces this `plan.md`; writes **no** `specs/truth/**`).
**Inputs**: `spec.md`, `research.md`, `truth-delta.md`, the current `specs/truth/**`.

## 1. Interface inventory

The round touches **one** system interface — the **CLI end** (a line-oriented terminal
end: commands, flags, stdin/stdout/stderr, exit codes, on-disk state). There is **no**
HTTP/REST API and **no** web UI. The observable CLI surface is unchanged in shape (no new
flag, no new output line on the answer/chrome paths, no new exit code); the new behaviour
is an **internal** prelude optimization with one on-disk artifact (the cache file).

| # | Interface | Kind | Description |
| --- | --- | --- | --- |
| I-1 | tellme CLI | `cli` | The existing prompt prelude consults/updates the cache; the observable offered tool set is unchanged. |

## 2. Planner delegation (CLI-streamlined)

| Planner | Applicability | Disposition |
| --- | --- | --- |
| `/axb-api-plan` | none — a standalone CLI has no OpenAPI surface | **NOOP** (`specs/truth/contracts/**` untouched). |
| `/axb-data-plan` | **conditional** — the round persists **local state** (the cache file `$TELL_ME_HOME/mcp-toolcache.json`) | **INVOKED (light)** — record the cache file + its shape as local state in `specs/truth/data/**`. |
| `/axb-ui-plan` | none — a plain line-oriented CLI (no TUI surface change) | **Skipped** (`ui/**` untouched). |
| `/axb-dsl-refine` | **the CLI end's contract owner** | **INVOKED** — writes `specs/truth/features/cli/**` + `dsl.md`. |

## 3. Dependency waves

| Wave | Order | Work |
| --- | --- | --- |
| W1 | 1 | `/axb-technical-research` updates `specs/truth/techstack.md` (the MCP rows) + **ADR 0058** (no wave dependency; independent of W2). |
| W2 | 2 (after W1) | `/axb-data-plan` records the cache file in `specs/truth/data/**`. |
| W3 | 3 (after W1) | `/axb-dsl-refine` writes the executable CLI truth (`specs/truth/features/cli/chat/**` + `dsl.md`). |

The **CLI end is carried forward to its contract owner `/axb-dsl-refine`** (no API/data/UI
planner applicable to it) — the round-032 pattern.

## 4. Handoff

- **Truth root**: `specs/truth/`
- **Truth-delta**: `specs/plans/087-mcp-tool-cache/truth-delta.md`
- **Truth owners**: `axb-technical-research` (`techstack.md`) · `axb-data-plan` (`data/**`) ·
  `axb-dsl-refine` (`features/cli/**`).

## 5. Domain model (ADR 0041 — load-bearing)

**Not modelled.** The cache is an **internal optimization**: the modelled `MCPServer` /
`MCPTool` entities, their invariants (`mcp-server-definition-preserved`,
`mcp-call-is-an-envelope`, `mcp-payload-forwarded-alone`), and the offered surface are
**unchanged**. No modelled behaviour changes, so `docs/domain-model/**` is **not** updated
(the ADR 0041 same-PR rule does not engage). `make modelith-check` must stay green (no drift
introduced).

## 6. Records

- **ADR 0058** (`docs/decisions/0058-mcp-tool-cache.md` + index).
- `specs/truth/techstack.md` (MCP rows + the *Not Introduced Yet* item).
- `specs/truth/data/data-model.dbml` (the cache file — local state).
- `specs/truth/features/cli/chat/{using-tools-from-a-remote-mcp-server.feature,dsl.md}`.
