# ADR 0048 — Recover from an unknown tool name (per-turn-bounded fold-back)

**Status**: Accepted (round 076)

**Date**: 2026-09-22

**Related**: issue [#154](https://github.com/gosharplite/tellme/issues/154) · round 008 (the tool-loop bound + the frozen tool-error phrase / exit 7, FR-010) · round 056 / [ADR 0025](0025-mcp-tool-call-reason.md) (the recoverable reason-less refusal this mirrors) · round 065 / [#132](https://github.com/gosharplite/tellme/issues/132) (the function-call/response pairing invariant) · round 057 / [ADR 0027](0027-tool-chrome-colour-and-payload-delta.md) (the loop diagnostic chrome) · issue [#155](https://github.com/gosharplite/tellme/issues/155) (MCP name presentation — a separate, out-of-scope issue) · `specs/truth/techstack.md` §Agent tool loop · `specs/truth/features/cli/chat/using-an-unknown-tool-name.feature`.

## Context

The agent tool loop looks a requested tool up in its registry, and on a miss **aborts the whole turn** (`internal/agent/agentloop.go:133-136`):

```go
tool, ok := a.Registry.Lookup(tc.Name)
if !ok {
    return agentport.Result{Steps: steps, Calls: calls}, &agentport.ErrIncomplete{Reason: fmt.Sprintf("tool %q is not available", tc.Name)}
}
```

The CLI maps `ErrIncomplete` to the frozen phrase `tellme: the tool request failed: …` and exit code **7** (round 008 FR-010). So a single wrong tool name kills the turn — while, in the *same* loop, a **real** tool's call-time error is folded back non-terminally (`error: …`, `agentloop.go:171-176`) and the reason-less refusal (round 056 / ADR 0025 D3) is likewise recoverable (log a `tool`-role result, append it, `continue`).

Observed live (butler / `deepseek-flash` on the `misc` repo, executing that repo's `SESSION-BOOTSTRAP.md`): the model asked for the GitHub MCP tool by its **bare** upstream name `get_me` — tellme only registers MCP tools under the **namespaced** wire name `mcp_github_get_me` — and the run aborted: `tellme: the tool request failed: tool "get_me" is not available` (**exit 7**).

An off-list name is a **recoverable model slip** (a typo, an alias, an upstream tool named by its bare name). It should not be fatal.

## Decision

1. **An unknown tool name is a recoverable fold-back, not terminal (D1).** On a registry miss the loop appends a `tool`-role result (with the call's `ToolCallID`) and `continue`s — the exact shape of the reason-less refusal. The unknown call **runs no tool**, records **no `history.Step`/`Signature`**, and writes **no** tool-usage record. *Rejected:* the terminal path (the defect); a provider-side retry (wrong layer).

2. **The fold-back names the unknown tool and a bounded list of the available wire names (D2).** `error: no tool named "<name>"; available tools: <n1>, <n2>, …` (built from `Registry.Tools()` in offer order; `no tools are available` when empty). A specific message breaks a fixation early; the list is bounded only by the **fixed registry size** (the loop's `clampBytes` does **not** run on this path — it applies only to an executed tool's result). *Rejected:* a bare `unknown tool` (sticky); a near-miss suggestion engine (overlaps #155, not needed).

3. **The recoverable path is bounded PER TURN (D3).** `maxUnknownToolFolds = 3`: after 3 unknown-name fold-backs in one turn the loop stops folding back and returns `ErrIncomplete` (frozen phrase + **exit 7**). **Why mandatory:** tellme has **no** repetition/runaway detection (the reference's SHA-256 / tool-repetition guards were deliberately not re-created), and the only other bound is `MaxLoops` (`MAX_TOOL_LOOP`, default **1000**; each round a **paid** provider call) — an unbounded fold-back would turn a fast fatal error into an up-to-~1000-call spend. The value is a named constant pinned by a unit test. *Rejected:* relying on `MaxLoops` alone.

4. **Counter home: a loop-local variable in `Run` (D4).** `Run` is one turn (the loop is constructed per turn), so a local counter resets every turn with no state to leak and no domain-model change. *Rejected:* a struct/`TurnState` field.

5. **Ordering: the unknown-name check stays BEFORE the reason gate (D5).** Single-owned classification: an unknown call that also lacks a reason is reported as **unknown** (one round trip corrects the name); since the unknown call never executes, the *no reason, no go* rule (which gates **execution**) is not violated.

6. **The round-008 FR-010 contract is NARROWED, not removed (D6).** The frozen phrase + **exit 7** still cover the genuine incomplete cases: the tool-loop **bound reached**, the per-turn **cap exhausted**, and **`no tools are registered`**. Only "one unknown name → immediate abort" changes.

7. **tellme's own robustness, not reference parity (D7).** Whether the reference aborts on an unknown name is not asserted; this is a **tellme-side robustness improvement**, recorded as such.

8. **Not modelled (ADR 0041 escape hatch) (D8).** The change is an error-handling classification in the loop; it introduces no modelled entity/attribute/relationship/invariant. `docs/domain-model/**` is unchanged (reason in `plan.md` §5).

## Consequences

- An off-list tool name no longer aborts the turn; the model is told the name is unknown (with the available names) and may correct itself within the same turn.
- The frozen tool-error phrase + exit 7 remain for the genuine incomplete cases; the E2E example that asserted "an unavailable tool is reported" still holds — it now resolves through the per-turn cap after the recoverable fold-backs.
- A model that fixates on an unknown name is stopped after `maxUnknownToolFolds` fold-backs, bound below `MAX_TOOL_LOOP` (no unbounded paid-call spend).
- **Every** call in a round still receives a paired `tool` result (I-1), so the Gemini/Vertex function-call/response pairing (round-065 / #132) is preserved; the terminal cap path `return`s before any request, so no unpaired wire state is sent.
- Family-local: the provider adapters, the wire envelopes, the registry, the reason gate, and the loop clamp are untouched (zero wire diff).

## Forward

- **RF-076-1** — the per-turn cap value (3) is a constant; if operator value later shows a different balance, it is a one-line change (with its unit pin).
- **RF-076-2** — the available-tools listing is bounded only by the fixed registry size (today ~9 names); a dedicated bound is possible if a huge registry ever makes it noisy.
- **RF-076-3** — general runaway protection for *valid* tools that keep erroring remains a separate, unaddressed decision (deliberately out of scope here).
- **RF-076-4** — a near-miss / "did you mean" suggestion for a bare MCP name overlaps [#155](https://github.com/gosharplite/tellme/issues/155) (MCP name presentation), which is a separate issue.
- **RF-076-5** — the reason × unknown ordering is pinned but the reasonless-refusal path is itself uncapped (a separate pre-existing surface; not changed here).
