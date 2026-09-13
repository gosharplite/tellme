# Truth Delta: 011-persona-and-payload-estimate

**Plan Package**: `specs/plans/011-persona-and-payload-estimate`
**Truth Root**: `specs/truth`

> Skeleton initialized by `/axb-specify`. Each truth owner replaces its placeholder row(s) with its own ADD / MODIFY / DELETE / NOOP entries during the round. Every owner must record at least one entry; a `NOOP` entry proves the area was checked.

## /axb-technical-research

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` | **Configuration**: added a *Persona instruction* row (`PERSON` sent as a leading `system` message; empty ⇒ no message). **Reasoning & Provider Transport**: extended *Request assembly* (leading persona `system` message; rides every request incl. tool-driven completions) and *Conversation context* (the persona precedes the conversation). **CLI Application**: widened the *Token estimator* inputs (persona + tool declarations + messages) and the *Payload status line* pre-flight estimate to the wire payload. **Testing & Verification**: extended the *E2E runner* (assert the leading `system` message + estimate growth) and *Pure-helper unit tests* (wire-payload estimate + persona assembly). **Not Introduced Yet**: added "a BPE tokenizer (tiktoken or equivalent)". Nothing else — one CLI end, no new dependency. | Round-011 research Decisions 1–8: send the configured persona on the wire (reference parity) and widen the estimator's inputs so the pre-flight estimate is comparable to the provider's measured count. |

## /axb-api-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/contracts/**` | Checked; left empty. tellme has a single CLI end and no HTTP/OpenAPI surface of its own; the outbound request content (adding the persona `system` message) is a CLI-transport concern recorded in `techstack.md`, not an authored contract. | `contract-authoritative` holds vacuously. |

## /axb-data-plan

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| NOOP | `specs/truth/data/**` | Checked; left empty. No persisted-state change — the estimate and token counts are display-only and recomputed; the session-history model (`history_entry`) is untouched. | `data-model-covers-all-state` holds vacuously. |

## /axb-dsl-refine

| Action | Truth Spec | Change Summary | Reason |
| --- | --- | --- | --- |
| ADD | `specs/truth/features/cli/chat/sending-the-configured-persona.feature` | New CLI `chat` feature: the configured `PERSON` is sent as the leading `system` message of every request the turn makes; an empty persona sends none; the persona is request-only (never printed). | Round 011 — the persona is parsed but never sent today; make it executable truth (`FR-001`–`FR-005`). |
| ADD | `specs/truth/features/cli/chat/estimating-the-wire-payload.feature` | New CLI `chat` feature: the pre-flight estimate measures the wire payload — it exceeds the conversation messages alone, grows with the wired payload, and is stable for identical inputs. | Round 011 — the estimate ignored most of the payload (`~5` vs `387`); make the wire-faithful estimate executable truth (`FR-006`–`FR-009`). |
| MODIFY | `specs/truth/features/cli/chat/dsl.md` | Added the Given rows `the runtime home holds a configuration whose persona is …` / `… with no persona` / `a previous run used the persona … and reported an estimated payload status`; the When row `the operator uses the persona … and starts tellme with the prompt …`; the Then rows `the request carried the persona …` / `the request carried no persona` / `the estimated payload exceeds the conversation messages alone` / `… is larger than the previous run's` / `… matches the previous run's`; and reworded `the request carried no earlier exchange` to allow the leading persona `system` message. | Round 011 — carry the persona request contract and the wire-faithful estimate into the executable DSL; the reword accounts for the now-sent persona (`FR-011`). |
| NOOP | `specs/truth/features/cli/dsl.md` | The interface root is unchanged — the persona and estimate rows are `chat`-module-specific; the frozen class-phrase vocabulary stays **10**. | No cross-module row was needed. |
