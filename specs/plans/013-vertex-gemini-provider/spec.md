# Feature Specification: tellme Vertex AI / Gemini Provider Support (round 013)

**Feature Branch**: `013-vertex-gemini-provider`

**Created**: 2026-09-13

**Status**: Draft — Clarify Round 1 resolved

**Input**: User request "tellme should support this provider" — the Vertex AI Gemini provider entry the operator is running with `tell-me-go`:

```yaml
  vertex-flash-3.8:
    TYPE: "gemini"
    MODEL: "gemini-3.8-flash"
    URL: "https://aiplatform.googleapis.com/v1/projects/masterspilitting-test/locations/global/publishers/google/models"
    API_KEY: "${NIFFLER_HOME}/secrets/key.json"
    THINKING_BUDGET: 32768
    THINKING_LEVEL: "HIGH"
    MAX_TOKENS: 40960
```

Today `tellme` adapts only the OpenAI-compatible family (`openai`/`deepseek`/`kimi`); a `TYPE: "gemini"` entry **loads and validates** but fails at turn time (`the provider request failed`, exit 6) because the transport factory rejects the family. This round **adds** the Vertex AI Gemini transport — behaviour intent **ADD**; it **MODIFIES** the factory's supported-family set and the provider request-assembly path.

The operator's live configuration (`$TELL_ME_HOME/configs/butler.yaml`) is a `tell-me-go`-shaped file that `tellme` already loads (its non-strict decode ignores the tell-me-go-only sections). It already contains `vertex-flash-3.8` plus eight further Vertex **gemini** entries (`vertex-flash`, `vertex-flash-lite`, `vertex-pro`, `dev`, `spilit`, `sit`, `sre`, `hub`) and three **anthropic** entries (`vertex-claude`, `vertex-claude-fable`, `claude`). The service-account file it points at exists and is a `"type": "service_account"` credential.

**Clarify Round 1 (2026-09-13)** resolved three high-impact decisions:

- **Q1 → Option 1 (scope)**: **Vertex AI only** — the entry's shape (`aiplatform.googleapis.com` + service-account). The Google Gemini API (`generativelanguage.googleapis.com`, inline key) and Application Default Credentials are deferred.
- **Q2 → Option 1 (credential)**: a `gemini` provider's expanded `API_KEY` ending in `.json` is a **service-account credential file**; any other shape is not supported this round (see FR-009). No config-schema change (reference parity).
- **Q3 → Option 1 (dependency)**: the service-account OAuth2 flow is **stdlib-only** (no provider SDK, no new module).

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Answering a prompt with a Vertex Gemini provider (Priority: P1)

As an operator, I want to point `SELECTED_PROVIDER` at a Vertex AI Gemini entry (`TYPE: "gemini"`, an `aiplatform.googleapis.com` URL) and have `tellme "<prompt>"` print the model's answer, so I can use the Gemini models my environment is already configured for.

**Why this priority**: it is the round's core capability and the reason the round exists; the credential handling (Story 2) only matters once the transport exists.

