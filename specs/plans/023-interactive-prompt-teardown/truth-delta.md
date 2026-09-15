# Truth Delta: 023-interactive-prompt-teardown

**Plan Package**: `specs/plans/023-interactive-prompt-teardown`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Interactive TUI prompt (`-i`)** row: the editor frame is **cleared** on submit (`Ctrl+S`/`Alt+Enter`) and abort (`Esc`/`Ctrl+C`) — reference parity — and the `-i` submit resumes the standard turn surface. **Turn chrome (operator)** row: the `-i` submit surface now emits the chrome and **echoes** the submitted (trimmed) prompt immediately before the input-capture acknowledgement; the positional / Ctrl+D surfaces stay echo-free. **Turn progress spinner (operator)** row: the `-i` submit surface is now a **positive** spinner surface (no separate gate). **E2E runner** row: `-i` drops from the spinner negatives (reclassified positive). **Interactive TUI prompt harness** row: the witness asserts the editor frame is **absent** after submit/abort and the standard surface **present**. **Pure-helper unit tests** row: adds the round-023 teardown/echo helpers. | Round-023 research Decisions 1–8 (operator-locked Q1 1a · Q2 2a · Q3 A′); a presentation-only change on the `-i` surface — no new dependency. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — tellme has a single CLI end and no OpenAPI/HTTP surface; the `-i` submit teardown authors no request/response contract. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/data-model.dbml` | Checked — the persisted session-history/usage records are unchanged; the `-i` teardown + echoed prompt are operator-facing output, not stored. | No record-shape change. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/continuing-the-interactive-prompt.feature` | New interface feature: the `-i` submit teardown (the editor clears on submit/abort) + the standard-surface handoff (echoed prompt → input-capture → rule/header → spinner). | FR-001..FR-010; research D1–D4. |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-turn.feature` | DELETE the Rule "The interactive prompt shows no turn chrome" (now false — the `-i` submit is a chrome surface); the header note re-scoped to include the `-i` submit surface. | FR-005/FR-006; research D2/D3. |
| MODIFY | `specs/truth/features/cli/chat/watching-the-tool-loop.feature` | MODIFY the round-022 negative Example → chrome-aware (`the pre-flight payload line is separated from the answer by a single blank line`). | The `-i` surface is now chrome, so the round-017 frame gap is the single blank (Option 1). |
| MODIFY | `specs/truth/features/cli/chat/presenting-the-progress-spinner.feature` | MODIFY the header note — drop the `-i` surface from the spinner negatives. | FR-006; research D3. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | ADD `the interactive prompt is cleared` · `the submitted prompt "{prompt}" is echoed on the diagnostic output` · `the prompt is not echoed on the diagnostic output` · `the pre-flight payload line is separated from the answer by a single blank line`; DELETE `the tool loop added no blank line before the answer`; round-019/round-022 notes updated + a round-023 note. | FR-003/FR-007/FR-008; the round-022 re-anchor. |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | The `the prompt is not echoed on the diagnostic output` row is re-tokenized to `the prompt "{prompt}" is not echoed on the diagnostic output` (the `{prompt}` parameter lets the reader-surface Example carry the prompt). | Round-023 implementation (PR #51 impl-review TD2): the reader When does not set `lastPrompt`, so the no-echo witness carries the prompt explicitly. |
| MODIFY | `specs/truth/features/cli/dsl.md` | MODIFY the interface-root `the run shows no turn chrome` + `the run shows no progress spinner` rows — drop the `-i` clause from `不該發生`. | FR-005/FR-006; the `-i` surface is no longer a negative. |
