# ADR 0058 — A cross-invocation MCP tool cache (the prompt prelude dials nothing in the steady state)

- **Status:** Accepted
- **Date:** 2026-09-25
- **Deciders:** tellme owner (issue #180)
- **Related:** [ADR 0006](https://github.com/gosharplite/aixbdd-tmg/blob/main/decisions/0006-claim-witness-obligation.md) (upstream — no normative clause reaches truth unwitnessed), [ADR 0041](0041-domain-model-drift-guard.md) (the load-bearing domain model), [ADR 0012](0012-deterministic-hermetic-tests.md) (determinism/hermeticity), [ADR 0056](0056-rollback-durability-witness.md) (the `fsync`-before-rename durability discipline this borrows), round 032 (`specs/plans/032-mcp-client` — the discovery this caches; its Q2 Option 2), round 044 (**ADR 0013** — the `deps.MCPDiscoverer` seam), round 076 (**ADR 0048** — the recoverable unknown-name fold-back); issue [#180](https://github.com/gosharplite/tellme/issues/180)

## Context

Every **prompt-bearing** tellme run dials **each enabled remote MCP server** (`initialize` +
`tools/list`) before the first provider request, because the loop derives `req.Tools` from the
registry at round 0 (`internal/cli` → `augmentRegistryWithMCP` → `deps.MCPDiscoverer` →
`mcp.Discover`). The dial is **bounded + concurrent** (one `context.WithTimeout(parent, 3 s)`
per server, then a `wg.Wait()`; `mcpDiscoveryBound` in `cmd/tellme/deps.go`) and a failure is a
**warn+skip**, so it never stalls a run (round-032 FR-008/FR-009/FR-010) — but it is an
**unconditional, per-invocation** network round-trip, even for a prompt that will never use a
tool (and even for `tellme --new hi`). Round 032 fixed the *stall*; the *unconditional dial*
remained, and was recorded as the forward item *"Cross-invocation MCP tool caching… (Q2
Option 2)"* (`specs/truth/techstack.md` → *Not Introduced Yet*).

The tool declarations are a **pure function of the server's tool list**: names are derived
deterministically (`mcp.NamespacedName` → `mcp_<server>_<tool>`, hash-truncated for length),
schemas are normalized/verified (`mcp.NormalizeMCPSchema` + the ADR 0031 floor), and servers
register in **sorted key order**. A cached copy is therefore **byte-reproducible** across runs
and can be offered **without contacting the server**. Two facts make this safe:

- a **stale** entry is harmless — a dropped/renamed tool surfaces as a round-076 **recoverable
  unknown-name fold-back**, never an abort;
- a **missing/corrupt** entry falls back to **one bounded discovery** — the existing path,
  unchanged.

Credentials must never persist (round-032 FR-017): the cache key is the server's **declaration**
(key + URL + auth mode), **never** the token.

## Decision

**D1 — choose Option A (a cross-invocation cache); record B/C/D as deliberate rejections.**
Options **B** (operator-authored static declarations in the config) moves schema maintenance
onto the operator and drifts when the server changes; **C** (one thin per-server proxy tool
resolved at call time) loses round-0 schema guidance and makes a dead server burn a round on the
**300 s** tool bound instead of a 3 s probe; **D** (remove discovery) makes the tools
unofferable. All three are **rejected and recorded** (§Alternatives).

**D2 — the store: `$TELL_ME_HOME/mcp-toolcache.json`, declaration-keyed.**
`{ "<server-key>": { url, auth, fetched_at, tools:[{name, description, input_schema}] } }`. An
entry is valid for a live config iff `entry.url == cfg.URL && entry.auth == cfg.EffectiveAuth()`.
The issue's "key + URL + auth mode" is encoded by **storing the declaration fields** rather than
a hash — same cold-key semantics (a new/edited declaration makes that key cold), no algorithm to
maintain, and directly testable. **The token is never a key part and never stored** (D7). The
file lives at the home **root** (not under `output/<mode>/`), so it is shared across modes and
**not cleared by `--new`**.

**D3 — the write: the round-084 durability discipline (temp file + `fsync` + atomic `rename`).**
`internal/infrastructure/mcp/toolcache.go` writes a same-directory temp file, `fsync`s it, then
atomically renames it over the active file (mirroring `history.fileStore.writeRaw`, ADR 0056).
Unlike the history rollback there is **no durability guarantee to assert** — a torn/absent cache
is merely a cold key (best-effort) — so the round asserts only that the **visible file is never
half-written** and records the `fsync` as a **mechanism**, not a claim.

**D4 — the lazy client: a cache hit must not dial.** `mcp.NewRemoteClient` connects at
construction, so a cached tool binds `mcp.lazyClient`, which resolves the credential and connects
**on the first `CallTool`** (bounded by the **call timeout**, not the discovery bound). Any
connect failure becomes the nil-error `error: …` recoverable result (round-032 TD1/R3). A cache
hit that runs no tool makes **zero** MCP connections and spawns **no** `gh`.

**D5 — per-run resolution: fresh ⇒ serve; stale ⇒ serve + refresh; cold ⇒ discover.** For each
enabled remote server key (sorted): a declaration-matching **fresh** entry
(`now − fetched_at < mcpToolCacheTTL = 24 h`) is served with **no dial**; a **stale** entry is
served with **no pre-request dial** and scheduled for refresh; an **absent / mismatched /
corrupt** key runs the existing bounded, concurrent discovery and is written to the cache. The
clock is an injected `now func() time.Time` seam.

**D6 — the refresh is post-answer on the same run, best-effort.** `deps.Discovery` gains
`Refresh func() []string`; `internal/cli` runs it from the deferred close hook (which fires
**after** the answer is rendered and the turn persisted), on a fresh `context.Background()` +
the fixed bound, dialing only the stale keys, rewriting their entries on success and **keeping
the prior entry** on failure. A detached background helper is out of scope (a one-shot process
has no long-lived home).

**D7 — no credential persisted or logged.** The cache stores only `url`/`auth` and the server's
tool definitions; token resolution is deferred to call time (D4). Round-032 FR-017 stands.

**D8 — no observable surface change.** The offered set/order, the naming contract, the schema
normalization/projection, the `{reason, MCP_PAYLOAD}` envelope, and the call-time error contract
are unchanged (rounds 032/056/061/077 hold). `make verify-no-network` is unchanged: the cache is
read/written **only** on the prompt path.

**D9 — best-effort / no new failure mode.** An unreadable/corrupt/absent file or a write error
degrades to live discovery; **no new class phrase, no new exit code** (the set stays ten);
stdlib-only; POSIX-only.

**D10 — records.** ADR 0058 (+ index); `specs/truth/techstack.md` (the *MCP tool discovery* row
MODIFY, a new *MCP tool cache* row ADD, the *Not Introduced Yet → Cross-invocation MCP tool
caching* item retired, and the credential-resolution rows noted); the `chat` CLI feature +
`chat/dsl.md` rows; `docs/domain-model/**` **NOT** updated (D11); `specs/truth/contracts/**`
NOOP; `specs/truth/data/**` records the cache file as local state.

**D11 — not modelled (ADR 0041 escape hatch).** The cache is an internal optimization; the
modelled `MCPServer`/`MCPTool` entities, their invariants, and the offered surface are unchanged.
Recorded in `plan.md` §5.

## Consequences

- In the steady state (a warm, fresh cache) a prompt prelude makes **no** MCP network call — the
  round-180 motivation. The cold-key case keeps the existing bounded ≤ one `mcpDiscoveryBound`
  concurrent delay, unchanged.
- The cache is **transparent**: the offered tools and their order are exactly what a live
  discovery would produce (pinned), so no model-visible behaviour changes.
- A server that has gone silent after its cache was written is discovered only when a tool is
  actually called — the call returns a recoverable `error: …` and the turn completes (the
  round-032 TD1/R3 contract, unchanged).
- Failure modes are **fail-open**: a corrupt/absent cache, a write error, a refresh failure, or a
  declaration mismatch all degrade to the pre-round behaviour; nothing new can abort a run.
- `--new` no longer implies a fresh MCP discovery — the cache is intentionally global, not
  session state; a stale entry refreshes after the answer, so the tool list tracks the server
  within the TTL.
- stdlib-only; POSIX-only; `go.mod`/`go.sum` unchanged.
- **Recorded divergence:** the reference (`tell-me-go`) has **no** cross-invocation MCP tool
  cache — this is a tellme-side capability, not parity (matching the round-032 Q2 deferral).

## Alternatives considered

- **Option B — operator-authored static declarations** (the config decides `req.Tools`): a
  legitimate simplification, but it moves schema maintenance onto the operator and drifts when
  the server changes. **Rejected** (recorded).
- **Option C — one thin per-server proxy tool** (`mcp_<server>` + a generic envelope, resolved
  lazily): zero steady-state dials but loses round-0 **schema guidance** (departs the
  round-032/077 per-tool contract) and makes a dead server burn a round on the **300 s** tool
  bound instead of a 3 s probe — a **worse** failure mode. **Rejected** (recorded).
- **Option D — remove discovery entirely**: tools become unofferable; relies solely on the
  unknown-name fold-back. **Rejected** (recorded).
- **A hash key instead of the stored declaration fields**: equivalent semantics, but adds an
  algorithm to keep stable (and to reimplement in tests). Chosen the stored-fields form (D2).
- **Store the credential / the `Authorization` header** to save the call-time resolution: refused
  — it violates round-032 FR-017 (a credential must never be written to disk).
- **An eager client on a cache hit** (dial to build the session): refused — it re-introduces the
  per-invocation dial the round exists to remove (D4).
- **A detached refresh helper / a background process**: out of scope (the one-shot process model).

## Forward (non-blocking)

> **⚠ Not open work.** A `§Forward` entry is a decision *deferred to a trigger* or a recorded
> divergence — **not** tasking. Do not re-raise absent its trigger.

- **RF-087-1** — TTL is a fixed constant (`24 h`); no operator config key. A per-server or global
  TTL key is a one-line addition if an operator ever wants a different refresh cadence.
- **RF-087-2** — no **explicit refresh flag** ships (the issue's design-point bullet lists one as
  *operator-chosen*); a `--refresh-mcp-tools` style flag is deferred.
- **RF-087-3** — the refresh is **post-answer on the same run**; a refresh on a *later* run
  (a different staleness policy) is not implemented.
- **RF-087-4** — the cache is a **single JSON file** rewritten whole; a large multi-server
  deployment with very large schemas rewrites the whole file per cold key. Fine for tellme's
  surface; not tuned.
- **RF-087-5** — the entry for a server **removed** from the config is ignored but not pruned
  from disk (EC-003); a bounded GC pass is deferred.
- **RF-087-6** — the `fsync` is a **mechanism**, not an asserted guarantee (a torn cache is just
  a cold key); the round does not claim crash-durability of the cache (unlike ADR 0056's
  calibrated rollback clause).
- **RF-087-7** — the **cold-key** dial is the deliberate price of "MCP tools must be in
  `req.Tools` for round 0" (the issue's own wording); eliminating it requires changing the
  offered surface (rejected Option C/D).