**Independent verification**: with `SELECTED_PROVIDER` selecting a `gemini` entry whose `URL` points at the in-process fake provider, run a prompt turn and confirm exactly one request is sent to the Vertex `:generateContent` endpoint shape (asserted from the fake's recorded request) and the model's answer is printed to `stdout`.

**Acceptance Scenarios**:

1. **Given** a configuration whose resolved provider has `TYPE: "gemini"` with a Vertex `aiplatform.googleapis.com` URL, **When** the operator runs `tellme "<prompt>"`, **Then** tellme sends exactly one Vertex `:generateContent` request (no family rejection) and prints the model's answer to `stdout`, exiting 0.
2. **Given** the same, **When** the provider returns a final answer, **Then** the candidate's text is normalized into the existing `llm.Response` and printed exactly as for the OpenAI-compatible family.
3. **Given** the entry declares `MODEL`, `MAX_TOKENS`, `THINKING_BUDGET`, and `THINKING_LEVEL`, **When** the request is sent, **Then** the URL is built from the configured project/location/publisher path and the model (not `<URL>/chat/completions`), and those limits map into the request's generation/thinking config.
4. **Given** the configuration sets `PERSON`, **When** the request is sent, **Then** the persona rides as the request's leading system instruction; an empty `PERSON` adds none.
5. **Given** tool definitions are registered (the round-008 registry), **When** the model requests a tool, **Then** the tool call is normalized into the existing `llm.ToolCall`, executed by the shared loop, and its result fed back — the same turn pipeline as the OpenAI-compatible family.
6. **Given** the provider reports usage metadata, **When** the turn completes, **Then** the measured payload status line is emitted exactly as for the OpenAI-compatible family; absent usage omits it.

**Functional Requirements**:

- **FR-001**: When the resolved provider's `TYPE` is `gemini` (or `google`), the system MUST route the turn to the Vertex AI Gemini transport instead of rejecting the family.
- **FR-002**: The transport MUST build a Vertex AI `:generateContent` request from the configured `URL` (project, location, publisher path), the resolved `MODEL`, and the assembled conversation; the `PERSON` MUST be sent as the leading system instruction when non-empty.
- **FR-003**: The transport MUST normalize the Vertex response — candidate content text, structured tool-call requests, and usage metadata — into the existing `llm.Response`, so answer printing, the tool loop, and the payload status line are unchanged downstream.
- **FR-004**: The transport MUST send `MAX_TOKENS` as the output-token cap and map `THINKING_BUDGET`/`THINKING_LEVEL` into the request's thinking config; an absent/zero value MUST fall back to the documented default and MUST NOT be sent as an override.
- **FR-005**: The transport MUST offer the registered tool definitions and normalize the model's tool-call requests, so the round-008 agent loop is consistent for a `gemini` provider.

**Non-Functional Requirements**:

- **NFR-001**: The Vertex request MUST be over stdlib `net/http` + `encoding/json` — no provider SDK.
- **NFR-002**: The transport MUST be exercisable hermetically in the E2E suite by pointing the provider `URL` at the in-process fake (no real egress), consistent with the existing fake-provider strategy.

---

### User Story 2 - Authenticating with the service-account key file (Priority: P2)

As an operator, I want `tellme` to authenticate to Vertex using the service-account JSON my configuration points at (`API_KEY: "${VAR}/secrets/key.json"`), so I do not have to mint or paste a short-lived token by hand.

**Why this priority**: it depends on Story 1 (the transport must exist) and is required for the provider to be usable at all.

**Independent verification**: with a service-account key file whose path ends in `.json`, confirm tellme exchanges it for an access token (against a fake token endpoint) and attaches `Authorization: Bearer <token>` to the Vertex request; confirm the token is reused within the run.

**Acceptance Scenarios**:

1. **Given** a `gemini` provider whose expanded `API_KEY` ends in `.json`, **When** a turn runs, **Then** tellme obtains an OAuth2 access token from the service-account credential and sends it as `Authorization: Bearer` on the Vertex request (the token-request target is injectable for the hermetic fake).
2. **Given** the run makes more than one request, **When** the requests are sent, **Then** the access token is obtained once and reused (cached) rather than re-minted per request.
3. **Given** a `gemini` provider whose credential is missing, not a `.json` path, or an unreadable/invalid key file, **When** a turn runs, **Then** the system fails with the frozen phrase `the provider request failed` and exit code 6 (with an actionable `stderr` message) and MUST NOT silently fall back to a bearer-key path.

**Functional Requirements**:

- **FR-006**: For a `gemini` provider, when the resolved `API_KEY` (after `${VAR}` expansion) ends in `.json`, the system MUST treat it as a service-account credential file and obtain a Vertex access token from it.
- **FR-007**: The obtained access token MUST be sent as `Authorization: Bearer <token>` on the Vertex request.
- **FR-008**: The access token MUST be obtained once and reused within the run (process-scoped cache), not re-minted per request.
- **FR-009**: A `gemini` provider whose `API_KEY` is empty, not a `.json` path, or an unreadable/invalid service-account file MUST fail with the frozen phrase `the provider request failed` (exit 6) — no silent bearer fallback, no unmapped provider error.
- **FR-010**: The service-account token exchange MUST be implemented with the standard library only (`crypto/*` JWT-RS256 + `net/http` POST to the token endpoint); no provider SDK and no new module.

**Non-Functional Requirements**:

- **NFR-003**: The OAuth token endpoint MUST be overridable through an injected seam so the E2E fake can serve the token exchange hermetically (offline-witness safe).
- **NFR-004**: The service-account file's secret material MUST NOT be logged; diagnostics may reference only its path and non-secret metadata.

---

### Edge Cases

- A `gemini` provider whose `URL` is **not** a Vertex (`aiplatform.googleapis.com`) endpoint is **out of this round's scope** (the Gemini API is deferred): a turn MUST fail with `the provider request failed` (exit 6), not a malformed request.
- An **Anthropic** provider (`vertex-claude`, `vertex-claude-fable`, `claude`) remains unsupported and MUST keep the existing unsupported-family failure (`the provider request failed`, exit 6).
- An **unknown/other** `TYPE` remains unsupported (existing behaviour).
- `PERSON` empty → no leading system instruction is sent.
- `MAX_TOKENS`/`THINKING_BUDGET` zero → the documented default applies; no override field is sent.
- The `${VAR}` in `API_KEY` is **unset** (`NIFFLER_HOME` not exported) → the existing expansion failure (`the provider configuration is invalid`, exit 3) is unchanged (it happens before the transport).
- The token endpoint is unreachable or rejects the credential → `the provider request failed` (exit 6).
- The model returns only a tool call (no text) → the loop continues unchanged; that step's answer text is empty.
- `usageMetadata` absent → no post-turn measured status line (existing no-usage behaviour).
- A `gemini` provider selected on the `-d` / boot / `-l` paths constructs **no** transport → unchanged (no network, no credential read).

## Requirements *(mandatory)*

> Story-specific FR / NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-011**: The round MUST NOT change the request shape or behaviour of the OpenAI-compatible family (`openai`/`deepseek`/`kimi`) — those requests MUST remain byte-identical to rounds 004–012.
- **FR-012**: The round MUST NOT regress rounds 001–012; all prior acceptance scenarios MUST remain green, and the frozen class-phrase vocabulary MUST be unchanged (`the provider request failed` = exit 6; `the provider configuration is invalid` = exit 3).
- **FR-013**: A non-`gemini`, non-OpenAI-family provider MUST keep its existing unsupported-family failure — the fresh family mapping MUST NOT accidentally widen support.

#### Non-Functional Requirements

- **NFR-005**: The round MUST NOT add a new module (`go.sum` unchanged) — the service-account OAuth2 flow and the Vertex transport are stdlib-only (Q3 = stdlib-only).
- **NFR-006**: All new assertions MUST be deterministic (no `time.Sleep`); the fake provider + the injectable token endpoint are the verification surface.

### Key Entities *(include if feature involves data)*

- **Vertex Gemini provider entry**: a `PROVIDERS` entry with `TYPE: "gemini"`, an `aiplatform.googleapis.com` `URL`, `MODEL`, a service-account `API_KEY` path, and optional `MAX_TOKENS`/`THINKING_BUDGET`/`THINKING_LEVEL`/`HEADERS`.
- **Service-account credential**: the `.json` key file the entry points at — exchanged for a Vertex access token.
- **Vertex access token**: the `Bearer` credential attached to each Vertex request; cached for the run.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With a `gemini` provider selected and its `URL` pointed at the fake, 100% of prompt runs send exactly one Vertex `:generateContent` request (fake-recorded) and print the model's answer to `stdout`, exit 0.
- **SC-002**: 100% of `.json` credentials produce an `Authorization: Bearer` header minted from a service-account access token, and the token is reused within the run.
- **SC-003**: A `gemini` provider with a missing / non-`.json` / unreadable credential fails with `the provider request failed` and exit 6 — never a silent bearer fallback.
- **SC-004**: OpenAI-compatible requests stay byte-identical to rounds 004–012, and every prior acceptance scenario remains green.
- **SC-005**: No new module is introduced (`go.sum` unchanged).
- **SC-006**: The Vertex transport is carried by at least one executable interface Rule in `specs/truth/features/cli/**`, and the Gherkin/DSL topology audit passes.

## Assumptions

- **A1 (scope)**: Vertex AI only — `aiplatform.googleapis.com` + service-account. The Google Gemini API (`generativelanguage.googleapis.com`, inline key) and Application Default Credentials are **deferred** (Q1 = Vertex-only).
- **A2 (credential detection)**: the `.json`-suffix heuristic on the expanded `API_KEY` distinguishes a service-account file from an inline key (reference parity) (Q2).
- **A3 (dependency)**: the OAuth2 service-account flow is stdlib-only; no new module (Q3).
- **A4 (credential timing)**: the credential file is read at **turn time** (transport construction); a missing/unreadable/invalid file surfaces as `the provider request failed` (exit 6). `-d`/boot/`-l` construct no transport and are unchanged.
- **A5 (request shape)**: the Vertex request follows the reference shape — project/location/publisher path parsed from the configured `URL`, the model appended, `:generateContent` — over stdlib HTTP.
- **A6 (testability)**: the provider `URL` (fake provider) and the OAuth token endpoint are both injectable, so the whole path is E2E-hermetic without real egress.
- **A7 (unchanged)**: single-turn execution, stdin piping, rendered/raw output, durable history, the agent tool loop (except the gemini request/response mapping), the payload status line, cross-stream ordering, persona-on-the-wire, and the interactive multi-line prompt are unchanged. No streaming, MCP, memory, pinning, pruning, `-b`/`--retry`, or `SafePath`/consent.
