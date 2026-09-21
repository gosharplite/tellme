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
