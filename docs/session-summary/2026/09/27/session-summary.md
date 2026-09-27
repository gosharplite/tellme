# Session Summary — 2026-09-27

**Project**: `tellme` — a disciplined BDD re-creation of `tell-me-go`
**Repo**: `github.com/gosharplite/tellme`
**Status file**: [`STATUS.md`](../../../../../STATUS.md) *(back-link — the single live-state source)*
**Workspace**: `…/mbp-johndoe-tellme/ait-tellme` (`$TELL_ME_HOME`); darwin/arm64 host (Go 1.26.6).
**Session mode**: `butler`.
**Branch**: `094-dead-code-hygiene` (off `dev` `ae300e9`); **PR [#199](https://github.com/gosharplite/tellme/pull/199) open**.
**Status at end of session**: round **094** `094-dead-code-hygiene` **PR OPEN** — remove the genuinely-dead exported symbols no shipped gate can see + add an **advisory** `make dead-code` carrier (never fails); anchor issue [#198](https://github.com/gosharplite/tellme/issues/198) (DoD = close it); **ADR 0064** (supersedes ADR 0042 §D5's wording; the §D4 percentage-coverage decline stands; corrects the `techstack.md` reachability clause — `RF-068-1` stays retired).

---

## 1. Session 84 — bootstrap, read #198, then round 094 `094-dead-code-hygiene` (anchor #198) → full AIxBDD pipeline → **PR [#199](https://github.com/gosharplite/tellme/pull/199) open**

The session began with a bootstrap (`SESSION-BOOTSTRAP.md` Steps 1–8; round 093 delivered/frozen;
active branch `dev`, tree clean at `ae300e9`), then the operator asked to read issue **#198**, then
directed *"Open a new aixbdd round, the goal is to close https://github.com/gosharplite/tellme/issues/198."*
A new branch **`094-dead-code-hygiene`** was created off `dev` `ae300e9`, the full AIxBDD pipeline ran,
and **PR [#199](https://github.com/gosharplite/tellme/pull/199)** was opened.

### At a glance

| Area | Outcome |
| --- | --- |
| Branch | **`094-dead-code-hygiene`** (off `dev` `ae300e9`) |
| Anchor | issue [#198](https://github.com/gosharplite/tellme/issues/198) — remove the genuinely-dead exports + add an advisory `make dead-code` carrier; **DoD = close it** |
| Theme | **Cleanup + carrier**: remove the exported-dead class the shipped gates cannot see, and give it a mechanical surface (advisory, never fails) |
| Clarify | **not escalated (0 questions)** — #198 locks the goals + the two decisions (D1 carrier(ii); D2 delete `EffectiveUseTUIPrompt`) |
| Pipeline | specify ✅ · spec-by-example **NOOP** (a Make target is not godog-drivable; no behaviour change) · technical-research ✅ (**ADR 0064** + the `techstack.md` rows) · system-analysis ✅ (the “interface” is the Makefile; api/data/UI NOOP) · dsl-refine **NOOP** · tasks ✅ (T001–T019 + the Claim→Witness ledger) · implement ✅ |
| The change | 8 grep-verified dead-symbol removals (`config.EffectiveUseTUIPrompt`, `openai.NewWithHTTPClient`, `agent.DefaultToolTimeout`, `di.TokenResolver`, `harness.RunWithStdin`/`RunBinary`, `mcptest.SchemaWithProperty`, `cli.noopCallObserver`) + the advisory `make dead-code` target + **ADR 0064** + index/back-pointer + `techstack.md` rows + the quality-model advisory member |
| Verification | `make verify` **OK** · `go test -count=1 ./...` **green** — E2E **330 scenarios · 2487 steps (unchanged)** · `make test-race` **no data races** · `gofmt`/`goimports`/`go vet`/`staticcheck`/`golangci-lint` clean · `go.mod`/`go.sum` unchanged |
| Witnesses (reproduced then reverted) | **W1** inject a synthetic dead export ⇒ `make dead-code` reports it + exits **0**; revert ⇒ clean (empty, exit 0) · **W2** the `verify:` aggregate line is unchanged (grep) · **W3** every removed symbol absent (grep) + gates green · **EC-001** absent tool ⇒ hint + exit 0 |
| Delivery | branch `094-dead-code-hygiene` → **PR [#199](https://github.com/gosharplite/tellme/pull/199) open** (no Copilot review; only a human merges) |

### The re-verification (do not trust a tool verdict alone)

The issue's inventory (7 symbols) was re-verified by **grep** and by the tool. Two measured
divergences, both recorded:

1. **The tool's clean-tree output** — vanilla `golang.org/x/tools/cmd/deadcode@v0.47.0 -test ./...`
   reports, on `dev` `ae300e9`: `cli.noopCallObserver.OnCallBegin/OnCallEnd`, `openai.NewWithHTTPClient`,
   `mcptest.SchemaWithProperty`, `ui.sharedSource.Suggest`, `harness.RunWithStdin`, `harness.RunBinary`.
   The issue's assumed "Unwrap-class" FP is **not** surfaced by `-test` (the RTA sees `errors.Is/As`).
2. **An added removal** — `cli.noopCallObserver` (a **production** symbol the issue's inventory missed)
   is grep-verified **never instantiated** (`compositeObserver` is built with a real renderer,
   `cli.go:875`); its only reference is its own `var _` conformance. Removed (a truly-dead type is
   removed, never filtered).
3. **The recorded FP** — `ui.sharedSource.Suggest` is **interface-conformance-only** (pinned by two
   compile-time `var _ I = T{}` assertions that guard the identical method sets of `domaintui.Source`
   and `prompt.Source`); **kept + filtered**, never deleted.

### Decisions locked (round 094 / ADR 0064)

| # | Decision |
| --- | --- |
| **D1** | Carry the exported-dead class with an **advisory** `make dead-code` (vanilla `x/tools/cmd/deadcode` pinned `v0.47.0`, a PATH dev-tool — **no** `go.mod`/`go.sum` change — run with `-test`), **never fails**, **not** a `make verify` member. |
| **D2** | "Starting state = clean": remove the dead set **first**, then add the target. |
| **D3** | The FP policy = **one documented exclusion predicate** (the **interface-conformance-only** class — the single recorded member `sharedSource\.Suggest`; **not** a name-wide `Unwrap` alternative), so a clean tree **reports no findings**; **not** a `NonFixCatalog` (avoids the ADR 0041 advisory-muse failure). |
| **D4** | An **absent** tool prints an install hint + exits **0** (a deliberate divergence from the fail-on-absent `modelith-check` gate). |
| **D5** | The carrier records the **provenance hazard** (the PATH name `deadcode` collides with the reference's heavy `tell-me-go/cmd/deadcode`): the vanilla x/tools binary is required. |
| **D6** | The round **supersedes ADR 0042 §D5's "no `dead-code`"** wording (the §D4 **percentage**-coverage decline **stands**); ADR 0042's body stays **verbatim**; its **index row** gains a forward pointer. |
| **D7** | `config.EffectiveUseTUIPrompt` is **deleted** (the inline duplicate in `cli.tuiRequested` is not a drop-in — it returns `true` for `-i` without loading config) and recorded as a single-owner candidate in ADR 0064 §Forward. |

### Commits (branch `094-dead-code-hygiene`)

| Commit | Note |
| --- | --- |
| `6322a16` | `feat(094)`: dead-code hygiene — remove the genuinely-dead exports + an advisory `make dead-code` carrier (ADR 0064) |
| *(this record, on the branch)* | `docs(094)`: STATUS + day log — pipeline complete; PR #199 open |

### Open items (non-blocking)

- **PR [#199](https://github.com/gosharplite/tellme/pull/199)** awaits a human review/merge → then the closeout (`SESSION-CLOSEOUT.md`): propagate `dev → main` (no-ff), tag **`round-094`**, refresh the binary; **close [#198](https://github.com/gosharplite/tellme/issues/198)**.
- **ADR 0064 §Forward** RF-064-1…6 (advisory-never-a-gate · the class-based predicate · provenance not machine-checked · host-GOOS only · structural-typing FPs inherent · the same-PR model update).
- **STATUS drift (recorded, operator-owned)** — `STATUS.md` said the tracker was `0 open`, but it holds **#194–#198** (from 2026-09-26/27). This round closes #198 at merge; the remaining live seeds (#194–#197) are round candidates (Bootstrap Agent Rule 11).

### Next steps

1. Human reviews + merges the round-094 PR; then the closeout (propagate `dev → main` no-ff, tag `round-094`; close #198).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `094-dead-code-hygiene` until merged, then `dev`).

### PM follow-ups

- None new (spec/acceptance complete; the round carries the falsifiable witnesses).

---

## 2. Session 84 (cont.) — round 094 review-fold loop (PR [#199](https://github.com/gosharplite/tellme/pull/199)) → `APPROVE WITH REQUIRED FOLDS` → folded

Dispatched the `architect` peer per `tm-chat-ingroup`: initialized **once** with `SESSION-BOOTSTRAP.md`
(`--new`), then continuations (no `--new`). The architect reproduced the gates on two out-of-tree
worktrees (`/tmp/pr199` head, `/tmp/pr199base` `dev` `ae300e9`) + a vanilla `deadcode@v0.47.0` in a
temp GOBIN, attacked the carrier by mutation, and independently verified the cleanup. Verdict:
**`APPROVE WITH REQUIRED FOLDS`** — no `[ARCHITECTURAL BLOCKER]`; every finding is **claim-accuracy /
one-line-recipe**. All were folded:

| Fold | Resolution |
| --- | --- |
| **F-094-1** the no-`-test` figure "~330" does not reproduce | Restated as the **measured 1151** (head) / **1157** (base) with the command (`spec.md` §1, `research.md` D2, ADR D2, this log). |
| **F-094-2** "a clean tree prints nothing" is **false in situ** (the dev host's PATH `deadcode` is the heavy tool ⇒ 284 lines, exit 0) + the banner asserted an unverified provenance | The target now **probes the binary's provenance** (`go version -m` → `path` must be `golang.org/x/tools/cmd/deadcode`) and **skips** (exit 0, naming the vanilla route); the claim restated as "**reports no findings**" across FR-4/SC-002/CLM-004/ADR/PR body. |
| **F-094-3** the `.*\.Unwrap$` alternative **swallows a genuine new finding** | Dropped it → the predicate is the single recorded member; the over-match recorded in ADR D5 / RF-064-2 (a live unwrap FP, if ever surfaced, is added as a **recorded symbol**, never a name-wide alternative). |
| **F-094-4** a **tool failure** is reported as `✓ no unreachable functions found` | The target now **checks the tool's exit status** and prints `analysis did not run (tool exit N)` (still exit 0). |
| **F-094-5** `techstack.md:172` left the **falsified** reachability clause present-tense | Reconciled **in place** (asserted → **falsified 2026-09-27 by round 094 / ADR 0064**). |
| **F-094-6** the governance banner mis-cited **`RF-068-1`** (it is the *unpaired-call diagnostic's E2E carrier*, not reachability) | The reopen restated precisely (supersede ADR 0042 §D5's wording; correct the `techstack.md` clause; `RF-068-1` stays **retired and inert**), aligned with D8. |
| **TD-094-1 / TD-094-2** | The predicate is **name-keyed + un-witnessed** (RF-064-2); the carrier has **no invocation occasion** → named as an on-demand **closeout / round-open** step (RF-064-7 + a `STATUS.md` env-note). |
| **N-094-1..4** | "prints nothing" → "reports no findings"; the 0042 index row names §D4's orphan half; the Makefile header aligned to D8; `DEADCODE_FP` quoted in the ADR + `research.md`. |
| **R-094-1 / R-094-2** | The heavy binary stays at `$GOPATH/bin/deadcode` → the `STATUS.md` env-note names the required provenance; the architect's race sweep was scoped to the 8 touched packages (the round's full-suite no-race claim stands). |

Witnesses re-measured at the fold head: clean tree (vanilla on PATH) → **no findings**, exit 0 · the
dev host's default PATH (heavy binary) → **skip**, exit 0 · injected dead `Unwrap` → **surfaces** ·
broken tree → **`analysis did not run (tool exit 1)`** · W1 top-level synthetic → reported, exit 0.

**State**: **PR [#199](https://github.com/gosharplite/tellme/pull/199) is ready for a human to review
and merge** (head `fb21879`; no Copilot review; only a human merges). On merge: propagate `dev → main`
(no-ff), tag `round-094`, refresh the binary, and **close [#198](https://github.com/gosharplite/tellme/issues/198)**.

### 2 (cont.) — review-fold loop CLOSED (the `architect` peer; 3 verification passes)

Loop history: `review` @ `77db4c1` → **`APPROVE WITH REQUIRED FOLDS`** (F-094-1..6 + TD-094-1/2 +
N-094-1..4 + R-094-1/2) → `fold` `27e122a` → fold-verification → **`FOLDS VERIFIED WITH RESIDUALS`**
(RES-094-FV-1..4 + N-094-FV-1/2) → residual fold `f9112b4` → residual verification → **WITH
RESIDUALS** (RES-094-FV-5 + N-094-FV-3) → residual fold 2 `fb21879` → **final verification →
`FOLDS VERIFIED — LOOP CLOSED`** (no residuals, no nits) — [`5852572109`](https://github.com/gosharplite/tellme/pull/199#issuecomment-5852572109).
The loop is **CLOSED**; the PR is ready for a human to merge.

---

## 3. Session 84 closeout (2026-09-27) — round 094 `094-dead-code-hygiene` **DELIVERED / FROZEN** (`SESSION-CLOSEOUT.md` Steps 1–8)

Round 094 was human-merged (PR [#199](https://github.com/gosharplite/tellme/pull/199) → `dev`
**`efdc614`**, **merge commit**); the remote branch was already gone (`git fetch --prune` →
`[deleted] origin/094-dead-code-hygiene`), the local branch tip (`dc8f9b6`) was an ancestor of
`origin/dev`, so the **local branch was deleted** (`git branch -d 094-dead-code-hygiene`, was
`dc8f9b6`), `dev` was fast-forwarded to the merge, and `SESSION-CLOSEOUT.md` Steps 1–8 ran.

| Step | Outcome |
| --- | --- |
| **1 — working tree** | `dev` clean; `dev == origin/dev == efdc614`; no delivered `specs/plans/**` package touched (093's package unmodified); no stray temp files (the architect staging files cleaned); round branch already deleted (local + remote) |
| **2 — gates** | **`make check` OK** (`make verify` OK + `go test -count=1 ./...` green) · E2E **330 scenarios · 2487 steps** · `make test-race` **no data races** · `verify-architecture` 0 issues · `verify-adr-index` consistent · `modelith-check` no drift ×3 · diff-level secret scan clean (prose-only matches: "token"/"TokenResolver") |
| **3 — STATUS.md** | header → round 094 **DELIVERED / FROZEN**; **Rule-12 split**: the **round-093 delivered-round detail + its env note** relocated **verbatim** into [`docs/archives/status/2026-09-27.md`](../../../../archives/status/2026-09-27.md); round-094 section added; delivered-rounds pointer → 001–094; roadmap candidates → **#194–#197** (round 094 closed #198); the round-094 env note + the round-close-tags line (`round-094` at this propagation) refreshed; **58 lines** (live state only) |
| **4 — daily summary** | this §3 (closeout) appended (the §1 + §2 records preserved) |
| **5 — reconcile** | `STATUS.md` ↔ this summary agree: no round in flight, `dev` active, branch heads match, tracker → **#198 closed at merge; #194–#197 open** |
| **6 — commit** | working `dev` committed + pushed |
| **7 — propagate + hand off** | `dev → main` (**no-ff**), tagged **`round-094`**; installed binary refreshed (`go install ./cmd/tellme`) |
| **8 — issue tracker** | **[#198](https://github.com/gosharplite/tellme/issues/198) CLOSED** (completed) with a linking comment naming PR #199 / `efdc614`; **[#194](https://github.com/gosharplite/tellme/issues/194)–[#197](https://github.com/gosharplite/tellme/issues/197)** verified **open** (live round seeds) — nothing to revise |

### Commits (branch `094-dead-code-hygiene`, then merged)

| Commit | Note |
| --- | --- |
| `6322a16` | `feat(094)`: dead-code hygiene — remove the genuinely-dead exports + an advisory `make dead-code` carrier (ADR 0064) |
| `77db4c1` | `docs(094)`: STATUS + day log — round 094 pipeline complete; PR #199 open (anchor #198) |
| `27e122a` | `fix(094)`: fold the architect review (F-094-1..6 + TD-094-1/2 + N-094-1..4 + R-094-1) |
| `f9112b4` | `fix(094)`: fold the fold-verification residuals (RES-094-FV-1..4 + N-094-FV-1/2) |
| `fb21879` | `fix(094)`: fold the fold-verification-2 residuals (RES-094-FV-5 + N-094-FV-3) |
| `dc8f9b6` | `docs(094)`: review-fold loop CLOSED — record the loop history + verdict |
| `efdc614` | PR [#199](https://github.com/gosharplite/tellme/pull/199) merge into `dev` (by the human, merge commit) |
| *(this closeout, on `dev`)* | `docs(094)`: day close — round 094 delivered + propagated; STATUS split + 09/27 summary §3 |

### Open items (non-blocking)

- **None new.** `STATUS.md` carries no open-items index: a deferred item lives in its `ADR 00NN §Forward` or a live GitHub issue. **ADR 0064 §Forward RF-064-1…7** are disclosures, not tasking.
- **Issue tracker**: **#194–#197 open** (live round seeds); **#198 closed**.

### Next steps

1. Open the next round off `dev` via `/axb-specify` — a theme from **operator value or a live issue** (the live seeds are [#194](https://github.com/gosharplite/tellme/issues/194)–[#197](https://github.com/gosharplite/tellme/issues/197); Bootstrap Agent Rule 11).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `dev`).

### PM follow-ups

- None new (spec/acceptance complete; the round carries the falsifiable witnesses).

*(Round 094 is fully closed out: PR #199 human-merged into `dev` (`efdc614`, merge commit); propagation `dev → main` **DONE (no-ff)**, tagged **`round-094`**; the installed binary refreshed; [#198](https://github.com/gosharplite/tellme/issues/198) closed.)*



### Process notes (durable)

- **Re-verify the inventory; do not trust a tool verdict alone.** The issue's "Unwrap-class" FP proved
  miscalibrated for the vanilla `-test` run; the measured clean-tree FPs are the **interface-conformance-only**
  class. Measure the clean tree, then design the predicate to it.
- **A truly-dead production type is removed, not filtered.** `cli.noopCallObserver` (uninstantiated)
  belongs in the cleanup; filtering it would hide a real finding (the advisory-muse failure).
- **`-test` is mandatory.** Without it the whole test surface reads as dead (**1151** items on the
  head `77db4c1`; **1157** on `dev` `ae300e9`); with it, the
  clean tree reduces to the recorded FPs.

---

## 4. Session 84 (cont.) — round 095 `095-modelith-upstream-main-route` **OPENED → full pipeline → PR open** (anchor issue [#200](https://github.com/gosharplite/tellme/issues/200); **ADR 0065**)

After the round-094 closeout, the operator filed issue **#200** (the `modelith` dev-tool MUST be installed
from the HEAD of the upstream `stacklok/modelith` `main` branch) and directed *"Open a new aixbdd round, the
goal is to close #200."* A new branch **`095-modelith-upstream-main-route`** was created off `dev` `248846c`,
the (toolchain/record-only) pipeline ran, and a PR was opened.

### At a glance

| Area | Outcome |
| --- | --- |
| Branch | **`095-modelith-upstream-main-route`** (off `dev` `248846c`) |
| Anchor | issue [#200](https://github.com/gosharplite/tellme/issues/200) — install `modelith` from the upstream `main` HEAD; **DoD = close it** |
| Theme | **Toolchain/record**: the route becomes `go install github.com/stacklok/modelith/cmd/modelith@main`; revise every **live** reference; **supersede ADR 0030 D2** (ADR 0065) |
| Clarify | **not escalated (0 questions)** — the directive locks the route; residual choices RD-owned |
| Pipeline | specify ✅ · spec-by-example **NOOP** (no behaviour change) · technical-research ✅ (**ADR 0065** + the `techstack.md` row) · system-analysis ✅ (1 toolchain end; api/data/UI NOOP) · dsl-refine **NOOP** · tasks ✅ (T001–T014 + the Claim→Witness ledger) · implement ✅ |
| The change | `Makefile` (`MODELITH_REF := main` / `MODELITH_INSTALL := go install github.com/stacklok/modelith/cmd/modelith@$(MODELITH_REF)`; `MODELITH_PIN` removed) · `docs/domain-model/README.md` (*Toolchain → Install*) · `specs/truth/techstack.md` (*Domain model* row, in place) · **ADR 0065** + `docs/decisions/README.md` (the 0065 index row + an ADR 0030 forward pointer) |
| Facts verified | upstream `main` HEAD = `9008354f19ff…` (`v0.5.1-0.20260927062055-9008354f19ff`); `go install …@main` builds it; `make modelith-check` green with it |
| Verification | `make verify` **OK** (adr-index consistent · modelith-check no drift ×3) · `go test -count=1 ./...` **green** — E2E **330 scenarios · 2487 steps (unchanged)** · `go.mod`/`go.sum` unchanged · no product code |
| Witnesses | **W1** absent `modelith` ⇒ `make modelith-check` exits non-zero **printing the new route** · **W2** `go install …@main` resolves/builds + gate green · **W3** no **live** fork/pin reference remains on the operational/shipping surfaces (the expected full-tree hits are the round package + ADR 0065 + the **ADR 0065 index row** `docs/decisions/README.md:91` + the immutable ADR 0030 body + frozen history) · **W4** `verify-adr-index` consistent |
| Delivery | branch `095-modelith-upstream-main-route` → **PR [#201](https://github.com/gosharplite/tellme/pull/201) open** (no Copilot review; only a human merges) |

### Decisions locked (round 095 / ADR 0065)

| # | Decision |
| --- | --- |
| **D1** | The route is the **upstream `main` HEAD** — `go install github.com/stacklok/modelith/cmd/modelith@main`; the fork + commit-pin clone route is superseded. |
| **D2** | The tracked ref is **`main`** (`MODELITH_REF`); the current HEAD hash (`9008354f19ff`) is recorded as **provenance, not a pin** (`go.mod`/`go.sum` unchanged; dev-tool binary). |
| **D3** | The accepted cost: tracking `main` is **non-hermetic** (a `main` advance may red `modelith-check` with no repo change — the ADR-0012 spurious-red class), **re-adopted deliberately**; recorded in the README, the Makefile comment, the truth row, and the ADR. |
| **D4** | **ADR 0030 D2 is superseded, not repealed** — its body stays verbatim; its index row gains a forward pointer; D3/D4/D5/D6 stand. |
| **D5** | The install route is **not modelled** — no `docs/domain-model/*.modelith.{yaml,md}` change (ADR 0041 escape hatch; recorded in `plan.md` §5). |
| **D6** | Frozen history (`specs/plans/060-*`, the 09/19 archive/summary) is **out of scope** (`plan-package-frozen` / Rule 12). |

### Open items (non-blocking)

- **PR** for round 095 awaits a human review/merge → then the closeout (`SESSION-CLOSEOUT.md`): propagate
  `dev → main` (no-ff), tag **`round-095`**, refresh the binary; **close [#200](https://github.com/gosharplite/tellme/issues/200)**.
- **ADR 0065 §Forward** RF-065-1…4 (the moving `@main` ref · provenance not machine-checked · the
  release-binary route not adopted · ADR 0030's own fork prose corrected forward only).

### Next steps

1. Human reviews + merges the round-095 PR; then the closeout (propagate `dev → main` no-ff, tag
   `round-095`; close #200).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `095-modelith-upstream-main-route` until
   merged, then `dev`).

### PM follow-ups

- None new (no PM-owned requirement gap; the round is a toolchain/record reconciliation).

### Process notes (durable)

- **A dev-tool route is a `TruthArtifact` clause, not just prose.** The `techstack.md` *Domain model* row
  is truth, so the route change is a **round** (truth + a superseding ADR), not a docs-only hygiene commit.
- **Supersede, don't edit.** ADR 0065 records the new route; ADR 0030's fork prose stays verbatim with an
  index forward pointer — the same shape as ADR 0064→0042.
- **A witness must exclude its own records.** W3's grep legitimately matches the round's plan package,
  ADR 0065 (which records what it supersedes), the ADR index row (`docs/decisions/README.md` — names the
  superseded fork), the immutable ADR 0030 body, and frozen history — the discriminating claim is scoped to
  the **operational/shipping** surfaces (README · Makefile · truth); the full exclusion set is recorded in
  `tasks.md` T010 (TD-095-1 / ADR 0065 RF-065-5).
