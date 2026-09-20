# ADR 0038 — Surface a Gemini/Vertex round's unpaired tool calls as a diagnostic

- **Status:** Accepted
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** [ADR 0037](0037-gemini-toolcall-id-provenance.md) (the unpaired-call account — this ADR **delivers its §Forward RF-067-1**), [ADR 0036](0036-toolcall-id-pairing.md) / [ADR 0035](0035-gemini-parallel-tool-call-batching.md) (the round-batching/id-link shape whose boundary drop this reports), [ADR 0020](0020-cli-ui-decoupling.md) (the `render.Lines` port), [ADR 0022](0022-offline-session-config-and-turns-log.md) (the `turns.log` narrowing), [ADR 0023](0023-list-default-and-chrome-colour.md) (the colour gate), round 067 (`specs/plans/067-toolcall-id-followups`), round 068 (`specs/plans/068-unpaired-call-diagnostic` — this ADR's round), issue [#36](https://github.com/gosharplite/tellme/issues/36) item 3

## Context

Round 067 (ADR 0037) made a Gemini/Vertex round's unpaired calls **accountable in code** via `UnpairedCallIDs` — but recorded that the accessor had **no live consumer**: at a round boundary a call that received no result (`M < N`) still contributes no part to the batched `user` turn and is dropped **silently** at runtime. ADR 0037 §Forward **RF-067-1** deferred surfacing it as a user-visible diagnostic, noting it was an **operator-gated scope addition**. The operator confirmed the scope addition (*"Open 068-*, the goal is to resolve RF-067-1"*).

**Honest premise:** the shipped agent loop appends exactly **one `tool` result per requested call** (`internal/agent/agentloop.go`), so in production **`M == N` always** and the diagnostic **cannot fire** through any shipped path. It is a **defensive** surface — for a partial/corrupted `prior`, or a future **out-of-order / concurrent** dispatch ([#36](https://github.com/gosharplite/tellme/issues/36) item 3), the exact motivation ADR 0035/0036 recorded.

## Decision

**D1 — The account gains a family-neutral single owner.** The round-067 algorithm moves to `llm.UnpairedToolCalls(msgs []llm.Message) []string` (a round boundary = a message carrying tool calls; each result binds by `ToolCallID`, else FIFO; a plain-text turn closes the round; a standalone media turn is not a boundary). The Gemini adapter's `UnpairedCallIDs` **delegates** to it and its per-round `dropped`/`unpaired()` machinery is deleted. *(The round-067 fold F-067-3 shared-one-pass property is **not** retained: the account is now a second walk, so the adapter's grouping and the mirror are tied by a **property pin** — `TestUnpairedCallIDs_AgreesWithEmittedBody` asserts `totalCalls − len(unpaired) == #functionResponse parts` on every fixture — not by construction; corrected at the review's TD-068-1.)*

**D2 — The diagnostic is surfaced by a CLI gateway decorator, informational.** The decorator recomputes over the **whole `prior`** on **every** `Complete`, so a short round anywhere in a resumed conversation is re-reported on each provider call of the turn and on each later turn of the session — a stated, accepted choice (a seen-set bound is a forward item). `internal/cli/unpaired_gateway.go` wraps the resolved `llm.Gateway`: before each `Complete` it computes `llm.UnpairedToolCalls(req.Messages)`; when non-empty it emits the diagnostic, then delegates **unchanged**. The request — hence the wire — is untouched; the turn proceeds (nothing aborts). This avoids widening the `deps.NewGateway` / factory / adapter-`Config` signatures.

**D3 — The surface is a `[Tool …]`-class `stderr` line, never `turns.log`; plain.** `ui.FormatUnpairedCalls(t, ids)` renders `[HH:MM:SS] [Tool Warning] <n> tool call(s) left unanswered: <id, id>`, exposed through the existing **`render.Lines`** port (`UnpairedCalls`) so `internal/cli` names no `internal/ui` type (ADR 0020). It is written to `stderr` only (never `stdout`) and is **not** routed to `turns.log` (`-t` unchanged; the `[Tool Output]`-block precedent, ADR 0022 D5). Its **id value is folded, control-free and rune-capped** by the same single-owned `[Tool …]` policy as its siblings (`sanitizeControl` → `oneLine` → `capRunes`, cap `unpairedIDsCap = 200`) — the ids are **provider-sourced** (round 067 prefers the provider's `functionCall.id`), so the control class matters independently of colour *(corrected at the review's B-068-1: "plain" is not the safety argument — colour and the terminal-control class are distinct, ADR 0008)*. The line is **colour-free** (no accent); colouring it is a deferred decision.

**D4 — Scope: no live producer today; harmless for the OpenAI-compatible family.** The decorator wraps every gateway, but the OpenAI-compatible family has no round-boundary drop, so its account is **always empty** — the OpenAI wire is untouched (I-1). The Gemini shipped path is always `M == N`, so the diagnostic is inert there too (I-7). No user-visible behaviour changes on any shipped path.

**D5 — Carriers are unit + CLI tiers (no E2E).** `M < N` has **no hermetic producer**, so no godog Example can drive the diagnostic; its carriers are the domain / ui / cli pins (the **round-059 documented-narrowing** class). `/axb-spec-by-example` and `/axb-dsl-refine` record the narrowing — never an Example-less Rule (RF-063-10 retired).

**D6 — Invariants held.** OpenAI-compatible wire byte-preserved (I-1); the round-066/067 batch + id-link shape preserved (I-2); the media-free text path byte-preserved (I-3); `stdout` untouched (I-4); replay fidelity holds (I-5); stdlib-only, POSIX-only, hermetic, no new dependency/config key (I-6); `M == N` is silent (I-7).

**D7 — Governance: this ADR extends ADR 0037 and delivers its §Forward RF-067-1.** It annotates that entry and the README index row. ADR 0037's body is otherwise not edited (the annotation precedent).

## Consequences

### Positive

- The unpaired-call account now has a **live consumer**: a short round is **reported**, not silently dropped — resolving RF-067-1.
- The account has **one owner** (`llm.UnpairedToolCalls`); a **property pin** (`TestUnpairedCallIDs_AgreesWithEmittedBody`) ties the mirror to the emitted body, so the adapter's wire behaviour and the CLI's diagnostic cannot silently drift.
- The diagnostic is **surgical**: no wire change, no exit codes, no `stdout`, no new dependency, no config key; the shipped happy path is **inert**.

### Negative / Accepted Trade-offs

- The diagnostic has **no live producer today** and **no E2E carrier** — it is defensive-only, verified at the unit/CLI tiers (the round-059 narrowing class). Disclosed, not hidden.
- It is computed eagerly (a cheap `O(messages)` walk) per `Complete`.
- `render.Lines` gains one method (a deliberate, backward-compatible port addition).

### Neutral

- **No user-visible behaviour change on any shipped path** (the loop is complete).
- No domain entity, persisted record, config, or tool change; the reason gate (ADR 0025) and the round-024 resource contract are untouched.

## Alternatives Considered

1. **Do nothing (leave the accessor with no consumer).** Rejected — that was RF-067-1 itself.
2. **Inject a diagnostic sink through the adapter `Config` / factory / `deps.NewGateway`.** Rejected (D2) — a much larger blast radius (signature widening + test churn) for no behaviour gain over the CLI decorator.
3. **Detect in the loop.** Rejected — the loop is family-blind; the round-boundary account is a provider-wire concept.
4. **Route the diagnostic to `turns.log` too (Q1 → B).** Not taken — `turns.log` follows the chrome's model-authored content; the `[Tool Output]` block is excluded (ADR 0022 D5).
5. **A distinct `[Warning]` prefix (Q1 → C).** Not taken — the `[Tool …]` family keeps the diagnostic legible with the rest of the tool chrome.
6. **A loud failure (Q2 → B).** Not taken — it would reverse the shipped "proceed" behaviour; recorded as a forward item.
7. **Fold into ADR 0037.** Rejected (D7) — the annotation precedent; the new decision is its own.

## Verification

- **Domain** — `TestUnpairedToolCalls` (all-paired/short-round/zero-results/multi-round/out-of-order/id-less-FIFO/duplicate-id/text-closes-round).
- **ui** — `TestFormatUnpairedCalls` (the exact `[Tool Warning]` line; plain, no ANSI).
- **cli** — `TestUnpairedGateway_EmitsOnShortRound` (emits the unpaired ids; the inner gateway still receives the request unchanged) · `…_SilentOnHappyPath` (I-7) · `TestUnpairedEmitter_WritesToStderr` (stderr only; no-op on empty) · `…_NilEmitterPassthrough`.
- **Falsifiability** — suppressing the emit reds the cli pin; suppressing the account reds the domain pin.
- **Regression** — the round-065/066/067 pins + the E2E journeys stay green; the OpenAI wire and the Gemini shapes are unchanged.
- **Live/closeout (non-gating)** — none required (no wire delta; the diagnostic has no live producer).

## References

- `internal/domain/llm/unpaired.go` · `internal/cli/unpaired_gateway.go` · `internal/ui/toolcall.go` (`FormatUnpairedCalls`) · `internal/domain/render/ports.go` · `internal/infrastructure/llm/gemini/client.go`.
- [ADR 0037](0037-gemini-toolcall-id-provenance.md) (§Forward RF-067-1) · [ADR 0020](0020-cli-ui-decoupling.md) · [ADR 0022](0022-offline-session-config-and-turns-log.md) · [ADR 0023](0023-list-default-and-chrome-colour.md).
- `specs/plans/068-unpaired-call-diagnostic/` (`spec.md` FR-001…FR-007; `research.md` D1…D7) · issue [#36](https://github.com/gosharplite/tellme/issues/36) item 3.

## §Forward (deferred, non-blocking)

- **RF-068-1** — the diagnostic has no live producer / E2E carrier today (`agentloop.go` is complete); a real carrier arrives with out-of-order/concurrent dispatch ([#36](https://github.com/gosharplite/tellme/issues/36) item 3).
- **RF-068-2** — the line is plain; colouring it (terminal-gated) is a separate decision.
- **RF-068-3** — the loud-failure variant (Q2 → B) is not taken.
- **RF-068-4** — the account is computed eagerly per `Complete` (unneeded memoisation at these sizes); a **seen-set** to avoid re-reporting the same historical short round across calls/turns is a deferred bound (D2).
- **RF-068-5** — the `[Tool Warning]` line is named in the `Chrome` domain entity + the *Turn chrome (operator)* truth row (N-068-1 record).
- **RF-068-1 (exit condition)** — when out-of-order/concurrent dispatch lands ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) and `M < N` gains a **live** producer, an **E2E carrier becomes mandatory** (the round-059 narrowing's exit), not optional.
