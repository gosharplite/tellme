# System Analysis — the `-h`/`-v` CLI shorthands (round 074)

**Plan Package**: `specs/plans/074-cli-help-and-version-shorthands`

## 1. Interfaces

| Interface | Kind | Planner | Result |
| --- | --- | --- | --- |
| The command-line flag surface (`-h`/`--help`, `-v`/`--version`) | `cli` | `/axb-dsl-refine` (contract owner) | **MODIFY** — the usage module gains the help Rule/Examples; the version feature gains the `-v` Example |
| API surface | — | `/axb-api-plan` | **NOOP** — a single CLI end; no OpenAPI/HTTP surface |
| Data surface | — | `/axb-data-plan` | **NOOP** — no persisted-state change |
| UI surface | — | `/axb-ui-plan` | **skipped** — a plain line-oriented CLI (no screen change) |

## 2. Waves

| Wave | Scope | Delegates to | Notes |
| --- | --- | --- | --- |
| **W1** | The flag surface | the implementation | `internal/cli/cli.go` — `parseFlags` gains `-h/--help` + the `-v` shorthand on `--version`; `run` prints `Usage of tellme:` + `fs.FlagUsages()` to `stdout` and returns `Success` (precedence: help → version → `-d` → …) |
| **W2** | The truth rows | `/axb-technical-research` (done) | `specs/truth/techstack.md` CLI-flag-parsing row MODIFY (`research.md` D1–D7) |
| **W3** | The executable CLI contract | `/axb-dsl-refine` | the `usage` module help Rule + the `diagnostics` version `-v` Example |
| **W4** | The plan-side acceptance | `/axb-spec-by-example` (done) | `features/acceptance/asking-for-help-and-version.feature` |

Every interface is delegated or carried to its contract owner — `wave-covers-interfaces` holds.

## 3. CLI contract (the `cli` interface)

The CLI end is a first-class truth interface; there is no API/data/UI planner for it, so `/axb-system-analysis` carries it forward to its contract owner `/axb-dsl-refine`. The change is **user-visible** (the flag surface on `stdout`), so `/axb-spec-by-example` is **NOT** NOOP.

## 4. Unchanged surfaces (invariants)

- The usage-error path is unchanged: an unrecognized flag still refuses with `tellme: the command-line usage is invalid` (`stderr`) and exit **2** (`spec.md` I-1).
- Help is offline + prompt-less (no provider request, no chrome, no spinner, no stdin read) (`spec.md` I-2).
- The frozen phrase vocabulary is unchanged — help emits **no** `tellme: …` line (`spec.md` I-3).
- `--version` output is byte-identical; `-v` is additive (`spec.md` I-4).
- No new dependency; stdlib + pflag; POSIX-only; hermetic (`spec.md` I-5).

## 5. Domain model (ADR 0041)

**Not modelled, and that is recorded here** (the same-PR rule's escape hatch): the round changes a **CLI flag surface**, not a modelled entity/invariant. The domain model has no CLI-flag entity (its `PromptInput`/`Config` entities describe behaviour, not the flag list), so there is nothing to update — `docs/domain-model/**` is **unchanged** and `modelith-check` stays green. (`techstack.md`'s *CLI flag parsing* row is the truth home for the flag list.)
