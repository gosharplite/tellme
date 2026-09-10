# tellme — Status

**Last updated**: 2026-09-10
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `001-cli-bootstrap-and-config` (→ `dev` → `main`)
**Daily log**: [`docs/2026/09/10/session-summary.md`](docs/2026/09/10/session-summary.md)

## Branch model

| Branch | Role |
| --- | --- |
| `main` | Stable / released line |
| `dev` | Integration line (round work lands here before `main`) |
| `001-cli-bootstrap-and-config` | This session's working branch |

## Current round — `001-cli-bootstrap-and-config`

**Scope (narrow foundation, locked)**: CLI boot; YAML configuration load + validation
(`config-valid-provider`); runtime home (`TELL_ME_HOME`) + per-mode session workspace
(`output/<mode>/`); build version (`--version`); offline setup diagnostic (`-d`, `-d --json`).
Configuration is a **slice-local input** — no `contracts/**`, no `data/**` truth this round.

### Artifacts

- [x] `specs/plans/001-cli-bootstrap-and-config/spec.md`
- [x] `specs/plans/001-cli-bootstrap-and-config/checklists/requirements.md`
- [x] `specs/plans/001-cli-bootstrap-and-config/truth-delta.md` (skeleton)
- [ ] `specs/plans/001-cli-bootstrap-and-config/features/acceptance/*.feature` — `/axb-spec-by-example`
- [ ] `specs/plans/001-cli-bootstrap-and-config/research.md` + `specs/truth/techstack.md` — `/axb-technical-research`
- [ ] `specs/plans/001-cli-bootstrap-and-config/plan.md` — `/axb-system-analysis`
- [ ] `specs/truth/features/**` + `dsl.md` — `/axb-dsl-refine`
- [ ] `specs/plans/001-cli-bootstrap-and-config/tasks.md` — `/axb-tasks`
- [ ] Implementation — `/axb-implement`

### Pipeline position

`/axb-specify` done → **next:** `/axb-spec-by-example`, then `/axb-technical-research`
→ `/axb-system-analysis` → `/axb-dsl-refine` → `/axb-tasks` → `/axb-implement`.

## Decisions locked this session

- Round 1 = **narrow foundation** (boot + config + home/workspace + version/diagnostics); no Q&A turn yet.
- **Not everything in tell-me-go will appear in tellme** — the scope is a deliberate subset.
- **`/axb-constitution` skipped** — the default constitution is used.
- Configuration treated as a **slice-local input**, not a truth artifact.
- **Niffler shell-env alignment folded into `spec.md`** (this round): `TELL_ME_*` env overrides (`TELL_ME_MODE`, `TELL_ME_SELECTED_PROVIDER`) take precedence over the YAML config — `FR-003`/`FR-007` now resolve the *effective* provider/mode and new `FR-015` records the cross-story rule (plus a matching edge case); default config path resolved to `$TELL_ME_HOME/configs/<mode>.yaml` (`<mode>` defaulting to `butler`); binary-name divergence (Niffler invokes `tell-me-go` vs tellme's `tellme`) recorded as an out-of-scope integration note.
- **No `/axb-clarify`** round needed (no gap changed story splitting, flows, acceptance, or success criteria).
- Artifacts written in **English** (project + working language).
- Working agreement: butler runs the `axb-*` skills directly, one phase at a time, with a review gate between phases.
- **Self-starting bootstrap**: `SESSION-BOOTSTRAP.md` (Step 7) and `STATUS.md` are present on the branches in play — `main`, `dev`, and the current working branch; a new working branch inherits it from its base. We do **not** fan out copies to every branch. Step 7 checks out the **Active branch** before any work, so a fresh session can continue.
- **Working style**: do the session's work on a local branch (where `STATUS.md` is present); updates flow up via the explicit two-step merge–merge (`working branch → dev → main`).

## Open items (non-blocking)

- Exact `-d --json` output schema.
- Exit-code numeric values (only distinctness is required).
- Error-message wording (`NFR-004`).

*(Resolved this round: default configuration path → `$TELL_ME_HOME/configs/<mode>.yaml`, `<mode>` defaulting to `butler` — see Decisions locked.)*

## Environment notes

- `origin` uses **SSH** (`git@github.com:gosharplite/tellme.git`). Authentication as `thptcnec`
  is confirmed working for both read and write (verified with a create/delete probe branch).
- Historical note: the initial `dev`/session pushes were routed through HTTPS via the `gh`
  credential helper (account `gosharplite`) while SSH write access was unavailable. That
  workaround is no longer needed; `origin` is back on SSH.
