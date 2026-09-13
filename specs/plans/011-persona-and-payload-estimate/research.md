# Phase 0 Research: tellme Persona-on-the-Wire & Wire-Faithful Payload Estimate (Round 011)

Topic: make `tellme`'s **outbound request faithful to its configuration** and its **pre-flight estimate faithful to the wire**. Two coupled defects: (i) the configured `PERSON` persona is **parsed but never sent**, so the model never receives the operator's instruction; (ii) the pre-flight estimate counts **only conversation message text**, so `b --new hi` reports `~5` while the provider measures `387`. Both defects were surfaced by comparing `tellme` against the reference `tell-me-go` on `deepseek-flash`: the reference injects the persona as a leading `system` message *and* its estimator counts the persona + tool declarations, which is why its estimate is comparable to the measured count.

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer (`gopkg.in/yaml.v3` + hand-written resolution), testing harness (`godog` + stdlib `testing`), provider transport (stdlib `net/http`), output rendering (glamour), session history (append-only JSON-Lines), the agent tool loop (round 008), the payload status line (round 009), and the cross-stream ordering contract (round 010) were locked in rounds 001–010. The system still has **one CLI end**, adds **no new system end**, **no new external service**, and **no new third-party dependency**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` and are **not re-decided here**: (1) single CLI end; (2) BDD techstack = `godog` running the built binary; (3) strategy = E2E black-box for the acceptance path plus fast unit tests for pure helpers. Clarify Round 1 (in `spec.md`) already locked the round's two high-impact decisions: Q1 estimate scope = persona **+ tool declarations + messages**; Q2 guarantee = pin the estimate's **inputs + determinism**, no numeric-equality/tolerance assertion. **IN this round**: the persona message on the wire, the estimator's widened inputs, and their executable assertions. **OUT**: any change to the answer stream, the status-line format, the measured line's source, or the request vocabulary beyond the added `system` message.

---

## Decision 1: The persona is a leading `system` message in the `messages` array

- **Decision**: The configured `PERSON` value is sent as a **separate leading `system` message** — the first element of the request `messages` array — ahead of the conversation (prior turns + current prompt). It is never merged into or prepended to the user prompt text, and never rendered to `stdout`.
- **Rationale**: This is the reference's shape for the OpenAI-compatible family (`maybeInjectInitialPersona` adds `{"role":"system","content":persona}` as the first message; the `developer` role is reserved for OpenAI reasoners, which tellme does not classify). A distinct message keeps the persona a first-class instruction and leaves the answer byte contract untouched.
- **Alternatives considered**:
  - **Prepend the persona to the user prompt text** — pollutes the prompt, is not a distinct instruction, and would corrupt the piped-stdin byte contract — rejected.
  - **Use the `developer` role** — only meaningful for OpenAI reasoner models; tellme ships only the OpenAI-compatible family and does no reasoner classification — rejected.
  - **Rely on the provider's own default system prompt** — the operator's configured instruction would be silently ignored — rejected (it is the defect).

## Decision 2: The persona rides every request the turn makes (transport-level injection)

- **Decision**: The persona is carried on **every** provider request made during the run — the main completion **and** any tool-driven completion (e.g. the session-summarisation tool's LLM call). It is injected at the **transport** level (the gateway/adapter) so a single construction covers all of them, rather than being re-supplied per call site.
- **Rationale**: The reference sets the persona once on the client (`openai.WithPersona`), so all calls inherit it. In tellme the gateway is built once per turn (`newGateway`) and the same instance is handed to the tool registry (`newToolRegistry(store, gw)`), so a transport-level persona covers the main loop and the summariser with no per-callsite plumbing.
- **Alternatives considered**:
  - **Add a `Persona` field to `llm.Request` and set it at each call site** — works but duplicates the assignment and lets a future call site silently omit it — rejected as the primary seam.
  - **Thread it only through the main chat path** — leaves the summariser's request without the persona, diverging from the reference and from `FR-004` — rejected.

## Decision 3: The persona is sourced from the resolved config; empty ⇒ no message

- **Decision**: The persona is the resolved `PERSON` value from the active configuration. When `PERSON` is empty/absent, **no** `system` message is sent (the conversation shape is unchanged).
- **Rationale**: `Config.Person` is already modeled (it is written by the harness's configuration Given today, just unused). Empty ⇒ no message mirrors the reference's `persona == ""` guard and keeps the no-persona request byte-identical to today.
- **Alternatives considered**:
  - **Send an empty `system` message** — an empty instruction is noise and changes the request shape for no benefit — rejected.
  - **Default a built-in persona** — would invent an instruction the operator did not configure — rejected.

## Decision 4: The estimator counts the wire payload — persona + tool declarations + messages

- **Decision**: The pre-flight estimate is computed over the **same input components sent on the wire**: the **persona message**, the **tool declarations** (name + description + parameter schema), and the **conversation messages** (prior turns + current prompt). The existing dependency-free heuristic is **extended in inputs**, exposed as a payload-level estimate used by the pre-flight line; the per-message term is unchanged.
- **Rationale**: The provider's `prompt_tokens` covers the system/persona message, the tools array, and the messages — the estimate must count the same three to be comparable (Clarify Q1 → Option 1). The tool declarations dominate a first-turn request, which is exactly what the old estimate ignored (`~5` vs `387`).
- **Alternatives considered**:
  - **Count persona + messages only** (Clarify Q1 Option 2) — keeps under-counting whenever tools are sent — rejected.
  - **Add a fixed base/framing allowance** (Clarify Q1 Option 3) — introduces an arbitrary constant with no wire source — rejected.
  - **Call a real BPE tokenizer** — a new dependency and network/asset weight the project forswore; the estimate must stay offline — rejected.

## Decision 5: Pin the estimate's inputs + determinism — not numeric equality

- **Decision**: Acceptance pins that the estimate is computed over the wire payload (Decision 4), is **deterministic** for identical inputs, and is **responsive** (a strictly larger wired payload ⇒ a strictly larger estimate). The estimate is **not** required to equal the provider's reported `prompt_tokens`, and **no** tolerance band is asserted.
- **Rationale**: The provider runs a real tokenizer over the exact request bytes; tellme uses an offline heuristic, so equality is impossible and a tolerance would be flaky across providers/heuristic changes (Clarify Q2 → Option 1). Pinning the inputs + determinism + responsiveness is stable, offline-testable, and makes the `~5 vs 387` defect class impossible.
- **Alternatives considered**:
  - **Assert the estimate within ±X% of the measured** (Clarify Q2 Option 2) — flaky and provider-frame-dependent — rejected.
  - **Pin determinism only** (Clarify Q2 Option 3) — would not close the "estimate ignores the wire" defect — rejected.

## Decision 6: Verification — reuse the fake, add the persona + estimate assertions

- **Decision**: The existing in-process fake provider (which already records the request's `messages` array — round 007 — and the sent tool definitions — round 008) is reused: the **persona message** is asserted from the recorded request's leading `system` message, and the **no-persona** case asserts its absence. The **estimate** is asserted by a pure-helper unit test (deterministic + responsive over an explicit persona/tools/messages input) plus an E2E check that the pre-flight line reflects a larger wired payload (the recorded request grows, the reported estimate grows). `stdout` stays byte-exact.
- **Rationale**: The verification surface already exists — no new harness dependency. The unit test pins the estimator's inputs cheaply; the E2E test pins the wired persona and the reported figure end-to-end.
- **Alternatives considered**:
  - **A new fake/recorder** — unnecessary; the round-007/008 recorder suffices — rejected.
  - **Assert the exact estimate number end-to-end** — couples the suite to a heuristic constant (fragile) — rejected.

## Decision 7: No new dependency; BDD techstack & strategy unchanged

- **Decision**: The round adds **no** third-party dependency (`go.mod` / `go.sum` unchanged) and does not change the BDD techstack (`godog` running the built binary) or the strategy (E2E acceptance path + fast unit tests for pure helpers). The persona injection and the estimator change are stdlib-only.
- **Rationale**: Adding a leading message and widening a heuristic's inputs are pure Go; the harness already records requests and tool definitions.
- **Alternatives considered**:
  - **Adopt a tokenizer / a provider SDK** — a new dependency with no need — rejected.

## Decision 8: The three AIxBDD must-ask questions remain settled

- **Decision**: No new system end; BDD techstack = `godog`; strategy = E2E + pure-helper units — unchanged this round, per the standing `techstack.md`. No `/axb-clarify` round is owed for them.
- **Rationale**: The round extends an existing CLI end's behaviour; it introduces no new interface, service, or test framework.
- **Alternatives considered**:
  - **Re-open the techstack questions** — no change to the ends or the runner — rejected.

---

## Residual risks / forward links

- **Persona plumbing seam (`FR-004`)**: the exact seam (extend the `gatewayFactory`/`openai.Config` with the persona vs a per-request field) is a `/axb-system-analysis` / implementation determination; Decision 2 fixes *that all requests carry it and where it is sourced*, not the parameter shape.
- **Estimation heuristic constants (A5)**: the chars/token ratio, the per-tool-declaration term, and whether a base term is used are **not** contract terms; they are chosen in implementation to make the estimate responsive without over-fitting. Only determinism, responsiveness, and the counted **inputs** are pinned (Decision 5).
- **No numeric parity**: the estimate will still under/over-count vs the provider's BPE count; this is accepted and asserted only as inputs + determinism + responsiveness.
- **Existing-truth rewrite (`FR-011`)**: the current assertion that a first-turn request "carries no earlier exchange / messages are exactly the current user prompt" must be reworded to account for the leading `system` message — a `/axb-dsl-refine` determination, recorded as a MODIFY toward the `chat` interface.
- **Persona assertion granularity**: whether the persona is asserted via a new Given/Then row or an extension of the existing fake-recorder rows is a `/axb-dsl-refine` determination.
- **Deferred, still out of scope**: streaming, pinning, `-b`/`--retry`, token-budget pruning, MCP, memory, TUI, and the reference's `developer`-role/OpenAI-reasoner path.
