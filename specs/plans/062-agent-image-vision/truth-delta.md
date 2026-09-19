# Truth Delta: 062-agent-image-vision

**Plan Package**: `specs/plans/062-agent-image-vision`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: skeleton initialized by `/axb-specify`. **Clarify CLOSED** (5 decisions, one at a time) — **Q1 → 1** (OpenAI-compatible family only; Gemini `inline_data` a forward item) · **Q2 → 1** (explicit per-provider `VISION` config key, default off) · **Q3 → 1** (`read_image` only) · **Q4 → 1** (`read_image` not offered when `VISION` off — the offered set is a function of capability) · **Q5 → 1** (inline only, 32 MiB ceiling, oversize = loud tool error). No residual `NEEDS CLARIFICATION`; the package is ready for the owner phases. Owner rows below are placeholders until the owners run.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/techstack.md` — *Image filesystem tool (`read_image`)* | The first media read tool: `read_image(filepath, reason)`; content-sniffed kind (JPEG/PNG/GIF/WebP magic bytes); 32 MiB inline ceiling with a loud oversize error; offered only for a vision-declaring provider; no path boundary; stdlib-only. | `spec.md` FR-001/FR-003/FR-005, S-4/S-6/S-7 |
| ADD | `specs/truth/techstack.md` — *Image content on the provider wire (OpenAI-compatible)* | The inline base64 `image_url` content array; the `user`-message-after-tool-result placement (media-first); the byte-identical text path; the 32 MiB inline ceiling; the Gemini family refuses media loudly (no silent drop). | `spec.md` FR-002/FR-004, S-2/S-6, `research.md` D4/D5/D7/D8 |
| MODIFY | `specs/truth/techstack.md` — *Provider entry schema* | Adds the `VISION` boolean (default false) — the explicit capability declaration; capability is never inferred from the model name/family. | `spec.md` FR-006, S-3 |
| MODIFY | `specs/truth/techstack.md` — *Agent tool schemas* | The offered set becomes a function of the selected provider's capability: the assembler takes a vision flag and offers `read_image` (through `resourceSchema`, so `required ⊆ properties` holds) only when `VISION: true`. | `spec.md` FR-007/FR-008/FR-009, S-5 |
| ADD | `docs/decisions/0032-agent-image-vision.md` (+ index row) | The capability key (not a name heuristic), the openai-only scope, the content sniff, the 32 MiB inline ceiling, the media placement, and §Forward RF-062-1…8. | `spec.md` S-8, `research.md` D10 |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/` (**no `contracts/**`**) | Single CLI end; no OpenAPI/HTTP surface exists or changed. | `spec.md` A3 (`contract-authoritative` holds vacuously) |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP (checked) | `specs/truth/data/data-model.dbml` | No persisted shape changes: a `read_image` step persists its **text** result only; the image is in-memory/in-flight (never written). The in-memory conversation is outside the dbml's modelled scope (persisted logs only). | `spec.md` FR-001, `research.md` D7, ADR 0032 RF-062-6 (`data-model-covers-all-state` holds — no persisted state changes) |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/reading-a-local-image.feature` | A new CLI-end feature: 3 Rules (the image reaches the wire, content-sniffed kind; the capability-dependent offered set; the loud refusals) + a documented-narrowing Rule. | `spec.md` US1/US2/US3 (`acceptance-coverage`) |
| MODIFY | `specs/truth/features/cli/chat/offering-the-agent-tools.feature` | The header notes the set is a **function of the selected provider's capability**; a new Example asserts a vision-enabled provider is additionally offered `read_image`. | `spec.md` FR-009, S-5 |
| ADD | `specs/truth/features/cli/chat/dsl.md` — `## Given (round 062)` (6 rows) + `## Then (round 062)` (7 rows) + a round-062 note | The new sentences: the workspace image-file Givens (kind / oversize / not-a-picture), the capability-declaring provider Givens, and the image/offered-set/loud-refusal Thens. | `spec.md` FR-001…FR-012 (`dsl-exact-one-match`) |
| NOOP (checked) | the interface-root `specs/truth/features/cli/dsl.md` + sibling modules | No cross-module row changed; the new rows are module-local to `chat`. | `dsl-single-authority` holds |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): refreshed for the round-062 capability (the `Provider.vision` attribute · the `Tool` capability gate + `tool-offered-only-when-capable` · the new `ImageContent` entity · a *Reading a local image* scenario); `make modelith-check` green (no drift).

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): 48 features · 6 modules · 16 root + 375 module rows · 1894 steps — the **same 5 pre-existing errors** as at round 061, **none new**; round 062's feature + Example match exactly one row each.
