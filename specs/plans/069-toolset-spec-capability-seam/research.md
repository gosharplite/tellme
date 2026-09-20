# Technical Research — the `ToolSetSpec` capability seam (round 069)

**Plan Package**: `specs/plans/069-toolset-spec-capability-seam`
**Spec**: `spec.md` (anchor issue [#140](https://github.com/gosharplite/tellme/issues/140))

**Input (operator)**: *"Create a detail new issue for this. Open round 069, the goal is to close this new issue."*

**Grounding**: measured on `dev` @ `5971822`. All file:line references below are to that tree.

---

## Problem restated

The agent tool-registry construction threads its **capability facts** as **positional scalars**:

```go
// internal/app/deps/deps.go:70
NewToolRegistry func(sink domaintools.OutputSink, vision bool, providerType string) domaintools.Registry
```

The producer (`cmd/tellme`) and the offline caller (`internal/cli`) must therefore remember the arity and the argument **order**, and the anonymous function type is **re-spelled** at the call site (`internal/cli/cli.go:998`). The shape grew one scalar per capability round: `vision bool` (round 062) then `providerType` (round 063) — one round after the round-062 review recorded the seam (**F-062-4** → **RF-062-10**); the round-063 review recorded it **overdue** (**RF-063-6**).

---

## Decisions

### D1 — The named value is `deps.ToolSetSpec`, declared in `internal/app/deps`

`internal/app/deps` is the **composition-root contract** package (round 044 / **ADR 0013**): it already imports `internal/domain/tools` (`domaintools`) and holds `Dependencies`; `internal/cli` already imports `deps`. So a struct declared there needs **no new import edge** anywhere, and it keeps the type next to the field that consumes it (`Dependencies.NewToolRegistry`).

*Alternative rejected:* `internal/domain/tools` — it would move a composition concern into the domain layer for no gain, and it would widen the domain package's public surface with an assembly-only type.

### D2 — The spec carries the **raw inputs**; the composition root resolves the family ceiling

```go
type ToolSetSpec struct {
    Sink         domaintools.OutputSink // nil on the offline --tool-usage path (as today)
    Vision       bool                   // the declared capability gate (config.Provider.Vision)
    ProviderType string                 // the provider label, resolved to a family by the root
}
```

**The resolved-ceiling form was rejected.** Resolving `imageCeiling` at the *caller* would require `internal/cli` to call `infrallm.Family` / `infratools.ImageCeilingForFamily` — an `internal/cli → internal/infrastructure` edge, which is a **RULE-B layer violation** the repo holds at a **0-violation baseline** (the round-051 R5.5 programme). The ceiling must be resolved where it is today: inside the composition-root builder (`cmd/tellme`), which is the only place permitted to name the infrastructure. So the spec carries `ProviderType` and the builder derives the ceiling.

*(This is why D2 is a **layer-safety** decision, not merely a taste one.)*

### D3 — `assembleAgentTools(spec)` resolves the ceiling lazily and drops **every** positional scalar

```go
func newToolRegistry(spec deps.ToolSetSpec) domaintools.Registry {
    return domaintools.NewRegistry(assembleAgentTools(spec)...)
}

func assembleAgentTools(spec deps.ToolSetSpec) []domaintools.Tool {
    ...
    if spec.Vision {
        ceiling := infratools.ImageCeilingForFamily(infrallm.Family(spec.ProviderType))
        tools = append(tools, infratools.NewReadImageTool(ceiling))
    }
    return tools
}
```

The resolution moves **into** the `if spec.Vision` branch (it is only ever used there), so `assembleAgentTools` needs **no** `imageCeiling int` parameter. After this change **no positional `bool`/`string`/`int` capability scalar remains** on the whole path — the SC-001 witness.

### D4 — `renderToolUsage` takes the **named** signature; its variants become named-field constructions

```go
func renderToolUsage(env runtimeEnv, newToolRegistry func(deps.ToolSetSpec) domaintools.Registry, ...) int {
    ...
    names := unionToolNames(newToolRegistry(deps.ToolSetSpec{}), newToolRegistry(deps.ToolSetSpec{Vision: true}))
```

The duplicated anonymous type (`func(domaintools.OutputSink, bool, string)`) is replaced by the named signature; the capability-blind union (round 062 fold **F-062-1**) is preserved (both variants, both with a nil sink), now with **named fields** instead of a positional `false/true/""`.

### D5 — A new capability is a **new field**, never a fourth scalar

The struct documents the invariant: a future gate (a third family, a new tool capability) extends `ToolSetSpec` as a **field** (or a sub-struct), and every call site keeps compiling with named fields. This is the seam's whole purpose (FR-004).

### D6 — Behaviour is byte-identical; no config, no UX

The offered set is still a function of `Vision`; the ceiling value is the same computation; the union, the chrome, `stdout`, exit codes, and the persisted records are untouched. The config schema is untouched (`VISION`/`TYPE` read exactly as today). The acceptance is **structural** (existing pins green with **no assertion changed**).

### D7 — Governance: a new ADR (0039), the ADR 0013/0019/0021 structural lineage

Record the seam as a durable structural decision; annotate **RF-062-10** (the seam half) and **RF-063-6** as **delivered**. The **media-channel refactor** (RF-062-10's other half — media returned from `Execute` instead of the per-call `context` collector) stays a **recorded forward item** and is **out of this round** (its blast radius is far larger and its motivation is cohesion, not this seam).

---

## Truth impact (techstack.md)

- **MODIFY** — *Composition root (dependency injection)*: the agent tool-registry build func now takes **one named `deps.ToolSetSpec`** (round 069 / ADR 0039).
- **MODIFY** — *Image filesystem tool (`read_image`)*: the parenthetical that names the `ToolSetSpec` seam as a next-round refactor is updated — the seam is **landed (round 069 / ADR 0039)**; only the media-channel half of **RF-062-10** remains.
- **NOOP** — every other row (no behaviour, no new dependency, no config key).

---

## Residual risks (forward items)

- **RF-069-1** — the media-channel refactor (RF-062-10's other half) remains open: media is attached via the per-call `context` collector, so a tool's full effect is not in its return value, and `internal/infrastructure/tools` imports `internal/domain/llm`. Its own round.
- **RF-069-2** — `ToolSetSpec.ProviderType` is the raw label; a future round may prefer a **typed** family value (the `infrallm.Family` classification) if a second consumer of the label appears. Today there is one consumer.
- **RF-069-3** — the spec is a **value** (not a pointer); a caller cannot accidentally mutate a shared spec. No shared instance exists today.
- **RF-069-4** — `assembleAgentTools` resolves the ceiling only in the vision branch; a future **non-vision** consumer of the ceiling (e.g. a media-limit check that runs without the tool) would need the resolution hoisted. Recorded, not needed today.

## Verification (expected — finalised in `tasks.md`)

- **Structural** — no positional capability scalar remains (`NewToolRegistry` / `newToolRegistry` / `assembleAgentTools` / `renderToolUsage`); grep-verifiable (SC-001).
- **Behaviour-identity** — the full unit suite + the godog E2E stay green with **no assertion changed** (SC-002).
- **Gates** — `make verify` green (the layer gate stays 0; `modelith-check` unchanged); `go.mod`/`go.sum` unchanged (SC-003).
- **Falsifiability** — dropping `spec.Vision` from the gate reds the existing offered-set pins; a wrong **generic** ceiling reds the E2E `reading-a-local-image` journey. *(Correction, PR #141 fold F3: the family-aware ceiling's **consumption** is carried by the `resolveImageCeiling` table pin added at fold F2 — the E2E's oversize fixture exceeds **both** family ceilings, so it cannot distinguish them. See `tasks.md` and ADR 0039 RF-069-5.)*
