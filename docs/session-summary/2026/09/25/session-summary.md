# Session Summary — 2026-09-25

**Project**: `tellme` — a disciplined BDD re-creation of `tell-me-go`
**Repo**: `github.com/gosharplite/tellme`
**Status file**: [`STATUS.md`](../../../../../STATUS.md) *(back-link — the single live-state source)*
**Workspace**: `…/beta-tellme/ait-tellme` (`$TELL_ME_HOME`); linux/amd64 host (Go 1.26.6).
**Session mode**: `butler`.
**Branch**: `087-mcp-tool-cache` (off `dev` `2162729`); **PR [#181](https://github.com/gosharplite/tellme/pull/181) open**.
**Status at end of session**: round **087** `087-mcp-tool-cache` **PR OPEN** — the prompt prelude's remote-MCP tool
discovery gains a **cross-invocation cache** so the steady-state prelude makes **no MCP network call**; anchor issue
[#180](https://github.com/gosharplite/tellme/issues/180) (DoD = close it); **ADR 0058**.

---

## 1. Session 75 — bootstrap, read #180, then round 087 `087-mcp-tool-cache` (anchor #180) → full AIxBDD pipeline → **PR [#181](https://github.com/gosharplite/tellme/pull/181) open**

The session began with a bootstrap (`SESSION-BOOTSTRAP.md` Steps 1–8; round 086 delivered/frozen; active branch `dev`,
tree clean at `2162729`) and a flag that the tracker had **1 open issue** (#180, opened 2026-09-24 after STATUS's
`0 open` was written). The operator asked to read #180, then *"Start a new round, the goal is to close #180."* A new
branch **`087-mcp-tool-cache`** was created off `dev`, the full AIxBDD pipeline ran, and **PR [#181](https://github.com/gosharplite/tellme/pull/181)** was opened.

### At a glance

| Area | Outcome |
| --- | --- |
| Branch | **`087-mcp-tool-cache`** (off `dev` `2162729`) |
| Anchor | [#180](https://github.com/gosharplite/tellme/issues/180) — *MCP discovery in the prompt prelude: add a cross-invocation tool cache so the common path makes no network call*; **DoD = close it** |
| Theme | **ADD** a cross-invocation MCP tool cache (**Option A**, the issue's recommendation + round-032 Q2 Option 2): a warm fresh entry dials **nothing**; stale ⇒ served + post-answer refresh; cold ⇒ the existing bounded discovery + an atomic write |
| Clarify | **not escalated (0 questions)** — the issue fixes the goals (FR-1…7, NFR-1…3, W1…W6) and recommends A; the residual mechanism choices are RD-owned |
| Pipeline | specify ✅ · spec-by-example ✅ · technical-research ✅ (**ADR 0058** + the techstack MCP rows + the retired forward item) · system-analysis ✅ (1 CLI end; api NOOP; **data conditional → the cache file**) · dsl-refine ✅ (4 Rules / 6 Examples + the round-087 DSL blocks) · tasks ✅ (T001–T017 + the Claim→Witness ledger) · implement ✅ |
| The change | `internal/domain/tools/toolcache.go` (`MCPToolCache` + entry) · `internal/infrastructure/mcp/{toolcache,lazyclient}.go` (file store; deferred connect) · `discovery.go` (`DiscoverCached` + `discoverKeys`) · `deps.Discovery.Refresh` + `MCPDiscoverer(ctx, home, servers)` · the CLI post-answer refresh hook · `cmd/tellme` wiring + `mcpToolCacheTTL = 24h` |
| Verification | `make verify` **OK** · `go test -count=1 ./...` **green** · `make test-race` **no data races** · `gofmt`/`goimports` clean · `modelith-check` no drift · topology audit **PASSED** (53 features · 21 root + 464 module rows) · `go.mod`/`go.sum` unchanged |
| Witnesses (reproduced then reverted) | **W1** force every key cold ⇒ the unit pin `dialed [u-shop]` + 3 E2E Examples fail at `never contacted`; **W3** skip the cache write ⇒ `saves=0` + the E2E fails at `remembered the tools`; **W5** return the lazy connect error ⇒ the unit pin reddens (`got boom`) |
| Delivery | branch `087-mcp-tool-cache` → **PR [#181](https://github.com/gosharplite/tellme/pull/181) open** (head `a1154c0`; awaiting a human review/merge; **no Copilot review**) |

### Decisions locked (round 087 / ADR 0058)

| # | Decision |
| --- | --- |
| **D1** | **Option A** — a cross-invocation cache; **B** (static declarations), **C** (per-server proxy tool) and **D** (remove discovery) recorded as deliberate rejections. |
| **D2** | The store: `$TELL_ME_HOME/mcp-toolcache.json`, **declaration-keyed** (`url` + effective `auth`; a mismatch is cold); the **token is never stored** (round-032 FR-017). |
| **D3** | The write: a same-directory temp file + `fsync` + atomic `rename` (the ADR 0056 discipline); the `fsync` is a **mechanism**, not an asserted guarantee (a torn cache is a cold key). |
| **D4** | A **lazy client** binds each cached tool and connects (resolving credentials) on the first `CallTool`, so a cache hit that runs no tool dials nothing; a call-time failure folds into the recoverable `error: …` result. |
| **D5** | Resolution per enabled server: **fresh** (`now − fetched_at < 24h`) ⇒ serve; **stale** ⇒ serve + refresh; **cold** ⇒ one bounded concurrent discovery + a write. |
| **D6** | The refresh is **post-answer on the same run** (best-effort; a failed refresh keeps the prior entry); a detached helper is out of scope. |
| **D7/D8/D9** | No credential persisted; no observable surface change (offer set/order == live; contracts unchanged); offline paths stay network-free; best-effort (no new phrase/exit code); stdlib-only. |
| **D10/D11** | Records: ADR 0058 (+ index), `techstack.md`, `data-model.dbml`, the `chat` feature + `dsl.md`; **`docs/domain-model/**` NOT modelled** (an internal optimization; ADR 0041 escape hatch, recorded in `plan.md` §5). |

### Commits (branch `087-mcp-tool-cache`)

| Commit | Note |
| --- | --- |
| `a1154c0` | `feat(087)`: cross-invocation MCP tool cache (ADR 0058) — code + truth + records + unit/E2E pins |

### Open items (non-blocking)

- **PR [#181](https://github.com/gosharplite/tellme/pull/181)** awaits a human review/merge → then the closeout (`SESSION-CLOSEOUT.md`): propagate `dev → main` (no-ff), tag **`round-087`**, refresh the binary; **close [#180](https://github.com/gosharplite/tellme/issues/180)**.
- **ADR 0058 §Forward** RF-087-1…7 (TTL constant · no explicit refresh flag · post-answer-only refresh · whole-file rewrite · no GC of a removed server's entry · the `fsync` is a mechanism · the cold-key dial is the deliberate price of round-0 `req.Tools`).
- Standing (advisory/records, not work): PR #16 **Obs 1**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; the topology audit is **0 errors / 0 warnings** (advisory).

### Next steps

1. Operator may dispatch the `architect` peer for the review-fold loop (the round-069…086 protocol) — **initialize once with `SESSION-BOOTSTRAP.md`, then continuations (no `--new`)** — or hand the PR straight to a human.
2. Human reviews + merges the round-087 PR; then the closeout (propagate `dev → main` no-ff, tag `round-087`, refresh the binary; close #180).
3. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `087-mcp-tool-cache` until merged, then `dev`).

### PM follow-ups

- None new (spec/acceptance complete; the round carries the falsifiable unit + E2E pins).

### Process notes (durable)

- **A cache must not dial to be correct**: `mcp.NewRemoteClient` connects at construction, so an eager client on a
  cache hit would re-introduce the very dial the round removes — hence the lazy client (resolve + connect at call time).
- **A stale-serving claim needs a server that cannot answer**: with a *working* server, "served from cache" and
  "revalidated synchronously" are observationally identical; the round closes the fake so a synchronous revalidation
  would drop the tool — that is the discriminating mutation.
- **A witness can be mechanism-tier**: the lazy client's connect fold is unit-pinned (the E2E Example still passes under
  its mutation because the loop also folds a non-nil tool error back) — recorded as a deliberate narrowing.
- **Lint is a gate, not a nicety**: `DiscoverCached` first failed `cyclop` (CC=24); it was refactored into
  `classifyCache`-style helpers (`cachedTools` / `discoverColdKeys` / `mergeCache` / `makeStaleRefresh`) — tellme has
  **no** `NonFixCatalog`, so a complexity finding is a refactor, not an acceptance record.
