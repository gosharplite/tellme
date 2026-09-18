# ADR 0017 — De-couple `internal/cli` from the TUI prompt: an injected domain `tui.Prompter` port (R5.2 of [#92](https://github.com/gosharplite/tellme/issues/92))

- **Status**: Accepted
- **Date**: 2026-09-18
- **Round**: 048 `048-cli-tui-prompt-decoupling` (R5.2 — the first de-coupling slice of [#101](https://github.com/gosharplite/tellme/issues/101))
- **Relates to**: **ADR 0011** (the layer-discipline gate + tier table + ratchet — *extended*, not superseded) · **ADR 0016** (RULE-E, the application import ceiling this round shrinks) · **ADR 0013** (the injected `Dependencies` seam + the composition root) · **ADR 0015** (the loop-presentation port — the precedent that a port is declared in `internal/domain/**`).

## Context

[#92](https://github.com/gosharplite/tellme/issues/92) **AC2** second clause: *"`internal/cli` depends only on `internal/domain/*`, stdlib, and application utilities."* Round 047 (R5.1) mechanised the ceiling as **RULE-E** (ADR 0016) and baselined the **3** residual unsanctioned edges:

```text
internal/cli -> internal/agent
internal/cli -> internal/ui
internal/cli -> internal/ui/tui/prompt
```

This round is **R5.2** — the first de-coupling slice. Clarify **Q1 → B** selected the smallest, most self-contained edge: **`internal/cli → internal/ui/tui/prompt`**. `internal/cli` uses the package in exactly two places (`cli.go`): `tuiprompt.Run(ctx, in, out, src, debounce)` and `tuiprompt.DefaultDebounceDuration`, both reached through the existing (widened) `tuiPromptRunner` func-type seam + `defaultRunTUIPrompt` nil-default.

The seam's shape is a problem in itself: PR [#104](https://github.com/gosharplite/tellme/pull/104) review recorded **F-4** — `cli.Options.RunTUIPrompt` is an **exported field of an unexported type** (`tuiPromptRunner`), so `go doc cli.Options` advertises a field no external package can set. Clarify **Q3 → A** folds F-4 into this round (it is the same seam being rewritten).

## Decision

1. **The port is a domain contract.** A new package `internal/domain/tui` declares:

   ```go
   type Source interface { Suggest(ctx context.Context, query string) []string }
   type Prompter interface {
       Run(ctx context.Context, in io.Reader, out io.Writer, src Source, debounce time.Duration) (string, bool, error)
       DefaultDebounceDuration() time.Duration
   }
   ```

   It references only stdlib (`context`/`io`/`time`) + domain types — **RULE-C-pure**; no `ui` type crosses in. (Q2 → A.)

2. **The adapter lives in the presentation tier.** A thin `internal/ui` type satisfies `domaintui.Prompter` by delegating to `tuiprompt.Run` + `tuiprompt.DefaultDebounceDuration`. `internal/ui` (tier 5) and `internal/ui/tui/prompt` (tier 5) are the same tier, so the import is **RULE-A-clean**; only tiers ≥ 5 may import `internal/ui/**`. Because `domaintui.Source` and `tuiprompt.Source` have identical method sets, the `Source` value is passed through with no wrapper.

3. **The wiring is an exported-type field, and F-4 closes.** `cli.Options` becomes `{ Deps deps.Dependencies; Prompter domaintui.Prompter }` — both fields of **exported** types. The `tuiPromptRunner` func type and the `RunTUIPrompt` field are **deleted**. `cmd/tellme` (the composition root, tier-table-exempt) injects the adapter. `internal/cli` retains the diagnostic hint + the suggestion-engine wiring (sanctioned imports) and loses its **nil-default** — the in-package fallback that imported the TUI package is removed. A nil `Prompter` is unreachable in production and returns a clear error, never a silent no-op.

4. **The RULE-E baseline ratchets 3 → 2.** `tools/arch/baseline.txt` is regenerated from the gate; the removed edge is gone; the other two residuals remain baselined for later slices. The round changes **no** user-facing behaviour (`stdout`/`stderr` byte-contracts, flags, exit codes, the TUI chrome are unchanged).

## What this change is *not*

- It does **not** touch the other two residual edges (`→ agent`, `→ ui`) — later R5 slices.
- It does **not** re-rule RULE-E's sanctioned set (ADR 0016 is unchanged).
- It does **not** change the tier table, the rule mechanism, or any existing rule's verdict (ADR 0011 stands; ADR 0016 stands).
- It is **not** a user-facing change (no flag, exit code, stream contract, or DSL vocabulary change) and adds **no** new dependency (`go.mod`/`go.sum` unchanged).

## Consequences

- **Positive**: `internal/cli` no longer imports `internal/ui/tui/prompt`; the RULE-E baseline shrinks (3 → 2, on the way to 0); F-4 is closed (`Options` exposes no field of an unexported type); the pattern — *domain interface + tier-≥5 adapter + composition-root injection* — is established for the remaining slices.
- **Cost**: one new zero-dependency domain package (`internal/domain/tui`) and one small `internal/ui` adapter; the `tuiPromptRunner` test seam is re-pointed to a fake port (test-only adaptation).
- **Faithfulness**: `internal/domain/tui`'s `Source`/`Prompter` mirror `tuiprompt.Source`/`tuiprompt.Run` verbatim, so the port is a thin structural bridge with no semantic reinterpretation.
- **Witness**: the gate (RULE-E reports **2**, 0 new / 0 stale; RULE-A/B/C 0; 0 cycles) + the adapted unit seams; falsifiability witnesses (a) re-introduced edge ⇒ red, (b) stale baseline line ⇒ red, (c) missing injection ⇒ loud failure — reproduced then reverted (ADR 0010). The E2E suite is regression, not the carrier (#92 AC5).

## Forward

- The **`internal/cli → internal/ui`** edge (the `ui.Format*`/`ui.New*` call sites) is the next natural slice; the port pattern is reusable.
- The **`internal/cli → internal/agent`** edge (the turn-path loop construction) is the deepest slice.
- **F-6/F-7/F-8** remain recorded on [#101](https://github.com/gosharplite/tellme/issues/101).
- At baseline **0** the ratchet has **no release valve** (ADR 0011/0016): a future legitimate ceiling violation must be refactored (or the rule amended), never baselined.
