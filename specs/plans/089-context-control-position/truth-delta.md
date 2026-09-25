# Truth delta — round 089 `089-context-control-position`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop`
proves the area was checked).

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` (*History summarisation* bullet, `Not Introduced Yet`) | Split the conflated clause: **remove `-b`/`--back`** from the settled-exclusion list (it is **delivered** — round 081 / ADR 0053; forward pointer to the *Session lifecycle flags* row) and **keep `--retry`** as a real non-introduction (roll back the last user message + resend; the reference ships it, tellme does not). The other exclusions (history pinning · token-budget pruning · `SafePath`/consent) are unchanged. **Semantics of the delivered capabilities are untouched** — only the *status* of `-b` moves from "excluded" to "delivered". | `truth-current` (round 089 / closes [#184](https://github.com/gosharplite/tellme/issues/184)): the bullet carried a **superseded present-tense claim** (round 081 delivered `-b`, but this row still excluded it — the round-088 F-088-1 class). |
| NOOP | `specs/truth/techstack.md` (all other rows) | Checked — no other row changes. The round-078 *Provider request retry* row already records the reference's `--retry` flag as **not adopted** (accurate); the *Session lifecycle flags* row already documents `-b`/`--back` as delivered. | No technology change (I-5). |

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/features/cli/**/*.feature` + `**/dsl.md` | Checked — **no change**. The "tellme offers no summarisation tool" claim is **already carried** by `chat/offering-the-agent-tools.feature` (Rule: *"a summarisation tool must not be offered"*, `chat/dsl.md`). The round records the decision (an ADR), which is not a Gherkin sentence pattern. | The decision is a durable record (`docs/decisions/**`), not interface truth; the carrier already exists (FR-004). |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface change (a CLI end). | CLI-streamlined pipeline (`/axb-api-plan` marked NOOP). |
| NOOP | `specs/truth/data/**` | Checked — no persisted-state change. | Truth/record-only round. |
| NOOP | `ui/**` | Checked — no user-facing UX surface change. | Plain line-oriented CLI; `/axb-ui-plan` skipped. |

## Records (not truth — noted for completeness)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| ADD | `docs/decisions/0060-context-control-position.md` | A durable decision: tellme does **not** port `summarize_history`; context control is the **operator-manual** `-l` (inspect) + `-b` (roll back) pair; token-budget pruning / history pinning / `SafePath`-consent stay settled exclusions; a revisit trigger is stated. | Records the position so future rounds can cite it (FR-003). Mirrors ADR 0030/0041 (a durable decision for a governed artifact). |
| MODIFY | `docs/decisions/README.md` | Add the **ADR 0060** index row (status `Accepted`) + a back-pointer on the **ADR 0053** index row. | `verify-adr-index` (`ADR-index-consistent`: every ADR indexed once); the back-pointer records that 0060 cites 0053. |
| NOOP | `docs/domain-model/**` | Checked — **not modelled** (ADR 0041 escape hatch). The model already records `-b`/`--back` (Session `session-rollback-stays-offline`) and has **no** summarisation/pruning entity; the round changes no modelled behaviour. | A decision about a **non-capability** + a truth *status* fix touch no entity/invariant/scenario. |
