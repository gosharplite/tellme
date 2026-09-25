# Truth delta — round 087 `087-mcp-tool-cache`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a
`noop` proves the area was checked).

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/using-tools-from-a-remote-mcp-server.feature` | **ADD** 4 Rules / 6 Examples: *A remembered tool list is reused without contacting the server* (a warm fresh cache ⇒ `tellme never contacted` + the tool still offered; the `--new` variant); *A server's tools are discovered with no memory yet, then remembered* (`tellme contacted` + `tellme remembered`); *An aged tool list is served before the request and refreshed afterwards* (a stale entry + a stopped server ⇒ the tool is still offered + the refresh warns); *A remembered tool that the server can no longer serve fails softly* (a transport-failing call and a stopped server each ⇒ the tool is offered + the run continues). | The executable interface truth for the cross-invocation cache. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | **ADD** a `## Given (round 087)` block (3 rows) + a `## Then (round 087)` block (2 rows) + the round-087 prologue note. | The DSL vocabulary for the new steps (`dsl-exact-one-match`). |
| NOOP | `specs/truth/features/cli/{configuration,diagnostics,history,usage,workspace}/**` and the interface root `cli/dsl.md` | Checked — untouched (their rows/features are unaffected; no new cross-module row is introduced). | Scope guard (`dsl-single-authority`). |

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` (*MCP tool discovery (non-stall)*) | The prelude consults the cross-invocation cache first: a warm fresh entry dials nothing; a cold/stale key degrades to the existing bounded concurrent discovery. | The discovery behaviour owner. |
| ADD | `specs/truth/techstack.md` (*MCP tool cache (cross-invocation)*) | The cache file, its declaration-keyed validation (never the token), the 24 h TTL, the durability write, the lazy client, the post-answer refresh, best-effort semantics. | A new recorded mechanism. |
| MODIFY | `specs/truth/techstack.md` (*Not Introduced Yet*) | Retire *Cross-invocation MCP tool caching* (delivered by round 087 / ADR 0058). | The round delivers the recorded forward item. |
| MODIFY | `specs/truth/techstack.md` (*MCP credential resolution*) | Record the **deferred, warn-less** resolution on a cache hit (the lazy client resolves the credential at first call; the cached path does not surface the `CredentialWarning` — a recorded limitation, ADR §Forward RF-087-8). | Fold F-087-6: the cached path's credential behaviour must be recorded, not left as a silent divergence from the row. |
| NOOP | all other `techstack.md` rows | Checked — no other technology change. | stdlib-only; POSIX-only. |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface (a pure CLI end). | CLI-streamlined pipeline (`/axb-api-plan` NOOP). |
| MODIFY | `specs/truth/data/data-model.dbml` | **ADD** `Table mcp_tool_cache_entry` + a project-note paragraph: the cache file at `$TELL_ME_HOME/mcp-toolcache.json` (server key → declaration + normalized tools; never the token; the freshness/refresh lifecycle; best-effort; not cleared by `--new`). | The round persists local state (`/axb-data-plan` conditional). |
| NOOP | `ui/**` | Checked — a plain line-oriented CLI; no UX plan surface. | `/axb-ui-plan` skipped. |

## Domain model (ADR 0041)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `docs/domain-model/**` | Checked — **not modelled**: the cache is an internal optimization; the modelled `MCPServer`/`MCPTool` entities, their invariants, and the offered surface are unchanged (recorded in `plan.md` §5). | The ADR 0041 same-PR rule does not engage (`modelith-check` stays green). |

## Records

| Action | Artifact | Summary |
| --- | --- | --- |
| ADD | `docs/decisions/0058-mcp-tool-cache.md` | The ADR (context, decisions D1–D11, consequences, alternatives, §Forward RF-087-1…7). |
| MODIFY | `docs/decisions/README.md` | The ADR 0058 index row. |
