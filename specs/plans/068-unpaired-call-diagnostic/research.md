# Technical Research — Round 068: surface a Gemini/Vertex round's unpaired tool calls as a diagnostic

**Plan package**: `specs/plans/068-unpaired-call-diagnostic`
**Truth owner**: `/axb-technical-research` → updates `specs/truth/techstack.md`
**Anchor**: **ADR 0037 §Forward RF-067-1** (operator-confirmed scope addition)
**Date**: 2026-09-20
**Inputs**: `spec.md` (US1 · FR-001…FR-007 · I-1…I-7 · S-1…S-6 · A1…A6; clarify **Q1 → A · Q2 → A**); the grounded current shape; `specs/truth/techstack.md`.

**Goal**: resolve **RF-067-1** — the round-067 unpaired-call account (`UnpairedCallIDs`) has **no live consumer**; surface it as a **user-visible `stderr` diagnostic** (Q1 → A), **informational** (Q2 → A).

---

## D1 — The account gets a single owner in the domain, and a live consumer in the CLI

The round-067 account lived in the Gemini adapter (`roundBuilder.unpaired()` + `UnpairedCallIDs`). Round 068 **relocates the algorithm to a family-neutral single owner** — `llm.UnpairedToolCalls(msgs []llm.Message) []string` (round boundary = a message carrying tool calls; each result binds by `ToolCallID`, else FIFO; a plain-text turn closes the round; a standalone media turn is not a boundary) — and the adapter's `UnpairedCallIDs` **delegates** to it (the round-067 pins stay green, invariant behaviour). The adapter's per-round `dropped`/`unpaired()` machinery is **deleted** as dead code.

**Why a domain owner:** the CLI must not import the concrete Gemini adapter (ADR 0020 layer discipline; the deps seam keeps adapters out of `internal/cli`). A domain-pure function lets the CLI compute the same account for the request it is about to send, with **one** algorithm (no drift between the adapter's wire behaviour and the CLI's diagnostic).

## D2 — The consumer is a CLI gateway decorator (no port/factory signature change)

`internal/cli/unpaired_gateway.go` wraps the resolved `llm.Gateway` in a small decorator: before each `Complete`, it computes `llm.UnpairedToolCalls(req.Messages)`; when non-empty it emits the diagnostic, then delegates to the inner gateway **unchanged** (Q2 → A: informational — the request, and thus the wire, is untouched). This avoids widening the `deps.NewGateway` signature / the factory / the adapter `Config` (a much larger blast radius) — the decorator is built at the one construction site (`runTurn`).

## D3 — The surface is a `[Tool …]`-class `stderr` line, never `turns.log` (Q1 → A)

`internal/ui.FormatUnpairedCalls(t, ids)` renders `[HH:MM:SS] [Tool Warning] <n> tool call(s) left unanswered: <id, id>`; it is exposed through the existing **`render.Lines` domain port** (`UnpairedCalls`) so `internal/cli` names no `internal/ui` type (ADR 0020). It is written to `env.stderr` only (never `stdout`) and is **not** routed to `turns.log` (`-t` unchanged; the `[Tool Output]`-block precedent, ADR 0022 D5). It is **plain** (no colour) — a warning is not a chrome accent, and a colour-free line makes the terminal gate trivially safe (ADR 0023).

## D4 — Scope: family-local in effect; no producer today (honest)

