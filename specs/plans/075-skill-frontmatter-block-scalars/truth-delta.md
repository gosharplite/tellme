# Truth Delta: 075-skill-frontmatter-block-scalars

**Plan Package**: `specs/plans/075-skill-frontmatter-block-scalars`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **`/axb-technical-research` + `/axb-dsl-refine` RUN (2026-09-21)** — `research.md` D1–D6 + **ADR 0047** + `techstack.md` MODIFY ×1. Clarify resolved at specify time (**0 questions** — not escalated). `/axb-api-plan` + `/axb-data-plan` record `NOOP`.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Skills catalog (load)* | the frontmatter reader resolves a YAML **block scalar** value (indicator `[>|][+-]?`; fold `>` / literal `|`; clip/strip/keep chomping; `TrimSpace`) so a skill authored with `description: >` lists by its **real text**, not the bare indicator; an inline value containing `>` stays literal; an indicator with no body resolves empty ⇒ the file is not a skill (best-effort skip, unchanged). **A deliberate divergence *beyond* the reference** — `tell-me-go`'s line-based `parseSkill` shares the limitation | `spec.md` US1/US2, FR-001…FR-006; `research.md` D1–D6 |
| ADD | `docs/decisions/0047-skill-frontmatter-block-scalars.md` (+ index row) | the decision record: stdlib-only reader extension; the indicator grammar; fold/literal + chomping; the divergence note; the excluded `yaml.v3` option | `spec.md` SC-003; `research.md` D1–D6 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface. | `plan.md` §1 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/**` | No persisted-state change (the catalog is not persisted). | `spec.md` NFR-003 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/features/cli/chat/listing-the-available-skills.feature` | **NEW** Rule + Example: a skill whose description is a folded block scalar is listed by its folded text | `spec.md` US1, FR-001…FR-003 |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | +2 rows — the `Given` that authors a folded block-scalar skill and the `Then` that reads the described text (+ a round-075 note) | `spec.md` US1/US2 |
