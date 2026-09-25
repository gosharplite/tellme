# System analysis — round 088 `088-mcp-cache-header-routing-and-mode-location`

**Owner**: `axb-system-analysis` (produces this `plan.md`; writes no `specs/truth/**`).

## 1. Interface inventory

One system interface — the **CLI end** (line-oriented). No HTTP/REST API, no web UI. The observable
CLI surface is unchanged in shape (no new flag, no new output line, no new exit code); the round
repairs an internal MCP-call path and moves an on-disk artifact.

| # | Interface | Kind | Description |
| --- | --- | --- | --- |
| I-1 | tellme CLI | `cli` | The prompt-path MCP prelude + a cached tool's call path; the offered set/order is unchanged. |

## 2. Planner delegation (CLI-streamlined)

| Planner | Applicability | Disposition |
| --- | --- | --- |
| `/axb-api-plan` | none | **NOOP** (`specs/truth/contracts/**` untouched). |
| `/axb-data-plan` | **conditional** — the round moves local persisted state | **INVOKED (light)** — `data-model.dbml`: the cache entry's location → `output/<mode>/`. |
| `/axb-ui-plan` | none | **Skipped** (`ui/**` untouched). |
| `/axb-dsl-refine` | the CLI end's contract owner | **INVOKED** — `specs/truth/features/cli/**` + `dsl.md`. |

## 3. Dependency waves

| Wave | Work |
| --- | --- |
| W1 | `/axb-technical-research` updates `specs/truth/techstack.md` + **ADR 0059**. |
| W2 (after W1) | `/axb-data-plan` records the cache location in `specs/truth/data/**`. |
| W3 (after W1) | `/axb-dsl-refine` writes the executable CLI truth (`specs/truth/features/cli/chat/**` + `dsl.md`). |

The CLI end is carried forward to `/axb-dsl-refine` (no API/data/UI planner for it).

## 4. Handoff

- **Truth root**: `specs/truth/`
- **Truth-delta**: `specs/plans/088-mcp-cache-header-routing-and-mode-location/truth-delta.md`
- **Truth owners**: `axb-technical-research` (`techstack.md`) · `axb-data-plan` (`data/**`) ·
  `axb-dsl-refine` (`features/cli/**`).

## 5. Domain model (ADR 0041 — load-bearing)

**Not modelled.** The modelled `MCPServer`/`MCPTool` entities, their invariants, and the offered
surface are unchanged; the round repairs an internal call path and moves an internal cache file. No
modelled behaviour changes; `docs/domain-model/**` is not updated. `make modelith-check` must stay
green.

## 6. Records

- **ADR 0059** (+ index; ADR 0058 back-pointer) · `specs/truth/techstack.md` (MCP rows) ·
  `specs/truth/data/data-model.dbml` · `specs/truth/features/cli/chat/{…feature,dsl.md}`.
