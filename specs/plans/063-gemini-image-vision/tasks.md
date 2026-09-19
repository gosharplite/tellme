# Tasks — Gemini image vision: `inlineData` on the Gemini/Vertex wire (round 063)

**Plan Package**: `specs/plans/063-gemini-image-vision`
**Execution**: `/axb-implement` — Red → Green → Refactor, one task at a time, marked `[X]` after verification.

## Core Inputs (Read before any task)

- `specs/plans/063-gemini-image-vision/spec.md` — US1–US3 · FR-001…FR-012 · SC-001…SC-006 · S-1…S-8 · I-1…I-5 · A1–A6.
- `specs/plans/063-gemini-image-vision/research.md` — D1…D11 + residual risks (RF-063-1…RF-063-6).
- `specs/plans/063-gemini-image-vision/plan.md` — the 1 interface, the wave, and the 9-site change surface.
- `specs/plans/063-gemini-image-vision/truth-delta.md` — the owner rows (techstack ADD/MODIFY ×5, api/data NOOP, dsl-refine MODIFY/ADD ×3 + 1 NOOP, the ADR ADD).
- `specs/truth/techstack.md` — *Image content on the provider wire (Gemini/Vertex)* (round 063, **new**) · *Image content on the provider wire (OpenAI-compatible)* · *Image filesystem tool (`read_image`)* · *Provider entry schema* · *Vertex/Gemini adapter*.
- `specs/truth/features/cli/chat/reading-a-local-image.feature` · `specs/truth/features/cli/chat/dsl.md` — the widened round-062 **Then** rows + `## Given (round 063)`.
- `docs/decisions/0033-gemini-image-vision.md` — the durable decision + §Forward RF-063-1…6 (and ADR 0032's annotation).
- Acceptance: `specs/plans/063-gemini-image-vision/features/acceptance/reading-a-local-image-with-a-gemini-provider.feature`.

## Phase 1 — Setup

- [ ] **T001** — Branch `063-gemini-image-vision` off `dev` `cff2515`; plan package created (spec + checklist + truth-delta; acceptance + research + ADR 0033 + techstack truth; `STATUS.md` phase-gate updated).
- [ ] **T002** — **No new technology**: no Setup phase for packages/toolchains — stdlib-only (`encoding/base64`, `encoding/json` already imported); `go.mod`/`go.sum` MUST stay unchanged (research D11). Smoke check: `go build ./...` + `go test -count=1 ./internal/infrastructure/llm/gemini/` green before touching anything.

## Phase 2 — Foundational (only scaffolding; NO product behaviour, NO assertions)

- [ ] **T003** — **Only**: the Gemini stepdef landing skeleton `tests/e2e/steps/step_r063_image.go` (package + imports + the 3 empty step functions registered through the shared `registrars`), plus the round-063 config helpers it will call if they do not already exist (`writeGeminiConfig` + `writeServiceAccountKey` exist — reuse). **Not**: any assertion or arrangement body.
- [ ] **T004** — **Only**: make the fake provider family-aware for tool NAMES — extend `tests/e2e/fakeprovider.Provider.ToolNamesAt` to also read `tools[].functionDeclarations[].name` (the Vertex wire shape) in addition to `tools[].function.name`. **Not**: any behaviour change to the OpenAI path; existing rounds-061/062 assertions stay green.
- [ ] **T005** — **Only**: refactor the round-062 image stepdef helpers into family-aware *shapes* (declared, no assertion change yet): a `recordedImageBlocks(body)` that returns both `image_url` URIs and `inlineData{mime,data}` pairs, and a `readImageToolResult(body)` that also resolves the Vertex `functionResponse.content`. **Not**: changing any Then's pass/fail semantics yet.
- [ ] **T006** — **Only**: unit-test landing files — `internal/infrastructure/llm/gemini/client_image_test.go` (the round-062 refusal pin **to be inverted**, plus the new wire pin) and a ceiling-boundary pin in `internal/infrastructure/tools/image_test.go` — with the new test functions declared as `t.Skip("RED phase")`. **Not**: assertions.

## Phase 3 — Test Alignment & Implementation (tests only; NO product code)

**DSL 参照** — every round-063 sentence's authority is `specs/truth/features/cli/chat/dsl.md`; read each row's `StepDef 實作語意` for the arrangement/assertion contract. The reused (widened) round-062 **Then** rows keep their authority rows; the 3 new **Given** rows live under `## Given (round 063)`.

**Markers** — `[BDD-ALIGN]` (reuse/widen an existing stepdef or unit pin) · `[BDD-RED]` (write a new stepdef that fails until the product lands) · `[UNIT]` (a hermetic unit pin) · `[P]` (independently parallelizable — lands in its own file/function) · `[REVIEW]` (a gate).

**Boundary** — Phase 3 writes **no** product file; it lands every new stepdef and unit pin RED.

**Parallel Hint** — T008–T013 are `[P]` (one task per DSL row / pin; each new function lives in its own landing file).

- [ ] **T007** `[BDD-ALIGN]` — widen the five round-062 **Then** stepdefs to judge **either** wire shape: `the request carried the image file "{name}"` (decode `image_url` **or** `inlineData`), `the image was attached as a "{mime}" picture` (`data:<mime>;base64,` **or** `inlineData.mimeType`), `the request carried no image` (neither wire), `the request offered the read-image tool` / `no read-image tool` (the family-aware `ToolNamesAt`), and the two `the tool result reported …` (the OpenAI `tool`-role result **or** the Vertex `functionResponse.content`). Rows unchanged in `dsl.md` beyond the round-063 widening note.
- [ ] **T008** `[P][BDD-RED]` — stepdef for the Given `a configured Gemini provider "{provider}" that can take images and whose endpoint asks tellme to read "{path}" and then answers with "{answer}"` (Vertex config + service account + `VISION: true` via `setSelectedProviderVision`; the Vertex-shaped fake scripts one `read_image` call).
- [ ] **T009** `[P][BDD-RED]` — stepdef for the Given `a configured Gemini provider "{provider}" that can take images whose endpoint answers with "{answer}"` (Vertex config + `VISION: true`; the fake just answers).
- [ ] **T010** `[P][BDD-RED]` — stepdef for the Given `a configured Gemini provider "{provider}" that cannot take images whose endpoint answers with "{answer}"` (Vertex config, no `VISION` key).
- [ ] **T011** `[P][UNIT]` — pin (RED) in `internal/infrastructure/llm/gemini/client_image_test.go`: a media-bearing request serializes to a `user` `contents` entry with an `inlineData` part (`mimeType` + base64) **after** the tool-result turn; a media-free request is **byte-identical** to the pre-round body (I-1 control). Remove the round-062 refusal assertion (inverted).
- [ ] **T012** `[P][UNIT]` — pins (RED) in `internal/infrastructure/tools/image_test.go`: the ceiling boundary both ways against an **injected** ceiling (at the ceiling passes; one byte over is the loud error), plus the single-owned family→ceiling resolver (OpenAI-compatible 32 MiB / Gemini 14 MiB).
- [ ] **T013** `[P][UNIT]` — pin (RED) in `internal/app/deps/deps_test.go` (or the composition-root test): the registry builder offers `read_image` iff vision **and** passes the resolved ceiling through (the widened seam).
- [ ] **T014** `[REVIEW]` — subagent review gate: confirm 0 undefined steps in the new feature, every round-063 DSL row matched exactly once (the 3 new Givens + the widened Thens), and the unit pins fail RED for the right reason.

## Phase 4 — Feature phases (product code; NO red/align here)

### Feature 4A — `reading-a-local-image.feature` MODIFY (the Gemini journeys; US1/US2)

- [ ] **T015** `[BDD-GREEN]` — the Gemini serializer: `buildContents` maps a media-bearing message to **one** `user` `contents` entry whose parts lead with a `{"inlineData":{"mimeType":…,"data":…}}` part per media part (after the tool result's `functionResponse` turn); base64 `StdEncoding`; the camelCase keys. **Test Scope**: T011's pin + the new feature's 4 Gemini scenarios + the rounds-062 scenarios (byte-identity, offered set, refusals) still green.
- [ ] **T016** `[BDD-REFACTOR]` — factor the blob builder into a small pure helper (`toInlineDataPart(mp)`) next to `toSDKBlob`-style helpers; keep the byte-identity control green. **Test Scope**: unchanged; all green.
- [ ] **T017** `[BDD-GREEN]` — the retired refusal: delete the `hasMedia` guard + helper from `internal/infrastructure/llm/gemini/client.go`; the family now carries media. **Test Scope**: T011's inverted pin green; the round-062 `offering-the-agent-tools` offered-set scenario still green.

### Feature 4B — the family-aware ceiling (US1 FR-004)

- [ ] **T018** `[BDD-GREEN]` — `read_image` takes the **resolved** ceiling (`NewReadImageTool(maxBytes int)`), replacing the hardcoded `imageMaxBytes`; the loud oversize error names the limit. **Test Scope**: T012's boundary pin green; the round-062 oversize scenario (OpenAI-compatible) still green.
- [ ] **T019** `[BDD-GREEN]` — the single-owned family→ceiling resolver + injection: the registry builder takes the resolved ceiling alongside the vision flag; the composition root resolves it from the selected provider's family. **Test Scope**: T013's pin green; the Gemini oversize scenario green (a >14 MiB image is refused before the wire).
- [ ] **T020** `[BDD-REFACTOR]` — de-duplicate the ceiling lookup (one table/owner; the tool holds one resolved scalar); all green.
- [ ] **T021** `[REGRESSION]` — full green: the round-062 feature + the round-063 feature + the rounds-013/030/061 Gemini scenarios + the unit pins.

### Feature 4C — CODE-REMOVE (the retired refusal surface)

- [ ] **T022** `[CODE-REMOVE]` — remove the now-dead refusal helpers/constants (the round-062 `hasMedia`; any unused media-refusal message) and the round-062 refusal pin's residual. **Test Scope**: no referencing test remains; `go build ./...` clean.

## Phase 5 — Verification

- [ ] **T023** — **Witness (a)**: remove the `inlineData` serialization → T011's wire pin RED → revert.
- [ ] **T024** — **Witness (b)**: move the injected ceiling (e.g. Gemini → 1 MiB) → the oversize boundary assertion RED → revert.
- [ ] **T025** — **Witness (c)**: re-add the Gemini text path drift (e.g. add a spurious key) → the byte-identity control RED → revert.
- [ ] **T026** — Gates: `gofmt` clean · `go vet ./...` clean · `go build ./...` clean · `go test -count=1 ./...` green (incl. the godog E2E) · `make verify` OK (incl. `modelith-check`) · topology audit **5 pre-existing, none new** · `go.mod`/`go.sum` unchanged.
- [ ] **T027** — Docs/ledger: the `docs/domain-model/tellme.modelith.{yaml,md}` refresh (the capability carried by both families) + `make modelith-check` green; `research.md` / `plan.md` / `tasks.md` / `truth-delta.md` owner rows finalised; the live spot-check (a real Vertex turn with an image — non-gating; witnesses RF-063-1/RF-063-2) recorded.
- [ ] **T028** — `SESSION-CLOSEOUT.md` Steps 1–8 after the human merge: propagate `dev → main`, tag `round-063` (operator approval), close nothing (operator request).

## Pre-Delivery Orphan Coverage Sweep

- `truth-delta.md` non-NOOP items: the 5 `techstack.md` rows (→ T015/T017/T018/T019/T004 read or deliver), ADR 0033 (+ the ADR 0032 annotation → Core Inputs + T015/T017/T019), the dsl-refine MODIFY (feature) + MODIFY (5 Then rows) + ADD (`## Given (round 063)`) (→ T008–T011/T015 Test Scope + Core Inputs) — **all covered**.
- `research.md` D1…D11: D1 (scope → T017), D2 (placement → T015), D3 (keys/base64 → T011/T015), D4 (family ceiling → T018/T019), D5 (capability unchanged → T004/T019), D6 (shared sniff → T018), D7 (seam → T018/T019), D8 (verification → T011/T012/T023), D9 (ADR → Core Inputs), D10 (byte-identity → T011/T016/T025), D11 (no new dep → T002/T026) — **all covered**.
- `specs/truth/techstack.md` round-063 rows: read in Core Inputs and delivered by T004/T015–T019 — **no orphan**.

## Implementation record (`/axb-implement`)

*(filled at implementation)*
