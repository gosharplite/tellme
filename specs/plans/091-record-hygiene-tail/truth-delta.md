# Truth delta — round 091 `091-record-hygiene-tail`

Per-owner ledger of the truth changes this round. Every owner records at least one row (a `noop`
proves the area was checked).

## axb-technical-research (owner: `specs/truth/techstack.md`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` (*In-file content search tool (`search_files`)* row, `:36`) | Drop the bare ordinal: *"the **ninth** agent tool — the missing half of the reader trio"* → *"the `search_files` tool — the missing half of the reader trio"*. The tool is identified by **name + round/ADR** (round 071 / ADR 0043), not by an ordinal. No semantics change. | `truth-current` (round 091 / closes [#188](https://github.com/gosharplite/tellme/issues/188)): the ordinal contradicted the shipped **offer order** (`search_files` is the **4th**) and the **base-set** count (**eight**; the gated `read_image` is the only ninth). The round-090 handling (drop the hand-count) applied to an ordinal. |
| MODIFY | `specs/truth/techstack.md` (*Provider request retry* row, `:105`) | *"… no new phrase, no new exit code; the vocabulary stays ten"* → *"… no new phrase, no new exit code; the **class-phrase vocabulary** is unchanged"*. Names the subject; drops the stale, drifting number. | `truth-current`: "ten" was a frozen-time count (round 007) that has drifted (the live vocabulary is **eleven**, the exit-code set **seven**), and the bare figure read as ten exit codes. |
| MODIFY | `specs/truth/techstack.md` (*Partial-turn persistence* row, `:31`) | *"the frozen class-phrase vocabulary + the ten-value exit-code set are unchanged"* → *"the frozen class-phrase vocabulary and the **exit-code set** are unchanged"*. Names the subject; drops the stale number. | Same class as `:105`; the exit-code set is **seven** (`internal/cli/exitcode.go`), never "ten". |
| NOOP | `specs/truth/techstack.md` (all other rows) | Checked — no other row changes. | No technology change (the round is record hygiene; NFR-1). |

## axb-dsl-refine (owner: `specs/truth/features/cli/**`)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` (the round-079 prologue note) | *"the frozen class-phrase vocabulary and the ten-value exit-code set are **unchanged**"* → *"the frozen class-phrase vocabulary and the **exit-code set** are unchanged"*. A **prose-note** fix only — **no** `DSLRow`, `不該發生` clause, step sentence, or `.feature` change. | `truth-current`: same stale-count conflation class as the `techstack.md` rows; the note names exit codes, so a bare "ten" misreads. |
| NOOP | `specs/truth/features/cli/**/*.feature` + all `dsl.md` rows | Checked — no `.feature` / row / step change; the offered-set row (`chat/dsl.md` `集合`) is already bound by the round-090 carrier and stays. | The round changes a record's **prose**, not the interface contract. |

## axb-api-plan / axb-data-plan / axb-ui-plan

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked — no API surface change (a CLI end). | CLI-streamlined pipeline (`/axb-api-plan` NOOP). |
| NOOP | `specs/truth/data/**` | Checked — no persisted-state change. | Truth/record-only round. |
| NOOP | `ui/**` | Checked — no user-facing UX surface change. | Plain line-oriented CLI; `/axb-ui-plan` skipped. |

## Records (not truth — noted for completeness)

| Action | Artifact | Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `docs/decisions/README.md` (the ADR 0043 index row) | *"… a ninth tool offered to every provider"* → an ordinal-free phrasing (*"… offered to every provider"*). The **ADR 0043 body is left verbatim** (an Accepted ADR is immutable — supersede, never edit). | FR-2: the curated index is a **live** surface; a bare ordinal there contradicts the offer order. |
| NOOP | `docs/decisions/0043-search-files-tool.md` | Checked — **not edited** (immutable Accepted ADR; a historical record of the state at decision time). Same principle applied to the **eight** `0051`/`0052`/`0053`/`0054`/`0055`/`0056`/`0057`/`0058` "ten" occurrences. | `docs/decisions/README.md` convention: ADRs are immutable once Accepted. |
| NOOP | `docs/domain-model/**` | Checked — **not modelled** (ADR 0041 escape hatch). No entity/invariant/scenario carries a tool ordinal or a class-phrase count. | No modelled behaviour change. |
| MODIFY (comments only) | `internal/infrastructure/tools/search.go` + `internal/infrastructure/tools/filesystem_test.go` | Comment-only edits (fold F-091-1 / TD-091-1): dropped the live **"ninth agent tool"** ordinal from the `search.go` doc comment (so no live product-code surface carries the ordinal) and the stale **"five tools"** count from the `TestToolSchemasRequireReason` comment. No behaviour, no logic, no schema change. | A live product-code comment and a test comment carried the stale count/ordinal the round removes — the round-090 comment-fix precedent (`[X]` a `.go` comment change is not a behaviour change). |
