# Technical research — round 087 `087-mcp-tool-cache`

**Owner**: `specs/truth/techstack.md` (the MCP rows + the *Not Introduced Yet* item).
**Inputs**: `spec.md`, the anchor issue **#180**, the frozen round-032 package
(`specs/plans/032-mcp-client/research.md` D3, Q2 Option 2), the live code
(`internal/infrastructure/mcp/{discovery,client,naming,tool}.go`, `internal/cli/cli.go`
`augmentRegistryWithMCP`, `cmd/tellme/deps.go`), and the round-084 durability precedent
(ADR 0056).

## Problem restatement (verified against `dev` `2162729`)

`internal/cli` → `augmentRegistryWithMCP` → `deps.MCPDiscoverer(ctx, servers)`
(`cmd/tellme/deps.go`) → `mcp.Discover` dials **every enabled remote server**
(`initialize` + `tools/list`) **before the loop is built**, because the loop derives
`req.Tools` from the registry at round 0. The dial is **bounded + concurrent** (one
`context.WithTimeout(parent, 3s)` per server, then `wg.Wait()`), and a failure is a
warn+skip — so it never stalls, but it is **unconditional** on the prompt path.

## Decisions

### D1 — Option A: a cross-invocation cache (the issue's recommended option; the recorded forward item)

`specs/truth/techstack.md` → *Not Introduced Yet* already records *"caching a tool list
across invocations (load instantly, refresh in the background)… (Q2 Option 2)"*. The
declarations are a **pure function of the server's tool list** (`mcp.NamespacedName` names,
`mcp.NormalizeMCPSchema` + the ADR 0031 projection, sorted server keys), so a cached copy is
byte-reproducible and can be offered without contacting the server. Alternatives **B**
(static config declarations), **C** (a per-server proxy tool), and **D** (remove discovery)
are **rejected and recorded** (ADR 0058 §Alternatives) — B moves schema maintenance onto
the operator; C/D change the offered surface and worsen the failure mode.

### D2 — the store: one JSON file under `$TELL_ME_HOME`, declaration-keyed

`$TELL_ME_HOME/mcp-toolcache.json`:

```json
{ "<server-key>": { "url": "…", "auth": "auto", "fetched_at": "RFC3339",
                    "tools": [ { "name": "…", "description": "…", "input_schema": { … } } ] } }
```

The entry is valid for a live config iff `entry.url == cfg.URL && entry.auth == cfg.EffectiveAuth()` —
the issue's "key + URL + auth mode" **semantics**, encoded by storing the declaration fields
rather than a hash (no algorithm to maintain; directly testable; the same cold-key
behaviour). The **token is never a key part and never stored** (round-032 FR-017). The file
lives at the home **root** (not under `output/<mode>/`), so it is shared across modes and
**not cleared by `--new`**.

### D3 — the write: the round-084 durability discipline, but a cache (not a guarantee)

Same-directory temp file + `fsync` + atomic `rename` (`internal/infrastructure/mcp/toolcache.go`),
mirroring `history.fileStore.writeRaw` (ADR 0056). Unlike the history rollback, the cache has
**no** durability guarantee to assert: a torn/absent cache is merely a cold key (best-effort),
so the round asserts only *atomicity of the visible file* (a reader never sees a half-written
payload) and records the fsync as a **mechanism**, not a claim.

### D4 — the lazy client: a cache hit must not dial

`mcp.NewRemoteClient` **connects at construction** (`c.Connect`). A cache hit that still built
an eager client would dial — killing FR-001. So a cached tool binds `mcp.lazyClient`: it
resolves the credential and connects **on the first `CallTool`** (bounded by the **call
timeout**, not the discovery bound), and any connect failure folds into the nil-error
`error: …` recoverable result (the round-032 TD1/R3 contract). A cache hit that executes no
tool therefore makes **zero** MCP connections and spawns **no** `gh`.

