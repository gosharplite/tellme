# System Analysis Plan — round 030 (`030-provider-truncation-guard`)

## Project Structure

### Document structure (this feature)

```text
specs/plans/030-provider-truncation-guard/
├── plan.md                        # this file — /axb-system-analysis
├── spec.md
├── research.md
├── truth-delta.md
├── checklists/
│   └── requirements.md
├── features/acceptance/*.feature  # /axb-spec-by-example ✓ done (refusing a reply cut off at the output limit)
└── tasks.md                       # produced by /axb-tasks

specs/truth/
├── techstack.md                   # /axb-technical-research — MODIFY ✓ done (the Provider output-cap truncation guard row)
└── features/cli/chat/**           # /axb-dsl-refine — ADD (contract owner)
```

*(No `contracts/**` change (`/axb-api-plan` `NOOP`), no `data/**` change (`/axb-data-plan` `NOOP`), and
no `ui/**` artifact — the guard is transport-facing, with no new operator UX surface.)*

### Repository structure (root)

```text
internal/infrastructure/llm/openai/client.go   # CHANGED — read choices[0].finish_reason; fail a "length" finish as a *llm.ProviderError instead of returning it
internal/infrastructure/llm/gemini/client.go   # CHANGED — read candidates[0].finishReason; fail a "MAX_TOKENS" finish (function-call-aware) instead of returning it
internal/infrastructure/llm/openai/*_test.go   # ADDED   — unit tests pinning the guard (truncated == length fails; stop / tool_calls healthy)
internal/infrastructure/llm/gemini/*_test.go   # ADDED   — unit tests pinning the guard (truncated == MAX_TOKENS fails; STOP / absent healthy; function-call-aware message)
tests/e2e/ (fakeprovider + steps + suite)      # CHANGED — the fake scripts a truncated response (both wire shapes); steps assert the refusal + exit 6 + the file/answer effects
specs/truth/techstack.md                        # MODIFY  — Provider output-cap truncation guard row + adapter/normalization/port rows ✓ done
go.mod / go.sum                                 # unchanged — stdlib-only (no new dependency)
```

**Structure Decision**: Round 030 changes **transport behaviour** *inside* the existing CLI end — it adds
**no** system boundary. The guard lives in the two adapters behind the **unchanged** `internal/domain/llm`
`Gateway` port, and the failure rides the **existing** `*llm.ProviderError`, which the CLI already maps to
the frozen class phrase `the provider request failed` + exit code 6. There is **no** new endpoint, **no**
persisted-state change, and **no** new dependency — consistent with `research.md` Decisions 1–7. The only
truth change beyond `techstack.md` (`/axb-technical-research`) is the **CLI interface truth** under
`specs/truth/features/cli/chat/**`, owned by the CLI end's contract owner `/axb-dsl-refine`.

---

## Analysis Plan

### System interface inventory

This requirement inventories **1** system interface — the **CLI end** (the `chat` capability: the
reasoning turn that dials the configured provider). The round **changes the CLI end's observable
behaviour**: a run whose provider **cuts the reply off at the output cap** now **refuses** (the frozen
class phrase + exit 6) instead of returning the truncated reply — so a truncated tool call is never
dispatched and a truncated answer is never printed. The interface must therefore be carried to its
contract owner for an interface-truth change.

There is **no** analysis planner for the CLI end (per the CLI-streamlined model): it is not delegated in a
`Wave`; it is carried forward to its **contract owner `/axb-dsl-refine`** at delivery
(`wave-covers-interfaces`'s carried-forward branch).

> **Scope notes**:
> - `/axb-api-plan` = **`NOOP`** (standalone CLI; no OpenAPI/HTTP surface; the guard authors no request/response shape).
> - `/axb-data-plan` = **`NOOP`** (the guard persists no state; the existing persisted shapes — `history.jsonl`, `tokens.log`, the prompt logs — are unchanged).
> - `/axb-dsl-refine` = **contract owner** (ADD a `chat` interface feature for the truncation guard — a reply cut off at the output cap is refused, both when it carries a tool call and when it carries only text — + the `chat/dsl.md` rows; possibly MODIFY `reporting-a-failed-provider-request.feature`'s header/scope note, since a truncation is a further cause of the same frozen provider class).
> - `/axb-ui-plan` = **skipped** (no UX surface change; the failure reuses the existing `stderr` class-phrase surface, and the operator chrome is unchanged).
> - `/axb-spec-by-example` = **done** (acceptance journey: refusing a reply cut off at the output limit).

### Analysis Wave schedule

**No waves.** The CLI end has no analysis planner, so there is no interface to order or delegate within a
wave; the round's contract-owner handoff (`/axb-dsl-refine`) happens at delivery, not inside a wave.

### Delegation order

1. **`/axb-api-plan`** — `NOOP` (no OpenAPI contract).
2. **`/axb-data-plan`** — `NOOP` (no persisted-state change).
3. **`/axb-dsl-refine`** — **contract owner**: ADD a `specs/truth/features/cli/chat/**` interface feature (Rules: a reply cut off at the output cap while asking tellme to change a file is refused and the file is untouched; a reply cut off at the output cap before the answer is finished is refused and no answer is printed) + `chat/dsl.md` rows (a provider whose reply is cut off at the output limit; the refusal reuses the frozen `the provider request failed` phrase + the provider error code; the file is not written / the answer is not printed); optionally MODIFY `reporting-a-failed-provider-request.feature`'s header/scope note. *(Exact file/rule/row names are `/axb-dsl-refine`'s call.)*

Not delegated:
- `/axb-ui-plan` — **skipped** (no UX surface change).

*Handoff payload (for the next phase)*: plan package `specs/plans/030-provider-truncation-guard`;
truth root `specs/truth`; truth-delta `specs/plans/030-provider-truncation-guard/truth-delta.md`;
interfaces: **1 (CLI end / `chat`)** carried to `/axb-dsl-refine`; the round's delivery is the finish-reason
guard in the two adapters (`internal/infrastructure/llm/{openai,gemini}/client.go`) and the `chat`
interface truth.

---

### Gating blockers

*(none — the two high-impact gaps were resolved in Clarify Round 1 (Q1 → universal trigger; Q2 → reuse the
provider class), and `research.md` Decisions 1–7 settle the guard's placement, trigger, failure class, the
no-retry posture, the unchanged request side, the guard ordering, and the verification strategy. No open
decision gates the round.)*
