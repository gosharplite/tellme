# Round 087 — `087-mcp-tool-cache`

**Theme**: the prompt prelude's remote-MCP tool discovery gains a **cross-invocation
cache** — a warm, fresh entry makes the prelude **dial no MCP server** and offer the
cached per-tool declarations, while a cold/corrupt/stale entry degrades to the existing
bounded discovery (and a stale entry serves cached, then refreshes **off the critical
path**). The offered set/order is **identical** to live discovery; the call-time error
contract is unchanged.

**Anchor**: live issue **#180** (DoD = close it) — *MCP discovery in the prompt prelude:
add a cross-invocation tool cache so the common path makes no network call*.
The remedy is the recorded forward item `specs/truth/techstack.md` → *Not Introduced Yet*
→ *Cross-invocation MCP tool caching* (round-032 Q2 Option 2).

---

## 1. Why this round

Every **prompt-bearing** `tellme` invocation dials **every enabled remote MCP server**
(`initialize` + `tools/list`) before the first provider request, purely to learn which
tools to offer the model (round-032 discovery; `internal/cli` → `augmentRegistryWithMCP`
→ `deps.MCPDiscoverer`). With servers declared (this workspace's four configs each carry
`github: {URL: https://api.githubcopilot.com/mcp/, …}`), that is an **unconditional,
per-invocation, avoidable** network round-trip — even for a prompt that will never use a
tool, and even for `tellme --new hi`.

Round 032 fixed the *stall* (a bounded ≤ **3 s** concurrent fast-fail), not the
*unconditional dial*. Round 087 makes the steady-state prelude **local**: the tool
declarations are a **pure function of the server's tool list** (`mcp_<server>_<tool>`
names via `mcp.NamespacedName`, schemas via `NormalizeMCPSchema`, sorted server keys), so
a cached copy is byte-reproducible across runs and can be offered **without contacting the
server**.

This is a **fresh round** against the live surface — never a re-open of the frozen
round-032 package (`plan-package-frozen` / `fresh-package-per-round`).

## 2. The change

1. **A cache store** — a JSON file `$TELL_ME_HOME/mcp-toolcache.json`
   (`{ "<server-key>": { url, auth, fetched_at, tools:[{name, description, input_schema}] } }`),
   written with the round-084 durability discipline (same-directory temp file + `fsync` +
   atomic `rename`); read best-effort (absent/corrupt ⇒ every key cold).
2. **The prelude resolution** (`internal/infrastructure/mcp` `DiscoverCached`): per enabled
   remote server, an entry whose stored **declaration** (`url` + `auth`) matches the live
   config is reused; a **fresh** entry (`now − fetched_at < TTL`) is served with **no dial**;
   a **stale** entry is served with no pre-request dial and marked for a **post-answer
   refresh**; a **cold** key (absent/mismatch) falls back to the existing bounded, concurrent
   discovery and is written to the cache.
3. **A lazy client** (`internal/infrastructure/mcp` `lazyClient`) — a cached tool binds a
   client that connects only on the **first `CallTool`**, so a cache hit that executes no
   tool makes zero MCP connections (the credential resolution is deferred with it).
4. **The composition root** (`cmd/tellme`) constructs the file cache for the resolved home
   and passes the home into the discoverer seam; the TTL is a named constant
   (`mcpToolCacheTTL = 24h`).
5. **The post-answer refresh** — `internal/cli` runs the returned refresh hook when the turn
   ends (post-answer, best-effort, bounded), so a stale entry is revalidated **off the
   critical path**; a failed refresh keeps the prior cache.
6. **Truth + records** — `specs/truth/techstack.md` (the *MCP tool discovery* row MODIFY +
   a new *MCP tool cache* row; the *Not Introduced Yet → Cross-invocation MCP tool caching*
   item retired), the `chat` MCP feature + `dsl.md` rows, **ADR 0058** (+ index).

## 3. Requirements

### 使用者故事 1 — 我的既有工作階段不再每次都撥打 MCP server (Priority: P1)

As an operator with declared remote MCP servers, I want a repeated prompt to reuse the
already-discovered tool list, so my common path makes no MCP network call.

**驗收情境**

1. **Given** a warm, fresh MCP tool cache for an enabled remote server, **When** the
   operator runs a prompt, **Then** the prelude dials **zero** MCP servers and the request
   still offers that server's per-tool declarations.

