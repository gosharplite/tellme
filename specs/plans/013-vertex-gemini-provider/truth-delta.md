# Truth Delta: 013-vertex-gemini-provider

**Plan Package**: `specs/plans/013-vertex-gemini-provider`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Reasoning & Provider Transport**: added *Vertex/Gemini adapter* (new `internal/infrastructure/llm/gemini`; Vertex `:generateContent` request/response mapping, stdlib `net/http`), *Provider family mapping* (factory `TYPE` → transport; `gemini`/`google` supported; other families keep the unsupported failure — no silent widening), and *Service-account authentication* (stdlib JWT-RS256 exchanged at the credential's `token_uri`, `Authorization: Bearer`, in-memory token cache) rows. **Configuration**: extended the *Provider entry schema* row with `API_KEY`'s per-family meaning (inline bearer vs service-account `.json` path) and added the *Provider credential resolution* row (read at turn time; missing/unreadable/invalid → `the provider request failed`, exit 6; no silent bearer fallback). **Testing & Verification**: extended *Local fake provider* (also serves the Vertex shape + the service-account token exchange) and *Pure-helper unit tests* (Vertex request/response mapping, service-account JWT + in-memory token cache, provider family mapping). **Not Introduced Yet**: replaced "Gemini/Vertex and Anthropic provider adapters" with Anthropic + the Google Gemini API family + Application Default Credentials (Gemini/Vertex no longer deferred). *(PR #33 architectural review D1–D4 folded in: Vertex tool wire format `role: "model"`/`"user"` + `functionResponse`; endpoint built from the configured URL — no hostname check; token cache `sync.RWMutex` + ~60s expiry margin; fakeprovider path-suffix dispatch.)* | Round-013 research Decisions 1–8: add the Vertex AI Gemini transport + a stdlib-only service-account OAuth2 flow (Clarify Q1 = Vertex-only, Q2 = `.json` credential detection, Q3 = stdlib-only); one CLI end, no new external system end, **no new module**. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and authors no HTTP/OpenAPI surface of its own; the outbound Vertex `:generateContent` request shape is described in `techstack.md` (a CLI-transport concern), not in an authored contract. | `contract-authoritative` holds vacuously — round 013 adds a provider transport, not a tellme-owned API. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. No persisted-state change — the session-history model is untouched and the Vertex access-token cache is **in-memory only** (research Decision 5). | `data-model-covers-all-state` holds vacuously. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/driving-a-vertex-gemini-model.feature` | New `chat` feature — atomic rules: a Vertex Gemini provider is driven (one `:generateContent` request → answer printed); the configured persona is carried; the configured output budget is carried; the model can request a tool and tellme runs it before answering. | Round 013 — carry the acceptance `answering-with-a-gemini-model.feature` rules into executable interface truth for the Vertex/Gemini transport. |
| ADD | `specs/truth/features/cli/chat/authenticating-to-a-vertex-gemini-model.feature` | New `chat` feature — the service-account key authenticates the request; a missing key file fails with the frozen `the provider request failed` (exit 6) and no fallback. | Round 013 — carry the acceptance `authenticating-with-a-service-account-key.feature` rules. |
| ADD | `specs/truth/features/cli/chat/refusing-a-provider-family-tellme-cannot-drive.feature` | New `chat` feature — a family tellme cannot drive (Anthropic example) keeps its existing refusal. | Round 013 — carry the acceptance `refusing-providers-tellme-cannot-drive.feature` rule; pin the family boundary (`FR-013`). |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Added the round-013 header note and **5 Given rows** (`a configured Gemini provider … answers with …`; `… asks tellme to read … and then answers with …`; `… whose key file is missing`; `the configured Gemini provider entry allows at most {tokens} output tokens`; `a configured provider … of a family tellme cannot drive`) and **2 Then rows** (`the request … carried the service-account access token`; `the request … allows at most {tokens} output tokens`). Reused existing rows (send-one-request / print-answer / exit-successfully / refuse-to-proceed / frozen-phrase / provider-error-code / persona / read_files / working-directory). | Round 013 — the new steps for the Vertex/Gemini provider, credential, and family-boundary contract. |
| NOOP | `specs/truth/features/cli/dsl.md` | The interface root is unchanged — the new rows are `chat`-module-specific; the frozen class-phrase vocabulary is unchanged. | No cross-module row was needed (the provider transport + credential are `chat`-module concerns, mirroring rounds 004/011). |