### D5 — resolution order per run: fresh → serve; stale → serve + refresh; cold → discover

For each enabled remote server key (sorted):
- **declaration matches + fresh** (`now − fetched_at < TTL`) ⇒ serve cached (lazy client), no dial;
- **declaration matches + stale** ⇒ serve cached (lazy client), no **pre-request** dial, mark for refresh;
- **absent / declaration mismatch / corrupt file** ⇒ the existing bounded, concurrent discovery for that key, then write the cache.

TTL = `mcpToolCacheTTL = 24h` (a named constant; no new config key — ADR §Forward). The clock
is the injected `now func() time.Time` seam (hermetic tests).

### D6 — the refresh: post-answer on the same run, best-effort

`deps.Discovery` gains `Refresh func() []string`. `internal/cli` runs it from the deferred
close hook — i.e. **after** the answer is rendered and the turn persisted (`defer closeMCP()`
at `cli.go:864` sits before the loop, so it fires at function exit). The refresh dials the
**stale** keys only (bounded, concurrent), rewrites their entries on success, and **keeps the
prior entry** on failure. It runs on a fresh `context.Background()` + the fixed bound, so a
SIGINT-cancelled turn context cannot break it. A detached background helper is out of scope
(the issue's note: the one-shot process has no long-lived home).

### D7 — the ordering/determinism pin

A warm-cache offer set must equal a live one. Cached tools are rebuilt in the server's stored
advertisement order, and the server keys are iterated **sorted**, so the assembled list is
`native… , mcp_<s1>_…, mcp_<s2>_…` — identical to `mcp.Discover`. A unit pin compares the two
directly.

### D8 — seams preserved

The SDK stays confined (`internal/infrastructure/mcp/**` only) — the file cache and the lazy
client are in that package; `internal/cli` sees only `deps.Discovery`. The discoverer seam
gains the resolved home: `MCPDiscoverer(ctx, home string, servers) Discovery` (the CLI already
holds `res.Home`). No new domain port is *required*; the round adds the domain-typed
`tools.MCPToolCache` seam so the orchestration (`DiscoverCached`) is unit-testable with a fake
(and the file adapter satisfies it).

### D9 — what is NOT modelled / NOT changed

`docs/domain-model/**` is **not** updated: the modelled `MCPServer`/`MCPTool` entities and the
offered surface are unchanged; the cache is an internal optimization (ADR 0041 escape hatch —
recorded in `plan.md` §5). `specs/truth/contracts/**` NOOP (a pure CLI end).
`specs/truth/data/**` gains the cache file as local state (`/axb-data-plan` conditional).

## Techstack rows (owner: `axb-technical-research`)

| Action | Row | Summary |
| --- | --- | --- |
| MODIFY | *MCP tool discovery (non-stall)* | The prelude consults the cross-invocation cache first; a warm fresh hit dials nothing; a cold/stale key degrades to the existing bounded discovery. |
| ADD | *MCP tool cache* | The file, its declaration-keyed validation, `24h` TTL, the durability write, the lazy client, the post-answer refresh, best-effort semantics. |
| MODIFY | *MCP client protocol library* / *MCP credential resolution* | Credential resolution is deferred to call time on a cache hit (the lazy client). |
| MODIFY | *Not Introduced Yet* | Retire *Cross-invocation MCP tool caching*. |
| NOOP | all other rows | No other technology change; stdlib-only. |

## Recorded choice (for ADR 0058) — the decision space

| Option | Offered from | Steady-state prelude network | Verdict |
| --- | --- | --- | --- |
| **A — cross-invocation cache** | cached enumeration, refreshed | **none** | **Chosen** |
| B — static config declarations | the config, by hand | none | Rejected — operator maintains schemas; drifts |
| C — per-server proxy tool | one thin tool per server | none | Rejected — round-0 schema-guidance loss; a dead server burns the 300 s tool bound |
| D — remove discovery | guess-the-name | none | Rejected — tools unofferable |
