# Phase 0 Research: tellme Vertex AI / Gemini Provider Support (Round 013)

Topic: add the **Vertex AI Gemini transport** to `tellme`, so a `PROVIDERS` entry with `TYPE: "gemini"` and a service-account key reaches the model instead of being refused. Today `tellme` adapts only the OpenAI-compatible family (`openai`/`deepseek`/`kimi`); `internal/infrastructure/llm/factory.go` rejects `gemini`, so the operator's `vertex-flash-3.8` entry loads, validates, and then fails at turn time with `the provider request failed` (exit 6). The reference `tell-me-go` ships Gemini/Vertex support (through the Google GenAI SDK); `tellme` deferred it in round 004.

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer (`gopkg.in/yaml.v3` + hand-written resolution/expansion), testing harness (`godog` + stdlib `testing`), the provider gateway port (`internal/domain/llm`), the OpenAI-compatible adapter, output rendering (glamour), session history (append-only JSON-Lines), the agent tool loop (round 008), the payload status line (round 009), cross-stream ordering (round 010), persona-on-the-wire + estimator (round 011), and the interactive prompt read (round 012) were locked in rounds 001–012. The system still has **one CLI end**. This round reaches **new external endpoints** (the Vertex AI `:generateContent` API and the Google OAuth token endpoint) but adds **no new module** (`go.sum` unchanged). The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the built binary; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers. Clarify Round 1 (in `spec.md`) locked the round's three high-impact decisions: Q1 = **Vertex AI only**; Q2 = **`.json`-suffix service-account detection**; Q3 = **stdlib-only OAuth2**. **IN this round**: the Vertex/Gemini transport, the service-account access-token flow, the credential boundary + failure taxonomy, and their executable assertions. **OUT**: the Google Gemini API family (`generativelanguage.googleapis.com`, inline key), Application Default Credentials, Anthropic, streaming, and anything beyond the gemini family.

---

## Decision 1: A new Vertex/Gemini adapter, selected by the factory's family mapping

- **Decision**: Add a new adapter (proposed `internal/infrastructure/llm/gemini`) implementing the existing `llm.Gateway` port over the Vertex AI `:generateContent` REST API, and extend `factory.go` so the resolved provider `TYPE` `gemini` (or `google`) maps to it. The adapter honours the same `llm.Request`/`llm.Response`/`llm.ProviderError` value types as the OpenAI adapter, so the answer path, the tool loop, and the payload status line are unchanged downstream.
- **Rationale**: The `llm.Gateway` port and the factory family-dispatch seam are the round-004 extension points designed exactly for this. A second adapter keeps the two wire families isolated (the OpenAI adapter stays byte-identical) and gives an un-adapted family a single actionable failure point.
- **Alternatives considered**:
  - **Extend the OpenAI adapter with a Vertex mode** — the Vertex `:generateContent` wire shape (contents/parts, `functionCall`, `usageMetadata`) shares almost nothing with OpenAI Chat Completions; a mode flag would entangle two protocols in one adapter — rejected.
  - **Adopt the Google GenAI SDK (`google.golang.org/genai`)** — the reference's choice; it would add a module and break the stdlib-transport precedent (round 004 Decision 2) — rejected (Q3 = stdlib-only).

## Decision 2: The Vertex `:generateContent` request/response mapping

- **Decision**: Build requests as `POST <base>/v1/projects/<project>/locations/<location>/<publisher-path>/<model>:generateContent` with a JSON body `{contents:[{role,parts:[{text}]}], systemInstruction:{parts:[{text}]}, generationConfig:{maxOutputTokens, thinkingConfig:{thinkingBudget, thinkingLevel}}, tools:[{functionDeclarations:[{name,description,parameters}]}]}`. Normalize the response from `candidates[0].content.parts[]` (text parts and `functionCall` parts) into `llm.Response.Text`/`ToolCalls`, and `usageMetadata{promptTokenCount,candidatesTokenCount,totalTokenCount}` into `llm.Usage`. The persona (`PERSON`) becomes the `systemInstruction` when non-empty; a zero `MAX_TOKENS`/`THINKING_BUDGET` omits the corresponding override field. Prior conversation messages are re-mapped to Vertex `contents`: the assistant's tool calls go under `role: "model"` with `functionCall` parts, and tool **results** under `role: "user"` with `functionResponse` parts — Vertex REST **rejects `role: "tool"`** (an OpenAI-only shape). The endpoint is built **directly from the configured `URL`** with **no hostname validation**, so a loopback Vertex-shaped fake passes transparently. *(PR #33 architectural-review directives D1/D2.)*
- **Rationale**: This is the standard Vertex AI REST shape and mirrors the reference's request/response semantics (persona as system instruction, max-output-tokens cap, thinking config, function declarations, usage metadata), which is what makes `tellme`'s downstream — the tool loop (round 008) and the measured status line (round 009) — work unchanged.
- **Alternatives considered**:
  - **`v1beta1` endpoint** — the operator's configured URL is `…/v1/…`; `v1` is the stable surface — rejected.
  - **Map tools through a flattened/opaque schema** — the round-008 loop needs structured `functionDeclarations` + `functionCall` round-trip parity — rejected.

