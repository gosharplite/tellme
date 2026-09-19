# Tasks — Agent image vision: `read_image` to the provider wire (round 062)

**Plan Package**: `specs/plans/062-agent-image-vision`
**Execution**: `/axb-implement` — Red → Green → Refactor, one task at a time, marked `[X]` after verification.

## Core Inputs (Read before any task)

- `specs/plans/062-agent-image-vision/spec.md` — US1–US3 · FR-001…FR-012 · SC-001…SC-006 · S-1…S-8 · I-1…I-5.
- `specs/plans/062-agent-image-vision/research.md` — D1…D11 + residual risks.
- `specs/plans/062-agent-image-vision/plan.md` — the 1 interface, the wave, and the 10-site change surface.
- `specs/plans/062-agent-image-vision/truth-delta.md` — the owner rows (techstack ADD/MODIFY ×4, api/data NOOP, dsl-refine ADD/MODIFY ×3, the ADR ADD).
- `specs/truth/techstack.md` — *Image filesystem tool (`read_image`)* · *Image content on the provider wire (OpenAI-compatible)* · *Provider entry schema* · *Agent tool schemas* (the round-062 rows).
- `specs/truth/features/cli/chat/reading-a-local-image.feature` · `…/offering-the-agent-tools.feature` · `…/dsl.md` (`## Given (round 062)` / `## Then (round 062)`).
- `docs/decisions/0032-agent-image-vision.md` — the durable decision + §Forward RF-062-1…8.

## Phase 1 — Setup

- [X] **T001** — Branch `062-agent-image-vision` off `dev` `902642a`; plan package created (spec + checklist + truth-delta skeleton; `STATUS.md` phase-gate updated).
- [X] **T002** — `/axb-clarify` CLOSED (Q1 → 1 · Q2 → 1 · Q3 → 1 · Q4 → 1 · Q5 → 1) recorded in `spec.md` (S-1…S-8) and `checklists/requirements.md`.
- [X] **T003** — **No new technology**: no Setup phase for packages/toolchains — stdlib only, `go.mod`/`go.sum` MUST stay unchanged (research D6/D11). Smoke check: `go build ./...` + `go test -count=1 ./internal/infrastructure/llm/openai/` green before touching anything.

## Phase 2 — Foundational (only scaffolding; NO product behaviour, NO tests)

- [X] **T004** — **Only**: the domain media carrier — add `MediaPart{ MIMEType string; Data []byte }` and `llm.Message.Media []MediaPart` in `internal/domain/llm/gateway.go` (a new optional field; doc it as "empty ⇒ the message serializes exactly as today"). **Not**: any serialization, sniffing, or loop logic. Compiles; existing tests still green (byte-identity holds trivially).
- [X] **T005** — **Only**: the config capability key — add `VISION bool` to `internal/config.Provider` (typed field; YAML `VISION`) and expose it on the resolved provider the factory derives. **Not**: any gate/serialization use. Compiles; a config with no `VISION` key still loads (default false).
- [X] **T006** — **Only**: the image-tool shell — create `internal/infrastructure/tools/image.go` with `newReadImageTool(...)` returning a `domaintools.Tool` whose `Name()`/`Description()`/`Parameters()`/`Contract()` are declared (schema built via the shared `resourceSchema` so `required ⊆ properties` holds for `filepath` + `reason`) and whose `Execute` returns an explicit `not implemented yet` error. **Not**: the read/sniff/size logic, not the registry wiring.
- [X] **T007** — **Only**: the stepdef landing skeleton file `tests/e2e/steps/step_r062_image.go` (package + imports + 13 empty step functions registered through the shared `register`), plus the fake-provider recording hooks needed later (declared, unused). **Not**: any assertion body.
- [X] **T008** — **Only**: the unit-test landing files — `internal/infrastructure/tools/image_test.go`, `internal/infrastructure/llm/openai/client_image_test.go` — with the test-function names declared as `t.Skip("RED phase")`. **Not**: assertions.

## Phase 3 — Test Alignment & Implementation (tests only; NO product code)

**DSL 参照** — every round-062 sentence's authority is `specs/truth/features/cli/chat/dsl.md`; read each row's `StepDef 實作語意` for the arrangement/assertion contract. Reused sentences keep their existing rows.

**Markers** — `[BDD-ALIGN]` (reuse an existing stepdef) · `[BDD-RED]` (write a new stepdef that fails until the product lands) · `[P]` (independently parallelizable — lands in its own file).

