# tellme — Status

**Last updated**: 2026-09-10
**Session mode**: `butler` (working directly with the user — no `pm`/`rd` delegation in this phase)
**Active branch**: `001-cli-bootstrap-and-config` (→ `dev` → `main`)

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
- **No `/axb-clarify`** round needed (no gap changed story splitting, flows, acceptance, or success criteria).
- Artifacts written in **English** (project + working language).
- Working agreement: butler runs the `axb-*` skills directly, one phase at a time, with a review gate between phases.

## Open items (non-blocking)

- Default configuration path when `-c`/`--config` is omitted.
- Exact `-d --json` output schema.
- Exit-code numeric values (only distinctness is required).
- Error-message wording.

## Environment notes

- `origin` uses **HTTPS** (`https://github.com/gosharplite/tellme.git`). The SSH key present
  authenticates as `thptcnec`, which cannot write to `gosharplite/tellme`; pushes go through the
  `gh` credential helper under the `gosharplite` account.
