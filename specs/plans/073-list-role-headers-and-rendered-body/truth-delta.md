# Truth Delta: 073-list-role-headers-and-rendered-body

**Plan Package**: `specs/plans/073-list-role-headers-and-rendered-body`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **skeleton — `/axb-specify` RUN (2026-09-21)**. Clarify **PENDING** (Q1 `-r` interplay · Q2 tool-activity scope · Q3 prompt-body rendering). No truth file written yet.

## Expected owner rows (to be filled by the owner skills)

| Owner skill | Expected action | Truth spec | Note |
| --- | --- | --- | --- |
| `/axb-technical-research` | MODIFY | `specs/truth/techstack.md` — the CLI-flags / history-listing row | the listing's presentation (role headers, glamour body, blank separator, `stdout`-gated colour) + the renderer/probe seams chosen by research (D-x) |
| `/axb-technical-research` | ADD | `docs/decisions/00NN-*.md` (+ index) | the decision record for the `-l` presentation parity (shape, colour axis, divergences) |
| `/axb-api-plan` | NOOP (checked) | `specs/truth/**` (no `contracts/**`) | Single CLI end; no OpenAPI/HTTP surface |
| `/axb-data-plan` | NOOP (checked) | `specs/truth/data/**` | The persisted `history.jsonl` shape is unchanged (I-2) |
| `/axb-dsl-refine` | MODIFY | `specs/truth/features/cli/history/inspecting-the-session-history.feature` | the listing Rules become the new per-message shape (header + rendered body + blank separator) |
| `/axb-dsl-refine` | MODIFY | `specs/truth/features/cli/history/dsl.md` | rewrite the listing Then-rows (`role: content` → the header/render/separator contract) |

*(Rows above are the expected shape pending clarify + research; the owner skills replace them with concrete entries.)*
