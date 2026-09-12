# Truth Delta: 004-first-reasoning-turn

**Plan Package**: `specs/plans/004-first-reasoning-turn`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **ADD** a new **Reasoning & Provider Transport** section: `Provider gateway port` = `internal/domain/llm` (`Gateway` interface + `Request`/`Response`/`ProviderError`); `OpenAI-compatible adapter` = `internal/infrastructure/llm/openai`; `HTTP transport` = stdlib `net/http` (+ `encoding/json`, no SDK); `Request assembly` (endpoint `<URL>/chat/completions`, `Authorization: Bearer`, `MODEL`, `MAX_TOKENS`, merged `HEADERS`, `reasoning_effort`); `Response normalization` (`choices[0].message.content`). | Round-004 research Decisions 1–5: the provider gateway architecture, transport, request assembly, response normalization, and the deterministic failure contract (class phrase + exit code `6`). |
| MODIFY | `specs/truth/techstack.md` | **CLI Application**: extended the `Project layout` row to add `internal/domain` (port) + `internal/infrastructure` (adapter); added a `Prompt input` row (first positional argument; dispatch precedence `--version` → `-d` → prompt turn → boot). | Round-004 Decision 1 (layered seam) + Decision 7 (positional prompt handling). |
| MODIFY | `specs/truth/techstack.md` | **Configuration**: updated the `Effective-value resolution & validation` row to note the resolved provider is now consumed by the round-004 reasoning path. | Round-004 Decision 3 consumes `resolution.Provider` (round-003 review finding #3). |
| MODIFY | `specs/truth/techstack.md` | **Testing & Verification**: added a `Local fake provider` row (`net/http/httptest`); **re-scoped** the `No-network verification` row from a whole-binary build-graph capability guard to offline-path witnesses (no-dial canary + differential no-egress sandbox) and recorded the guard's retirement; extended the `Pure-helper unit tests` row with request assembly + response normalization. | Round-004 Decision 6: the network-path test strategy and the amendment of round-001 Decision 5's capability guard (the chat path legitimately links `net/http`). |
| MODIFY | `specs/truth/techstack.md` | **Not Introduced Yet**: reworded the provider-technology bullet (SDKs excluded — the transport uses stdlib `net/http`); added Gemini/Vertex + Anthropic adapters, streaming (SSE), and the full provider-agnostic `Thought` model / tool-call shapes. | Round-004 Decisions 2 & 4: scope of the first provider transport and the deferred `Thought` model. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and no HTTP/OpenAPI surface of its own; the only outbound HTTP is the external provider call, which is not an API contract tellme exposes. | `contract-authoritative` holds vacuously. The provider request/response shape is CLI-end behaviour owned by `/axb-dsl-refine` (`specs/truth/features/cli/**`). |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. Round 004 is a single in-memory, non-streaming turn (Clarify Q1) — no `history.jsonl`, no session state, and no persisted or in-memory model is introduced. | `data-model-covers-all-state` holds vacuously. Session durability is deferred to a later slice. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/dsl.md` | New **chat** module DSL: 5 Given rows (a configured provider whose endpoint answers with an answer / two configured providers that answer / answers with an error status / endpoint unreachable / no usable answer), 2 When rows (`the operator starts tellme with the prompt "{prompt}"`, `the operator runs tellme's diagnostic with the prompt "{prompt}"`), 3 Then rows (sends exactly one request to the provider / prints the provider's answer / exits with the provider error code `6`). | Round-004 CLI-end + provider-gateway interfaces (`plan.md`); carries the two acceptance features under a new `chat` module. |
| ADD | `specs/truth/features/cli/chat/answering-a-single-prompt.feature` | New interface feature — Rule *A prompt sends exactly one request to the selected provider and prints the answer* (2 Examples) + Rule *A run outside the normal prompt turn makes no provider request* (2 Examples). | Carries acceptance `answering-a-single-prompt.feature` (prompt turn + offline-preservation edges). |
| ADD | `specs/truth/features/cli/chat/reporting-a-failed-provider-request.feature` | New interface feature — Rule *An unreachable provider or an error response is reported with the frozen class phrase* (2 Examples) + Rule *A response that cannot be understood as an answer fails the same way* (1 Example). | Carries acceptance `reporting-a-failed-provider-request.feature`. |
| MODIFY | `specs/truth/features/cli/dsl.md` (interface root) | (i) Added the ninth frozen class phrase `the provider request failed` to the `tellme explains on stderr that "{reason}"` vocabulary (8 → 9). (ii) **Moved** the Given row `the selected provider override is "{provider}"` from `configuration/dsl.md` to the root (now cross-module: configuration + chat; **semantics unchanged**; old position `configuration/dsl.md` → new position `cli/dsl.md`). (iii) **Moved** the Then row `tellme performs no network access` from `diagnostics/dsl.md` to the root (now cross-module: diagnostics + chat) and **MODIFIED its semantics** to the offline-path canary (recording sink + differential witness), recording that round 004 retired the whole-binary capability guard. (iv) MODIFIED `the operator starts tellme` to note **no prompt**. | Round-004 FR-009 / Clarify Q3 (new failure class + exit `6`) and the re-scoped offline-path no-network witness (research Decision 6). |
| MODIFY | `specs/truth/features/cli/configuration/dsl.md` | Removed the Given row `the selected provider override is "{provider}"` (promoted to the interface root — DSL single-authority). | The row is now used by two modules, so its authority moved to the interface root. |
| MODIFY | `specs/truth/features/cli/diagnostics/dsl.md` | Removed the Then row `tellme performs no network access` (promoted to the interface root); updated the module header note to point at the root row and record the retired capability guard. | DSL single-authority + the re-scoped offline-path witness (research Decision 6). |
