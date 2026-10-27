# Research — round 098 `098-param-swap-guard`

**Owner**: `/axb-technical-research` (RD) · **Truth owner for**: `specs/truth/techstack.md`
**Input**: `spec.md` · issue [#210](https://github.com/gosharplite/tellme/issues/210) · the pre-round tree
(`dev` `0231f64`)

Every figure below comes from `make verify-params-report`
(`TestParameterShapeReport`), never from prose. That rule is itself one of the round's decisions (D8): in
the five drafts behind issue #210, every wrong number came from memory and every right one came from a
script.

---

## D1 — Guard-first, because a refactor has no falsifiable carrier

A refactor-only round cannot fail. A guard that is green before the tree is fixed cannot fail either.
The round therefore runs in one order: **commit the script → run it (red) → fix the tree → run it (green)**,
and records both runs. The obligation is **ADR 0041 D2** (*a check must demonstrate a catch*) plus upstream
`aixbdd-tmg` ADR 0006. **Decision: one round, guard + fixes (option A).**

## D2 — The rule: four arms, two hazards

| Arm | Rule | Class |
| --- | --- | --- |
| (a) | `n ≥ 4` **and** an adjacent run of ≥ 2 identical-type parameters | swap |
| (b) | an adjacent run of ≥ 3 identical-type parameters | swap (arity-independent) |
| (c) | an adjacent run of ≥ 2 `func`-typed parameters | **DI idiom** |
| (d) | a same-package named-struct literal with ≥ 2 unkeyed elements | swap at the call site |

**Exempt**: `context.Context` runs · `interface{}`/`any` runs · a lone trailing variadic.
**Rejected**: a bare parameter-count arm (a count is a smell, not a hazard; the earlier `≥6` draft was
wrong twice — today's tree measures 22–23 at `n ≥ 6`, and the `discover` it named is `n = 5`).

**Type identity is STRUCTURAL**: slices, channels, maps and func signatures are expanded, never flattened
to `"func"`. The flattened key produced **four** wrong measurements across the issue's drafts (once
creating reds, three times in re-verification). **Decision: structural key, runs flattened (`a, b, c
float64` is three parameters).**

## D3 — Arm (c) is not a swap rule (option 2: keep, re-derived)

Measured: **0** identical-signature `func` runs in `internal/**` (figure R) — every adjacent closure pair
in the five injected-closure lists differs in signature, so swapping them does not compile. Arm (c)
therefore claims **no** swap. It is kept on the DI rationale: *an injected-closure list belongs in a spec
struct* (`turnOptions` / `LoopSpec` / `ListingSpec` / `ToolSetSpec`). Red set: **40** (option 2) vs **35**
(option 1).

**Verdict against issue #210's `Open verification` §3: option 2.**

## D4 — Scan interface methods (top-level declarations only)

**4 of the 40** signature reds are interface methods — `render.Lines.{PayloadMeasured, PayloadEstimate,
Ready}` and `suggestions.WorkspaceSource.Entries`. They are the *contracts* that force implementations to
change: change only the implementation and the port still invites the swap. Scope: **top-level `TypeSpec`
interfaces**; the one inline interface type (`mcp/client.go`, `interface{ UseCallTimeout() }`) has zero
parameters and is out of scope by construction.

## D5 — Arm (d): the call site is the other half, and `go vet` cannot see it

`govet`'s `composites` analyser (in the `standard` lint set; `.golangci.yml` has **no** override) flags
unkeyed fields **only for struct types from another package** — verified empirically on a throwaway
two-package module (`go vet` flagged the cross-package literal, not the same-package one;
`golangci-lint` reported `govet: 1`). So a signature-only fix could leave
`ComputeCost(costInputs{4, 6, 3, 2})` compiling and swapping silently. **Decision: arm (d) in the same AST
walk, scanning `_test.go` too** (the same-package construction sites `vet` misses are largely tests).

**And arm (d) must attribute ELIDED literal types.** A literal's `Type` is nil when its container implies
it — `[]T{{1, 2}}`, `map[K]T{"k": {1, 2}}` — which is the **test-table shape**. The first implementation
matched only `Type.(*ast.Ident)` and skipped those, so its reported `L = 0` was **not** what the arm's
definition says it measures; the corrected walk (container-element expectation + whole-file struct
collection + local types) found **6** live sites on the pre-round tree. Caught in review, not by the
round's own witnesses.

**Verdict against issue #210's `Open verification` §5: implementable in one walk, and required.**

## D6 — The `n ≥ 4` boundary is kept, and its remainder is measured

| Remainder | Pre-round | Post-round |
| --- | --- | --- |
| interface methods at `n ≤ 3, run ≥ 2` | 6 | 6 |
| `FuncDecl`s at `n ≤ 3, run = 2` | 38 | 33 |
| `FuncDecl`s with an identical-type **result** run | 11 | 11 |
| type aliases (`type A = B`) | 0 | 0 |

`ActionLine(t, tool, arguments string)` is the same hazard as `Entries(ctx, dir, prefix string, limit
int)`; only arity differs. Note **`n` counts exempt parameters**: `Entries` is `n = 4` only because `ctx`
counts toward arity while being run-exempt — which is exactly why `ActionLine` (n=3) and `Entries` (n=4)
straddle the boundary. **Decision: do not widen; record with the measurement (ADR 0068 §Forward).**

## D7 — Enforcement: a build-tagged test, three commands, `-count=1`

Mirror `verify-architecture` exactly (`Makefile`), **all three** commands — `go vet -tags=quality` ·
`golangci-lint run --build-tags=quality` · `go test -count=1 -tags=quality -run
TestVerifyParameterShapes` — because the middle one lints the guard's own code (a draft quoted only two).

**`-count=1` is kept — but its rationale is NOT the architecture gate's, and the first form of this
paragraph got that wrong.** `tools/arch` shells a `go list` **child process**, which the test cache
genuinely cannot see; this gate reads the tree **in-process** (`filepath.WalkDir` + `parser.ParseFile`),
and `go help test` is explicit that the cache matches on exactly the files a test opens — so here the
cache **does** see the tree, and a violation reds on a plain `go test` run. `-count=1` is belt-and-braces
+ recipe parity. The inherited-verbatim version reached the ADR and `techstack.md` while contradicting
its own witness (whose runs re-ran every time); review caught it. **An inherited rationale is not a
verified one.**

**A `go/parser` walk, not a `go vet` analyser** — a **design property, not a catch**: it needs no
`go list`, no host build context, and is cheaper than `verify-architecture`. *(The earlier build-tag
justification was false: both telemetry reds live in `system_metrics_math.go`, which has no build
constraint; the four tagged files hold 13 functions and 0 reds.)*

## D8 — The script owns the numbers; the allow-list is a fail-on-stale ratchet

Per site the scan emits `file:line · name · arms · the type string it keyed on · n · longest run`; the
report prints every figure **with its label and universe**. `verify-params-report` is the reporting entry
(never fails); `verify-params` is the gate.

Clearing a violation: fix it, or add an allow-list entry **with a recorded reason**. The list is
**fail-on-stale**, and `-update-allowlist` **refuses to GROW** it (**ADR 0020 fold R-51-1**) — the
`tools/arch` D3 ratchet, restated for a hand-maintained list. **At delivery the list is empty** (option A).

## D9 — The refactor the guard forces

Named seam structs (the house idiom: `ToolSetSpec`, `LoopSpec`, `turnOptions`):

| Seam | Replaces |
| --- | --- |
| `render.ReadyStats` | 7 positional figures (3 adjacent `float64` + 3 adjacent `int`) |
| `render.PayloadStatus` | `tokens, budget int, mode, model string` and its estimated sibling |
| `sessionRef` | `homeDir, configPath string` (and `renderTurn`'s three-string head) |
| `cliDeps` | the **five** injected-closure lists (arm (c)) |
| `spinnerLine` / `loopSignals` | the frame's two strings + two ints / the two `chan struct{}` |
| `actionText` | the action line's two strings (the colour variant only — the plain one is `n = 3`) |
| `tokenBound` · `WorkspaceQuery` · `CostInputs` · `usageCall` | the remaining scalar/string runs |
| `retrySpec` · `mcpFieldRef` | a closure pair / a three-string triple |
| `treeSpec` · `captureTarget` · `capturedPipes` | the tree walk's cursor / the capture target / the pipe pair |
| `darwinMachVM` / `darwinSysctlVM` · `mcp.RemoteSpec` · `requestSpec` (×2) | the telemetry page counts / the MCP connection / the two wire bodies |

Plus the `internal/cli` split (1682 → 474 lines, same package ⇒ **zero import-graph movement** ⇒
`verify-architecture`'s header-only baseline untouched, which is what proves the slice legal).

**`requestBody` × issue #195**: `requestSpec` carries `MaxTokens` as a **named field with ownership
undecided**, which pre-empts nothing and actively helps #195 (a per-(provider, model) budget-field choice
is a *named* choice). **Preferred order: this round before #195.**

## D10 — Model impact (resolved, not assumed — ADR 0041)

| Model | Verdict |
| --- | --- |
| `docs/domain-model/tellme.modelith.yaml` (product) | **No change.** `usage-summed-for-the-ready-summary` concerns **values**, not parameter shapes (`internal/domain/model…:504`); the round changes no behaviour. |
| `docs/domain-model/quality.modelith.yaml` (process) | **MODIFY** — the `QualityPipeline` gate list gains `verify-params`, and the `QualityGate` examples name the parameter-shape gate. |

## D11 — Stale-link drive-by (issue #210 §7)

`docs/decisions/0056-…` and `0058-…` linked `0012-deterministic-hermetic-tests.md`, which does not exist
(the real file is `0012-hermetic-make-go-env.md`). `verify-adr-index` checks index membership and
duplicate numbers only, **never links**. **Decision: fix both** (a two-word edit beats a forward item).

---

## Residual risks

- **Syntactic type identity** can miss an alias-mediated swap — measured **0** aliases today (D6).
- **Result lists** stay unscanned — **11** measured (ADR 0068 RF-068-1).
- **Arm (d) allow-list keys** carry a line number, so an allow-listed literal would stale when unrelated
  lines move; the list is empty, so this is a property of the mechanism (RF-068-3).
- **Build-tag union**: the scan parses files **as written**, so two darwin-tagged files declaring the same
  symbol are both scanned (RF-068-4).
