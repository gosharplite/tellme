# ADR 0039 — The `ToolSetSpec` capability seam

- **Status:** Accepted
- **Date:** 2026-09-20
- **Deciders:** tellme owner
- **Related:** [ADR 0032](0032-agent-image-vision.md) (agent image vision — this ADR **delivers its review fold F-062-4 / §Forward RF-062-10**, the seam half), [ADR 0033](0033-gemini-image-vision.md) (Gemini image vision — this ADR **delivers its §Forward RF-063-6**), [ADR 0013](0013-composition-root-injection.md) (the composition root + `deps.Dependencies`), [ADR 0019](0019-agentloop-port.md) (`LoopSpec` — the same "name the construction inputs" pattern), [ADR 0021](0021-tool-output-ctor-injection.md) (the ctor-injected `[Tool Output]` sink this path carries), round 062 (`specs/plans/062-agent-image-vision`), round 063 (`specs/plans/063-gemini-image-vision`), round 069 (`specs/plans/069-toolset-spec-capability-seam` — this ADR's round), issue [#140](https://github.com/gosharplite/tellme/issues/140)

## Context

The agent tool-registry construction threaded its capability facts as **positional scalars**:

```go
// internal/app/deps/deps.go
NewToolRegistry func(sink domaintools.OutputSink, vision bool, providerType string) domaintools.Registry
```

The shape grew **one scalar per capability round**: `vision bool` (round 062 — the `read_image` + `VISION` round) and then `providerType` (round 063 — so the composition root could resolve the family-aware inline ceiling). The round-062 review recorded **F-062-4** (*"a `ToolSetSpec` seam instead of a bare positional `vision bool`"*) as **RF-062-10**; the round-063 review recorded **RF-063-6**, annotating it **overdue** — *"`NewToolRegistry` gained a third positional scalar one round after F-062-4 recorded the seam … the recommended next-round landing is a `ToolSetSpec` (capability set + resolved limits) instead of a fourth scalar."*

Two costs followed: (1) the producer (`cmd/tellme`) and the offline caller (`internal/cli`) must remember the **arity and order**, and (2) the anonymous function type is **re-spelled** at the call site (`internal/cli/cli.go`, `renderToolUsage`). Both are maintenance hazards whose probability rises with each added capability.

## Decision

**D1 — One named value (`deps.ToolSetSpec`) replaces the positional scalars.** It is declared in `internal/app/deps` — the composition-root **contract** package (ADR 0013) — which already imports `internal/domain/tools` and is already imported by `internal/cli`, so the change adds **no import edge**. It carries the `[Tool Output]` sink, the `vision` capability gate, and the provider **label**.

**D2 — The spec carries the raw inputs; the composition root resolves the family ceiling.** Resolving the ceiling at the *caller* would force `internal/cli` to name `infrallm.Family` / `infratools.ImageCeilingForFamily` — an `internal/cli → internal/infrastructure` edge, a **RULE-B layer violation** the repo holds at a **0-violation baseline** (the round-051 R5.5 programme). The ceiling therefore stays resolved inside the composition-root builder, exactly as before. *(This is a layer-safety constraint, not a preference.)*

**D3 — The builder resolves the ceiling lazily; no positional scalar remains.** `assembleAgentTools(spec)` computes the ceiling **inside** the `if spec.Vision` branch (its only use), so it needs no `imageCeiling int` parameter. After this change **no positional `bool`/`string`/`int` capability scalar remains** on the path (`NewToolRegistry` / `newToolRegistry` / `assembleAgentTools` / `renderToolUsage`).

**D4 — The callers name fields.** The prompt path (`internal/cli`) constructs `deps.ToolSetSpec{Sink: prog.ToolOutput, Vision: res.Provider.Vision, ProviderType: res.Provider.Type}`; the offline `--tool-usage` path takes the **named** signature `func(deps.ToolSetSpec) domaintools.Registry` and builds its two capability-blind union variants as `deps.ToolSetSpec{}` and `deps.ToolSetSpec{Vision: true}` — replacing the duplicated anonymous type and the positional `false`/`true`/`""`.

**D5 — A new capability is a field, never a fourth scalar.** A future gate (a third family, a new tool capability) extends `ToolSetSpec` as a **field** (or sub-struct); every call site keeps compiling with named fields. This is the seam's purpose.

**D6 — Behaviour is byte-identical; no config, no UX.** The offered set is still a function of `Vision`; the ceiling computation is unchanged; the `--tool-usage` union, the chrome, `stdout`, exit codes, the persisted records, and the **config schema** (`VISION`/`TYPE`) are untouched. The acceptance is **structural** — the existing pins stay green with **no assertion changed**.

**D7 — Scope: the seam only.** **RF-062-10 bundles two items** — this seam **and** the media-channel refactor (media returned from `Execute` instead of the per-call `context` collector). **Only the seam lands here.** The media-channel refactor is a separate, larger, cohesion-motivated item (its own round; recorded).

## Consequences

### Positive

- The registry-construction path takes **one self-describing value**; the next capability is a **field**, not a fourth positional argument plus a re-spelled function type at each call site.
- The seam recorded twice (**F-062-4 / RF-062-10**, **RF-063-6**) is closed.
- **No new import edge**, **no config change**, **no UX change**; behaviour is provably identical.

### Negative / Accepted Trade-offs

- A small struct is introduced; the constructor calls change shape (call-site churn in four files + three test doubles). Purely mechanical.
- The spec carries the **raw** `ProviderType` (the ceiling is resolved downstream by the root). A future round may prefer a typed family value if a second consumer appears (recorded, **RF-069-2**).

### Neutral

- No domain entity, persisted record, config key, tool, or wire change; the reason gate (ADR 0025), the round-024 resource contract, and the layer baseline are untouched.

## Alternatives Considered

1. **Keep the positional scalars (do nothing).** Rejected — it is the overdue debt (#140), and each capability round makes it worse.
2. **A `map[string]any` capability bag.** Rejected — untyped; it trades positional fragility for stringly-typed fragility.
3. **Declare `ToolSetSpec` in `internal/domain/tools`.** Rejected (D1) — it moves a composition concern into the domain layer and widens the domain surface for no gain.
4. **Carry the resolved `ImageCeiling` on the spec (resolve at the caller).** Rejected (D2) — a layer violation (`internal/cli → internal/infrastructure`).
5. **Fold in the media-channel refactor (RF-062-10's other half).** Rejected (D7) — a larger blast radius, its own round; the spec's I-4 keeps it out.

## Verification

- **Structural** — no positional capability scalar remains on the construction path (`NewToolRegistry` / `newToolRegistry` / `assembleAgentTools` / `renderToolUsage`); the call sites name fields. Grep-verifiable.
- **Behaviour-identity** — the full unit suite + the godog E2E (`offering-the-agent-tools`, `reading-a-local-image`, `accounting-for-the-tool-use`, the offline `--tool-usage` pin) stay green with **no assertion changed**.
- **Gates** — `make verify` green (the layer gate stays at 0; `modelith-check` unchanged); `go.mod`/`go.sum` unchanged.
- **Falsifiability** — dropping `spec.Vision` from the gate reds the existing offered-set pins; a wrong ceiling reds the image-ceiling pin.

## References

- `internal/app/deps/deps.go` (the `NewToolRegistry` field + the new `ToolSetSpec`) · `cmd/tellme/deps.go` (`assembleAgentTools`, `newToolRegistry`) · `internal/cli/cli.go` (the prompt path + `renderToolUsage`) · `internal/config/config.go` (`Provider.Vision` / `.Type`).
- [ADR 0013](0013-composition-root-injection.md) · [ADR 0019](0019-agentloop-port.md) · [ADR 0021](0021-tool-output-ctor-injection.md) · [ADR 0032](0032-agent-image-vision.md) (RF-062-10) · [ADR 0033](0033-gemini-image-vision.md) (RF-063-6).
- Round 062: PR [#129](https://github.com/gosharplite/tellme/pull/129) · round 063: PR [#130](https://github.com/gosharplite/tellme/pull/130) · `specs/plans/069-toolset-spec-capability-seam/` (`spec.md` FR-001…FR-006; `research.md` D1…D7).

## §Forward (deferred, non-blocking)

> **⚠ Not open work.** A `§Forward` entry is a decision *deferred to a trigger* or a recorded divergence — **not** tasking. Do not re-raise absent its trigger.

- **RF-069-1** — the **media-channel refactor** (RF-062-10's other half): return media from `Execute` (a `domain/tools` media type) instead of the per-call `context` collector (ADR 0032 D7a). Its own round; larger blast radius.
- **RF-069-2** — `ToolSetSpec.ProviderType` is the raw label; a typed family value may be preferred if a second consumer appears. One consumer today.
- **RF-069-3** — the spec is a value type (no shared instance to mutate).
- **RF-069-4** — the ceiling is resolved inside the vision branch; hoist it if a non-vision consumer appears.
- **RF-069-5** — **CLOSED (fold F2, PR #141 review)** — the family-aware ceiling **consumption** now has a witness: `resolveImageCeiling(spec)` (a named seam in `cmd/tellme`) is pinned by `TestResolveImageCeilingPinsTheFamilyAwareConsumption` (gemini/google → 14 MiB; deepseek/"" → 32 MiB; the families must differ). The E2E `reading-a-local-image` journey alone could not witness it (its oversize fixture exceeds **both** ceilings), which is what this item recorded; the pin closes the gap.