## Decision 3: Service-account OAuth2 with the standard library only

- **Decision**: For a `gemini` provider, obtain the access token from the service-account JSON with stdlib only: parse `client_email`/`private_key`/`token_uri`, build a JWT claim set (`iss`=client_email, `scope`=`https://www.googleapis.com/auth/cloud-platform`, `aud`=token_uri, `iat`/`exp`), sign it **RS256** (`crypto/rsa` + `crypto/x509` PEM parsing + `crypto/sha256`), and exchange it at the token endpoint (`POST` `grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=<jwt>` with `net/http`), reading `access_token` + `expires_in`. Attach `Authorization: Bearer <token>` to every Vertex request.
- **Rationale**: Q3 = stdlib-only. The service-account flow is a bounded, well-specified JWT-bearer exchange; keeping it stdlib preserves the round-004 "no provider SDK" precedent and the `go.sum`-unchanged norm (round 012).
- **Alternatives considered**:
  - **`golang.org/x/oauth2/google`** — the smallest helper dependency; still a new module and a new transitive graph — rejected (Q3).
  - **`google.golang.org/api` / the GenAI SDK** — heavier; same objection — rejected.

## Decision 4: Credential detection (`.json`) and the failure taxonomy (exit 6, no fallback)

- **Decision**: A `gemini` provider's expanded `API_KEY` whose value ends in `.json` (case-insensitive) is a **service-account credential file**; any other `gemini` credential shape is **unsupported this round**. A credential that is missing, not a `.json` path, or an unreadable/invalid key file fails at **turn time** with the frozen phrase `the provider request failed` (exit 6) and an actionable `stderr` message — **no** silent fallback to a bearer-key path and no unmapped provider error.
- **Rationale**: Q2 = `.json`-suffix detection (reference parity; no config-schema change). The credential read is a **request-path** concern (transport construction), so its failure belongs to the existing provider-failure class (exit 6) — not the boot-resolvable configuration-invalid class (exit 3), which stays reserved for values like an unset `${VAR}` in `API_KEY` (that expansion failure is unchanged). Crucially, `tellme` must **not** copy the reference's quirk (a failed `os.Stat` silently falling back to `APIKeyAuth`, sending the path text as a key).
- **Alternatives considered**:
  - **A dedicated config field (`API_KEY_FILE`/`CREDENTIALS_FILE`)** — explicit, but changes the provider config contract and diverges from the operator's existing entry — rejected (Q2).
  - **Classify a bad credential as `the provider configuration is invalid` (exit 3)** — would require a boot-time credential check, contradicting "the credential is read at turn time" and splitting the auth path across two classes — rejected.
  - **Silent bearer fallback** (the reference's `resolveGoogleAuth` fall-through) — sends the path string as a key and masks the real problem — rejected.

## Decision 5: Process-scoped in-memory token cache

- **Decision**: Cache the minted access token **in memory** on the adapter, keyed by the credential, and reuse it across requests within the run until near expiry (re-minting on expiry or a 401). The token is **not** persisted to disk. Access is guarded by a `sync.RWMutex` with double-checked locking and a **~60s expiry safety margin**, so a token at the edge of expiry is re-minted rather than used. *(PR #33 architectural-review directive D3.)*
- **Rationale**: `FR-008` requires reuse within the run; in-memory caching satisfies it without introducing a new persisted state artifact (which would pull in the data truth and a cache-invalidation surface). The reference persists a token cache, but that is a larger surface than this round needs.
- **Alternatives considered**:
  - **No cache (re-mint per request)** — an extra token round-trip on every tool-loop iteration — rejected (violates `FR-008`).
  - **A persisted disk cache** — a new state artifact + invalidation semantics; deferred to a later round if ever needed — rejected for now.

## Decision 6: Hermetic verification via the existing fake + the credential's `token_uri`

- **Decision**: Verify the whole Gemini path end-to-end without real egress: point the provider `URL` at the existing in-process fake (as rounds 004+) **and** point the test service-account credential's `token_uri` at a fake token endpoint, so the JWT-bearer exchange is served locally too. No product-side test-only env knob is introduced — the credential's own `token_uri` field is the natural token-endpoint seam. The request shape (Vertex body + `Authorization: Bearer`) is asserted from the fake's recorded request.
- **Rationale**: `NFR-003` + `NFR-006`; the fake-provider strategy already records requests. Honouring the credential's `token_uri` is both correct production behaviour (the field is part of the SA JSON) and the seam tests need — no new knob. The fake dispatches by **request path** — `…/token` → the OAuth2 exchange JSON, `…:generateContent` → the Vertex body, default → the existing OpenAI handler — so the existing OpenAI scenarios do not regress. *(PR #33 architectural-review directive D4.)*
- **Alternatives considered**:
  - **A `TELL_ME_MOCK_URL`-style env override for the token endpoint** — an extra product env knob purely for tests; the reference uses one for the base URL, but tellme's provider `URL` is already configurable and the token endpoint has a natural field — rejected as unnecessary.
  - **A pty / real-egress integration test** — forsworn and offline-hostile — rejected.

## Decision 7: No new *module*; BDD techstack & strategy unchanged

- **Decision**: The round adds **no** third-party dependency (`go.mod` / `go.sum` unchanged) and does not change the BDD techstack (`godog` running the built binary) or the strategy (E2E acceptance path + fast unit tests for pure helpers). The transport and the OAuth2 flow are stdlib-only.
- **Rationale**: Both the Vertex REST mapping and the JWT-bearer exchange are plain Go over `net/http` + `crypto/*`; the harness already records requests.
- **Alternatives considered**:
  - **A provider SDK / `x/oauth2`** — a new module with no need — rejected (Q3).

## Decision 8: The three AIxBDD must-ask questions remain settled

- **Decision**: No new system end; BDD techstack = `godog`; strategy = E2E + pure-helper units — unchanged this round, per the standing `techstack.md`. No `/axb-clarify` round is owed for them.
- **Rationale**: The round extends an existing CLI end's provider transport; it introduces no new interface, service, or test framework, and no new system end.
- **Alternatives considered**:
  - **Re-open the techstack questions** — no change to the ends or the runner — rejected.

---

## Residual risks / forward links

- **Exact Vertex field names (Decision 2)**: the JSON field names (`contents`/`parts`/`systemInstruction`/`generationConfig`/`thinkingConfig`/`functionDeclarations`/`functionCall`/`usageMetadata`) follow the Vertex REST reference but are **not yet verified against the live API on this host**; the implementation/DSL phase must confirm them (a `/axb-dsl-refine`/implementation determination).
- **Thinking-config mapping (Decision 2)**: how `THINKING_LEVEL` (`"HIGH"`) and `THINKING_BUDGET` map into `thinkingConfig` for a `gemini-3.8-flash` model is model-dependent; the round pins that they are **sent** (or omitted when zero), not that the API accepts every combination — a residual to verify live.
- **`google` TYPE alias (FR-001)**: whether `TYPE: "google"` is accepted in addition to `"gemini"` is a `/axb-dsl-refine`/implementation determination (held open here).
- **Credential-failure class placement (Decision 4)**: exit 6 (`the provider request failed`) is chosen; revisit only if the implementation reveals a cleaner boundary with the configuration-invalid class (exit 3).
- **Token expiry / 401 handling (Decision 5)**: the exact skew and the 401-triggered re-mint are implementation details; only "obtained once and reused within the run" is pinned.
- **Auth seam shape**: whether the OAuth2 flow lives in the gemini adapter or a shared `internal/infrastructure/auth` helper is an implementation determination (like round-011's persona-plumbing seam); Decision 3 fixes *what it must do*, not its package shape.
- **Deferred, still out of scope**: the Google Gemini API family (inline key), Application Default Credentials, Anthropic, streaming, MCP, memory, pinning, pruning, `-b`/`--retry`, `SafePath`/consent.