**Boundary** — Phase 3 writes **no** product file; it lands every stepdef and unit pin RED.

**Parallel Hint** — T010–T022 are `[P]` (one task per DSL row; each new stepdef is a distinct function in `step_r062_image.go`; the unit pins land in the two named files).

- [X] **T009** `[BDD-ALIGN]` — reuse-detect: the reused sentences (`the operator has a runnable tellme installation`, `the runtime home is "ait-tmg"`, `the operator starts tellme with the prompt "…"`, `tellme prints the provider's answer "…"`, `tellme exits successfully`, `the request offered exactly the agent tools`) already have stepdefs — no new work; confirm the two feature files name them byte-exactly.
- [X] **T010** `[P][BDD-RED]` — stepdef for the Given `the workspace holds an image file "{name}" that is a {kind} picture` (write a minimal valid PNG/JPEG/GIF/WebP by magic bytes into the run's working directory).
- [X] **T011** `[P][BDD-RED]` — stepdef for the Given `the workspace holds an image file "{name}" larger than the image size limit` (> 32 MiB; header + padding, written sparsely).
- [X] **T012** `[P][BDD-RED]` — stepdef for the Given `the workspace holds a file "{name}" whose content is not a supported picture` (plain text bytes under an image-like name).
- [X] **T013** `[P][BDD-RED]` — stepdef for the Given `a configured provider "{provider}" that can take images and whose endpoint asks tellme to read "{path}" and then answers with "{answer}"` (config entry declares `VISION: true`; fake records the body + scripts one `read_image` call).
- [X] **T014** `[P][BDD-RED]` — stepdef for the Given `a configured provider "{provider}" that can take images whose endpoint reports the offered tools and then answers with "{answer}"`.
- [X] **T015** `[P][BDD-RED]` — stepdef for the Given `a configured provider "{provider}" that cannot take images whose endpoint reports the offered tools and then answers with "{answer}"` (no `VISION` key).
- [X] **T016** `[P][BDD-RED]` — stepdef for the Then `the request carried the image file "{name}"` (decode the recorded `image_url` block; bytes equal the file's, exactly).
- [X] **T017** `[P][BDD-RED]` — stepdef for the Then `the image was attached as a "{mime}" picture`.
- [X] **T018** `[P][BDD-RED]` — stepdef for the Then `the request carried no image`.
- [X] **T019** `[P][BDD-RED]` — stepdef for the Then `the request offered the read-image tool` (the recorded `tools[]` carries a well-formed `read_image` declaration).
- [X] **T020** `[P][BDD-RED]` — stepdef for the Then `the request offered no read-image tool`.
- [X] **T021** `[P][BDD-RED]` — stepdef for the Then `the tool result reported the image is too large`.
- [X] **T022** `[P][BDD-RED]` — stepdef for the Then `the tool result reported the content is not a supported picture`.
- [X] **T023** `[P][UNIT]` — pins (RED) in `internal/infrastructure/tools/image_test.go`: the sniff table (JPEG/PNG/GIF/WebP magic bytes → MIME; unknown/empty → the loud error); the 32 MiB boundary both ways.
- [X] **T024** `[P][UNIT]` — pins (RED) in `internal/infrastructure/llm/openai/client_image_test.go`: a media-bearing message serializes to a content **array** with a leading text part + the `image_url` data-URI; a text-only message serializes identical to the pre-round body (I-1 byte-identity control); a media-bearing message on the **gemini** adapter returns a loud `*llm.ProviderError`.
- [X] **T025** `[REVIEW]` — subagent review gate: confirm 0 undefined steps (`go test ./tests/e2e/ -run TestE2ESuite` reports no undefined), every round-062 DSL row matched exactly once, and the two unit-pin files fail RED for the right reason.

## Phase 4 — Feature phases (product code; NO red/align here)

### Feature 4A — `reading-a-local-image.feature` ADD (US1/US2/US3)

- [X] **T026** `[BDD-GREEN]` — implement the read path end to end: the sniffer (magic bytes), the 32 MiB ceiling, `read_image`'s read+attach, the domain `MediaPart` on the result channel, the loop's `user`-message placement (media-first, after the `tool` result), and the openai content-array serialization; the gemini loud refusal. **Test Scope**: the two unit-pin files (T023/T024) + the 5 new E2E scenarios of the new feature.
- [X] **T027** `[BDD-REFACTOR]` — factor the sniffer + the data-URI builder into small pure helpers (single-responsibility, table-driven); keep the byte-identity control green. **Test Scope**: unchanged; all green.

### Feature 4B — `offering-the-agent-tools.feature` MODIFY (the capability-dependent set)

- [X] **T028** `[BDD-GREEN]` — the vision flag: `agentTools()` gains a vision input and assembles `read_image` only when the selected provider declares `VISION: true`; wire the flag from the resolved provider at `cmd/tellme`; keep the round-031 well-formedness gate green for both variants. **Test Scope**: the offered-set E2E scenarios (both features) + `TestAgentToolSchemasAreWellFormed`.
- [X] **T029** `[BDD-REFACTOR]` — de-duplicate the two assembler variants (one builder parameterised by the flag, or two thin wrappers) so the round-031 gate reads one production path; E2E green.

## Phase 5 — Verification

- [X] **T030** — **Witness (a)**: remove the image serialization → `client_image_test.go` RED → revert.
- [X] **T031** — **Witness (b)**: invert the capability flag → the offered-set assertion RED → revert.
- [X] **T032** — **Witness (c)**: drop the 32 MiB ceiling → the oversize E2E scenario RED → revert.
- [X] **T033** — Gates: `gofmt` clean · `go vet ./...` clean · `go build ./...` clean · `go test -count=1 ./...` green (incl. the godog E2E) · `make verify` OK (recorded in the round record) · topology audit **5 pre-existing, none new** · `go.mod`/`go.sum` unchanged.
- [X] **T034** — Docs/ledger: `research.md`, `plan.md`, this `tasks.md`, `truth-delta.md` owner rows marked; the live spot-check (a real vision endpoint, non-gating) recorded.
- [X] **T035** — `SESSION-CLOSEOUT.md` Steps 1–8 after the human merge: propagate `dev → main`, tag `round-062` (operator approval), close nothing (operator request).

## Pre-Delivery Orphan Coverage Sweep

- `truth-delta.md` non-NOOP items: the 4 `techstack.md` rows (→ T004/T005/T006/T026/T028 read or deliver), the ADR 0032 (→ Core Inputs + T026), the dsl-refine ADD/MODIFY ×3 (→ T026/T028 Test Scope + Core Inputs) — **all covered**.
- `research.md` D1…D11: D1 (entry point → T006), D2/D5 (wire/serialization → T004/T024/T026), D3 (capability → T005/T028), D4/D7 (placement → T026), D6 (sniff → T023/T026), D8 (gemini refusal → T024/T026), D9 (offered set → T028), D10 (ADR → Core Inputs), D11 (verification → T023/T024/T030) — **all covered**.
- `specs/truth/techstack.md` round-062 rows: read in Core Inputs and delivered by T004–T028 — **no orphan**.

## Implementation record (`/axb-implement`)

All 35 tasks `[X]`. Delivered sites: `internal/domain/llm/media.go` (MediaPart + collector) · `llm.Message.Media` · `config.Provider.Vision` · `internal/infrastructure/tools/image.go` (sniff + 32 MiB ceiling + attach) · `openai.messageContent` (content array / byte-identical text path) · `gemini` loud media refusal · the loop's per-call collector + `user`-message placement · `cmd/tellme` vision-gated assemblage (+ `deps.NewToolRegistry` widened) · the E2E stepdefs.

**Gates**: `gofmt` clean · `go vet ./...` clean · `go test -count=1 ./...` **green** · `make verify` **OK** (arch 0 · modelith-check up to date · lint 0 · govulncheck clean · cross-compile 4/4) · topology audit 5 pre-existing, **none new** · E2E **259 scenarios / 1918 steps, all passed** · `go.mod`/`go.sum` unchanged.

**Witnesses (reproduced RED → reverted)**: (a) removing the image serialization reds `TestMessageContentPinsTheWireShape` + `TestCompleteCarriesImageOnTheWire`; (b) inverting the capability flag reds the offered-set Examples (19 E2E failures); (c) disabling **both** ceiling checks (the `Stat` size check and the read-length check) reds the oversize scenario.

**Domain model**: `docs/domain-model/tellme.modelith.{yaml,md}` refreshed (ADR 0030) — the `Provider.vision` attribute, the `Tool` capability gate + `tool-offered-only-when-capable`, the new `ImageContent` entity, and a *Reading a local image* scenario; `make modelith-check` green.
