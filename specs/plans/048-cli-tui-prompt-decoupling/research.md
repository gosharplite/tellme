# Phase 0 Research: de-couple `internal/cli` from the TUI prompt — R5.2, first de-coupling slice (Round 048)

Topic: remove the residual layer-discipline edge **`internal/cli -> internal/ui/tui/prompt`** (RULE-E baseline, ADR 0016) by inverting the CLI's call into the interactive TUI prompt into an **injected domain port** — a behaviour-preserving structural refactor that ratchets the RULE-E baseline **3 → 2** ([#101](https://github.com/gosharplite/tellme/issues/101) → [#92](https://github.com/gosharplite/tellme/issues/92) AC2, second clause).

Scope note: the language (`Go 1.26`), module, CLI flag layer (`spf13/pflag`), config layer, testing harness (`godog` on the built binary; stdlib `testing` for units), provider transports, skills, MCP, the **Bubble Tea** TUI family, and the presentation packages were locked in rounds 001–047. This round adds **no new system end**, **no external service**, and **no new third-party dependency**; the system keeps **one CLI end**. The three AIxBDD must-ask questions remain answered by the standing `techstack.md` (single CLI end; BDD techstack = `godog` over the built binary; E2E black-box + unit strategy) and are **not re-decided** (D8). **IN**: the domain port, the `internal/ui` adapter, the CLI rewiring, the `cmd/tellme` injection, the baseline 3 → 2, ADR 0017, the truth row, and the F-4 closure. **OUT**: the other two residual edges (`→ agent`, `→ ui`), F-6/F-7/F-8, and any `tellme` binary behaviour.

