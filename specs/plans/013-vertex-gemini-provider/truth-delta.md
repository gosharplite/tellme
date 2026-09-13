# Truth Delta: 013-vertex-gemini-provider

**Plan Package**: `specs/plans/013-vertex-gemini-provider`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Reasoning & Provider Transport**: added *Vertex/Gemini adapter* (new `internal/infrastructure/llm/gemini`; Vertex `:generateContent` request/response mapping, stdlib `net/http`), *Provider family mapping* (factory `TYPE` → transport; `gemini`/`google` supported; other families keep the unsupported failure — no silent widening), and *Service-account authentication* (stdlib JWT-RS256 exchanged at the credential's `token_uri`, `Authorization: Bearer`, in-memory token cache) rows. **Configuration**: extended the *Provider entry schema* row with `API_KEY`'s per-family meaning (inline bearer vs service-account `.json` path) and added the *Provider credential resolution* row (read at turn time; missing/unreadable/invalid → `the provider request failed`, exit 6; no silent bearer fallback). **Testing & Verification**: extended *Local fake provider* (also serves the Vertex shape + the service-account token exchange) and *Pure-helper unit tests* (Vertex request/response mapping, service-account JWT + in-memory token cache, provider family mapping). **Not Introduced Yet**: replaced "Gemini/Vertex and Anthropic provider adapters" with Anthropic + the Google Gemini API family + Application Default Credentials (Gemini/Vertex no longer deferred). | Round-013 research Decisions 1–8: add the Vertex AI Gemini transport + a stdlib-only service-account OAuth2 flow (Clarify Q1 = Vertex-only, Q2 = `.json` credential detection, Q3 = stdlib-only); one CLI end, no new external system end, **no new module**. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/contracts/**` | _To be recorded by the truth owner._ | _Round in progress — placeholder._ |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/data/**` | _To be recorded by the truth owner._ | _Round in progress — placeholder._ |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _pending_ | `specs/truth/features/cli/**` | _To be recorded by the truth owner._ | _Round in progress — placeholder._ |
