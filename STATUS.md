# tellme — Status

**Last updated**: 2026-09-11 (**session 13 — roadmap 003/004 ([#9](https://github.com/gosharplite/tellme/issues/9), [#10](https://github.com/gosharplite/tellme/issues/10)) + docs reorg (session-summary path · STATUS split + archive · decisions under docs/)** · session 12 closeout — round 002 `002-followup-cleanups` delivered, merged, propagated**: removed the `--json` flag (reverses round-001 FR-013); froze the operator-facing contract — the failure **class phrase** (`tellme: {phrase}`; the tail is contract-free) + exit codes `0/2/3/4/5` (the pin is falsifiable via `TestExitCodesMatchPinnedContract`); added F9 pure-helper unit tests and the `golangci-lint`/`govulncheck` gates. Delivered as **PR [#7](https://github.com/gosharplite/tellme/pull/7)** — two architecture-review rounds (F1–F7, N1–N3) **plus Grill Round #6** (issue [#8](https://github.com/gosharplite/tellme/issues/8), *closed*) — **merged** (`f2a058f`, head branch deleted) and **propagated** `002-followup-cleanups → dev → main` (`dev` `2b50cfd`, `main` `7caa4f1`); `make verify` OK · godog **20/20** · orphan sweep 0 · topology audit 126 steps; `SESSION-CLOSEOUT.md` executed. *Prior — session 11: round-001 implementation delivered & propagated ([PR #6](https://github.com/gosharplite/tellme/pull/6), merge `62f217f`); **round 001 delivered / frozen**.*)
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `dev` (between rounds — post-round status/docs commit to `dev`; the round-002 branch `002-followup-cleanups` is restored to its delivered tip `21d990a`). The next round starts a fresh `003-*` branch off `dev`.
**Daily log**: [`docs/session-summary/2026/09/11/session-summary.md`](docs/session-summary/2026/09/11/session-summary.md)
**Archive**: [`docs/archives/status/2026-09-11.md`](docs/archives/status/2026-09-11.md) — historical status (round 001, all closed grill/upstream/clarify records, the accumulated decisions log, propagation + environment history), cut at 2026-09-11 (session 13).

## Branch model

| Branch | State | Role |
| --- | --- | --- |
| `main` | merged up from `dev` | Stable / released line |
| `dev` | merged up from the working branch | Integration line (round work lands here before `main`) |
| `001-cli-bootstrap-and-config` | delivered / frozen (round 001) | Round-001 working branch — PR [#6](https://github.com/gosharplite/tellme/pull/6) merged; round 001 is delivered / frozen history |
| `002-followup-cleanups` | merged & propagated | Round-002 base branch — PR [#7](https://github.com/gosharplite/tellme/pull/7) merged (`f2a058f`); propagated `→ dev → main` |
| ~~`002-implement-followup-cleanups`~~ | **deleted** | Round-002 implementation head — merged into `002-followup-cleanups` and deleted (remote + local) |

> **Branch convention (session 13)**: post-round status/docs (`STATUS.md` + daily logs) are committed to
> `dev` and propagated to `main`; each round works on its own `NNN-*` branch off `dev`, and a delivered
> round branch never receives post-round commits.

> `001-implement-cli-bootstrap-and-config` (the PR #6 head branch) was **merged into `001-cli-bootstrap-and-config` and deleted** (remote + local) at the session-11 closeout.

> **Heads are intentionally not pinned here** — the working branch is the moving tip and every commit
> on it is followed by the two-step merge `working → dev → main`. Read live heads with
> `git rev-parse --short main dev HEAD` rather than trusting a snapshot.

## Roadmap — next slices (agreed 2026-09-11, session 13)

Two follow-on slices were agreed with the user and opened as tracking issues. Both are **not started**; each begins a **fresh** plan package (`fresh-package-per-round`), and round 001/002 packages stay frozen.

| Slice | Issue | Scope | Primary truth owners |
| --- | --- | --- | --- |
| **003 — Provider-registry completeness** | [#9](https://github.com/gosharplite/tellme/issues/9) | Grow the boot-subset `PROVIDERS` entry to the **real provider fields** the first turn needs (`API_KEY` with `${VAR}` expansion, `HEADERS`, `THINKING_BUDGET`/`THINKING_LEVEL`, …) + deterministic validation — still **offline** (`internal/config/config.go`'s documented "full config schema" gap). | `/axb-dsl-refine` (MODIFY `features/cli` configuration), `/axb-technical-research` (techstack) |
| **004 — First reasoning turn** | [#10](https://github.com/gosharplite/tellme/issues/10) | `tellme "<prompt>"` → one provider request → printed response; provider domain port + one adapter; deterministic failure class; network-path test strategy. **Depends on 003.** | `/axb-dsl-refine` (new chat/turn module), `/axb-technical-research` (transport amendment) |

> **Roadmap correction**: the originally-discussed slice **A (effective-provider resolution)** is **already implemented** — round-001 `resolve()` Step 5 (`EffectiveSelectedProvider` → `ProviderInRegistry` → provider-mismatch → config error 3) plus the round-002 F9 unit tests (`internal/config/config_test.go`). 003 therefore targets the **remaining** provider-config gap, not the resolution behaviour.

## Current round — `002-followup-cleanups`

**Scope (small follow-up cleanup, locked by the round-002 clarify round)**: (1) **remove the `--json`
flag entirely** — no machine-readable diagnostic mode (reverses round-001 **FR-013**; resolves review
finding **F4** by deletion, because `--json` becomes an unrecognized flag handled by the existing usage
contract); (2) **freeze the operator-facing failure contract** — exact stderr messages (`NFR-004`) and
numeric exit codes (`FR-014`); (3) **F9** — fast unit tests for the pure resolution helpers;
(4) **quality-gate hardening** — an ignored-error gate and a dependency-vulnerability gate.

> **Round-001 package stays frozen**: `specs/plans/001-cli-bootstrap-and-config/**` is history and is
> not modified. Round 002 supersedes the affected behaviour in `specs/truth/**` (a `DELETE` intent).

### Artifacts

- [x] `specs/plans/002-followup-cleanups/spec.md` — `/axb-specify`
- [x] `specs/plans/002-followup-cleanups/checklists/requirements.md`
- [x] `specs/plans/002-followup-cleanups/truth-delta.md` (skeleton)
- [x] `specs/plans/002-followup-cleanups/features/acceptance/*.feature` — `/axb-spec-by-example`
- [x] `specs/plans/002-followup-cleanups/research.md` + `specs/truth/techstack.md` — `/axb-technical-research`
- [x] `specs/plans/002-followup-cleanups/plan.md` — `/axb-system-analysis`
- [x] `specs/truth/features/cli/**` — `/axb-dsl-refine`
- [x] `specs/plans/002-followup-cleanups/tasks.md` — `/axb-tasks`
- [x] Implementation — `/axb-implement`
- [x] Review + delivery — two architecture-review rounds (F1–F7, N1–N3) + **Grill Round #6** (issue [#8](https://github.com/gosharplite/tellme/issues/8), *closed*; gist <https://gist.github.com/gosharplite/a9042dd85a246bd667de2bb40a6226fc>); **PR [#7](https://github.com/gosharplite/tellme/pull/7) merged** (`f2a058f`) + **propagated** (`dev` `2b50cfd`, `main` `7caa4f1`); `SESSION-CLOSEOUT.md` executed

### Pipeline position

All phases **done** — `/axb-specify` → `/axb-spec-by-example` → `/axb-technical-research` → `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → **`/axb-implement` (19/19 tasks `[X]`)** — then two architecture-review rounds + **Grill Round #6** (issue [#8](https://github.com/gosharplite/tellme/issues/8), *closed*) + re-certification. **Round 002 delivered, merged, propagated:** PR [#7](https://github.com/gosharplite/tellme/pull/7) **merged** (`f2a058f`) and propagated `002-followup-cleanups → dev → main` (`dev` `2b50cfd`, `main` `7caa4f1`). `make verify` OK · godog **20/20** · orphan sweep 0 · topology audit 126 steps. **Next: slice 003 — provider-registry completeness ([#9](https://github.com/gosharplite/tellme/issues/9)); then 004 — first reasoning turn ([#10](https://github.com/gosharplite/tellme/issues/10)).**

### Decisions locked (round 002)

- **Clarify Q1** — the `--json` flag is **removed entirely** (no flag, no machine-readable diagnostic
  mode). Reverses round-001 FR-013; **F4** resolves by deletion — with `--json` no longer a flag, its
  rejection is covered by the existing unrecognized-flag usage rule (no new mechanism).
- **Clarify Q2** — include (1) **quality-gate hardening** (ignored-error gate + dependency-vulnerability
  gate), (2) **pin the `NFR-004` error-message wording**, (3) **pin the `FR-014` exit-code numeric
  values** (currently `0/2/3/4/5`).
- **F9** — in scope; reframes round-001 research **Decision 5** (E2E-only) to permit fast unit tests for
  pure helpers (`EffectiveMode` / `EffectiveSelectedProvider` / `ProviderInRegistry` / `EnsureWorkspace`),
  complementary to (never replacing) the E2E acceptance path. Owner: `/axb-technical-research`.

## Open items (non-blocking)

- **Round 002 delivered & closed out (2026-09-11)** — F4 (`--json` removed), F9 (pure-helper unit tests),
  the exit-code pin, the `NFR-004` class-phrase wording, and the quality gates are all landed, merged
  (PR [#7](https://github.com/gosharplite/tellme/pull/7)), and propagated (`dev` `2b50cfd`, `main`
  `7caa4f1`). See the **Current round — `002-followup-cleanups`** section above. The bullets below are
  kept as the origin record.
- **Future-package candidates (not started)** — (a) **CI**: a workflow to run `make verify`
  automatically (round 002 kept the manual gate); (b) **F9 extension**: unit tests for `internal/cli`
  flag parsing (round 002 covered only the pure resolution helpers); (c) **PM-4** `tellme init`
  (first-run config scaffold); (d) revisit the `--json` **reference divergence** if a downstream
  consumer ever needs machine-readable diagnostic output.
- **Round-001 PR #6 review — deferred (recorded; next slice)** — carried from the architecture review ([round 1](https://github.com/gosharplite/tellme/pull/6#issuecomment-5628564073), [round 2 verified](https://github.com/gosharplite/tellme/pull/6#issuecomment-5628657071)). *Now-set F1+F8 / F2 / F3 + in-area F5 / F6 / F7 and the 3 nits are **fixed** (commits `defd416` + nits commit); the below are deliberately deferred:*
  - **F4** — `tellme --json` **without `-d`** is a silent no-op (falls through to the boot path). Needs a **CLI-contract decision** (treat as a usage error **vs.** document the no-op) → **PM acceptance rule** + `/axb-dsl-refine` (`usage/dsl.md`); when decided, record it in `usage/dsl.md` **and** the next slice's `truth-delta.md`. **Owner:** PM + `/axb-dsl-refine`. **When:** next slice.
  - **F9** — no unit tests for `internal/{cli,config,home}`. **Reframed:** the ratified **E2E-only D5 governs the acceptance path**; it does **not** forbid **table-driven unit tests for pure helpers** (`EffectiveMode` / `EffectiveSelectedProvider` precedence, `EnsureWorkspace` idempotency) — those are complementary. **Owner:** `/axb-technical-research` (re-decide against D5). **When:** next slice.
- **Blocking (grill #2)** — **CLEARED.** Both PM acceptance gaps are resolved (PM-1/PM-2) and the
  cross-repo `aixbdd-tmg` CLI-seat decision is **closed** (PR
  [#2](https://github.com/gosharplite/aixbdd-tmg/pull/2)). No gating blockers remain.
- Exact `-d --json` output schema.
- Exit-code numeric values (incl. the new dedicated "diagnostic: unresolved" code).
- Error-message wording (`NFR-004`).
- **Unchecked-error coverage** (round-001 residual): closed next slice by `golangci-lint` + `errcheck`.
- **Host-rule residuals (grill #4)** — **(R1) the atomicity convention** (Rule 2's fold/split boundary is not mechanically derivable) and **(R2) the English override** (STANDARDS §2/§3 are unconditional; no ratified warrant in either repo). Both are **out of `/axb-dsl-refine`'s writ** → route `/axb-clarify` → `aixbdd-tmg` issue → PR (the PR #2 route). **Opened upstream 2026-09-11**: [aixbdd-tmg#5](https://github.com/gosharplite/aixbdd-tmg/issues/5) (R1) and [aixbdd-tmg#6](https://github.com/gosharplite/aixbdd-tmg/issues/6) (R2). They **do not gate** `/axb-tasks`; round-001 ships the CLI contract unchanged on these two points. **⟦RESOLVED (session 8)⟧** — R1 closed by [aixbdd-tmg PR #7](https://github.com/gosharplite/aixbdd-tmg/pull/7) (entailment criterion), R2 by [PR #8](https://github.com/gosharplite/aixbdd-tmg/pull/8) (Project Language clause); tellme followed up with `docs/decisions/0001-project-language.md` (English home) and the W2/D2 re-decision (fold → recorded as NOOP in `truth-delta.md`). Nothing outstanding here.
- **Upstream methodology tracking (grill #5)** — [aixbdd-tmg#9](https://github.com/gosharplite/aixbdd-tmg/issues/9) (`ParallelHint` same-file merge risk under concurrent dispatch) and [aixbdd-tmg#10](https://github.com/gosharplite/aixbdd-tmg/issues/10) (`axb-tasks` Phase 5 missing Pre-Delivery Orphan Coverage Sweep gate). **⟦RESOLVED (session 10)⟧** — closed by [aixbdd-tmg PR #11](https://github.com/gosharplite/aixbdd-tmg/pull/11) (Zero Shared Edits + ParallelHint concurrency arbitration, ADR 0003) and [PR #12](https://github.com/gosharplite/aixbdd-tmg/pull/12) (mandatory Pre-Delivery Orphan Coverage Sweep, ADR 0004). `tellme` round-001 `tasks.md` verified already conformant (45 independent stepdef files; sweep 19/19 PASSED) — no local change required.
- **DECIDED (grill #4)**: exit-code **numeric values** = an **implementation** choice (FR-014 requires only *distinct + deterministic*); the `--json` **key schema** is now **pinned** in `specs/truth/features/cli/diagnostics/dsl.md`.

## Environment notes

- **2026-09-11 (session 13)**: docs/metadata-only session — executed `SESSION-BOOTSTRAP.md` (Steps 1–8),
  then agreed and opened the **003/004 roadmap issues** ([#9](https://github.com/gosharplite/tellme/issues/9),
  [#10](https://github.com/gosharplite/tellme/issues/10)) and recorded them in `STATUS.md` +
  `docs/session-summary/2026/09/11/session-summary.md` §16. **No product/truth change.** Gates: `gofmt -l .` clean ·
  `go vet ./...` clean · secret scan clean (the lone pattern hit was the config field name `API_KEY` in prose).
- **2026-09-11 (session 13 — doc-tree reorganization, same session)**: moved `docs/2026` → `docs/session-summary/2026` (+ refs; back-link depth fixed); **split `STATUS.md`** — live state kept, history archived to `docs/archives/status/2026-09-11.md`; moved `decisions/` → `docs/decisions/` (+ refs; upstream `aixbdd-tmg` refs and frozen packages untouched). Committed on `dev` (`3222635`).
- `origin` uses **SSH** (`git@github.com:gosharplite/tellme.git`). Authentication as `thptcnec`
  is confirmed working for read **and** write.
- **No-network sandbox limitation (verified 2026-09-10)**: `unshare -n` / `unshare -rn` fail
  `Operation not permitted` on this dev host — the privileged netns sandbox is a CI/privileged-Linux
  mechanism; local SC-004 uses the unprivileged hostile-DNS/proxy fallback + the build-graph guard.
- **2026-09-11 (session 11 — human tooling)**: `go install ./cmd/tellme` → **`/home/pos/go/bin/tellme`**; created a **full-fidelity `…/beta-niffler/tellme.sh`** (drives `tellme`; exports `TELL_ME_HOME`; skips `completion` — `niffler.sh` untouched) and wired **`alias tm="source …/tellme.sh"`** into `~/.bashrc` (beside `nf`/`fp`/`wk`/`db`/`tb`), so a **human** can use/test `tellme` from a fresh terminal. A follow-on to the round-001 **Niffler ↔ `tellme` binary-name alignment** note; recorded in daily-log §14. *(Access registrations: read+write for `…/beta-niffler/` + `~/.bashrc`.)*
