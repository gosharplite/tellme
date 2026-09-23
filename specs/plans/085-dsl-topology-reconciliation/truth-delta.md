# Truth delta — round 085 `085-dsl-topology-reconciliation`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop`
proves the area was checked).

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/dsl.md` (interface root) | **ADD 4 cross-module rows** promoted from module DSLs: `a configured provider "{provider}" whose endpoint answers with "{answer}"` (Given) · `the effective mode is "{mode}"` (Given) · `the operator starts tellme with the prompt "{prompt}"` (When) · `tellme exits with the configuration error code` (Then). Semantics **unchanged**; only the authority level moves (module → root). | `dsl-single-authority` + `dsl-exact-one-match`: a row used by two or more modules MUST live at the interface root so the merged lookup sees it from every module (round 085 / closes #173, Class A). |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | (i) **REMOVE** the two promoted rows (`a configured provider "{provider}" whose endpoint answers with "{answer}"`; `the operator starts tellme with the prompt "{prompt}"`) — now at the root. (ii) **ADD** the missing `## Given (round 083)` / `## Then (round 083)` headings + `DSL 句型` header rows so the five round-083 rows are **parseable**. (iii) **ADD** the composite projection row `a configured provider "{provider}" whose endpoint asks tellme to read "{path}" with the reason "{reason}" and then answers with "{answer}" and reports the token usage:` (DataTable `prompt/cached/completion/thinking`). | Class A (single authority) · Class B (restore table structure) · Class C (the missing composite carrier for `colouring-the-session-chrome.feature`) — round 085 / closes #173. |
| MODIFY | `specs/truth/features/cli/configuration/dsl.md` | **REMOVE** the two promoted rows (`the effective mode is "{mode}"`; `tellme exits with the configuration error code`) — now at the root. Semantics unchanged. | `dsl-single-authority` — the rows are shared by the `configuration` and `history` modules; round 085 / closes #173. |
| NOOP | `specs/truth/features/cli/**/*.feature` | Checked — **no `.feature` change**. The failing steps' *text* is correct; only the DSL-row resolution changes. | The E2E step definitions are unchanged, so the reconciled rows must not require a feature edit (`dsl-exact-one-match` is a row-lookup fix, not a step rewrite). |
| NOOP | `specs/truth/features/cli/{diagnostics,usage,workspace}/**` | Checked — untouched (their rows are unaffected by the promotion). | Scope guard (I-1). |

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Retire** the `## Not Introduced Yet` bullet *"`make help` `verify-no-network` text (R8d) … a recorded tidy-up."* | The tidy-up lands this round (FR-006); `recorded-forward-item → delivered`. A note-level row removal (no contract/invariant/behaviour change). |
| NOOP | `specs/truth/techstack.md` (all other rows) | Checked — no other row changes. | No technology change (I-5). |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface change (a CLI end). | CLI-streamlined pipeline (mark `/axb-api-plan` NOOP). |
| NOOP | `specs/truth/data/**` | Checked — no persisted-state change. | No state change (truth-only round). |
| NOOP | `ui/**` | Checked — no user-facing UX surface changes (a `dsl.md`/help-string reconciliation). | CLI, plain line-oriented; `/axb-ui-plan` skipped. |
