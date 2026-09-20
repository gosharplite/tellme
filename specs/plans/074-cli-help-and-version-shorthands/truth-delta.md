# Truth Delta: 074-cli-help-and-version-shorthands

**Plan Package**: `specs/plans/074-cli-help-and-version-shorthands`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **skeleton — `/axb-specify` RUN (2026-09-21)**. Clarify **PENDING** (Q1 `--help` long form? · Q2 help stream + exit code?). No truth file written yet.

## Expected owner rows (to be filled by the owner skills)

| Owner skill | Expected action | Truth spec | Note |
| --- | --- | --- | --- |
| `/axb-technical-research` | MODIFY | `specs/truth/techstack.md` — the *CLI flag parsing* row | `-h`/`-v` added; the help block's source + the dispatch precedence chosen by research (D-x) |
| `/axb-technical-research` | ADD | `docs/decisions/00NN-*.md` (+ index) | the decision record for the help/version shorthands |
| `/axb-api-plan` | NOOP (checked) | `specs/truth/**` (no `contracts/**`) | Single CLI end; no OpenAPI/HTTP surface |
| `/axb-data-plan` | NOOP (checked) | `specs/truth/data/**` | No persisted-state change |
| `/axb-dsl-refine` | MODIFY | `specs/truth/features/cli/usage/**` + `specs/truth/features/cli/diagnostics/version-and-setup-diagnostic.feature` | the help/version-shorthand Rules + rows |

*(Rows above are the expected shape pending clarify + research; the owner skills replace them with concrete entries.)*
