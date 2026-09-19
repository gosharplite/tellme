# Truth Delta: 063-gemini-image-vision

**Plan Package**: `specs/plans/063-gemini-image-vision`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify IN PROGRESS** — **Q1 → A (reuse the single `VISION` key) LOCKED** (operator, 2026-09-19); **Q2** (the Gemini inline size ceiling) is pending, asked one at a time. **Q3** (the `inline_data` wire placement) is deliberately left to `/axb-technical-research`. Owner rows below are **expected** shapes, not yet recorded.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Image content on the provider wire* | the Gemini/Vertex family carries the image as an `inline_data` blob (`mime_type` + base64 `data`) on a `user` turn; the round-062 loud Gemini refusal is retired; a family-appropriate inline ceiling (Q2). | `spec.md` FR-001…FR-005, S-2/S-5 |
| MODIFY | `specs/truth/techstack.md` — *Provider entry schema* / capability note | the single `VISION` boolean is now honoured by **both** families (no second key). | `spec.md` FR-007/FR-008, S-1 |
| ADD | `docs/decisions/00NN-*.md` (+ index row) | the Gemini `inline_data` path; extends ADR 0032; supersedes `RF-062-1`. | `spec.md` S-8, `research.md` (TBD) |
| NOOP (checked) | the OpenAI-compatible image row | unchanged by this round (the round-062 `image_url` shape stands). | `spec.md` S-7 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface exists or changed. | `spec.md` A3 |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked, expected) | `specs/truth/data/data-model.dbml` | No persisted shape changes: the image stays in-memory/in-flight; the history step carries the tool's **text** result only (round-062 limitation). The round-062 `Message.Media` is in-memory. | `spec.md` A3, round-062 RF-062-6 |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY (expected) | `specs/truth/features/cli/chat/reading-a-local-image.feature` | the Gemini journey (a `gemini` provider carrying the image) + the retired refusal Rule. | `spec.md` US1/US2 |
| ADD/MODIFY (expected) | `specs/truth/features/cli/chat/dsl.md` | the new/changed sentences (a gemini family provider; the image on the gemini wire; the retired "cannot carry images" Then). | `spec.md` FR-001…FR-012 |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): to be refreshed if the capability fact's wording changes (e.g. `Provider.vision` "honoured by both families", the *ImageContent* entity's wire note); `make modelith-check` must stay green.

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; each new/changed step must match exactly one `DSLRow`.
