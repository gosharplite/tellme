# Truth Delta: 063-gemini-image-vision

**Plan Package**: `specs/plans/063-gemini-image-vision`
**Truth Root**: `specs/truth`

> Plan package truth-delta. Owner rows are recorded by the truth-owner skills (`/axb-technical-research`, `/axb-api-plan`, `/axb-data-plan`, `/axb-dsl-refine`). Each owner records at least one entry; a `NOOP` entry proves the area was checked.
>
> **Status**: **Clarify CLOSED** — **Q1 → A** (reuse the single `VISION` key) and **Q2 → B** (a **family-aware** inline ceiling enforced by `read_image` as a loud tool error) locked by the operator, one at a time (2026-09-19). **Q3** (the `inlineData` wire placement) was left to `/axb-technical-research` and is decided in `research.md` **D2**. Owner rows: `research`/`api`/`data` **recorded below**; `dsl-refine` rows are **recorded** (the phase has run — see the `/axb-dsl-refine` section).

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
| MODIFY | `specs/truth/features/cli/chat/reading-a-local-image.feature` | ADD-ed the round-063 Gemini journeys: *A Gemini model that can see is shown a local image* (2 Examples) · *One declaration of the capability serves both families* (2 Examples) · *An image the Gemini family cannot carry fails loudly* (2 Examples); the header notes the second family and the family-aware ceiling. Existing rounds-062 Rules/Examples unchanged. | `spec.md` US1/US2/US3; acceptance `reading-a-local-image-with-a-gemini-provider.feature` (`acceptance-coverage`) |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — 5 round-062 **Then** rows | the image/offered/refusal sentences (`the request carried the image file "{name}"` · `the image was attached as a "{mime}" picture` · `the request carried no image` · `the request offered [no] the read-image tool` · the two `the tool result reported …` rows) are **widened family-aware** — they judge **either** wire shape (`image_url` **or** `inlineData`); semantics preserved, scope widened. | `spec.md` FR-001…FR-005; `research.md` D2/D3 |
| ADD | `specs/truth/features/cli/chat/dsl.md` — `## Given (round 063)` (3 rows) + a round-063 note | the Gemini-provider Givens: a seeing Gemini provider that asks tellme to read a path · a seeing Gemini provider that just answers · a blind Gemini provider; and the module note (the `inlineData` blob, the retired refusal, the family-aware ceiling). **No new Then row** — the round-062 Then sentences are reused (widened). | `spec.md` FR-001…FR-012 (`dsl-exact-one-match`; each new step matches one row) |
| NOOP (checked) | `specs/truth/features/cli/chat/offering-the-agent-tools.feature` + the interface-root `dsl.md` | the offered-set truth is **unchanged** (Q1 → A: the gate is family-agnostic); no cross-module row changed. | `spec.md` FR-008, S-1 (`dsl-single-authority` holds) |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` — the round-062 **note** | the note's stale family-scope clause ("the Gemini family has no `inline_data` path and returns a loud provider error") is **annotated in place** with a supersession pointer to the round-063 note (the round-057 TF-057-1 class: publish truth must not assert retired behaviour). | `truth-current`; `spec.md` FR-005/FR-009 |

**Domain model** (`docs/domain-model/tellme.modelith.{yaml,md}`, ADR 0030 — descriptive, subordinate to truth): to be refreshed if the capability fact's wording changes (e.g. `Provider.vision` "honoured by both families", the *ImageContent* entity's wire note); `make modelith-check` must stay green.

**Topology audit** (`axb-gherkin-and-dsl` script over `specs/truth/features/cli`): must keep the **same 5 pre-existing errors**, **none new**; each new/changed step must match exactly one `DSLRow`.

## PR #130 review folds (APPROVE WITH REQUIRED FOLDS — 2026-09-19)

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `docs/decisions/0032-agent-image-vision.md` | the §Forward **RF-062-1** entry annotated **SUPERSEDED — landed by ADR 0033**; the *Family asymmetry* trade-off marked **CLOSED** (the stale "only an openai-family provider until RF-062-1 lands" guidance removed). | review **F-063-1** (a Forward item must not read as open after it lands — `residual-homed-on-a-durable-surface`) |
| MODIFY | `specs/truth/techstack.md` — *Provider family mapping* | names the exported **`Family(typeLabel)`** as the single owner of the label→family classification (same tables `NewGateway` dispatches on); the composition root consumes it for the ceiling. | review **F-063-3** (the round's new load-bearing classifier had no owning row) |
| MODIFY | `docs/decisions/0033-gemini-image-vision.md` | §Forward gains **RF-063-7** (the multi-call media placement + the widened two-image live check) · **RF-063-8** (the family-blind oversize fixture ⇒ acceptance stronger than runner) · **RF-063-9** (`ImageCeilingForFamily` fails open / stringly-typed / package home) · **RF-063-10** (the comment-only protection Rule, the 4-round recurrence); **RF-063-6** annotated as overdue. | review **TD-063-1…TD-063-4** + the RF-062-10 note |
| MODIFY | `internal/infrastructure/tools/image.go` + `image_test.go` | `NewReadImageTool` guards a non-positive ceiling (falls back to the OpenAI-compatible default); `TestNewReadImageToolDefaultsCeiling` pins it. | review nit (the `agentTools()` magic `0` latent trap) |
| MODIFY | `tests/e2e/steps/tool_usage.go` + `tests/e2e/fakeprovider/fakeprovider.go` | the name-only enumerator passes `0` (no literal provider label + real ceiling); `ToolNamesAt` documents the deliberate both-shapes concatenation. | review nits |
| ADD | `internal/infrastructure/llm/gemini/client_image_test.go` | `TestRequestBody_MultiCallRound_MediaTurnsInterleave` — the two-call round shape (model, frA, mediaA, frB, mediaB), both blobs, in order. | review **TD-063-1** (the shape had no carrier) |
