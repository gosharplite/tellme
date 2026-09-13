# System Analysis Plan — round 013 (`013-vertex-gemini-provider`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/013-vertex-gemini-provider/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/
│   └── acceptance/
│       ├── answering-with-a-gemini-model.feature
│       ├── authenticating-with-a-service-account-key.feature
│       └── refusing-providers-tellme-cannot-drive.feature
└── tasks.md                       # produced later by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — updated this round (Vertex/Gemini transport + auth)
├── data/**                        # /axb-data-plan — NOOP this round (no persisted-state change)
└── features/                      # executable CLI Gherkin + DSL — /axb-dsl-refine
    └── cli/{chat,providers}/…     # ADD/MODIFY — the Vertex/Gemini provider contract
```

*(No `contracts/**` truth artifact in this round — see the `NOOP` note for `/axb-api-plan` below. `data/**` is **`NOOP`**: the token cache is in-memory only (research Decision 5), so no persisted-state change.)*

### Source-code structure (repository root)

```text
cmd/tellme/
└── main.go                        # unchanged — entrypoint

internal/
├── config/
│   ├── config.go                  # unchanged — `Provider` already models every field of the entry
│   ├── expand.go                  # unchanged — `${VAR}` expansion of `API_KEY` (the `.json` path)
│   └── *_test.go                  # unchanged
├── domain/
│   └── llm/
│       └── gateway.go             # unchanged — the `Gateway` port + `Request`/`Response`/`ProviderError`;
│                                  #   the Gemini adapter implements the SAME port (no port change needed)
├── infrastructure/
│   └── llm/
│       ├── factory.go             # CHANGED — map `gemini`/`google` → the new adapter; every other family
│       │                          #   keeps the existing unsupported-provider failure (no silent widening)
│       ├── factory_test.go        # EXTENDED — gemini/google → the Vertex adapter; a family tellme cannot
│       │                          #   drive is still refused
│       ├── gemini/               # NEW package — the Vertex/Gemini transport adapter
│       │   ├── client.go          #   Vertex `:generateContent` request assembly + response normalization
│       │   ├── auth.go            #   service-account credential read + JWT-RS256 → access token (+ cache)
│       │   ├── client_test.go     #   request/response mapping units (fake HTTP)
│       │   └── auth_test.go       #   credential/JWT/token-cache units (fake token endpoint)
│       └── openai/                # unchanged — the OpenAI-compatible adapter stays byte-identical
└── cli/
    └── cli.go                     # unchanged — dispatch is family-agnostic; `newGateway` already routes
                                   #   through the factory seam

tests/e2e/
├── fakeprovider/fakeprovider.go   # EXTENDED — also serve the Vertex `:generateContent` shape AND the
│                                  #   service-account token exchange (targeted by the test credential's
│                                  #   `token_uri`)
└── steps/                          # NEW step files — the Gemini answer journey, the credential journey,
                                   #   and the family boundary

specs/truth/features/cli/
└── {chat,providers}/…             # ADD/MODIFY — the Vertex/Gemini provider contract + DSL rows

go.mod / go.sum                     # unchanged — stdlib-only; no new module
Makefile                            # unchanged (no new gate)
```

**Structure Decision**: Round 013 adds one **provider-transport capability to the existing CLI end**. A new adapter package (`internal/infrastructure/llm/gemini`) implements the **already-existing** `llm.Gateway` port over the Vertex AI `:generateContent` REST API, and the factory's family-dispatch seam (`internal/infrastructure/llm/factory.go`) is extended to map `gemini`/`google` to it. The service-account OAuth2 flow (`crypto/rsa`/`crypto/x509`/`crypto/sha256` + `net/http`, with an in-memory token cache) lives with the adapter. There is **no** new domain port, config field, persisted state, CLI flag, or third-party dependency — consistent with `research.md` Decisions 1–8 and Clarify Round 1. The `internal/cli` dispatch needs **no** change (it already builds the gateway through the factory seam). The **executable contract** (a `gemini` provider is driven; the service-account credential authenticates; a family tellme cannot drive is refused) is pinned in `specs/truth/features/cli/**` by `/axb-dsl-refine`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface. Round 013 does not introduce a new system end; it adds a **provider-transport family** to the **CLI end**.

1. `CLI end (operator terminal interface)`
   - Endpoint type: `CLI / terminal endpoint`
   - Primary interface: when the resolved provider's `TYPE` is `gemini`/`google` (a Vertex `aiplatform.googleapis.com` URL), `tellme "<prompt>"` drives the Vertex AI `:generateContent` API instead of refusing the family: it builds the request from the configured project/location/publisher path + `MODEL`, sends the persona as the system instruction, maps `MAX_TOKENS`/`THINKING_BUDGET`/`THINKING_LEVEL` into the generation/thinking config, offers the registered tool definitions, and normalizes the response (text, tool-call requests, usage) into the existing `llm.Response`. Authentication uses the operator's configured **service-account JSON** (an `API_KEY` ending in `.json`): a stdlib JWT-RS256 assertion is exchanged at the credential's `token_uri` for an access token (`Authorization: Bearer`, cached in memory for the run). A missing/unusable credential fails with the frozen `the provider request failed` (exit 6) and no silent fallback; a family tellme cannot drive (e.g. Anthropic) keeps its existing refusal. The answer stream, the frozen `tellme: {phrase}` vocabulary, the exit-code table (`0/2/3/4/5/6/7`), and the OpenAI-family requests are **unchanged**.
   - Requirement evidence: `FR-001`–`FR-013`, `NFR-001`–`NFR-006`; acceptance features `answering-with-a-gemini-model.feature`, `authenticating-with-a-service-account-key.feature`, `refusing-providers-tellme-cannot-drive.feature`.
   - Planner: **none** — terminal endpoints have no analysis planner. Its **contract owner** is **`/axb-dsl-refine`**, which updates the executable Gherkin features and DSL rows under `specs/truth/features/cli/**` at delivery (carried forward per `wave-covers-interfaces`).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no HTTP endpoints or OpenAPI specification of tellme's own). The outbound **Vertex** request shape is described in `techstack.md` (a CLI-transport concern), not in an authored `contracts/**` artifact.
> - `/axb-data-plan` = **`NOOP`** — no persisted-state change; the session-history model is untouched and the access-token cache is **in-memory only** (research Decision 5).
> - `/axb-ui-plan` = **skipped** (CLI-streamlined workflow; no HTML mockups).
> - The **Vertex/Gemini transport** and the **service-account auth** are **not** separate system interfaces: they are behaviour of the **CLI end** (its outbound transport). Their executable contract is carried forward to `/axb-dsl-refine`.
> - The **Vertex AI API** and the **Google OAuth token endpoint** are **external dependencies reached outbound by the CLI**; round 013 adds these outbound endpoints but authors **no** tellme-owned external contract beyond the executable CLI truth.
> - **Truth amendments carried to `/axb-dsl-refine`**: **ADD/MODIFY** a `cli` feature (likely `chat` for the answer/tool journey, plus a `providers` module for the family/credential rules) + `dsl.md` rows pinning — (a) a `gemini` provider is driven (a Vertex request is sent, the model's answer printed), (b) the configured service-account key authenticates (and a missing/unusable key fails with the frozen provider phrase, exit 6), and (c) a family tellme cannot drive is refused; and confirm **every round-013 acceptance rule is carried** by an interface feature (`acceptance-coverage`). The root `cli/dsl.md` class-phrase vocabulary stays **10**.

---

### Analysis Wave schedule

#### Wave 1 (single wave)

- Parallel-analyzed interfaces:
  - `CLI end (operator terminal interface)`
- Analysis focus:
  - **API** → **`NOOP`**: no OpenAPI/HTTP surface of tellme's own; the outbound Vertex request shape is a CLI-transport concern recorded in `techstack.md`.
  - **Data** → **`NOOP`**: no persisted-state change (the token cache is in-memory).
  - **CLI end** → handoff to contract owner **`/axb-dsl-refine`**: pin the Vertex/Gemini provider contract as mechanically assertable interface Rules — (a) a `gemini` provider is driven (Vertex request sent; answer printed; tool-call round-trip; persona/limits mapped), (b) the service-account key authenticates and a missing/unusable key fails with `the provider request failed` (exit 6, no fallback), (c) a family tellme cannot drive is refused — while `stdout` stays byte-exact and the class-phrase vocabulary is unchanged.
- Scheduling rationale: there is **one** interface and the behaviour is settled by `research.md` Decisions 1–8 and Clarify Round 1, so there is nothing to sequence; a single wave suffices. Per `Wave依賴排序與平行分組判準.md` Rules 1–2, a sole interface forms one wave.

---

### Delegation order

1. **`/axb-api-plan`** — Wave 1 (`CLI end`) → **`NOOP`** (no OpenAPI contract).
2. **`/axb-data-plan`** — Wave 1 → **`NOOP`** (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract-owner handoff at delivery** for the `CLI end` → pin the Vertex/Gemini provider contract (drive, authenticate, refuse) in `specs/truth/features/cli/**`, and confirm `acceptance-coverage` for all three round-013 acceptance features.

*Handoff payload (for the next phase)*: plan package `specs/plans/013-vertex-gemini-provider`; truth root `specs/truth`; truth-delta `specs/plans/013-vertex-gemini-provider/truth-delta.md`; interfaces `CLI end`; analysis focus as above; acceptance features `features/acceptance/answering-with-a-gemini-model.feature` + `features/acceptance/authenticating-with-a-service-account-key.feature` + `features/acceptance/refusing-providers-tellme-cannot-drive.feature`.

Not delegated:
- `/axb-ui-plan` (skipped — no UI).
- `/axb-api-plan` and `/axb-data-plan` are invoked only to record their `NOOP`.

---

### Gating blockers

*(none — Clarify Round 1 settled the scope (Vertex-only), the credential detection (`.json`), and the dependency posture (stdlib-only); `research.md` Decisions 1–8 settled the transport, the request/response mapping, the OAuth2 flow, the credential-failure taxonomy, the token cache, the hermetic verification, and the no-dependency posture. **Open (non-blocking):** the exact Vertex REST field names / thinking-config mapping (verify live), the `google` TYPE alias, the credential-failure class placement (exit 6 chosen), the token-expiry skew / 401 re-mint, and the auth seam's package shape — all held as research/DSL/task-level defaults, none gating this round.)*
