# Session Summary — 2026-09-20

**Project**: `tellme` — a disciplined BDD re-creation of `tell-me-go`
**Repo**: `github.com/gosharplite/tellme`
**Status file**: [`STATUS.md`](../../../../../STATUS.md) *(back-link — the single live-state source)*
**Workspace**: `…/mbp-johndoe-niffler/ait-tellme` (`$TELL_ME_HOME`); darwin/arm64 host (Go 1.26.6).
**Session mode**: `butler`.
**Branch**: `dev` @ `a076b33` — round 064 delivered via **PR [#131](https://github.com/gosharplite/tellme/pull/131)** (merged **fast-forward**); round branch deleted (local + remote); propagation `dev → main` **DONE (no-ff, `8914fee`)**, tagged `round-064`.
**Status at end of day**: round **064** `064-prompt-suggestion-parity` **DELIVERED / FROZEN** — the `-i` suggestion **recent-prompt candidate pool deepens from the newest 10 distinct prompts to the newest 50** (the reference's `LoadTopN(ctx, 50)`), the surfaced list **still capped at 10**; **ADR 0034**; operator request (no anchor issue).

---

## 1. Session 39 — round 064 `064-prompt-suggestion-parity`: opened → full pipeline → review loop (4 passes) → merged (fast-forward) → closeout

Bootstrapped (`SESSION-BOOTSTRAP.md` Steps 1–8; round 063 delivered/frozen; active branch `dev`), then answered an operator question about the `-i` suggestion depth, opened round **064** as an **operator request**, ran the full AIxBDD pipeline, took **PR [#131](https://github.com/gosharplite/tellme/pull/131)** through a four-pass review loop to **CERTIFIED MERGE-READY**, saw the **human merge** (fast-forward), and ran `SESSION-CLOSEOUT.md` (Steps 1–8).

### At a glance

| Area | Outcome |
| --- | --- |
| Theme | operator request — **no anchor issue**: at the `-i` prompt, `commit` showed only **3** suggestions (the reference shows **10**); the cause is the candidate-pool depth |
| Clarify (one at a time) | **Q1 → A** — **pool-only** parity (deepen the pool to 50; the cap stays 10; no empty-query-first-5, no session source, so the CLI keeps its *no-startup-disk-I/O* property) |
| Pipeline | specify ✅ · clarify ✅ (Q1) · spec-by-example ✅ · technical-research ✅ (**ADR 0034** + `techstack.md`) · system-analysis ✅ (1 CLI interface; api/data NOOP) · dsl-refine ✅ · tasks ✅ (T001–T011) · implement ✅ |
| The change | `internal/app/suggestions/service.go` splits `maxSuggestions = 10` into **`promptPoolDepth = 50`** (history-source depth) + **`maxSuggestions = 10`** (surface cap); `addPrompts` requests `promptPoolDepth`. Nothing else changes |
| Review chain (PR #131) | review `5257404315` (APPROVE WITH REQUIRED FOLDS; TD-1…TD-4 + R-1…R-3) → fold `c8fd54a` → fold verification `5744921433` (**FOLDS VERIFIED**) → fold-back `1d1be61` (**R-4** + nits + the depth **literal pin**) → final verification `5744957939` (**CERTIFIED MERGE-READY**, loop CLOSED) → fold-back `c39ad21` (**R-5**) → STATUS `7c917b5`/`63aaa62` |
| Merge | PR [#131](https://github.com/gosharplite/tellme/pull/131) merged **`63aaa62`** (**fast-forward**); remote branch deleted → local branch deleted after an ancestor check |
| Closeout | `gofmt`/`go vet` clean · `go test -count=1 ./...` **green** (268/268 E2E) · `make verify` **OK** · topology audit **5 pre-existing, none new** (48 features · 381 module rows · 1962 steps) · diff secret scan clean · `STATUS.md` split (round-063 detail → `docs/archives/status/2026-09-20.md`) · **nothing to close** (operator request) |

### Decisions locked (round 064)

| # | Decision |
| --- | --- |
| **Q1 → A** | **Pool-only** behavioural parity: the recent-prompt candidate pool deepens to the newest **50** distinct prompts; the surfaced list stays **capped at 10**; no empty-query-first-5, no session source (keeps the round-016 *no-startup-disk-I/O* property). |
| **D1** (research) | One named owner for the depth: `promptPoolDepth = 50` and `maxSuggestions = 10` are **distinct** constants (the two concepts were conflated) + a **literal pin** (review R-4/RF-064-5) binding the parity value. |
| **D2** | Read shape unchanged — tellme keeps its lazy, result-bounded `Recent(ctx, n)` read (not the reference's pre-load). |
| **D3** | Source set / ordering / empty-query behaviour unchanged; the `-i` rendered surface untouched. |
| **D6** | **ADR 0034** records the depth + the kept divergences (tool source, `~/.tellme/` log, no `WorkspacePolicy`, no compaction) + the unadopted forward items. |

### Commits (branch `064-prompt-suggestion-parity`, then merged fast-forward to `dev` `63aaa62`)

| Commit | Note |
| --- | --- |
| `e45b112` | `docs(064)`: plan package + spec |
| `1306d07` | `docs(064)`: STATUS in flight |
| `5a02f72` | `docs(064)`: fold clarify **Q1 → A** (drop US2/FR-006…008) |
| `3e5c6b5` | `docs(064)`: acceptance + research + **ADR 0034** + `techstack.md` |
| `6fc1ba9` | `feat(064)`: implementation (T001–T011) — pool depth 50 |
| `a5b4957` | `docs(064)`: STATUS PR open |
| `c8fd54a` | `docs(064)`: fold review (TD-1…TD-4 + R-1…R-3) |
| `1d1be61` | `docs(064)`: fold-back R-4 + nits + the depth literal pin |
| `7c917b5` | `docs(064)`: STATUS fold verification |
| `c39ad21` | `docs(064)`: fold-back R-5 |
| `63aaa62` | `docs(064)`: STATUS final verification (merged fast-forward to `dev`) |

### Artifacts / truth

- Plan package: `spec.md` · `checklists/requirements.md` · `features/acceptance/finding-a-recent-prompt-beyond-the-shallow-window.feature` · `research.md` (D1–D6) · `plan.md` · `tasks.md` (T001–T011) · `truth-delta.md`.
- Truth: `techstack.md` MODIFY (the *Prompt suggestion engine* row: pool 50 / cap 10; the stale "active session" source clause corrected) · `chat/prompting-with-suggestions.feature` MODIFY (a depth Rule/Example **and** an executable cap Rule/Example) · `chat/dsl.md` MODIFY (2 Given rows + 1 negative Then row + the round-064 note) · `contracts/**` + `data/**` NOOP · domain model `tellme.modelith.{yaml,md}` (the session-source clause corrected; `modelith-check` green).
- Code: `internal/app/suggestions/service.go` · `internal/app/suggestions/service_test.go` (unit pin) · `tests/e2e/steps/step_r064_suggestions.go` + `step_r064_then_no_recent_prompt.go`.
- **ADR 0034** (`docs/decisions/0034-prompt-suggestion-pool-depth.md` + index).

### Falsifiability witnesses (reproduced then reverted)

`promptPoolDepth = 10` ⇒ the unit pin RED (the literal-pin diagnostic `promptPoolDepth = 10, want the reference's 50 (ADR 0034)`) + the E2E depth Example RED · `maxSuggestions = 20` ⇒ the E2E cap Example RED (the value-binding cap carrier) · `promptPoolDepth = 51` ⇒ the unit pin RED (the literal pin binds the parity value).

## 2. Open items (non-blocking)

- **Round-064 forward items** — **RF-064-1** full source parity (empty-query-first-5 + session source) · **RF-064-2** `WorkspacePolicy` parity · **RF-064-3** log self-compaction · **RF-064-4** pre-load vs per-query read · **RF-064-5** depth-constant review (attested by the literal pin) · **RF-064-6** a streaming prompt-log read · **RF-064-7** prompts-first crowding. All in **ADR 0034 §Forward**.
- **PM note** — round 064 repeated the comment-only-Rule class at authoring (its cap Rule was a comment) but **fixed it in-round** at review **TD-1** (the cap is now executable truth). **RF-063-10** stays open for the frozen round-063 journey.
- Carried: PR #16 **Obs 1**; round-006 **Obs 3**; sequential tools / no pruning / no `flock`; [#91](https://github.com/gosharplite/tellme/issues/91) / [#13](https://github.com/gosharplite/tellme/issues/13).

## 3. Next steps

1. Open round **`065-*`** off `dev` via `/axb-specify` (candidates: [#91](https://github.com/gosharplite/tellme/issues/91) self-development umbrella · [#13](https://github.com/gosharplite/tellme/issues/13) coverage tooling · a PM pass to retire **RF-063-10**).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

*(Round 064 is fully closed out: PR #131 human-merged into `dev` (`63aaa62`, fast-forward); propagation `dev → main` **DONE** (**no-ff**, `8914fee`), tagged **`round-064`** with operator approval; the installed binary was refreshed from the `dev` head.)*

## 4. PM follow-ups

- None new (spec/acceptance complete; the review's TD-1 folded the only PM-owned gap this round introduced).

---

## 5. Session 40 (2026-09-20, cont.) — round-063 closeout **live check** performed: single-image Vertex vision verified (RF-063-2 CLOSED); a pre-existing, media-agnostic multi-call defect surfaced and filed (#132)

The operator confirmed the round-063 live check by hand ("I have done the above. It works."), then asked for first-party evidence produced **through a peer agent** on the gemini `dev` provider. The check ran via the **`coder`** peer (per `tmg-chat-ingroup`: remote-party-marked prompt staged in `/tmp`, `env -u TELL_ME_MODE`, `TELL_ME_SELECTED_PROVIDER=dev` — `gemini-3.8-flash`, Vertex `websc-dev-433809`, `VISION: true`), with two tiny prepared PNGs.

### At a glance

| Area | Outcome |
| --- | --- |
| Run A — **1 × `read_image`** (control/baseline) | **PASS, `exit 0`** — coder: *"CONFIRM: read and attached. CONTENT: Solid red (#FF0000). ANOMALY: None."* → the model **sees** the image on the Vertex wire |
| Run B — **2 × `read_image`** in one round (the widened check) | **FAIL, `exit 6`** — both calls ran; the **next** Vertex request 400'd |
| Run C — **2 × `read_files`**, one round, **no media** | **FAIL, `exit 6`** — the **identical 400** ⇒ the defect is **media-agnostic** |
| 400 (verbatim) | `provider dev: provider returned status 400: Please ensure that the number of function response parts is equal to the number of function call parts of the function call turn.` |
| Root cause | `internal/agent/agentloop.go` appends **one `tool` message per call**; `gemini.buildContents` maps **each** to its **own** `user` turn with **one** `functionResponse` part — Vertex requires a function-call turn's responses in **one** turn |
| Provenance | per-call tool message = round **008** (`5a37fe4`); per-message `functionResponse` user turn = round **013** (`de79fc0`); round **063** (`e8e0880`) added only the `inlineData` branch. **Pre-existing** — not introduced by round 063 |
| Blast radius | **Gemini/Vertex only**, any round with **≥2 tool calls** (OpenAI-compatible unaffected — separate `role:"tool"` messages). Not hit when a round has exactly one tool call (the shipped single-image usage) |
| Filed | **[#132](https://github.com/gosharplite/tellme/issues/132)** (low priority) — the adapter must batch a round's `functionResponse` parts into one `user` turn (the ADR 0033 RF-063-7 "round-scoped placement") |

### Forward-item disposition

- **RF-063-2** — standalone-`user`-turn placement — **CLOSED (live-verified)** by run A.
- **RF-063-7** — multi-call round placement — **reframed**: moot for the shipped one-tool-call-per-round usage; the two-image variant 400s on the **media-agnostic** defect above ([#132](https://github.com/gosharplite/tellme/issues/132)).
- **RF-063-1** — the 14 MiB Gemini ceiling — **still open** (the 427-byte fixtures did not exercise the magnitude).

### Records

- `docs/decisions/0033-gemini-image-vision.md` — §Forward RF-063-1/2/7 annotated + the Verification *Live* bullet records the outcome (the ADR-0032 in-place forward-annotation precedent).
- `STATUS.md` — header (session 40) + the Round-063 forward-items line + a live-check bullet + the #132 tracker note.
- Evidence (transient, `/tmp`): `/tmp/tellme-livecheck/{single,send,ctrl}.{stdout,stderr}` + `{alpha-red,beta-blue}.png`.

### Next steps

1. Open round **`065-*`** off `dev` (candidates: [#91](https://github.com/gosharplite/tellme/issues/91) · [#13](https://github.com/gosharplite/tellme/issues/13) · RF-063-10 PM clean-up · the `ToolSetSpec` seam RF-062-10/RF-063-6). **[#132](https://github.com/gosharplite/tellme/issues/132)** is a low-priority candidate (fix only if parallel tool rounds are used).

---

## 6. Session 42 (2026-09-20, cont.) — round 065 `065-gemini-parallel-tool-calls`: anchor **#132** → full pipeline → architect review-fold loop (4 passes) → **human-merged (PR #133 → `dev` `b904281`)** → branch cleanup → closeout (Steps 1–8) + `go install`

A later session on the same calendar day: opened **round 065** from anchor issue **[#132](https://github.com/gosharplite/tellme/issues/132)** (the media-agnostic Gemini multi-call 400 the round-063 live check had found), ran the full AIxBDD pipeline, took **PR [#133](https://github.com/gosharplite/tellme/pull/133)** through a **review → fold → fold-verification → fold → final fold-verification** chain (the `architect` peer, initialised by `SESSION-BOOTSTRAP.md`), saw the **human merge** (a merge commit), deleted the branch (local + remote), and ran `SESSION-CLOSEOUT.md` Steps 1–8 with `go install`.

**Workspace**: `$TELL_ME_HOME` = `…/mbp-johndoe-niffler/ait-tellme`; **darwin/arm64** host (Go 1.26.6). **Session mode**: `butler`.

### At a glance

| Area | Outcome |
| --- | --- |
| Theme | anchor [#132](https://github.com/gosharplite/tellme/issues/132): a Gemini/Vertex turn with **≥2 parallel tool calls** 400'd; the adapter now **batches a round's tool results into one `user` turn** |
| Clarify | **not escalated (0 questions)** — the goal (close #132) and the fix were unambiguous; the residual choices were technical (`/axb-technical-research`) |
| Pipeline | specify ✅ · clarify ✅ (not escalated) · spec-by-example ✅ · technical-research ✅ (**ADR 0035** + `techstack.md` ×3) · system-analysis ✅ (1 CLI end; api/data NOOP) · dsl-refine ✅ · tasks ✅ (T001–T010) · implement ✅ |
| The change | `internal/infrastructure/llm/gemini/client.go` `buildContents` batches — a `model` turn's **N** calls ⇒ the round's **N** `functionResponse` parts in **one** `user` turn (call order); the round's media turns follow (round-scoped placement). **Family-local**: the loop, ports, tools, config, and the OpenAI-compatible wire are untouched |
| Review chain (PR #133, the `architect` peer) | review `5745379718` (APPROVE WITH REQUIRED FOLDS; **TD-065-1/2** + **R-065-1/2/3** + **N-065-1/2**) → fold `6ec36a8` → fold verification `5745407336` (**FOLDS VERIFIED WITH RESIDUALS** — the drop was correct only for `M ≥ N/2`) → fold `35a662c` (`pending[:0]` + the short-round pin) → re-verification `5745432548` (fix correct; carrier residual) → fold `588a2c5` (round-2 call renamed `search`) → **final fold verification `5745449629` — FOLDS VERIFIED WITH RESIDUALS (mergeable)** |
| Merge | PR [#133](https://github.com/gosharplite/tellme/pull/133) **human-merged** into `dev` (`b904281`, **merge commit**); remote branch deleted by the human, then the **local branch deleted** after an ancestor check (`git branch -d`, was `588a2c5`) |
| Issue | [#132](https://github.com/gosharplite/tellme/issues/132) was **manually closed (completed)** — the PR body's `Closes **#132**` did **not** auto-close (the bold markers defeated the keyword) |
| Closeout | `gofmt`/`go vet`/`go build` clean · `go test -count=1 ./...` **green** (271/271 E2E) · `make verify` **OK** · topology audit **5 pre-existing, none new** (49 features · 384 module rows · 1989 steps) · diff-level secret scan clean · `STATUS.md` split (round-064 detail + row + env note → `docs/archives/status/2026-09-20.md`) · `go install ./cmd/tellme` · **propagated `dev → main` (no-ff)** + tag **`round-065`** |

### Work done

1. **Grounding + round opened** — `/axb-specify` created `specs/plans/065-gemini-parallel-tool-calls/` off a new branch; the defect was already reproduced (round-063 live check).
2. **Pipeline** — acceptance Gherkin (3 Rules) → `/axb-technical-research` (D1–D9 + **ADR 0035** + `techstack.md` ×3 rows) → `/axb-system-analysis` (1 CLI end; api/data NOOP) → `/axb-dsl-refine` (`chat/calling-several-tools-in-one-round.feature` + `dsl.md` round-065 rows) → `/axb-tasks` → `/axb-implement`.
3. **The fix** — `pending = pending[:0]` at a round boundary; a family-local batching in `buildContents`; new unit pins + the E2E journey. Witness reproduced and reverted.
4. **The review-fold loop (the `architect` peer)** — 7 required folds, then two *real* residual catches (the `M ≥ N/2` arithmetic and the coincident cookie name) fixed; final verdict mergeable.
5. **Merge + cleanup + closeout + `go install`** — PR #133 merged (`b904281`); branch deleted (local + remote); `SESSION-CLOSEOUT.md` Steps 1–8.

### Decisions locked (round 065)

| # | Decision |
| --- | --- |
| Fix shape | A model round's **N** tool results share **one** `user` turn (call order); the pairing stays the FIFO name match (`ToolCallID` optional); the round's media turns follow the batch |
| Family-local | The change is confined to `internal/infrastructure/llm/gemini`; the OpenAI-compatible wire, the loop, ports, tools and config are untouched |
| Byte/shape identity | **I-1/I-2/I-3** — the OpenAI wire, the single-call Gemini path, and the media-free text path are unchanged |
| ADR | **ADR 0035** records it; extends **ADR 0033**, **supersedes its RF-063-7**, annotates its D2 note |

### Commits (branch `065-gemini-parallel-tool-calls`, then merged)

| Commit | Note |
| --- | --- |
| `f96e8d8` | `docs(065)`: plan package + spec |
| `4c2b78b` | `docs(065)`: acceptance + research + ADR 0035 + techstack truth |
| `e7903d5` | `feat(065)`: batch a Gemini round's tool results into one `user` turn |
| `6ec36a8` | `fix(065)`: fold the architect review (TD-065-1/2 + R-065-1…3 + N-065-1/2) |
| `35a662c` | `fix(065)`: fold the fold-verification residual (drop all unpaired names) |
| `588a2c5` | `test(065)`: strengthen the short-round pin (round-2 call renamed) |
| `b904281` | PR [#133](https://github.com/gosharplite/tellme/pull/133) merge into `dev` (by the operator) |
| *(this closeout, on `dev`)* | `docs(065)`: day close — round 065 delivered + propagated; STATUS split + 09/20 summary §6 |

### Verification (2026-09-20, on `dev` @ `b904281`)

- `gofmt -l .` clean · `go vet ./...` clean · `go build ./...` clean · `go test -count=1 ./...` **green** (24 packages incl. the godog E2E, **271/271 scenarios**) · `make verify` **OK** (arch gate 0 · modelith-check ×3 · lint 0 · govulncheck clean · cross-compile 4/4) · topology audit **5 pre-existing, none new** (49 features · 384 module rows · 1989 steps) · `go.mod`/`go.sum` unchanged · diff-level secret scan clean.
- **`go install ./cmd/tellme`** refreshed → `$(go env GOPATH)/bin/tellme`; `--version` → `dev`.

### Open items (non-blocking)

- **Round-065 live check (pending, non-gating)** — one real Vertex turn with two tool calls (not run hermetically).
- **RF-065-1…5** in **ADR 0035 §Forward**; the review residual (the `N=2 M=1` sub-case's inherent equivalence) recorded + accepted as mergeable.
- **RF-063-7** — **SUPERSEDED by ADR 0035** (round 065); **#132 closed**.

### Next steps

1. Open round **`066-*`** off `dev` (candidates: [#91](https://github.com/gosharplite/tellme/issues/91) self-development umbrella · [#13](https://github.com/gosharplite/tellme/issues/13) coverage tooling · the `ToolSetSpec` seam RF-062-10/RF-063-6 · RF-063-10 the meta-Rule clean-up).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

*(Round 065 is fully closed out: PR #133 human-merged into `dev` (`b904281`, merge commit); propagation `dev → main` **DONE (no-ff, `3b9bf8c`)**, tagged **`round-065`**; the installed binary refreshed; [#132](https://github.com/gosharplite/tellme/issues/132) closed.)*

---

## 7. Session 43 (2026-09-20, cont.) — round-065 live check **PERFORMED AND PASSED**: one real Vertex turn with two tool calls in one round completes cleanly (the #132 fix witnessed on a real endpoint)

A later session on the same calendar day: the round-065 closeout left exactly one open, non-gating item — the **live re-check** (a real Vertex/Gemini turn with two tool calls), which the hermetic closeout could not run for want of a live credential. At the operator's direction the check was run **through the `coder` peer on the `dev` provider** (per `tmg-chat-ingroup`: marker-led prompt staged in `/tmp`, `env -u TELL_ME_MODE`, `TELL_ME_SELECTED_PROVIDER=dev`), and the outcome was **recorded durably** (ADR 0035 + `STATUS.md`).

### At a glance

| Area | Outcome |
| --- | --- |
| Theme | witness the **round-065 / ADR 0035** batching fix on a **real Vertex** endpoint (the one open, non-gating closeout item) |
| Method | `tmg-chat-ingroup` orchestrator side: staged a remote-party-marked prompt in `/tmp/tellme-prompt.K5stxk` (`write_file`, not a heredoc) → `env -u TELL_ME_MODE TELL_ME_HOME=$WS TELL_ME_SELECTED_PROVIDER=dev tellme --new -r -c $WS/configs/coder.yaml < $STAGING` → retrieve via `-l 1` → `rm` the staging file |
| Provider | `dev` = `gemini-3.8-flash`, **Vertex** (real endpoint), as **explicitly requested** by the operator (the peer otherwise inherits the orchestrator's own model) |
| The turn | **2 tool calls in ONE round**, same instant `05:44:39`: `read_files([note1.txt, note2.txt])` + `list_files(/tmp/tellme-livecheck)` — the media-free pair that 400'd as *Run C* in the round-063 live check |
| Result | **PASS** — follow-up request completed (`Payload: 11890/1000000 tokens`), **`exit 0`**, self-reported **`ANOMALY: None`**, **zero `400`s**; chrome header read `coder - gemini-3.8-flash`; `wc -c output/coder/history.jsonl` non-zero |
| Fixtures honoured | `NOTE1: ALPHA note: the sky is green.` · `NOTE2: BETA note: the sky is orange.` (read from `/tmp/tellme-livecheck/`) |
| Recorded | **ADR 0035** — `## Verification` *Live* bullet now carries the outcome; §Forward gains a *Live check — CLOSED (verified)* line. **`STATUS.md`** — header (session 43) + the open item flipped **pending → DONE (live-verified)** + the round-065 env note. Evidence (transient): `/tmp/tellme-livecheck/r065_send.{stdout,stderr}` |

### Why this closes it

The round-063 live check had produced a three-run matrix whose **Run C** — *2 × `read_files`, no media, one round* — failed with the verbatim `400 … Please ensure that the number of function response parts is equal to the number of function call parts …` (a defect proven **media-agnostic**, per-call `tool` message since round 008). Round 065 (ADR 0035) batched a model round's `functionResponse` parts into **one** `user` turn. This session re-ran the identical *Run C* shape on a real Vertex endpoint and it now completes cleanly — the fix is verified **on the wire**, not only by the unit pin + fake-provider E2E.

### Decisions / records

| # | Item |
| --- | --- |
| — | The provider override (`dev`) was used **only** because the operator explicitly named it (the `tmg-chat-ingroup` §5 rule); the check is otherwise a standard peer dispatch. |
| — | The live-check outcome is homed on **durable surfaces** — **ADR 0035** (the decision record) + **`STATUS.md`** (live state) — mirroring the session-40 RF-063-2 precedent. |
| — | **No** product code, truth, or `specs/plans/**` file changed (the round-065 package stays frozen); the **only** remaining round-065 forward items are **RF-065-1…5** (ADR 0035 §Forward), all untouched by this check. |

### Next steps

1. Open round **`066-*`** off `dev` (candidates: [#91](https://github.com/gosharplite/tellme/issues/91) self-development umbrella · [#13](https://github.com/gosharplite/tellme/issues/13) coverage tooling · the `ToolSetSpec` seam RF-062-10/RF-063-6 · RF-065-1 the `ToolCallID` pairing · RF-063-10 the meta-Rule clean-up).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

