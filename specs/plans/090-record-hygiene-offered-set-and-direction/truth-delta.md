# Truth delta — round 090 `090-record-hygiene-offered-set-and-direction`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop`
proves the area was checked).

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` (the `the request offered exactly the agent tools` row) | The `集合` cell now enumerates the **eight** base tools (adds `search_files`), the stale hand-count is dropped, and the cell names the **single owner** (the live agent-tool registry that `agentTools()` builds). The `不該發生` clause is kept. **Step text unchanged** (the sentence pattern is identical). | `truth-current` (round 090 / closes [#186](https://github.com/gosharplite/tellme/issues/186)): the row asserted a **seven**-tool set that **omitted `search_files`** (round 071 / ADR 0043) and **contradicted** the sibling `offering-the-agent-tools.feature` ("eight") — the round-088 **F-088-1** class. |
| NOOP | `specs/truth/features/cli/chat/offering-the-agent-tools.feature` | Checked — **no change**. The feature already states "exactly the **eight** agent tools" and names `search_files`; it is the **current** surface. The round reconciles the *other* surface to it and adds a carrier. | The feature is correct; only `dsl.md` drifted (A2). |
| NOOP | `specs/truth/features/cli/{history,configuration,workspace,diagnostics,usage}/**` | Checked — untouched (no offered-set claim elsewhere). | Scope guard (I-1). |

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` (the `spf13/cobra` **Not Introduced Yet** bullet) | Correct the wording: `cobra` is deferred "until subcommands (e.g. `browse`) and further flag surfaces exist" — removing the false implication that `retry` is a subcommand (the reference's `retry` is a **flag**, `--retry`). | Record accuracy (FR-006). A note-level wording fix (no contract/invariant/behaviour change). |
| NOOP | `specs/truth/techstack.md` (all other rows) | Checked — no other row changes. | No technology change (I-5). |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface change (a CLI end). | CLI-streamlined pipeline (`/axb-api-plan` NOOP). |
| NOOP | `specs/truth/data/**` | Checked — no persisted-state change. | Record-only round. |
| NOOP | `ui/**` | Checked — no user-facing UX surface change. | Plain line-oriented CLI; `/axb-ui-plan` skipped. |

## Records (not truth — noted for completeness)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0061-operator-declared-direction.md` | Records the operator-declared direction (no security layer · no Windows · bash-first · POSIX-only) with rationale, consequences, and the accepted-risk note. | The direction had **no** durable home; the operator decided it belongs in an ADR, not `README.md` (FR-005). |
| MODIFY | `docs/decisions/README.md` | Add the **ADR 0061** index row (status `Accepted`). | `verify-adr-index` (`ADR-index-consistent`). |
| MODIFY | `README.md` | (i) reconcile the tool-surface enumeration to the shipped set (B); (ii) demote the direction paragraph to a **summary that points at ADR 0061** (the line claiming "direction changes are recorded here" is corrected). | FR-004 / FR-005 — README is not the direction's source of truth. |
| MODIFY | `STATUS.md` | The *Direction* line points at **ADR 0061** (not README). | FR-005 — one direction home; the live-state file cites it. |
| MODIFY | `tests/e2e/steps/step_r021_t026_chat_then_offered_tools.go` | Correct the stale history **comment** (record round 071's `search_files`); **no behaviour change**. | FR-003 — the comment falsely said the set "grew to seven". |
| ADD | `cmd/tellme/*_test.go` (the offered-set doc-consistency carrier) | A **test-only** check: parse the `chat/dsl.md` offered-set cell and assert set-equality with `agentTools()`. | FR-001 — the single-sourced carrier that reddens on drift (W-A). |
| NOOP | `docs/domain-model/**` | Checked — **not modelled** (ADR 0041 escape hatch). The model already records the tool surface concepts; the round changes no modelled behaviour (a doc *count* fix + a decision record + a comment). | A record/direction round touches no entity/invariant/scenario. |
