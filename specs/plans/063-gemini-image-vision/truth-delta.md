# Truth Delta: 063-gemini-image-vision

**Plan Package**: `specs/plans/063-gemini-image-vision`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **Clarify CLOSED** — **Q1 → A** (reuse the single `VISION` key) and **Q2 → B** (a **family-aware** inline ceiling enforced by `read_image` as a loud tool error) locked by the operator, one at a time (2026-09-19). **Q3** (the `inlineData` wire placement) was left to `/axb-technical-research` and is decided in `research.md` **D2**. Owner rows: `research`/`api`/`data` **recorded below**; `dsl-refine` rows are **expected** (that phase has not run).

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/techstack.md` — *Image content on the provider wire (Gemini/Vertex)* | the new row: a media-bearing message becomes one `user` `contents` entry with `inlineData` blob parts (`mimeType` + base64 `data`, camelCase proto-JSON), emitted after the tool-result `functionResponse` turn (the reference's #1441 hazard avoided by construction); the round-062 loud Gemini refusal **retired**; the family-aware inline ceiling (**14 MiB**, derived from the documented ≈20 MB inline request bound). | `spec.md` FR-001…FR-005, S-2/S-5/S-6; `research.md` D1/D2/D3/D4 |
| MODIFY | `specs/truth/techstack.md` — *Image content on the provider wire (OpenAI-compatible)* | corrected the now-stale clause ("the Gemini adapter has no `inline_data` path … a loud `*llm.ProviderError`") to point at the new Gemini row; the `image_url` block itself is unchanged. | `spec.md` S-7; `research.md` D1 |
| MODIFY | `specs/truth/techstack.md` — *Image filesystem tool (`read_image`)* | the tool takes the **selected provider's resolved inline ceiling** as a construction input (alongside the vision gate) and enforces it as a loud tool error on **both** families. | `spec.md` FR-004, S-6 (Q2 → B); `research.md` D4/D7 |
| MODIFY | `specs/truth/techstack.md` — *Provider entry schema* | the single `VISION` boolean is now honoured by **both** families (no second key). | `spec.md` FR-007/FR-008, S-1 (Q1 → A); `research.md` D5 |
| MODIFY | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | the adapter now carries media (`inlineData` blobs) instead of refusing it. | `spec.md` FR-001/FR-005; `research.md` D1/D10 |
| ADD | `docs/decisions/0033-gemini-image-vision.md` (+ the index row; and the **Status** + index-row annotation of ADR 0032) | the decision: the `inlineData` blob, the placement, the camelCase keys, the family-aware ceiling + its derivation/residual; **extends ADR 0032**, **supersedes its RF-062-1**, narrows its D4/D8. | `spec.md` S-8; `research.md` D9 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface exists or changed. | `spec.md` A3 (`contract-authoritative` holds vacuously) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | No persisted shape changes: the image stays in-memory/in-flight; the history step carries the tool's **text** result only (round-062 limitation). The round-062 `Message.Media` is in-memory; this round is adapter-side. | `spec.md` A3; round-062 RF-062-6 (`data-model-covers-all-state` holds) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY (expected) | `specs/truth/features/cli/chat/reading-a-local-image.feature` | the Gemini journey (a `gemini` provider carrying the image) + the retired refusal Rule. | `spec.md` US1/US2 |
| ADD/MODIFY (expected) | `specs/truth/features/cli/chat/dsl.md` | the new/changed sentences (a gemini family provider; the image on the gemini wire; the retired "cannot carry images" Then). | `spec.md` FR-001…FR-012 |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): to be refreshed if the capability fact's wording changes (e.g. `Provider.vision` "honoured by both families", the *ImageContent* entity's wire note); `make modelith-check` must stay green.

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; each new/changed step must match exactly one `DSLRow`.
