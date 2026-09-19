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
