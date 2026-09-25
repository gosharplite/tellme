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
| Witnesses (reproduced then reverted; *figures restated at fold-verification — see §2*) | **W1** every key cold ⇒ the unit pin `dialed [u-shop]` + **4** E2E Examples (2 at `never contacted`, 2 at `the request offered the tool`); **W3** skip the cache write ⇒ `saves=0` (2 pins) + the E2E fails at `remembered the tools`; **W5** return the lazy connect error ⇒ the unit pin reddens (`got boom`) |
| Delivery | branch `087-mcp-tool-cache` → **PR [#181](https://github.com/gosharplite/tellme/pull/181) open** (round-open head `a1154c0`; awaiting a human review/merge; **no Copilot review**) |

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

## 2. Session 75 (cont.) — the `architect` review-fold loop (PR #181)

Dispatched the `architect` peer per `tm-chat-ingroup`: initialized **once** with `SESSION-BOOTSTRAP.md`
(`--new`), then **continuations** (no `--new`). The architect reviewed PR #181 and posted
[`pull/181#issuecomment-5825601267`](https://github.com/gosharplite/tellme/pull/181#issuecomment-5825601267) —
**`APPROVE WITH REQUIRED FOLDS`**, no `[ARCHITECTURAL BLOCKER]`; it reproduced `make verify` / `go test` / race /
topology and a set of mutation probes on an out-of-tree copy of the head.

| Fold | Resolution |
| --- | --- |
| **F-087-1** the lazy client's successful delegation had no carrier | Added the live-server E2E Example `A remembered tool against a live server is actually run` + `TestLazyClient_DelegatesOnFirstCall`; the mutation (never delegate) now reddens both. |
| **F-087-2** the recorded W1 figures did not reproduce (4 Examples, not 3; 2 pins per W2/W3) | Restated in `tasks.md` §Claim→Witness + §Falsifiability, this summary, and the PR body. |
| **F-087-3** FR-006 had no carrier/ledger row and named the wrong mechanism | Reworded FR-006 to the actual mechanism (a recoverable `error: …` from a dropped/renamed cached tool; the round-076 fold-back is a registry-**miss** case) + added the dropped-tool Example + the ledger row. |
| **F-087-4** the temp-file + `fsync` + `rename` was asserted but undistinguished from an in-place write | Added the `cacheFS` seam + `TestFileToolCache_SaveUsesTempThenRename` + `TestFileToolCache_SaveFailureKeepsPrior`; an in-place `os.WriteFile` now reddens. |
| **F-087-5** the checklist's carrier column over-claimed on three rows | Added `TestDiscoverCached_WriteErrorIsBestEffort` + `TestDiscoverCached_MismatchedSiblingStaysWarm`; the E2E remember Then scans the cache for credentials; corrected the checklist. |
| **F-087-6** the credential-resolution deferral was un-reconciled and the cached path dropped `CredentialWarning` | Recorded the deferred/warn-less resolution in the techstack row + ADR RF-087-9; reconciled `research.md` ↔ `truth-delta.md`. |
| **TD-087-1** a stale entry whose refresh keeps failing re-dials every run | Recorded (ADR §Consequences + RF-087-10 + a tasks narrowing). |
| Nits **N-087-1…6** | FR-005 mutation probabilistic (recorded); memoised failed connect (D4); the EC-003 pin asserts the requirement not preservation; the cached schema is re-normalized; the head figure refreshed; `issue #180` wording. |

Post-fold: `make verify` **OK** · `go test -count=1 ./...` **green** · `make test-race` **no data races** · topology
audit **PASSED** (53 features · 21 root + 464 module rows · 2445 steps). Fold commit `d6def99`.

### 2 (cont.) — fold verification (the `architect` peer) → `FOLDS VERIFIED WITH RESIDUALS` → folded → re-verification

The architect verified the folds at `c92cb45` (read-only, scratch copy): gates green, topology **PASSED** (2445 steps),
every new carrier reproduced by mutation (F-087-1 red, F-087-4 red, F-087-5 red, corrected W1 = 4 Examples, the reworked
EC-003 pin non-vacuous, the N-087-4 re-normalization idempotent). Verdict **`FOLDS VERIFIED WITH RESIDUALS`** with six
residuals — folded in one commit:

- **RES-087-FV-1** — ADR §Context still asserted the superseded FR-006 mechanism → reconciled to the recoverable `error: …`.
- **RES-087-FV-2** — the dropped-tool Example is a **companion guard** (the recoverable fold is structural/loop-owned; no in-round mutation reddens it) → recorded as such in `spec.md` + `tasks.md`.
- **RES-087-FV-3** — the N-087-4 "untrusted on both axes" truth clause had no carrier → added `TestDiscoverCached_CachedNameAndSchemaReValidated`.
- **RES-087-FV-4** — PR body `+5 Rules` → the measured **4 Rules / 8 Examples**.
- **RES-087-FV-5** — checklist FR-002 named non-existent steps → corrected.
- **RES-087-FV-6** — head-figure vintage drift → refreshed.
- Nits: day-summary §2 wording.

### 2 (cont.) — review-fold loop CLOSED (PR #181 ready for a human to merge)

The architect verified the residual folds at `4d32d2c` → **`FOLDS VERIFIED WITH RESIDUALS`** (all six verified closed;
one figure residual **RES-087-FV-7** + nits). RES-087-FV-7 (W1 = **3** unit pins, not 1) + **N-087-11** (the round-open
head qualifier) were folded at `0844540`; the final verification returned:

> **`FOLDS VERIFIED — LOOP CLOSED (no residuals)`** — [`pull/181#issuecomment-5825845663`](https://github.com/gosharplite/tellme/pull/181#issuecomment-5825845663).

Loop history: review @ `0986c8e` → `APPROVE WITH REQUIRED FOLDS` (F-087-1…6 + TD-087-1 + N-087-1…6) → fold-verification @
`c92cb45` → `WITH RESIDUALS` (RES-087-FV-1…6) → residual-verification @ `4d32d2c` → `WITH RESIDUALS` (RES-087-FV-7 + nits)
→ **final @ `0844540` → LOOP CLOSED**. A final cosmetic fold tidied the W1 carrier cell (`0844540` → this step).

**State**: **PR [#181](https://github.com/gosharplite/tellme/pull/181) is ready for a human review and merge** — no Copilot
review; only a human merges (an agent never merges or pushes to `dev`/`main` except the closeout propagation). On merge:
propagate `dev → main` (no-ff, tagged **`round-087`**) and **close issue #180**.

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
