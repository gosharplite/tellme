# Session Summary — 2026-09-22

**Project**: `tellme` — a disciplined BDD re-creation of `tell-me-go`
**Repo**: `github.com/gosharplite/tellme`
**Status file**: [`STATUS.md`](../../../../../STATUS.md) *(back-link — the single live-state source)*
**Workspace**: `…/mbp-johndoe-niffler/ait-tellme` (`$TELL_ME_HOME`); darwin/arm64 host (Go 1.26.6).
**Session mode**: `butler`.
**Branch**: `076-recoverable-unknown-tool` (off `dev`) → **PR [#156](https://github.com/gosharplite/tellme/pull/156)** → human-merged into `dev` (`851466e`), propagated, tagged **`round-076`**.
**Status at end of day**: round **076** `076-recoverable-unknown-tool` **DELIVERED / FROZEN** — an **unknown tool name** becomes a **recoverable, per-turn-bounded fold-back**; **ADR 0048**; anchor issue [#154](https://github.com/gosharplite/tellme/issues/154) (**closes it**).

---

## 1. Session 62 — round 076 `076-recoverable-unknown-tool`: the unknown-tool-name bug (`misc` bootstrap) → two confirmed live issues → #154 spec'd → full pipeline → `architect` review-fold loop (3 passes, CLOSED) → **human-merged (PR #156 → `dev` `851466e`, merge commit)** → branch deleted → closeout (Steps 1–8)

The session began with the operator asking about a live failure: a butler session running the `misc` repo's `SESSION-BOOTSTRAP.md` aborted with `tellme: the tool request failed: tool "get_me" is not available` (**exit 7**) — the model had asked for the GitHub MCP tool by its **bare** upstream name (`get_me`; the callable wire name is `mcp_github_get_me`). We traced it to the loop's **terminal** unknown-name branch, filed **two** issues (the loop fix [#154](https://github.com/gosharplite/tellme/issues/154) + the MCP-name presentation [#155](https://github.com/gosharplite/tellme/issues/155)), and — on the operator's direction — opened **round 076** with **#154 as the anchor (DoD = close it)**. The round ran the full AIxBDD pipeline, took **PR [#156](https://github.com/gosharplite/tellme/pull/156)** through the `architect` peer's review-fold loop to **FOLDS VERIFIED — CLEARED FOR HUMAN MERGE**, then the human-merged it and `SESSION-CLOSEOUT.md` Steps 1–8 ran.

### At a glance

| Area | Outcome |
| --- | --- |
| Theme | anchor [#154](https://github.com/gosharplite/tellme/issues/154): an **unknown tool name** becomes a **recoverable fold-back** (bounded per turn), not a terminal abort |
| Clarify | **not escalated (0 questions)** — the design was locked by the issue (recoverable + per-turn cap); residuals (message, counter home, ordering, ADR shape) → `/axb-technical-research` |
| Pipeline | specify ✅ · spec-by-example ✅ · technical-research ✅ (**ADR 0048** + `techstack.md` ×2) · system-analysis ✅ (1 CLI end; api/data NOOP; domain model **not modelled**) · dsl-refine ✅ (NEW feature + 2 Rules; the moved Example) · tasks ✅ (T001–T010) · implement ✅ |
| Before → after | `tellme: the tool request failed: tool "get_me" is not available` (**exit 7**) → `error: no tool named "get_me"; available tools: …, mcp_github_get_me, …` fed back + **continue**; after `maxUnknownToolFolds = 3` the phrase + exit 7 return |
| The change | `internal/agent/agentloop.go` — on a registry miss: `logAction` + fold back `unknownToolResult(tc.Name, a.Registry.Tools())` (a `tool`-role message carrying `ToolCallID`) + `continue`; a loop-local `unknownFolds` counter capped by `maxUnknownToolFolds = 3`. Family-local: **zero wire diff** |
| Review chain (PR #156, the `architect` peer) | `review` (**APPROVE WITH REQUIRED FOLDS** — no `[ARCHITECTURAL BLOCKER]`; the loop change was right, the folds were **falsifiability**: **F-1** pairing/no-usage unwitnessed · **F-2** the cap Example non-discriminating + value unpinned · **F-3** ordering unwitnessed · **F-4** a false `clampBytes` claim · **F-5** chrome unpinned · **F-6** a stale sibling feature; + TD-076-1…4, N-1…N-5) → fold `19a73bd` → `FOLD-VERIFICATION` (**FOLDS VERIFIED WITH RESIDUALS**; R1/R2/R3) → residual fold `4fb14a1` → `FOLD-VERIFICATION` (**FOLDS VERIFIED — CLEARED FOR HUMAN MERGE**) |
| Merge | PR [#156](https://github.com/gosharplite/tellme/pull/156) **human-merged** into `dev` (`851466e`, **merge commit**); remote branch deleted by the human, then the **local branch deleted** after an ancestor check |
| Closeout | `make verify` **OK** · `go test -count=1 ./...` green (24 pkgs; E2E 52 features) · `make test-race` green · topology audit **5 pre-existing, none new** (52 features · 409 module rows · 2110 steps) · `STATUS.md` split (round-075 detail + env note → `docs/archives/status/2026-09-22.md`) · propagation `dev → main` (**no-ff**) + tag **`round-076`** · `go install` · **[#154](https://github.com/gosharplite/tellme/issues/154) CLOSED** |

### Decisions locked (round 076 / ADR 0048)

| # | Decision |
| --- | --- |
| **D1** | An unknown tool name is a **recoverable fold-back** (log the action block + append a paired `tool` result + `continue`) — mirroring the round-056 reason-less refusal; the unknown call runs no tool and records no step/usage. |
| **D2** | The fold-back names the unknown tool + a **bounded** list of the available wire names (bounded by the fixed registry size — **not** by `clampBytes`, which does not run on this path). |
| **D3** | The recoverable path is **bounded per turn** (`maxUnknownToolFolds = 3`); tellme has **no** repetition detection and the only other bound is `MAX_TOOL_LOOP` (1000, paid) — so this cap is mandatory. |
| **D4** | Counter home: a **loop-local** variable in `Run` (one turn per `Run`) — no state to leak, no domain-model change. |
| **D5** | Ordering: the unknown-name check stays **before** the reason gate (single-owned classification: an unknown call with a blank reason is reported as *unknown*). |
| **D6** | Round-008 **FR-010 narrowed, not removed** — the phrase + exit 7 cover bound reached / cap exhausted / no tools registered. |
| **D7/D8** | A **tellme-side robustness** improvement (reference behaviour not asserted); **not modelled** (`docs/domain-model/**` unchanged — ADR 0041 escape hatch; `plan.md` §5). |

### Commits (branch `076-recoverable-unknown-tool`, then merged)

| Commit | Note |
| --- | --- |
| `5454b8a` | `docs(076)`: plan package + spec |
| `c56bfd0` | `feat(076)`: recover from an unknown tool name (bounded per-turn fold-back) — acceptance + research + ADR 0048 + truth + implementation |
| `19a73bd` | `fix(076)`: fold the architect review (F-1…F-6 + TDs/nits) |
| `4fb14a1` | `fix(076)`: fold the fold-verification residuals (R1/R2/R3) |
| `851466e` | PR [#156](https://github.com/gosharplite/tellme/pull/156) merge into `dev` (by the operator) |
| *(this closeout, on `dev`)* | `docs(076)`: day close — round 076 delivered + propagated; STATUS split + 09/22 summary |

### Artifacts / truth

- Plan package: `spec.md` (US1/US2 · FR-001…FR-006 · NFR-001…NFR-004 · I-1…I-6 · edge cases · SC-001…SC-003 · A-1…A-5) · `checklists/requirements.md` · `features/acceptance/recovering-from-an-unknown-tool-name.feature` · `research.md` (D1–D8) · `plan.md` · `tasks.md` (T001–T010 + the fold + residual ledgers) · `truth-delta.md`.
- Truth: `specs/truth/techstack.md` *Agent tool loop* MODIFY (+ a *Tool-usage accounting* parenthetical correction) · `specs/truth/features/cli/chat/using-an-unknown-tool-name.feature` (**NEW**, 2 Rules) · `failing-the-tool-loop.feature` (Example relocated) · `offering-the-agent-tools.feature` (sibling Example synced) · `chat/dsl.md` (+2 rows) · `contracts/**` + `data/**` NOOP · `docs/domain-model/**` **unchanged** (not modelled).
- Code: `internal/agent/agentloop.go` · `internal/agent/agentloop_test.go` (the fold-back pin set: pairing, no-usage, mixed round, the cap + its value, ordering, chrome, reset) · `tests/e2e/steps/step_r076_unknown_tool.go`.
- **ADR 0048** (`docs/decisions/0048-recoverable-unknown-tool-name.md` + index).

### Falsifiability witnesses (reproduced then reverted)

Restore the terminal `return` ⇒ the unit pin **and** the E2E `A one-off unknown tool name is recovered and the turn finishes` redden · drop the fold-back's `ToolCallID` ⇒ the pairing pin reddens (`ToolCallID = ""`) · `maxUnknownToolFolds = 10` ⇒ the value pin reddens · `= 0` (the pre-076 behaviour) ⇒ the three unknown-tool E2E Examples redden. The `architect` independently re-ran each probe on a scratch copy.

## Open items (non-blocking)

- **Round-076 forward items** — **RF-076-1** the cap value (3) is a constant · **RF-076-2** the available-tools listing is bounded only by the fixed registry size · **RF-076-3** general runaway protection for *valid* tools (out of scope) · **RF-076-4** a near-miss suggestion for a bare MCP name overlaps [#155](https://github.com/gosharplite/tellme/issues/155) · **RF-076-5** the reason-less refusal path is still uncapped. All in **ADR 0048 §Forward**.
- **Round-075/074/073/072 forward items** — RF-075-* / RF-074-* / RF-073-* / RF-072-* in their `ADR §Forward`.
- **Issue tracker** — **[#154](https://github.com/gosharplite/tellme/issues/154) CLOSED (completed)** (fixed by round 076 / ADR 0048) · **[#155](https://github.com/gosharplite/tellme/issues/155) stays OPEN** (MCP name presentation — research-gated, a separate issue).
- **PM follow-ups** — **none open.**
- Carried: PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; round-011 forward items; the **5 pre-existing** Gherkin/DSL topology-audit errors.

## Next steps

1. Open the next round off `dev` via `/axb-specify` — a theme from **operator value or a live issue** (the only live issue is [#155](https://github.com/gosharplite/tellme/issues/155), research-gated and may be NOOP).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

*(Round 076 is fully closed out: PR #156 human-merged into `dev` (`851466e`, merge commit); propagation `dev → main` **DONE (no-ff)**, tagged **`round-076`** with operator approval; the installed binary refreshed from the `dev` head; [#154](https://github.com/gosharplite/tellme/issues/154) closed.)*

## PM follow-ups

- None new (spec/acceptance complete; the `architect`'s F-1…F-6 were falsifiability/record folds, not PM-owned gaps).

## Process notes (durable)

- **"A claimed invariant with no witness is a round-trip waiting to happen"** — the round-076 review's whole weight was falsifiability: the loop change was correct, yet the **pairing**, **no-usage**, **cap-discrimination**, **ordering** and **chrome** claims were all unpinned (a mutation stayed green). Add the pin with the claim; a witness that cannot fail is not a witness.
- **The bug reproduced itself on the reviewer's own session** — the `architect` peer called a hallucinated `exec_command` and the **installed (pre-076)** binary aborted with the exact `tool … is not available` exit-7; re-staging with an explicit tool list recovered it (the session-61 precedent). A live repro on a peer is strong grounding.
- **A `clampBytes` claim must be checked against the path** (review F-4): the loop's clamp runs only on the *executed*-tool result path, not on a fold-back — an unbounded list was claimed bounded.
- **Cap constants deserve a literal pin** (review F-2): asserting `N+1` symbolically proves the mechanism, not the value — a literal pin (round-064's precedent) forces a reviewed change.

---

## 39. Session 63 (2026-09-22, cont.) — round 077 `077-mcp-tool-name-presentation` **OPENED**: make the callable MCP wire name discoverable → full pipeline → **PR open** (anchor [#155](https://github.com/gosharplite/tellme/issues/155); **ADR 0049**)

Bootstrap re-run (Steps 1–8; round 076 delivered/frozen; active branch `dev`). The operator directed *"Open a new round, the goal is to close issue 155."* A new branch **`077-mcp-tool-name-presentation`** was created off `dev` `ab8fab7`; the full AIxBDD pipeline ran to a green PR.

### At a glance

| Area | Outcome |
| --- | --- |
| Branch | **`077-mcp-tool-name-presentation`** (off `dev` `ab8fab7`) |
| Anchor | [#155](https://github.com/gosharplite/tellme/issues/155) — MCP tool presentation does not surface the namespaced wire name; **research-gated, may be NOOP**; **DoD = close it** |
| Theme | MODIFY (MCP tool presentation): the offered MCP declaration's **description** names the **callable wire name** (`mcp_<server>_<tool>`), and the empty-description fallback names the callable name — description-only, schema unchanged, server text never rewritten |
| Clarify | **not escalated (0 questions)** — the issue delegated the decision to `/axb-technical-research` |
| **Verify-first (the issue's decisive step)** | a **live** `initialize` + `tools/list` against `api.githubcopilot.com/mcp/` (2026-09-22): **45** tools, **0** with an empty description; `get_me`'s description names no tool → tellme's **fallback did NOT fire** (it was not the observed cause). **A fix IS warranted** (the presentation gap is real) — the NOOP verdict was **not** taken |
| Pipeline | specify ✅ · clarify ✅ (0) · **spec-by-example ✅** (`features/acceptance/discovering-the-callable-mcp-tool-name.feature`) · technical-research ✅ (**ADR 0049** + a `techstack.md` ADD row) · system-analysis ✅ (1 CLI end; api/data NOOP) · dsl-refine ✅ (a Rule + 2 Then rows) · tasks ✅ (T001–T012) · implement ✅ |
| The change | `internal/infrastructure/mcp/tool.go` — a named `callNameNoteFormat` constant + a computed `description` in `NewTool`: `Call this tool as "<t.name>". <server text verbatim>`; the empty-description fallback now names the **callable** name (not `def.Name`) |
| Verification | `gofmt`/`goimports` clean · `go vet` clean · `go test -count=1 ./...` **green** (24 pkgs; E2E **289 scenarios**) · `make verify` **OK** · `go.mod`/`go.sum` unchanged · topology audit **the same 5 pre-existing errors, none new** (52 features · 412 module rows · 2124 steps) |
| Witnesses (reproduced then reverted) | **W1** neutralize the note ⇒ the E2E `The offered declaration names the callable wire name` reddens at the intended Then (+ a `vet` printf build red); **W2** note the bare name ⇒ the callable-name + truncated-name pins red; **W3** bare fallback ⇒ the fallback pin reds |
| Delivery | branch `077-mcp-tool-name-presentation` → **PR open** — awaiting a human review/merge (no Copilot review) |

### Decisions locked (round 077 / ADR 0049)

| # | Decision |
| --- | --- |
| **D1** | Ship a fix (not NOOP) — the verify-first result shows the gap is real. |
| **D2** | Every offered MCP tool's **description** is **prefixed** with a tellme-authored call-name note naming the callable wire name (`t.name`). |
| **D3** | The note is added **outside** the server's own text, which is relayed **unchanged** (ADR 0025 D1); description-only, **schema bytes unchanged**; **no** prompt change (ADR 0025 D4). |
| **D4** | The **fallback** names the **callable** name, never the bare upstream name (a latent-correctness fix; the reference has no fallback). |
| **D5** | The note names the **post-truncation** wire name (`Name()`), so it can never lie. |
| **D6/D8** | Uniform + control-free; a **recorded divergence *beyond* the reference** (which neither advertises nor falls back). |

### Commits (branch `077-mcp-tool-name-presentation`)

| Commit | Note |
| --- | --- |
| `2178efa` | `docs(077)`: plan package + spec |
| `999f51f` | `feat(077)`: make the callable MCP wire name discoverable (ADR 0049) + unit/E2E pins + truth |

### Open items (non-blocking)

- **RF-077-1** the note is a small steady token cost per MCP declaration · **RF-077-2** hermetically unprovable model-efficacy · **RF-077-3** recorded divergence *beyond* the reference · **RF-077-4** a redundant note if the server text already names the tool · **RF-077-5** non-MCP tool kinds out of scope. All in **ADR 0049 §Forward**.
- Carried: PR #16 **Obs 1**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; the 5 pre-existing topology-audit DSL errors.

### Next steps

1. Human reviews + merges the PR (**no Copilot review**; only a human merges); then propagate `dev → main` (no-ff), tag `round-077`, refresh the binary; **close [#155](https://github.com/gosharplite/tellme/issues/155)**.
2. Operator may dispatch the `architect` peer for the review-fold loop (the round-069…076 protocol).
3. Re-read `SESSION-BOOTSTRAP.md` next session.

### PM follow-ups

- None new (the acceptance journey is a short presentation Rule; the round carries the falsifiable unit + E2E pins).

---

## 40. Session 63 closeout (2026-09-22) — round 077 `077-mcp-tool-name-presentation` **DELIVERED / FROZEN** (`SESSION-CLOSEOUT.md` Steps 1–8)

Round 077 was human-merged (PR [#157](https://github.com/gosharplite/tellme/pull/157) → `dev` `273b5a9`, **merge commit**); the remote branch was already gone, the operator confirmed, and the local branch was deleted after an ancestor check. Then the closeout ran.

### At a glance

| Area | Outcome |
| --- | --- |
| Step 1 — working tree | clean; on `dev`; `dev == origin/dev`; no delivered `specs/plans/**` touched (frozen history intact); staging files removed |
| Step 2 — gates | `gofmt`/`goimports` clean · `go vet ./...` clean · `go test -count=1 ./...` **green** (24 pkgs; E2E **289 scenarios**) · `make verify` **OK** · `make test-race` **no data races** · topology audit **the same 5 pre-existing errors, none new** (52 features · 412 module rows · 2124 steps) · `go.mod`/`go.sum` unchanged |
| Step 3 — `STATUS.md` | header + round-in-flight (`none`) + active branch (`dev`) refreshed; **Rule-12 split**: the **round-076** delivered-round detail + its env note relocated **verbatim** into [`docs/archives/status/2026-09-22.md`](../../../../archives/status/2026-09-22.md); the **round-077** delivered-round section added; the delivered-rounds index gains 077; branch model + propagation history + roadmap (077 → Delivered) + open items (round-077 forward items added; round-072 compacted to a pointer) + env notes (round-077 note; topology counts) updated |
| Step 4 — daily summary | this §40 appended (the §1–§39 record preserved) |
| Step 5 — reconcile | `STATUS.md` ↔ this summary agree: no round in flight, `dev` active, 0 open issues |
| Step 6 — commit | working `dev` committed + pushed |
| Step 7 — propagate + hand off | `dev → main` (**no-ff**), tagged **`round-077`**; installed binary refreshed (`go install ./cmd/tellme`) |
| Step 8 — issue tracker | **[#155](https://github.com/gosharplite/tellme/issues/155)** **CLOSED (completed)**; tracker now **0 open** |

### Commits (branch `077-mcp-tool-name-presentation`, then merged fast-forward-equivalent)

| Commit | Note |
| --- | --- |
| `2178efa` | `docs(077)`: plan package + spec |
| `999f51f` | `feat(077)`: make the callable MCP wire name discoverable (ADR 0049) + unit/E2E pins + truth |
| `ca5a908` | `docs(077)`: name the round PR (#157) |
| `24be271` | `fix(077)`: fold the architect review (F-1…F-4 + TD-1 + nits N-1…N-5) |
| `a37e838` | `fix(077)`: fold fold-verification residuals (R-1 record hygiene; R-2 home RF-077-6) |
| `6977d5c` | `docs(077)`: review-fold loop CLOSED (cleared for human merge) |
| `273b5a9` | PR [#157](https://github.com/gosharplite/tellme/pull/157) merge into `dev` (by the operator) |
| *(this closeout, on `dev`)* | `docs(077)`: day close — round 077 delivered + propagated; STATUS split + 09/22 summary §40 |

### Open items (non-blocking)

- **Round-077 forward items (RF-077-1…6)** in **ADR 0049 §Forward** (the note's token cost · hermetically unprovable efficacy · a divergence *beyond* the reference · a redundant note · non-MCP tool kinds · the deliberately un-pinned literal wording).
- Carried: PR #16 **Obs 1**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; the 5 pre-existing topology-audit DSL errors.

### Next steps

1. Open the next round off `dev` via `/axb-specify` — a theme from **operator value** (the tracker is **0 open**; the tool surface is settled).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

### PM follow-ups

- None new (the acceptance journey is a short presentation Rule; the round carries the falsifiable unit + E2E pins).

## 41. Session 64 (2026-09-22, cont.) — round 078 `078-provider-transport-retry` **DELIVERED / FROZEN**: bounded provider transport retry (operator request) → full pipeline → `architect` review-fold loop (3 passes, CLOSED) → **human-merged (PR #158 → `dev` `c291e1d`)**, propagated, tagged **`round-078`**; closeout (Steps 1–8)

The operator reported seeing a remote-server connection drop in tellme, and asked for a small bounded fix: *"wait 1 sec → retry → wait 3 sec → retry → error"*. We traced the failure to tellme having **no retry layer at all** (one provider call, one failure, terminal), agreed the **retryability predicate** (transport + HTTP 429/5xx — *not* 4xx/auth/decode/the round-030 truncation), opened **round 078** as an **operator request** (no anchor issue), ran the full AIxBDD pipeline, and took **PR [#158](https://github.com/gosharplite/tellme/pull/158)** through the `architect` peer's **review → fold → fold-verification** loop to **FOLDS VERIFIED — CLEARED FOR HUMAN MERGE**; the operator merged it, deleted the remote branch, and `SESSION-CLOSEOUT.md` Steps 1–8 ran (local branch deleted after an ancestor check; local `dev` fast-forwarded).

### At a glance

| Area | Outcome |
| --- | --- |
| Branch | `078-provider-transport-retry` (off `dev` `dd9b94b`) |
| Theme | **operator request** (no anchor issue, rounds 073/074/075 precedent) — a **bounded transport retry** so a momentary network blip does not cost a re-run |
| Clarify | **not escalated (0 questions)** — the theme/schedule are operator-given and the retryability predicate (transport + 429/5xx) was **operator-locked** in-session; the residual choices (mechanism / seam / typed classification / hermetic test seam / record shape) → `/axb-technical-research` |
| Pipeline | specify ✅ · spec-by-example ✅ · technical-research ✅ (**ADR 0050** + `techstack.md` MODIFY ×2 + ADD) · system-analysis ✅ (1 CLI end; api/data NOOP; not modelled) · dsl-refine ✅ (a new Rule + 4 Examples; +4 Given / +4 Then rows) · tasks ✅ (T001–T012) · implement ✅ |
| Before → after | a retryable failure was terminal (`the provider request failed` + exit 6, one request) → it is now retried **twice** (1 s, 3 s) and, on success, the turn completes; after exhaustion the failure surface is **byte-identical** (`the provider request failed` + exit 6) |
| The change | `internal/domain/llm/gateway.go` (`ProviderError.Status`/`Transport` + `llm.Retryable`) · `internal/infrastructure/llm/{openai,gemini}/client.go` (the typed facts) · `internal/cli/retry_gateway.go` (a `llm.Gateway` decorator: fixed `retryDelays = {1s, 3s}`, `ctx`-aware, the `TELL_ME_FORCE_RETRY_DELAY_MS` seam, a chrome-styled `stderr` retry line that yields/restores the spinner) + `internal/cli/cli.go` (wired **inner** to `withUnpairedDiagnostic`) |
| Review chain (PR #158, the `architect` peer) | `review` (**APPROVE WITH REQUIRED FOLDS** — no `[ARCHITECTURAL BLOCKER]`; the design was right, the folds were **witness / record quality**: **F-1** the `tellme: ` prefix quoted on three record surfaces · **F-2** the accounting MUST unwitnessed · **F-3** the notifier's yield/restore dead · **F-4** the retry line had no E2E carrier; + TD-078-1…3, RF-078-1…4, N-078-1…4) → fold `9ac7e6a` → `FOLD-VERIFICATION` (**FOLDS VERIFIED WITH RESIDUALS**; R-1…R-3) → residual fold `657ae2a` → `FOLD-VERIFICATION` (**FOLDS VERIFIED — CLEARED FOR HUMAN MERGE**; no residual) |
| Merge | PR [#158](https://github.com/gosharplite/tellme/pull/158) **human-merged** into `dev` (`c291e1d`, **merge commit**); remote branch deleted by the human; the **local branch deleted** after an ancestor check |
| Closeout | `make check` **OK** · `go test -count=1 ./...` green (E2E **293 scenarios**) · `make test-race` green · topology audit **5 pre-existing, none new** (52 features · 424 module rows · 2154 steps) · `STATUS.md` split (round-077 detail + env note → `docs/archives/status/2026-09-22.md`) · propagation `dev → main` (**no-ff**) + tag **`round-078`** · `go install` · **no anchor issue to close** (tracker stays 0 open) |

### Decisions locked (round 078 / ADR 0050)

| # | Decision |
| --- | --- |
| **D1** (operator-locked) | **Retryable** = a **transport** failure (dial / EOF / reset / broken pipe / GOAWAY / TLS / timeout) **or** HTTP **429/5xx**; **not** 4xx / auth / content-filter / decode / the round-030 truncation. |
| **D2** | **Typed classification** — `ProviderError` gains `Status`/`Transport` (facts set at the adapter); the **domain-owned** `llm.Retryable` predicate (`Transport || Status == 429 || 500 ≤ Status ≤ 599`) — **never** string-matched. |
| **D3** (operator-locked) | Fixed schedule `retryDelays = {1 s, 3 s}` (an array) → **≤2 retries / 3 attempts**; a `TELL_ME_FORCE_RETRY_DELAY_MS` hermetic seam **scoped to the 078 E2E scenarios**; a **literal** unit pin (the order is a constant-pin). |
| **D4** | The seam is a `llm.Gateway` **decorator** in `internal/cli`, **inner to** `withUnpairedDiagnostic`; one seam covers **both** families. |
| **D5** | The retry is announced once per retry on `stderr` only — chrome-styled `[HH:MM:SS] retrying …`, carrying **neither** the `tellme: ` prefix nor the class phrase; the spinner frame is yielded/restored. |
| **D6** | Cancellation aborts immediately — the decorator checks `ctx.Err()` before each attempt **and before announcing** a retry; a client-side 300 s timeout stays retryable, a parent cancellation does not. |
| **D7** | A retried call is **one** AI-endpoint call — one `calls`/usage/frame; a failed attempt writes no history. |
| **D8** | **ADR 0050** + `techstack.md` MODIFY ×2 (incl. the round-030 "no retry layer" reconciliation) + ADD; the CLI feature + DSL rows; **not modelled** (`docs/domain-model/**` unchanged — ADR 0041 escape hatch; `plan.md` §5). |
| **D9** | Scope: the provider path only — MCP is out (discovery is warn+skip; an MCP tool-call failure is already a recoverable fold-back); `-b`/`--retry` is unrelated; **no** new class phrase; **no** new exit code. |

### Commits (branch `078-provider-transport-retry`, then merged)

| Commit | Note |
| --- | --- |
| `489c4b7` | `docs(078)`: plan package + spec |
| `b098058` | `docs(078)`: acceptance Gherkin |
| `77b835b` | `docs(078)`: technical research + ADR 0050 + techstack truth |
| `d61935e` | `feat(078)`: retry a transient provider transport failure (bounded, 1s/3s) |
| `9ac7e6a` | `fix(078)`: fold the architect review (F-078-1…4 + TD-078-1…3 + RF/N) |
| `657ae2a` | `docs(078)`: fold the fold-verification residuals (R-1…R-3) |
| `c291e1d` | PR [#158](https://github.com/gosharplite/tellme/pull/158) merge into `dev` (by the operator) |
| *(this closeout, on `dev`)* | `docs(078)`: day close — round 078 delivered + propagated; STATUS split + 09/22 summary §41 |

### Artifacts / truth

- Plan package: `spec.md` (US1/US2 · FR-001…FR-007 · NFR-001…NFR-004 · SC-001…SC-005) · `checklists/requirements.md` · `features/acceptance/recovering-from-a-momentary-provider-failure.feature` · `research.md` (D1–D9) · `plan.md` · `tasks.md` (T001–T012 + the fold and residual ledgers) · `truth-delta.md`.
- Truth: `specs/truth/techstack.md` *Provider gateway port* MODIFY + *Provider output-cap truncation guard* MODIFY + *Provider request retry (transient transport)* **ADD** · `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` (a new Rule, 4 Examples) · `chat/dsl.md` (+4 Given / +4 Then rows; **reconciled** the round-030 "no retry layer" note and the round-004 "exactly one provider request" note) · `contracts/**` + `data/**` NOOP · `docs/domain-model/**` **unchanged** (not modelled).
- Code: `internal/domain/llm/gateway.go` · `internal/infrastructure/llm/{openai,gemini}/client.go` · `internal/cli/retry_gateway.go` + `internal/cli/cli.go` · `internal/domain/llm/retryable_test.go` · `internal/cli/retry_gateway_test.go` · `tests/e2e/fakeprovider/fakeprovider.go` (the `DropFirst` transport drop) · `tests/e2e/steps/scenario_context.go` · `tests/e2e/steps/step_r078_retry_{givens,thens}.go`.
- **ADR 0050** (`docs/decisions/0050-bounded-provider-transport-retry.md` + index).

### Falsifiability witnesses (reproduced then reverted — by us and by the architect)

Change the retry-line text ⇒ the E2E `tellme announces …` reddens · persist `calls = len+1` ⇒ the E2E `the turn was recorded as a single provider call` reddens (`persisted calls = 2, want 1`) · drop the `YieldIndicator()` call ⇒ the unit `…YieldsAndRestoresAroundTheLine` reddens · remove the pre-notify `ctx.Err()` check ⇒ the unit `…NoAnnounceWhenCancelledMidCall` reddens. The `architect` independently re-ran all four.

### Open items (non-blocking)

- **Round-078 forward items** — **RF-078-1** the delays/count are constants · **RF-078-2** no jitter (deliberate) · **RF-078-3** the retry covers transport/status only · **RF-078-4** the added worst-case latency · **RF-078-5** the retry-line yield/restore + no-prefix (resolved) · **RF-078-6** MCP unchanged · **RF-078-7** offline readers untouched · **RF-078-8** the byte-identity re-send pin (resolved) · **RF-078-9** the split retry policy (deliberate) · **RF-078-10** the `DropFirst` hijack fallback (defensive). All in **ADR 0050 §Forward**.
- **Compaction**: the round-076/075/074/073 forward-item batches are now **compacted to a single pointer row** (each ≥2 rounds old) per the curation rule.
- **Issue tracker** — round 078 had **no anchor issue** (an operator request); the tracker stays **0 open** (`gh issue list --state open` = empty) — nothing to close or revise.
- **PM follow-ups** — **none open.**
- Carried: PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; round-011 forward items; the **5 pre-existing** Gherkin/DSL topology-audit errors.

### Next steps

1. Open the next round off `dev` via `/axb-specify` — a theme from **operator value** (the tracker is **0 open**; the tool surface is settled).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

*(Round 078 is fully closed out: PR #158 human-merged into `dev` (`c291e1d`, merge commit); propagation `dev → main` **DONE (no-ff)**, tagged **`round-078`** with operator approval; the installed binary refreshed from the `dev` head; no anchor issue to close.)*

### PM follow-ups

- None new (the acceptance journey is a short recovery Rule; the round carries the falsifiable unit + E2E pins).

### Process notes (durable)

- **A claim must have a witness that can redden (the round's whole review weight).** The retry change was correct, yet the **`tellme: ` prefix** was quoted on three record surfaces the code must NOT emit (F-1), the **accounting MUST** had no witness (F-2), the notifier's **yield/restore** branch was dead (unit passed `nil`; no 078 E2E terminal) (F-3), and the **retry line** had no E2E carrier (F-4). Add the pin with the claim; a witness that cannot fail is not a witness (the round-075/077 lesson, again).
- **A retry announcement must precede a retry that WILL run (TD-1).** `notify` fired before the interruptible sleep, so a SIGINT mid-call printed "retrying …" for a retry that never happened; the fix checks `ctx.Err()` **before** announcing.
- **Scope a hermetic seam to the scenarios that need it (TD-2).** A global `TELL_ME_FORCE_RETRY_DELAY_MS=0` would have silently retry-enabled *every* E2E scenario; the override now lives in the 078 Givens only.
- **Reconcile, don't just append (the round-030 reconciliation).** Round 078 makes round-030's "tellme adds **no** retry layer (there is none to change)" false; the round MODIFY-ed the clause (truth + `dsl.md`) rather than leaving a contradiction — and the fold also reconciled the round-004 "exactly one provider request" note (TD-3).

---

## 42. Session 65 (2026-09-22, cont.) — round 079 `079-interrupted-turn-partial-persistence` **DELIVERED / FROZEN**: preserve completed tool steps on an operator-interrupted turn (anchor issue #159) → full pipeline → `architect` review-fold loop (3 passes, CLOSED) → **human-merged (PR #160 → `dev` `30fb420`)**, propagated, tagged **`round-079`**; closeout (Steps 1–8)

The operator asked to read issue **#159** and *"open a new round, the goal is to close this issue."* We bootstrapped (`SESSION-BOOTSTRAP.md` Steps 1–8; round 078 delivered/frozen; active branch `dev`), opened **round 079** off `dev` `4f9c96d` with **#159 as the anchor (DoD = close it)**, ran the full AIxBDD pipeline, took **PR [#160](https://github.com/gosharplite/tellme/pull/160)** through the `architect` peer's **review → fold → fold-verification → residual fold → final fold-verification** loop to **FOLDS VERIFIED — loop CLOSED, cleared for human merge**, the human merged it and deleted the remote branch, and `SESSION-CLOSEOUT.md` Steps 1–8 ran (local branch deleted after an ancestor check; `dev` == the merge commit `30fb420`).

### At a glance

| Area | Outcome |
| --- | --- |
| Branch | `079-interrupted-turn-partial-persistence` (off `dev` `4f9c96d`) |
| Anchor | issue [#159](https://github.com/gosharplite/tellme/issues/159) — *Preserve completed tool steps on Ctrl+C interruption … with a synthetic turn-closing answer*; **DoD = close it** |
| Theme | **MODIFY (session-history persistence)**: an operator-interrupted turn (`SIGINT`/`SIGTERM` → `context.Canceled`) with **≥1** completed tool step is persisted as ONE `history.Entry` closed with a synthetic assistant answer; **zero** steps writes nothing (today's clean abort) |
| Clarify | **not escalated (0 questions)** — the core behaviour is anchored by the issue; the issue's *Open Design Decisions* were routed by the operator to `/axb-technical-research` |
| Pipeline | specify ✅ · spec-by-example ✅ (2 Rules / 2 Examples) · technical-research ✅ (**ADR 0051** + `techstack.md` MODIFY + ADD + the **domain-model** invariant amendment) · system-analysis ✅ (1 CLI end; api/data NOOP) · dsl-refine ✅ (+2 Given / +6 Then rows) · tasks ✅ (T001–T010) · implement ✅ |
| Before → after | a `Ctrl+C` mid-tool-turn was terminal (`the provider request failed` + exit 6) and **discarded** the prompt + all completed steps → the partial turn is **kept** (one entry, synthetic close), the run exits **0** with `stdout` empty + one informational `stderr` line; the **zero-step** case is unchanged (phrase + exit 6) |
| The change | `internal/domain/history/history.go` (`InterruptedTurnAnswer`) · `internal/cli/cli.go` (`runTurn`: typed `errors.Is(err, context.Canceled) && len(result.Steps) > 0` ⇒ `persistInterruptedTurn`) · `internal/cli/interrupted_turn_test.go` · `tests/e2e/steps/step_r079_interrupt.go` (a scripted `execute_command` running `kill -INT $PPID; sleep 30` — a **real** `SIGINT` to the turn process) |
| Review chain (PR #160, the `architect` peer — init once with `SESSION-BOOTSTRAP.md`, continuations) | `review` (**APPROVE WITH REQUIRED FOLDS** — no `[ARCHITECTURAL BLOCKER]`; record/witness quality: **F-079-1** the hermetic-seam record described a stall mechanism that does not exist · **F-079-2** the persisted-`calls` claim had no E2E carrier · **F-079-3** the kill-path result is a nil-error `Exit Code: -1`, not `error: …` · **F-079-4** the interrupted path leaves a dangling `╭─⠿ Turn N` frame; + TD-079-1/2, N-079-1/2) → fold `3e6564c` → `FOLD-VERIFICATION` (**FOLDS VERIFIED WITH RESIDUALS**; RES-079-FV-1…5) → residual fold `936125a` → `FOLD-VERIFICATION` (**FOLDS VERIFIED — loop CLOSED, cleared for human merge**; post-closure cosmetic table fix `4be021c`, product code byte-identical) |
| Merge | PR [#160](https://github.com/gosharplite/tellme/pull/160) **human-merged** into `dev` (`30fb420`, **merge commit**); remote branch deleted by the human; the **local branch deleted** after an ancestor check |
| Closeout | `make check` **OK** · `go test -count=1 ./...` green (E2E **295 scenarios · 2201 steps**) · `make test-race` green · topology audit **5 pre-existing, none new** (52 features · 6 feature modules · 17 root + 432 module rows · 2175 steps) · diff secret scan clean · `make modelith-check` ×3 green · `STATUS.md` split (round-078 detail + env note → `docs/archives/status/2026-09-22.md`) · propagation `dev → main` (**no-ff**) + tag **`round-079`** · `go install` · **[#159](https://github.com/gosharplite/tellme/issues/159) CLOSED** |

### Decisions locked (round 079 / ADR 0051)

| # | Decision |
| --- | --- |
| **D1** | Persist the completed steps on an operator interruption: one atomic `history.Entry` closed with the synthetic `history.InterruptedTurnAnswer`. *Rejected:* persisting without a closing answer (breaks role alternation). |
| **D2** | **Typed** detection — `errors.Is(err, context.Canceled)` (unwraps through the adapter's `*llm.ProviderError` + the round-078 retry decorator); **never** string-matched. |
| **D3** | The synthetic answer is the one stored constant `[Turn interrupted by operator via Ctrl+C]`; a `SIGTERM` shares the wording. |
| **D4** | **Always-on**; no config toggle / no double-Ctrl+C discard (declined alternatives recorded). |
| **D5** | Interrupted-with-work ⇒ exit **`Success` (0)**, `stdout` empty, one informational `stderr` line (no `tellme: ` prefix, not in `turns.log`); zero-step ⇒ today's surface (phrase + exit 6); append failure ⇒ the environment phrase (4). The exit-code set stays **ten**. |
| **D6** | **No** usage persisted for the interrupted turn (the existing `persistTurnUsage` gate already yields nothing; a failed turn writes none today). |
| **D7** | The trigger is `len(result.Steps) > 0`; a "completed tool step" is one whose tool **returned** a result (a success, or a nil-error kill/timeout `Exit Code: -1`) — the loop's `error: ` prefix applies only to a non-nil tool error. |
| **D8** | The `history.Entry` / `history.Step` JSON shapes are **unchanged** (the synthetic answer is the `answer` value); `calls` = the completed inference rounds. |
| **D9** | Scope: the provider/tool-loop turn only (MCP steps are ordinary steps); `-b`/`--retry` + the offline readers untouched. The live interruption is hermetic (a scripted `kill -INT $PPID`, no pty). |
| **D10** | Records: **ADR 0051** + index; `techstack.md` MODIFY + ADD; the CLI truth feature + `dsl.md` rows; and the **domain-model** amendment (`history-append-after-complete` + `turn-answer-stored-verbatim`) per the ADR-0041 same-PR rule. |

### Commits (branch `079-interrupted-turn-partial-persistence`, then merged)

| Commit | Note |
| --- | --- |
| `eb27570` | `docs(079)`: plan package + spec |
| `074abe9` | `docs(079)`: STATUS — round 079 in flight |
| `a68d06d` | `docs(079)`: technical research + **ADR 0051** + techstack truth + domain-model amendment |
| `beb3670` | `feat(079)`: acceptance + plan + dsl-refine truth + tasks + implementation (persist the interrupted turn's completed steps) |
| `48a92ef` | `docs(079)`: STATUS — pipeline complete, PR open |
| `3e6564c` | `fix(079)`: fold the architect review (F-079-1…4 + TD-1/2 + N-1/2) |
| `936125a` | `docs(079)`: fold the fold-verification residuals (RES-079-FV-1…5) |
| `4be021c` | `docs(079)`: fix the round-079 DSL table layout (RES-079-FV-6, cosmetic) |
| `30fb420` | PR [#160](https://github.com/gosharplite/tellme/pull/160) merge into `dev` (by the operator) |
| *(this closeout, on `dev`)* | `docs(079)`: day close — round 079 delivered + propagated; STATUS split + 09/22 summary §42 |

### Artifacts / truth

- Plan package: `specs/plans/079-interrupted-turn-partial-persistence/**` — `spec.md` (US1/US2 · FR-001…FR-008 · NFR-001…NFR-004 · SC-001…SC-005 · I-1…I-8 · S-1…S-12) · `checklists/requirements.md` · `features/acceptance/preserving-an-interrupted-turns-completed-work.feature` · `research.md` (D1–D10) · `plan.md` · `tasks.md` (T001–T010 + the fold ledger + the recorded narrowings) · `truth-delta.md`.
- Truth: `specs/truth/techstack.md` *Session history store* **MODIFY** (append-after-complete qualified) + *Interrupted-turn persistence* **ADD** · `specs/truth/features/cli/chat/remembering-the-conversation.feature` (2 new Rules / 2 Examples) · `chat/dsl.md` (+2 Given / +6 Then rows + the round-079 note) · `contracts/**` + `data/**` NOOP.
- Domain model: `docs/domain-model/tellme.modelith.{yaml,md}` — `history-append-after-complete` + `turn-answer-stored-verbatim` **MODIFY** (the operator-interrupted exception; re-rendered; `modelith-check` green). A **recorded divergence** (the reference discards the partial turn).
- Code: `internal/domain/history/history.go` · `internal/cli/cli.go` · `internal/cli/interrupted_turn_test.go` · `tests/e2e/steps/step_r079_interrupt.go`.
- **ADR 0051** (`docs/decisions/0051-interrupted-turn-partial-persistence.md` + index).

### Falsifiability witnesses (reproduced then reverted)

**M1** (reviewer) remove the interrupted-turn branch ⇒ the E2E `A turn interrupted after a tool step keeps it` reddens — proving the **typed** detection on the real `SIGINT` → `*llm.ProviderError` → retry-decorator path. **M2** (`Answer: ""`) ⇒ the synthetic-close Then reddens. **F-079-2** (`Calls: 0`) ⇒ the persisted-`calls` Then reddens (`persisted calls = 0, want 1`). **TD-079-1** (leak the class phrase / write stdout) ⇒ the phrase-absence / `wrote nothing to standard output` Thens redden.

### Open items (non-blocking)

- **Round-079 forward items** — **RF-079-1** the synthetic answer is stored as the turn's `answer` · **RF-079-2** a killed-tool step is counted (nil-error `Exit Code: -1`) · **RF-079-3** no usage persisted · **RF-079-4** the exit-0/zero-step-exit-6 asymmetry · **RF-079-5** `SIGTERM` shares the "via Ctrl+C" wording · **RF-079-A** no scenario chains the product-written interrupted entry into a resume · **RF-079-B** the dangling `╭─⠿ Turn N` frame reaches `stderr`/`turns.log` · **RF-079-7** no archive interaction · **RF-079-8** the informational line is `stderr`-only · **RF-079-9** the `ind.Stop()`↔append window is unexercised. All in **ADR 0051 §Forward**.
- **Compaction**: the round-077 forward-item batch is now **compacted to a pointer row** (2 rounds old) per the curation rule.
- **Issue tracker** — **[#159](https://github.com/gosharplite/tellme/issues/159) CLOSED (completed)** (fixed by round 079 / ADR 0051, PR [#160](https://github.com/gosharplite/tellme/pull/160)); the tracker is **0 open**.
- **PM follow-ups** — **none open.**
- Carried: PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; round-011 forward items; the **5 pre-existing** Gherkin/DSL topology-audit errors.

### Next steps

1. Open the next round off `dev` via `/axb-specify` — a theme from **operator value** (the tracker is **0 open**; the tool surface is settled).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

*(Round 079 is fully closed out: PR #160 human-merged into `dev` (`30fb420`, merge commit); propagation `dev → main` **DONE (no-ff)**, tagged **`round-079`** with operator approval; the installed binary refreshed from the `dev` head; [#159](https://github.com/gosharplite/tellme/issues/159) closed.)*

### PM follow-ups

- None new (spec/acceptance complete; the `architect`'s F-079-1…4 were record/witness folds, not PM-owned gaps).

### Process notes (durable)

- **A recorded mechanism must match the shipped one (F-079-1).** Three durable surfaces (the `techstack.md` row, ADR §9, `research.md` D9) described a **fake-provider stall + harness signal helper** that was never built — the shipped seam is a scripted `execute_command` running `kill -INT $PPID`. Worse, that stale paragraph gave the *narrowing* the wrong reason. When a seam changes during implementation, reconcile **every** surface that names it (the round-078 "reconcile, don't append" lesson, again).
- **A claimed E2E assertion must have a carrier that can redden (F-079-2).** SC-001 claimed the E2E asserts the persisted `calls`; it did not (`Calls: 0` left the E2E green) — the claim now has a Then, and the mutant reddens it.
- **Describe the kill path exactly (F-079-3).** A cancelled `execute_command` returns a **nil-error** `Exit Code: -1`, not an `error: …` string; the loop's `error: ` prefix applies only to a non-nil tool error. "Completed tool step" was defined where it is used.
- **An unavoidable artifact must be recorded, not denied (F-079-4).** The interrupted path emits the aborted call's `╭─⠿ Turn N` frame before the cancellation lands (round-040 `OnCallBegin` ordering) — recorded as RF-079-B rather than claimed absent.
- **The live signal can be produced from inside the tool (the round's design win).** A scripted `execute_command` running `kill -INT $PPID; sleep 30` lands a **real** `SIGINT` on the turn process hermetically — no pty, no stall harness — so the typed detection is witnessed end-to-end (M1), not merely on a unit double.

---

## 43. Session 66 (2026-09-22, cont.) — round 080 `080-failed-turn-partial-persistence` **DELIVERED / FROZEN**: persist a FAILED turn's completed tool steps (anchor issue #161) → full pipeline → `architect` review-fold loop (CLOSED) → **human-merged (PR #162 → `dev` `4eec76e`)**, propagated, tagged **`round-080`**; closeout (Steps 1–8)

The operator asked to fix the remaining loss: *"retry succeeds → all executed steps persist; retry fails → none do. We need to fix the above."* We created a detailed issue [#161](https://github.com/gosharplite/tellme/issues/161), opened **round 080** off `dev` `b1c929e` with **#161 as the anchor (DoD = close it)**, ran the full AIxBDD pipeline, took **PR [#162](https://github.com/gosharplite/tellme/pull/162)** through the `architect` peer's **review → fold → fold-verification** loop to **FOLDS VERIFIED — loop CLOSED**, the human merged it and deleted the remote branch, and `SESSION-CLOSEOUT.md` Steps 1–8 ran (local branch deleted after an ancestor check; `dev` == the merge commit `4eec76e`).

### At a glance

| Area | Outcome |
| --- | --- |
| Branch | `080-failed-turn-partial-persistence` (off `dev` `b1c929e`) |
| Anchor | issue [#161](https://github.com/gosharplite/tellme/issues/161) — *Persist completed tool steps when a turn FAILS mid-tool-loop (provider failure or retry exhaustion), not only on Ctrl+C*; **DoD = close it** |
| Theme | **MODIFY (session-history persistence)**: a turn that **fails** with ≥1 completed step is persisted as ONE `history.Entry` closed with a class-specific synthetic answer, while the failure surface stays **unchanged** (frozen phrase + exit 6/7); zero steps ⇒ nothing |
| Clarify | **not escalated (0 questions)** — the core behaviour is anchored by the issue; the issue's *Open Design Decisions* were routed by the operator to `/axb-technical-research` |
| Pipeline | specify ✅ · spec-by-example ✅ (3 Rules / 4 Examples) · technical-research ✅ (**ADR 0052** + `techstack.md` MODIFY ×2 + the domain-model invariant extension) · system-analysis ✅ (1 CLI end; api/data NOOP) · dsl-refine ✅ (+3 Given / +5 Then rows) · tasks ✅ (T001–T012) · implement ✅ |
| Before → after | a failed turn (retry exhausted / non-retryable 4xx / the tool-loop exit-7 class) discarded the prompt + every completed step → it keeps them (one entry, a class-specific synthetic close), still exiting **6**/**7** with the frozen phrase + one informational `stderr` line; the **zero-step** failure is unchanged |
| The change | `internal/domain/history` (`ProviderFailedTurnAnswer` / `ToolFailedTurnAnswer`) · `internal/cli/cli.go` (`failTurn` extracted; the interruption predicate broadened to `ctx.Err() != nil \|\| errors.Is(err, context.Canceled)` — folding issue **hole #2**) · `internal/cli/failed_turn_test.go` · `tests/e2e/fakeprovider/fakeprovider.go` (a per-reply `Drop`/`ErrorStatus`) · `tests/e2e/steps/step_r080_failed_turn.go` |
| Review chain (PR #162, the `architect` peer — init once with `SESSION-BOOTSTRAP.md`, continuations) | `review` (**APPROVE WITH REQUIRED FOLDS** — no `[ARCHITECTURAL BLOCKER]`; witness/record quality: **F-080-1** the hole-#2 fix had no witness (a mutation left the suite green) · **F-080-2** the persisted-`calls` claim had no E2E carrier · **F-080-3** an unrecorded interruption-over-failure precedence; + TD-080-1/2, N-080-1…4) → fold `b8092dc` → `FOLD-VERIFICATION` (**FOLDS VERIFIED — loop CLOSED, cleared for human merge**; RES-080-FV-1…3) → residual fold `46fba48` (product code byte-identical) |
| Merge | PR [#162](https://github.com/gosharplite/tellme/pull/162) **human-merged** into `dev` (`4eec76e`, **fast-forward**); remote branch deleted by the human; the **local branch deleted** after an ancestor check |
| Closeout | `make check` **OK** · `go test -count=1 ./...` green (E2E **299 scenarios · 2239 steps**) · `make test-race` green · topology audit **5 pre-existing, none new** (52 features · 6 feature modules · 17 root + 440 module rows · 2213 steps) · diff secret scan clean · `make modelith-check` ×3 green · `STATUS.md` split (round-079 detail + env note → `docs/archives/status/2026-09-22.md`) · propagation `dev → main` (**no-ff**) + tag **`round-080`** · `go install` · **[#161](https://github.com/gosharplite/tellme/issues/161) CLOSED** |

### Decisions locked (round 080 / ADR 0052)

| # | Decision |
| --- | --- |
| **D1** | Persist a failed turn with completed steps: one atomic `history.Entry` closed with a synthetic answer. *Rejected:* persisting without a closing answer (breaks role alternation). |
| **D2** | **Scope (B): any failed turn** with ≥1 completed step — provider failures **and** the tool-loop failure (`ErrIncomplete`, exit 7). One predicate. |
| **D3** | The **failure surface is unchanged** — the frozen phrase + exit **6** / **7**; only one informational `stderr` line is added. *Rejected:* exit 0 / 130. |
| **D4** | A **distinct** synthetic answer per class: `history.ProviderFailedTurnAnswer` = `[Turn ended early: the provider request failed]`; `history.ToolFailedTurnAnswer` = `[Turn ended early: the tool loop did not complete]`. |
| **D5** | The interruption predicate is broadened to `ctx.Err() != nil \|\| errors.Is(err, context.Canceled)` (both structural) → folds issue **hole #2** (a `Ctrl+C` during a round-078 retry wait). **Precedence:** an interruption outranks a same-turn failure (only when steps > 0). |
| **D6** | The failed-turn `Append` is **best-effort** — an append failure never masks the failure (no keep line claimed). Divergence from the operator path (which routes an append failure to exit 4) recorded. |
| **D7** | Records: **ADR 0052** + index; `techstack.md` (the round-079 row broadened to *Partial-turn persistence (operator interruption or failure)* + the *Session history store* clause extended); the CLI truth feature + `dsl.md` rows; and the `docs/domain-model` invariant extension (same-PR, ADR 0041). |

### Commits (branch `080-failed-turn-partial-persistence`, then merged)

| Commit | Note |
| --- | --- |
| `77fdf2f` | `docs(080)`: plan package + spec + STATUS — round 080 in flight |
| `ff9e00b` | `feat(080)`: acceptance + research + **ADR 0052** + truth + domain-model + implementation (persist a failed turn's completed steps) |
| `f555ddb` | `docs(080)`: STATUS — pipeline complete, PR #162 open |
| `b8092dc` | `fix(080)`: fold the architect review (F-080-1…3 + TD-1/2 + N-1…4) |
| `46fba48` | `docs(080)`: fold the fold-verification residuals (RES-080-FV-1…3, cosmetic) |
| `4eec76e` | `docs(080)`: STATUS — review-fold loop CLOSED, PR #162 ready for human merge (the merge; fast-forward) |
| *(this closeout, on `dev`)* | `docs(080)`: day close — round 080 delivered + propagated; STATUS split + 09/22 summary §43 |

### Artifacts / truth

- Plan package: `specs/plans/080-failed-turn-partial-persistence/**` — `spec.md` (US1/US2 · FR-001…FR-007 · NFR-001…NFR-004 · SC-001…SC-005 · I-1…I-8 · S-1…S-8) · `checklists/requirements.md` · `features/acceptance/preserving-a-failed-turns-completed-work.feature` · `research.md` (D1–D7) · `plan.md` · `tasks.md` (T001–T012 + the fold ledger + the recorded narrowings) · `truth-delta.md`.
- Truth: `specs/truth/techstack.md` (the *Interrupted-turn persistence* row broadened + the *Session history store* clause extended) · `specs/truth/features/cli/chat/remembering-the-conversation.feature` (3 new Rules / 4 Examples) · `chat/dsl.md` (+3 Given / +5 Then rows + the round-080 note) · `contracts/**` + `data/**` NOOP.
- Domain model: `docs/domain-model/tellme.modelith.{yaml,md}` — `history-append-after-complete` + `turn-answer-stored-verbatim` **MODIFY** (extend the exception from *operator-interrupted* to *operator-interrupted or failed*; re-rendered; `modelith-check` green). A **recorded divergence** (the reference discards the partial turn in both cases).
- Code: `internal/domain/history/history.go` · `internal/cli/cli.go` · `internal/cli/failed_turn_test.go` + `internal/cli/interrupted_turn_test.go` (reworked) · `tests/e2e/fakeprovider/fakeprovider.go` · `tests/e2e/steps/step_r080_failed_turn.go` (+ the shared `readSessionEntries` helper in `history_fixture.go`).
- **ADR 0052** (`docs/decisions/0052-failed-turn-partial-persistence.md` + index).

### Falsifiability witnesses (reproduced then reverted)

`kept := false` ⇒ the two round-080 E2E Examples redden at `tellme stored the failed turn …` · `kept := true` ⇒ the zero-step Example reddens at `the session history is empty` · `Calls: 0` ⇒ the persisted-`calls` Then reddens · the keep line gated on `kept` instead of `stored` ⇒ the unit pin reddens (D6) · the failed-but-kept path returning `Success` ⇒ the E2E reddens (`exit code = 0, want 6`) · deleting the `ctx.Err() != nil` term ⇒ the `failTurn`-seam pin reddens (`code = 6, want 0`).

### Open items (non-blocking)

- **Round-080 forward items** — **RF-080-1** a small steady disk cost · **RF-080-2** a silent best-effort append failure on the failure path · **RF-080-3** no usage persisted for the failed-but-kept turn · **RF-080-4** the synthetic answers are stored as the `answer` · **RF-080-5** the exit-7 case is unit-carried · **RF-080-A** the `errors.Is` term is unreachable in production · **RF-080-B** nothing asserts what `-l` shows after a failed turn · **RF-080-7** no archive interaction. All in **ADR 0052 §Forward**.
- **Compaction**: the round-078 forward-item batch is now **compacted to a pointer row** (2 rounds old).
- **Issue tracker** — **[#161](https://github.com/gosharplite/tellme/issues/161) CLOSED (completed)** (fixed by round 080 / ADR 0052, PR [#162](https://github.com/gosharplite/tellme/pull/162)); the tracker is **0 open**.
- **PM follow-ups** — **none open.**
- Carried: PR #16 **Obs 1** stdout TTY probe **OPEN**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; round-011 forward items; the **5 pre-existing** Gherkin/DSL topology-audit errors.

### Next steps

1. Open the next round off `dev` via `/axb-specify` — a theme from **operator value** (the tracker is **0 open**; the tool surface is settled).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

*(Round 080 is fully closed out: PR #162 human-merged into `dev` (`4eec76e`, fast-forward); propagation `dev → main` **DONE (no-ff)**, tagged **`round-080`** with operator approval; the installed binary refreshed from the `dev` head; [#161](https://github.com/gosharplite/tellme/issues/161) closed.)*

### PM follow-ups

- None new (spec/acceptance complete; the `architect`'s F-080-1…3 were witness/record folds, not PM-owned gaps).

### Process notes (durable)

- **A fix that cannot fail under any test is not a fix (F-080-1).** The round's headline — the broadened predicate folding issue #161 hole #2 — had **no witness**: `runTurn` builds its own ctx, so no unit could see a cancelled ctx, and the round-079 E2E's real `SIGINT` made **both** predicate terms true. The reviewer proved it by mutation (deleting the new term left the whole suite green). The fold added a direct `failTurn`-seam pin (the function already takes `ctx`) — the seam's testability was the fix.
- **A claimed E2E assertion must have a carrier (F-080-2).** SC-001's persisted-`calls` claim had no E2E Then; `Calls: 0` left the E2E green. The fold added the carrier (and a class-swap-proof parameterisation).
- **A widening can create a new decided surface (F-080-3).** Broadening the interruption predicate introduced an interruption-over-failure precedence that the round did not record. Record it, or the next reader assumes the old semantics.
- **A best-effort side effect must not lie (D6).** The failed-turn append never masks the failure and never prints a keep line it did not earn — a deliberate divergence from the operator path (where exit 0 would be a lie if the work were lost).
- **Widening a predicate is how a seam becomes testable-unreachable (RF-080-A).** The union's `errors.Is` term is now unreachable in production; a future ctx seam would restore unit-testability of the shipping mechanism.