> **Provenance of the locked scope (#101 + this round's clarify):** round 048 is **R5.2** — the **first de-coupling slice** of the R5 programme (round 047 = R5.1, the RULE-E gate). Clarify **Q1 → B** (slice = the TUI prompt edge), **Q2 → A** (the port lives in `internal/domain/**`), **Q3 → A** (F-4 folded in). The mandatory questions (D8) are unchanged from the standing truth.

---

## Decision 1: The DoD is the RULE-E baseline **3 → 2**; the round is behaviour-preserving (no product behaviour)

- **Decision**: the round removes exactly the **`internal/cli -> internal/ui/tui/prompt`** edge from the production **and test** import graph; the committed `tools/arch/baseline.txt` is regenerated from the gate and lists the remaining two residuals (`internal/cli -> internal/agent`, `internal/cli -> internal/ui`). RULE-A/B/C stay **0**; cycles **0**; the gate is **green on `dev` at delivery**.
- **Rationale**: the round's success is machine-checked by the RULE-E gate (ADR 0016) — #92 **AC5** (the witness is the gate, not the E2E suite). Removing one edge is the smallest falsifiable increment of the R5 programme.
- **Alternatives considered**: **all three edges at once** (Q1-A) — #101 calls R5 "a multi-round programme, not a slice"; **the `agent`/`ui` edges** (Q1-C/D) — later slices, larger behavioural surfaces — both rejected.

## Decision 2: The injected port is a **domain-declared `Prompter`** carrying the `Run` call + the debounce default

- **Decision**: declare a small domain port (a new package `internal/domain/tui`) with:

  ```go
  // Source yields candidate suggestions for the current query (ctx-cancellable).
  type Source interface { Suggest(ctx context.Context, query string) []string }

  // Prompter runs the interactive prompt for one invocation.
  type Prompter interface {
      Run(ctx context.Context, in io.Reader, out io.Writer, src Source, debounce time.Duration) (string, bool, error)
      DefaultDebounceDuration() time.Duration
  }
  ```

  The port carries **exactly** the two things `internal/cli` needs from the TUI package today: the `Run` call (`cli.go:202`) and the debounce fallback (`cli.go:214`). Every other TUI-related construction (the prompt tracker, the three-reader registry, the multi-source suggestion engine, the diagnostic hint) stays in `internal/cli` — it references only **sanctioned** imports (`internal/domain/**`, `internal/app/**`, `internal/config`/`home`).
- **Rationale**: the port must be **narrow** (the widened `tuiPromptRunner` shape carried cli-internal types `resolution`/`runtimeEnv`/`deps.Dependencies`, which a domain port cannot name — a self-referential, cli-shaped seam). A minimal `Run`-shaped port is the tightest contract that removes the import, and it matches the existing `tuiprompt.Run`/`tuiprompt.Source` verbatim.
- **Alternatives considered**: **a port carrying the whole `defaultRunTUIPrompt`** (incl. the suggestion wiring) — impossible without leaking cli-internal types, or moving the suggestion engine into the TUI package (re-opening round 015/044 decisions) — rejected; **a `func`-typed port** rather than an interface — the repo's port convention is a named interface (ADR 0013/0015), and an interface gives an obvious test double — rejected.

## Decision 3: The port lives in **`internal/domain/**`** — RULE-C-pure (clarify Q2 → A)

- **Decision**: the port is a **domain-owned contract** (`internal/domain/tui`), referencing only **stdlib** (`context`/`io`/`time`) + domain types (**no** `ui` type crosses in). `internal/domain/tui` imports nothing but `context`/`io`/`time` — RULE-C holds trivially.
- **Rationale**: a port is a domain-facing contract (ADR 0015 declared the loop-presentation port in `internal/domain/agent`); [#101](https://github.com/gosharplite/tellme/issues/101) open-decision 3 recommends the domain home. The `Source` interface is satisfied **structurally** by `cli.tuiSource` (and by `internal/app/suggestions`) — no adapter needed, because the method set is identical to `tuiprompt.Source` (D4).
- **Alternatives considered**: **`internal/app/**`** (Q2-B) — an application *utility*, whereas this is a plain outbound contract; the domain home is more faithful and keeps `internal/cli` importing only domain — rejected.

## Decision 4: The adapter lives in **`internal/ui`** and delegates verbatim (RULE-A)

- **Decision**: add a thin adapter in the presentation tier (`internal/ui`, a new small type, e.g. `ui.TUIPrompter`) that satisfies `domaintui.Prompter` by delegating to `tuiprompt.Run` + `tuiprompt.DefaultDebounceDuration`. `internal/ui` is **tier 5**; `internal/ui/tui/prompt` is also **tier 5** — a same-tier import is RULE-A-clean.
- **Rationale**: **RULE-A** (ADR 0011 D1) forbids an *upward* import, so only tiers ≥ 5 may import `internal/ui/**` — the adapter must live inside a tier-5 package (or the exempt `cmd/tellme`). `internal/ui` already imports several `internal/domain/**` packages (`llm`, `metrics`, `agent`), so importing `internal/domain/tui` is consistent. Because `domaintui.Source` and `tuiprompt.Source` have **identical method sets** (`Suggest(ctx context.Context, query string) []string`), a `domaintui.Source` value is assignable to `tuiprompt.Source` — the adapter needs **no** source-wrapping.
- **Alternatives considered**: **a closure in `cmd/tellme`** (exempt) — works, but leaves the presentation adapter in `main` and duplicates a per-invocation closure; a named `internal/ui` type is the reusable home — rejected; **`internal/ui/tui/prompt` satisfying the port itself** — the package function `Run` is not a method; a receiver type in `prompt` would be fine too, but `internal/ui` already owns the presentation seam conventions (round 046 `ToolLineRenderer`) — `internal/ui` chosen for locality.

## Decision 5: The wiring is a **second, exported-type field** on `cli.Options`; `TODO` F-4 is closed (clarify Q3 → A)

- **Decision**:
  - `cli.Options` becomes `{ Deps deps.Dependencies; Prompter domaintui.Prompter }` — **both fields are of exported types**.
  - The `tuiPromptRunner` **func type** and the `Options.RunTUIPrompt` field are **deleted** (F-4: an exported field of an unexported type is gone).
  - `cmd/tellme/buildOptions()` wires the port: `cli.Options{Deps: buildDeps(), Prompter: ui.TUIPrompter{}}`.
  - `internal/cli`'s `defaultRunTUIPrompt` is **retained as logic** (the diagnostic hint + the suggestion-engine wiring) but its `tuiprompt.Run` call + debounce fallback go through the injected port (`opts.Prompter.Run(...)` / `opts.Prompter.DefaultDebounceDuration()`); the **nil-default is removed** (no in-package fallback that imports the TUI package — NFR-001).
  - A **nil** `Prompter` is a **programming error** (unreachable in production — the composition root always injects): `cli` returns the existing environment-error path (a clear message, never a silent no-op).
- **Rationale**: F-4 exists precisely because `Options` advertised an unsettable field; making the field type an **exported domain interface** resolves it cleanly, and the port replaces the widened cli-shaped func type. Deleting the nil-default is what removes the `internal/cli → internal/ui/tui/prompt` import.
- **Alternatives considered**: **keep the field name `RunTUIPrompt`** — fine, but `Prompter` names the seam's role (a runner) more directly — either is acceptable; RD may pick the name. **Put the port in `deps.Dependencies`** — `deps` holds **factory funcs**; the prompter is one stateless value; `Options` already owned the presentation seam pre-round-044 — rejected.

## Decision 6: The test seam is re-pointed from a fake func to a fake **port** (test-only adaptation)

- **Decision**: `tui_dispatch_test.go` and `tui_submit_chrome_test.go` (and the `testdeps_test.go` comment) construct `Options{Prompter: fakePrompter{...}}` instead of `Options{RunTUIPrompt: func…}`; the fake implements `domaintui.Prompter` (returning canned `(text, ok, err)`). No production assertion changes; the **behaviour** asserted (the dispatch reaches the seam; `stdout` stays empty; the submit path) is identical.
- **Rationale**: the seam's *shape* changes but its *contract* does not; the existing tests already target the contract (a canned `(text, ok, err)`), so re-pointing them is a pure adaptation. Keeping them proves behaviour preservation at the unit layer (the round's witness, NFR-004).
- **Alternatives considered**: **delete the tests** — loses the round's unit witness — rejected; **new tests only, leave the old failing** — impossible (they reference a deleted field) — rejected.

## Decision 7: ADR **0017** + the truth MODIFY (the durable record)

- **Decision**: record the de-coupling durably in a **new ADR 0017** (`docs/decisions/0017-cli-tui-prompt-decoupling.md` + the `docs/decisions/README.md` index row): the domain port, its home, the adapter home, the behaviour-preservation claim, the baseline movement **3 → 2**, the **F-4 closure**, and its relation to ADR **0011/0016** (the rule + ratchet it advances), **0013** (the injected seam + composition root it extends), and **0015** (the port-in-domain precedent). Update `specs/truth/techstack.md`: the **Layer-discipline gate** row's baseline figure becomes **RULE-E 2** (naming the removed edge); the **Composition root (dependency injection)** row's `RunTUIPrompt` clause is restated as the injected **domain** `tui.Prompter`; the **Interactive TUI prompt (`-i`)** row notes the port de-coupling.
- **Rationale**: per `docs/decisions/README.md`, a structural change future slices must cite gets an ADR; the R5.x slices must cite the port pattern. The gate row is the sole truth home for the baseline figure (a real MODIFY, no unevidenced NOOP).
- **Alternatives considered**: **truth prose only, no ADR** — leaves the pattern non-citable — rejected (mirrors rounds 042/047).

## Decision 8: Testing & BDD techstack unchanged; `/axb-api-plan`/`/axb-data-plan`/`/axb-dsl-refine` are `NOOP`

- **Decision**: no new system end, no BDD-techstack change, no test-strategy change. `/axb-spec-by-example` = **NOOP** (no user-facing business journey); `/axb-api-plan` = **NOOP** (no OpenAPI surface); `/axb-data-plan` = **NOOP** (the refactor moves wiring, not persisted/runtime state); `/axb-dsl-refine` = **NOOP** (a `cli`-reader refactor is not the `tellme` CLI contract); `/axb-ui-plan` = skipped. The three AIxBDD must-ask questions stay answered by the standing `techstack.md` and are **not** re-decided.
- **Rationale**: the round changes no observable CLI behaviour, so no Gherkin scenario is the carrier; the E2E suite runs green as **regression**. Authoring CLI Gherkin for an internal wiring change would risk an `acceptance-coverage` mismatch (the round-042 A3/A6 precedent).
- **Alternatives considered**: **author a plan-side acceptance Feature** — no user-facing behaviour to express — rejected.

## Decision 9: The witness is the **gate + unit seams**, with reproduced-then-reverted falsifiability witnesses

- **Decision**: prove the de-coupling by (i) the gate (RULE-E reports the baseline **2**, 0 new / 0 stale; RULE-A/B/C 0; 0 cycles) and (ii) the adapted unit seams (dispatch + submit), **plus** three falsifiability witnesses reproduced then reverted (ADR 0010): (a) re-introduce `internal/cli → internal/ui/tui/prompt` ⇒ RULE-E reds and names the edge; (b) leave the now-fixed TUI edge in `baseline.txt` ⇒ the gate reds (stale); (c) inject a missing/false prompter ⇒ the TUI dispatch unit seam (or the build) fails loudly. The E2E suite is run green as **regression**, but it is **not** the acceptance carrier.
- **Rationale**: #92 **AC5** — the witness is the gate + unit seams. The gate's fail-on-stale + the import scan are the machine-checkable claims; the witnesses prove both the rule fires (a) and the ratchet has teeth (b).
- **Alternatives considered**: **rely on the green E2E suite** — a passing suite says nothing about whether the import is gone — rejected (the round-009 trap).

## Decision 10: Inherited mechanism, determinism, and the ratchet's terminal state are unchanged

- **Decision**: RULE-E and its machinery are **not** re-litigated (ADR 0011 D2–D9, ADR 0016): module-root-anchored `go list`, the `CROSS_TARGETS` union + filtered child env, the merged production+test graph, the production-only SCC pass, the sorted baseline, the fail-on-stale ratchet, default-deny. The **sanctioned set** (ADR 0016) is **unchanged** — this round **removes an edge**, it does not re-rule the allow-list. At the terminal state (baseline 0, after the later slices) the **no-release-valve** policy applies equally.
- **Rationale**: the round only shrinks the baseline; re-deriving the rule would re-open settled decisions and risk divergence.
- **Alternatives considered**: **re-rule the sanctioned set as part of this round** — out of scope (a separate, higher-impact decision) — rejected.

---

## Residual risks / forward links

- **The other two residual edges** (`internal/cli → internal/agent`, `internal/cli → internal/ui`) — later R5 slices; recorded on **#101**.
- **F-6/F-7/F-8** (PR #104 review) — out of scope; recorded on **#101**. **F-4 closes** this round (Q3).
- **The `internal/cli → internal/ui` edge still carries `ui.Format*`/`ui.New*`** — after this round, that residual is the next natural slice; the port pattern established here (domain interface + `internal/ui` adapter + composition-root injection) is directly reusable.
- **Adapter home** (`internal/ui` vs a `cmd/tellme` closure) — RD may choose either; both are RULE-A-clean. If a future `internal/app` consumer needs the prompter, the domain port already makes it reachable.
- **A nil `Prompter`** is unreachable in production (the composition root always injects); `cli` returns a clear error rather than a silent no-op — a defensive path, recorded.
- **Custom build-tag-gated files** stay out of scope (ADR 0011 D6).
- **No behaviour change** — `stdout`/`stderr`, class-phrase vocabulary, flags, exit codes, the TUI chrome, and every existing rule's verdict are unchanged; the round touches `internal/cli/**`, a new `internal/domain/tui` package, `internal/ui/**` (the adapter), `cmd/tellme` (wiring), `tools/arch/baseline.txt`, an ADR, `specs/truth/techstack.md`, the plan package, and `STATUS.md` + the day's summary.