**功能需求（FR）**

- **FR-001**: With a **warm, fresh** cache entry for every enabled remote server, the
  prompt prelude MUST make **zero** connections to those servers, and `req.Tools` MUST
  carry their per-tool declarations (namespaced names). [Verification Intent: observable → 驗收情境 1 + the E2E zero-connection Then + the unit pin]
- **FR-002**: A **cold** key (absent entry, a corrupt cache, or an entry whose stored
  `url`/`auth` no longer matches the config) MUST fall back to **exactly one** bounded
  discovery for that server — the existing round-032 path, unchanged — and the result MUST
  be written to the cache. [Verification Intent: observable → the E2E cold-discovery Then]
- **FR-003**: A **stale** entry (`now − fetched_at ≥ TTL`) MUST be **served** (zero
  pre-request connections; the tool is still offered) and refreshed **after** the answer
  (or on the next prompt-bearing run); a **failed** refresh MUST leave the prior entry
  intact. [Verification Intent: observable → the E2E stale-offered-from-cache Then]
- **FR-004**: A **new or edited** server declaration (changed `URL` or `AUTH`) MUST make
  that key cold — one bounded discovery for that key (while an unchanged sibling key stays
  warm). [Verification Intent: unobservable → unit pin]
- **FR-005**: The **offered set and order** from a warm cache MUST be **identical** to live
  discovery (same names, same order: native tools, then MCP tools in sorted server-key
  order, each server's tools in advertised order). [Verification Intent: observable → the E2E offered-name Then + the unit determinism pin]
- **FR-006**: A tool the server has since **dropped/renamed** (a cached declaration whose server no
  longer knows the tool) MUST yield a **recoverable** tool result — the cached (lazy) call forwards
  the name to the server, whose tool-level error is folded back as `error: …` (the round-032 TD1/R3
  contract) and the turn completes, never aborts. (The round-076 unknown-name fold-back fires only
  on a **registry miss** — a name the cache never offered — not for a cached tool the server has
  dropped.) [Verification Intent: observable → the E2E `A remembered tool the server has since dropped fails softly` (a companion guard — the recoverable fold is structural/loop-owned, round-032 TD1/R3)]
- **FR-007**: An MCP server that is **down when the tool is called** (a warm-cache entry
  whose server no longer answers) MUST yield a **recoverable** `error: …` tool result; the
  turn completes. [Verification Intent: observable → the E2E recoverable-error Then]

### 使用者故事 2 — 快取永遠不會讓我的執行變慢或外洩憑證 (Priority: P2)

**功能需求（FR）**

- **FR-008**: The cache MUST be **best-effort**: an unreadable/corrupt/absent file, or a
  write error, MUST degrade to live discovery with **no new failure mode** (no new class
  phrase, no new exit code). [Verification Intent: unobservable → unit pins]
- **FR-009**: The cache MUST NOT store or log any **credential** (a `TOKEN`, a derived
  `Authorization` header, or the service-account material); only the declaration
  (`url`/`auth`) and the server's tool definitions are persisted. [Verification Intent: observable → a file-content Then + unit pin]
- **FR-010**: `--new` MUST NOT clear the cache (it is not session state — it lives at the
  `$TELL_ME_HOME` root, outside `output/<mode>/`). [Verification Intent: observable → the E2E `--new` Then]

### 邊界情況

- **EC-001**: A cache file whose entry's stored declaration mismatches the config MUST be
  treated as **cold** for that key (never offered from a stale declaration).
  [Verification Intent: unobservable → unit pin]
- **EC-002**: A **stale** entry served with a **non-answering** server MUST still offer the
  cached tool (the pre-request path never dials); only the post-answer refresh dials (and
  warns). [Verification Intent: observable → the stale-with-a-never-answering-server Then]
- **EC-003**: An entry for a server **not present** in the current config MUST be ignored
  (it is never offered) but MAY remain on disk. [Verification Intent: unobservable → unit pin]
- **EC-004**: With **no** `MCP_SERVERS` (or all disabled) the cache MUST be neither read for
  effect nor required — behaviour is exactly as before the round.
  [Verification Intent: unobservable → unit pin]

### 關鍵實體

- **`MCPToolCacheEntry`** — one server's cached tool definitions + provenance: the
  declaration (`url`, `auth`), `fetched_at`, and the server's advertised
  `MCPToolDefinition`s (name, description, `input_schema`). Never a credential.

## 4. Invariants

- **I-1 — No observable surface change.** The offered tool set, its order, the naming
  contract, the schema normalization, the reason envelope, and the call-time error contract
  are unchanged (rounds 032/056/061/077 hold).
- **I-2 — Offline stays network-free.** The cache is read/written **only** on the
  prompt-bearing path; `--version`/`-d`/`-l`/`-t`/prompt-less `--new`/boot make no network
  contact (`verify-no-network` unchanged).
- **I-3 — No credential persisted or logged** (round-032 FR-017 stands).
- **I-4 — Cold-key delay stays bounded** by the existing fixed concurrent bound (≈ one
  `mcpDiscoveryBound`) regardless of server count (round-032 FR-010 holds).
- **I-5 — Best-effort.** A cache failure is never a run failure; the vocabulary and exit
  code set stay ten.
- **I-6 — stdlib-only / POSIX-only / hermetic.**

## 5. Scope

**In**: the cache store + its JSON shape; the cache-aware prelude resolution; the lazy
client; the `deps.Discovery` refresh hook + the `MCPDiscoverer` home parameter; the
composition-root wiring + TTL constant; the CLI post-answer refresh; the truth rows
(`techstack.md`, the `chat` MCP feature + `dsl.md`); ADR 0058 (+ index); the unit + E2E
carriers.

**Out**: a re-open of round 032 (frozen); the stdio (`COMMAND`) transport; the MCP `-d`
diagnostic; MEMORY/PLUR; the tool-call envelope/naming/schema/auth/call-time-error
contracts; an explicit refresh flag (recorded as an ADR §Forward item); a detached
refresh helper; `go.mod`/`go.sum`.

## 6. Success criteria

- **SC-001** With a warm fresh cache the prelude makes zero MCP connections and the request
  offers the cached per-tool declaration (FR-001).
- **SC-002** With a missing/corrupt cache the prelude makes exactly one bounded discovery +
  a write (FR-002).
- **SC-003** With a stale cache the prelude serves cached and refreshes after the answer; a
  failed refresh keeps the prior cache (FR-003).
- **SC-004** The offered set/order from a warm cache equals live discovery (FR-005).
- **SC-005** `make check` (verify + test) is green; `go.mod`/`go.sum` unchanged; the offline
  paths stay network-free.

## 7. Assumptions

- **A1** The cache is keyed by the server key and validated by the stored declaration
  (`url` + effective `auth`); a mismatch is cold. The token is never part of the key or the
  entry (FR-009). *(The issue suggests a "hash of the declarations"; the round stores the
  declaration fields and compares — same cold-key semantics, no hash to maintain, and
  directly testable. Recorded in ADR 0058.)*
- **A2** The refresh is **post-answer on the same run** (the one-shot process has no
  long-lived background; a detached helper is out of scope) — the issue's option (a).
- **A3** `TTL = 24 h` (a named constant `mcpToolCacheTTL`); the round adds **no** operator
  config key for it (recorded as an ADR §Forward item).
- **A4** The cache lives at `$TELL_ME_HOME/mcp-toolcache.json`, shared across modes (the
  entry is declaration-keyed, so two modes declaring the same server share it) and **not**
  cleared by `--new` (FR-010).
- **A5** No `NEEDS CLARIFICATION` that changes the story split or the acceptance logic: the
  issue fixes the goals (FR-1…FR-7, NFR-1…NFR-3, W1…W6) and recommends option A; the
  residual mechanism choices (store shape, key validation, ttl, refresh placement) are
  RD-owned and routed to `/axb-technical-research`.
- **A6** This is a **CLI** capability round: `/axb-system-analysis` records the CLI end
  carried to `/axb-dsl-refine`; `/axb-api-plan` NOOP; `/axb-data-plan` is **conditional** —
  the round persists local state (the cache file), so it records that state in the data
  truth (or an explicit NOOP with rationale) — see `plan.md`.
- **A7** `docs/domain-model/**` is **not modelled** by this round (the cache is an internal
  optimization; the modelled `MCPServer`/`MCPTool` entities, their invariants, and the
  offered surface are unchanged) — recorded in `plan.md` §5 (ADR 0041 escape hatch).
