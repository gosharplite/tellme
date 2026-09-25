# Truth delta — round 088 `088-mcp-cache-header-routing-and-mode-location`

Per-owner ledger. Every owner records at least one row (a `noop` proves the area was checked).

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` (*MCP tool cache*) | The cache file moves to the per-mode workspace (`$TELL_ME_HOME/output/<mode>/mcp-toolcache.json`); the lazy client warms the SDK `tools/list` cache on the first call. | Repairs the round-087 defects (issue #182; ADR 0059). |
| MODIFY | `specs/truth/techstack.md` (*MCP tool discovery*) | Note that a cached tool's session must have issued `tools/list` for `x-mcp-header` routing. | The discovery/normalizer owner. |
| NOOP | all other `techstack.md` rows | No other technology change. | stdlib-only. |

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` | Add the round-088 Rule/Example (a remembered header-routed tool is actually run) + the cache-location Example. | The executable interface truth. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Add the `## Given/Then (round 088)` rows. | DSL vocabulary. |
| NOOP | the other `cli/**` modules + the root `cli/dsl.md` | Untouched. | `dsl-single-authority`. |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | No API surface. | CLI-streamlined pipeline. |
| MODIFY | `specs/truth/data/data-model.dbml` | The cache entry's **location** note → `output/<mode>/mcp-toolcache.json`. | `/axb-data-plan` conditional (local state). |
| NOOP | `ui/**` | No UX plan surface. | `/axb-ui-plan` skipped. |

## Domain model (ADR 0041)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `docs/domain-model/**` | **Not modelled** (an internal optimization; the modelled entities/invariants are unchanged). | ADR 0041 escape hatch. |

## Records

| Action | Artifact | Summary |
| --- | --- | --- |
| ADD | `docs/decisions/0059-mcp-cache-header-routing-and-mode-location.md` | The ADR (supersedes ADR 0058 D2 placement; qualifies its D4 lazy-connect). |
| MODIFY | `docs/decisions/README.md` | The ADR 0059 index row (+ a back-pointer on 0058). |