The decorator wraps every gateway, but the OpenAI-compatible family has **no round-boundary drop**, so the account is **always empty** for it — the diagnostic is inert there (I-1 unaffected: the OpenAI wire is untouched). The shipped Gemini loop **appends one result per call** (`agentloop.go`), so **`M == N` always** in production and the diagnostic **does not fire** on any shipped path (I-7). It is a **defensive** surface for a partial/corrupted `prior` or a future out-of-order/concurrent dispatch ([#36](https://github.com/gosharplite/tellme/issues/36) item 3). This is disclosed on every surface; the round does **not** claim a live-surface change.

## D5 — Carriers: unit + CLI tiers (no E2E), the round-059 narrowing class

Because `M < N` has **no hermetic producer** through the built binary, **no godog Example can drive the diagnostic**. Its carriers are: a **domain** pin (`llm.UnpairedToolCalls` table test), a **ui** pin (`FormatUnpairedCalls`), and a **cli** pin (the decorator emits on a short round; silent on `M == N`; the emitter writes to `stderr`). `/axb-spec-by-example` and `/axb-dsl-refine` record a **documented narrowing** (the round-059 cadence-narrowing precedent; never an Example-less Rule — RF-063-10 retired).

## D6 — Verification

- **Domain** pin: all-paired (nil), short-round, zero-results (M=0 per round), multi-round, out-of-order, id-less FIFO, duplicate-id, text-closes-round.
- **ui** pin: the exact `[Tool Warning]` line, plain (no ANSI).
- **cli** pins: emit on `M < N` (the inner gateway still receives the request unchanged); silent on `M == N`; `unpairedEmitter` writes to `stderr` and no-ops on empty; nil-emitter passthrough.
- **Falsifiability**: suppressing the emit reds the cli pin; suppressing the account reds the domain pin.
- **Regression**: the round-065/066/067 pins + the E2E journeys stay green; the OpenAI wire and the Gemini batch/id shapes are unchanged (I-1/I-2/I-3); `stdout` byte-exact (I-4).

## D7 — Governance: a new ADR extending ADR 0037

**ADR 0038** records the diagnostic + the detection seam; it **delivers ADR 0037 §Forward RF-067-1** and annotates it. `specs/truth/techstack.md`: the *Vertex/Gemini adapter* row gains the "surfaced as a diagnostic" clause; the *Agent tool loop* row may note the decorator. `/axb-api-plan` **NOOP**; `/axb-data-plan` **NOOP**.

---

## Truth impact (semantic units)

| Action | Truth Spec | Summary |
| --- | --- | --- |
| MODIFY | `specs/truth/techstack.md` — *Vertex/Gemini adapter* | the `M < N` unpaired-call account is surfaced as a plain `stderr` `[Tool Warning]` diagnostic via the family-neutral `llm.UnpairedToolCalls` (adapter delegates) and a CLI gateway decorator; informational; not in `turns.log`; no live producer today. |
| ADD | `docs/decisions/0038-unpaired-call-diagnostic.md` (+ index; an annotation on ADR 0037's §Forward RF-067-1) | the diagnostic decision + the seam. |
| NOOP (checked) | the OpenAI-compatible rows | untouched (I-1; the account is always empty there). |

## Residual risks / forward items (non-blocking)

- **RF-068-1** — no E2E carrier: `M < N` has no hermetic producer (the round-059 narrowing class); a future out-of-order/concurrent dispatch ([#36](https://github.com/gosharplite/tellme/issues/36) item 3) would give it a live one.
- **RF-068-2** — the diagnostic is **plain** (no colour); colouring it is a deferred, separate decision.
- **RF-068-3** — the loud-failure variant (Q2 → B) is not taken; recorded.
- **RF-068-4** — `llm.UnpairedToolCalls` is computed eagerly per `Complete` (a cheap walk); a memoisation is unneeded at these sizes.

## References

- `internal/domain/llm/unpaired.go` · `internal/cli/unpaired_gateway.go` · `internal/ui/toolcall.go` (`FormatUnpairedCalls`) · `internal/domain/render/ports.go` (`Lines.UnpairedCalls`) · `internal/infrastructure/llm/gemini/client.go` (`UnpairedCallIDs` delegates).
- `docs/decisions/0037-gemini-toolcall-id-provenance.md` (§Forward RF-067-1) · `docs/decisions/0022-offline-session-config-and-turns-log.md` (the `turns.log` narrowing) · `docs/decisions/0023-list-default-and-chrome-colour.md` (the colour gate) · `docs/decisions/0020-cli-ui-decoupling.md` (the render.Lines port).
