# Truth Delta: 014-session-replay-fidelity

**Plan Package**: `specs/plans/014-session-replay-fidelity`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **CLI Application → Session history store**: the tool-step record gains an optional provider-agnostic **`signature`** (`{tool, arguments, result, signature?}`), omitted when empty so existing lines stay byte-identical, replayed on resume. **CLI Application → Agent tool loop**: records each executed step's signature alongside its result. **Reasoning & Provider Transport → Vertex/Gemini adapter**: the captured `thoughtSignature` is now also **persisted with the step** so a resumed session replays it across processes. **Testing & Verification → E2E runner / Local fake provider / Pure-helper unit tests**: the two-process resume witness (E2E) + the `Step` signature round-trip and `BuildMessages` replay (unit). | Round-014 research Decisions 1–6: persist the tool-step signature and replay it on resume (Clarify Q1 = dedicated `signature` field; Q2 = unchanged best-effort for legacy histories); stdlib-only, no new module. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and authors no HTTP/OpenAPI surface of its own; round 014 changes no request shape — it only replays a provider token the round-013 adapter already re-emits. | `contract-authoritative` holds vacuously — round 014 changes the persisted record + resume behaviour, not a tellme-owned API. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/data/data-model.dbml` | `history_step` gains a nullable, provider-agnostic **`signature`** (`varchar`) — the opaque token the model emitted for the call (for example the Gemini 3 `thoughtSignature`), persisted only when the provider supplies one and omitted when empty (so existing lines stay byte-identical). The `history_step` Note and the project Note are reconciled to the widened shape `{prompt, answer, steps:[{tool, arguments, result, signature?}]}` and to the replay semantics (echoed verbatim on resume). | Round-014 research Decisions 1/3 + spec `FR-001`/`FR-002`/`FR-005`/`FR-007`: persist the per-step signature so a resumed Gemini session replays its tool steps faithfully (`data-model-covers-all-state`). |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| _(placeholder)_ | `specs/truth/features/cli/**` | | |
