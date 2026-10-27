# Tasks — round 098 `098-param-swap-guard`

**Owner**: `/axb-tasks` (RD) · **Input**: `spec.md` · `research.md` · `plan.md` · `truth-delta.md` ·
`specs/truth/**`

Execution order follows the analysis waves in `plan.md`. The **script lands first** (T001), so its red run
is the specification; every later task is arbitrated by the compiler or by the gate itself.

---

## Phase 1 — Setup

- [X] **T001** Commit the scan **first**: `tools/params/{doc.go,params_test.go,allowlist.txt}` — the
      `go/parser` walk, the structural type key, arms (a)–(d), the exemptions, the labelled-report
      printer, the allow-list reader/writer (fail-on-stale + refuse-to-grow), the module-root resolver,
      the enumeration assertion, and the three synthetic self-tests (`arms` · `unkeyed` · `allow-list`)
      with a name-set assertion. **No fix yet** — the tree is red on purpose.
      *Verification*: the gate reds with exactly the **46** sites its report prints (W1).

- [X] **T002** Run the report on the pre-round tree and record the labelled baseline (every figure with
      its label and universe) into `spec.md` / `research.md`. Reconcile every figure against issue #210's
      published table; where a post-fix figure moved (because the fix collapsed the shape), record the
      **post-fix** measurement too rather than smoothing the pre-fix one.

## Phase 2 — Foundational

- [X] **T003** **Wave 1 — the domain seam shapes.** `internal/domain/render`: `PayloadStatus`,
      `ReadyStats`, and the three `Lines` port methods; `internal/domain/llm`: `CostInputs` +
      `ComputeCost(p, in)`. Update `internal/ui`'s adapters (`metrics.go`, `status.go`,
      `render_ports.go`) keeping each `Format*` body's format string **verbatim** (only field access
      changes), and `internal/cli/call_renderer.go`'s three call sites.
      *Verification*: `go test ./internal/...` green; the four `domain/render` + three `ui/render_ports`
      reds are gone.

- [X] **T004** **Wave 2 — the application seam shapes.** `internal/app/suggestions.WorkspaceQuery`;
      `internal/cli`: `sessionRef` (replacing the `homeDir, configPath` pair at every red site),
      `cliDeps` (one bundle for the **five** injected-closure lists), `spinnerLine`, `loopSignals`,
      `actionText`, `tokenBound`, `usageCall`, `retrySpec`.
      *Verification*: the eight `internal/cli` reds + the two `suggestions` reds are gone.

- [X] **T005** Split `internal/cli/cli.go` (**1682** lines) into cohesive same-package files —
      `boot.go` · `turn.go` · `reporting.go` · `workspace.go` · `diagnostics.go` · `interactive.go` — by
      moving whole contiguous concern blocks, then `goimports -w` (the aliased imports re-added by hand).
      **The pre-existing `internal/cli/prompt.go` name collision is resolved explicitly**: the new
      interactive file is `interactive.go`, not `prompt.go`.
      *Verification*: largest production file **474** lines; `go build` + `go test ./internal/cli` green
      with **no behavioural edit**.

- [X] **T006** **Wave 3 — the infrastructure seam shapes.** `internal/config.mcpFieldRef`;
      gemini/openai `requestSpec` (28 test call lines, **signature-only**); `mcp.RemoteSpec` + the
      `ClientFactory` seam (both `NewRemoteClient`s, `discovery.go`, `lazyclient.go`, `di/mcp_factory.go`);
      `telemetry.darwinMachVM` / `darwinSysctlVM`; `tools.treeSpec` / `captureTarget` / `capturedPipes`.
      *Verification*: the remaining twelve reds are gone; `go test ./internal/...` green.

- [X] **T007** Key the two arm-(d) literals in `internal/agent/callhooks_test.go`.
      *Verification*: the gate's report shows **0** sites.

## Phase 3 — Test alignment & implementation (Red → Green → Refactor)

- [X] **T008** Wire the gate into `make verify`: `verify-params` (three commands, `-count=1` — the
      `verify-architecture` recipe shape), `verify-params-report`, `verify-params-update`.
      *Verification*: `make verify` OK with **twelve** members.

- [X] **T009** Truth: `specs/truth/techstack.md` — the *Parameter-shape gate* row + `verify-params` in
      the *Task runner* aggregate. Quality model: `quality.modelith.yaml` (the `QualityPipeline` list +
      the `QualityGate` examples), re-rendered. The product model resolved as **no change** (ADR 0041).
      *Verification*: `modelith-check` no drift; `modelith-lint` 0/0.

- [X] **T010** Record **ADR 0068** + its index row; fix the two dead `0012-deterministic-hermetic-tests`
      links in ADR 0056/0058.
      *Verification*: `verify-adr-index` consistent; no dead link in `docs/decisions/`.

- [X] **T011** **Regression** — full E2E contract.
      *Verification*: **334 scenarios · 2519 steps**, all green — byte-identical behaviour.

## Phase 4W — Witnesses (unobservable claims pinned)

| # | Claim | Run | Result |
| --- | --- | --- | --- |
| **W1** | The guard is **red first** (the pre-round tree) | the committed gate in a `git worktree` at `dev` `0231f64` | **FAIL — 46 sites**, every pre-fix figure reproduced: J 31 · J′ 4 · K 36 · K′ 4 · **L 6** · M 9 · N 6 · O 38 · P 11 · Q 0 · R 0 · I 69. *(First reported as 42 with **L 2**; the arm-(d) elision bug — see the fold below — hid 4 of the 6 literals.)* |
| **W2** | The round's tree is **clean** | the gate on this tree | **ok** — 0 sites, allow-list empty |
| **W3** | Arm (a)/(b) fire on a **leaf**, arity-independent | `a, b, c float64` added to `formatReadyColour` | **1 site**: `internal/ui/metrics.go:43 formatReadyColour arms=ab n=5 run=3 float64` |
| **W4** | The **non-scalar** run reds | `a, b, c io.ReadCloser` added to `abortCapture` | **1 site**: `…command.go:292 abortCapture arms=ab n=6 run=3 io.ReadCloser` |
| **W5** | **Interface** scanning is live | `a, b, c int` added to `render.Lines.Ready` | **1 site**: `…render/ports.go:43 Ready arms=ab n=4 run=3 int` |
| **W6** | The allow-list is **fail-on-stale** | a synthetic entry for the already-fixed `ui/metrics.go:FormatReady` | **red**: `1 stale allow-list entr(ies) … remove them` |
| **W7** | `-count=1` / the test cache — **re-recorded by design, per review** | a cacheable green run, then a violation appended to `internal/domain/llm/pricing.go` | **CONFIRMED BY DESIGN, and the claim is the opposite of the one first written.** The cacheable run is `(cached)` on a repeat, and the violation **reds on a plain `go test` run** (no `-count=1`): this gate reads the tree **in-process**, and `go help test` keys the cache on exactly the files a test opens. So the `tools/arch` rationale (a `go list` **child process** the cache cannot see) does **not** apply here; `-count=1` is kept for **parity**, not because a cached green would be vacuous. *(The first version of this row called the re-runs "not reproducible … incidental (VCS stamping / timing)" — that was backwards, and it was contradicted by its own quoted result. Review item 1.)* |
| **W8** | **Behaviour identity** | full E2E contract | **334 scenarios · 2519 steps — unchanged and green** |
| **W9** | **No import edge moved**; no model drift; no races | `verify-architecture` baseline (**0** non-comment lines) · `modelith-check` no drift · `make test-race` | all clean |
| **W10** | The **call-site** half reds, and `go vet` cannot see it | `CostInputs{4, 6, 3, 2}` unkeyed in `pricing_test.go` (same package) | gate: **1 site** `…pricing_test.go:10 CostInputs arms=d 4 unkeyed elements`; `go vet` **0 findings**; `golangci-lint` (govet) **0 issues** — the G1 gap |

### Post-fix labelled figures (the script's numbers, not prose)

| # | Figure | Universe | Pre-round | Post-round |
| --- | --- | --- | --- | --- |
| A | `FuncDecl` total | `internal/**`, `_test.go` excluded | 656 | 658 |
| B | …zero-parameter | same | 199 | 199 |
| C | …`n ≥ 1` | same | 457 | 459 |
| C′ | …`n ≥ 1` — interface methods | top-level interfaces | 49 | 49 |
| D | `n ≥ 5` | `FuncDecl` | 43 | 19 |
| E | `n ≥ 6` | `FuncDecl` | 22 | 8 |
| F | `n ≥ 4` AND run ≥ 2 | `FuncDecl` | 30 | **0** |
| F′ | `n ≥ 4` AND run ≥ 2 — interface | interface methods | 4 | **0** |
| G | `n ≥ 5` AND run ≥ 2 | `FuncDecl` | 23 | **0** |
| H | longest run ≥ 3 | `FuncDecl` | 10 | **0** |
| I | any run ≥ 2 | `FuncDecl` | 69 | 33 |
| J | **RED** arms (a)/(b) | `FuncDecl` | 31 | **0** |
| J′ | **RED** arms (a)/(b) — interface | interface methods | 4 | **0** |
| K | **RED** arms (a)/(b)/(c) | `FuncDecl` | 36 | **0** |
| K′ | **RED** arms (a)/(b)/(c) — interface | interface methods | 4 | **0** |
| L | **RED** arm (d) | `internal/**` incl. `_test.go`, one row per `file:line:Type` | 6 | **0** |
| M | remainder — `_test.go` arms (a)/(b)/(c) | `*_test.go` (out of scope) | 9 | 5 |
| N | remainder — interface below the arity | `n ≤ 3`, run ≥ 2 | 6 | 6 |
| O | remainder — `FuncDecl` `n ≤ 3`, run = 2 | `FuncDecl` | 38 | 33 |
| P | remainder — RESULT run (G2) | `FuncDecl`, `error` exempt | 11 | 11 |
| Q | type aliases (G3) | `internal/**` | 0 | 0 |
| R | identical-signature func runs | `FuncDecl` | 0 | 0 |

**46 = 40 signature sites (36 `FuncDecl` + 4 interface) + 6 arm-(d) call-site literals.** The `FuncDecl` total
rose by 2 because two helper constructors were added (`cliDepsFrom`, `reportSpec` accessors were inlined
instead); `n ≥ 5`/`n ≥ 6` fell because the seam structs absorbed the wide positional heads.

---

## DoD checklist (issue #210)

| # | DoD | Status |
| --- | --- | --- |
| 1 | `verify-params` is a `make verify` member, **all three** recipe commands, `-count=1`, red pre-fix on exactly the committed script's output, green after | ✅ (W1 / W2 / T008) |
| 2 | All sites fixed (option A) — allow-list empty and fail-on-stale, witnessed | ✅ (T003–T007 / W6) |
| 3 | `cli.go` split into same-package files; the `prompt.go` collision resolved explicitly; no file > ~600 lines; baseline header-only | ✅ (T005: 1682 → 474; W9) |
| 4 | **Both** domain port contracts changed; import graph unchanged; the model question resolved | ✅ (T003 / T004 / T009) |
| 5 | The guard scans interface methods (top-level), witnessed as a run-of-3 mutation | ✅ (W5) |
| 6 | W3/W4/W5 are arity-independent (run-of-3) | ✅ |
| 7 | Unit-test edits are signature-only; `pricing_test.go`'s positional call + comment replaced by the keyed form; no seam struct constructed unkeyed | ✅ (T003 / T006 / W10) |
| 8 | `make verify` OK · E2E 334 / 2519 · `test-race` clean · `modelith-check` no drift · `modelith-lint` 0/0 · topology audit unchanged · `govulncheck` clean | ✅ (W8 / W9) |
| 9 | W1–W10 recorded with the runs quoted | ✅ (this table; **W7 re-recorded as confirmed-by-design** after review) |
| 10 | The scan script is committed first, emits arms + type string per site, prints every figure with its label and universe; its output matches every figure in the plan | ✅ (T001 / T002 / the figure table — **every figure re-measured from the script after the fold**) |
| 11 | One ADR recording the two hazards, the arm-(c) decision, arm (d), the interface scope, the `n ≥ 4` boundary with 6 + 33, results out of scope (11), syntactic identity (0 aliases), `n` counts exempt params, the allow-list design, the measurement-bug lesson; index row; techstack row (incl. `-count=1` + its `doc.go` origin); the quality model names the gate | ✅ (T009 / T010 / ADR 0068) |
| 12 | The two dead `0012-deterministic-hermetic-tests.md` links fixed or a forward item recorded | ✅ **fixed** (ADR 0056 / 0058) |

---

## PR-review fold (2026-10-27; reviewer comment [`6436228558`](https://github.com/gosharplite/tellme/pull/211#issuecomment-6436228558))

Six findings, **all accepted** — two of them (1 and 2) load-bearing, because both were claims the round had
recorded **without a witness that could fail**.

| # | Finding | Disposition |
| --- | --- | --- |
| **1** | The `-count=1` rationale is **false for this gate** — `go help test` says the cache matches on the files a test opens, and this gate reads them in-process; W7's own quoted runs re-ran every time, which is the documented behaviour, not a host quirk | **Accepted.** Reworded in **ADR 0068 D9**, `tools/params/doc.go`, the `Makefile` comment and the `specs/truth/techstack.md` row; **W7 re-recorded as confirmed-by-design** with the cacheable-run reproduction. `-count=1` is **kept** — as parity with `verify-architecture`, whose `go list` **child process** the cache genuinely cannot see. The inherited rationale is now explicitly marked as previously wrong. |
| **2** | Arm (d) **misses elided-type literals** (`[]T{{…}}`, `map[K]T{"k": {…}}`) and **function-local** struct types — with a live instance in the tree, so the reported `L = 0` was not what the arm's definition measures | **Accepted and fixed.** `unkeyedLits` now carries an **expected element name** down the tree and `declaredStructs` walks the **whole file** (local types count); generic instantiations resolve to their base identifier. Re-measured: **L = 6** pre-round (was reported 2 → **W1 is 46, not 42**), **0** after keying the 4 surviving sites. ADR **D4** records the class; **RF-068-3b** records what a parser walk still cannot attribute (a literal implied by a func parameter/field). |
| **3** | Allow-list key `file:name` collides across receivers (`writer.go` has two `Execute`s, `mcp/client.go` two `Close`s) | **Accepted.** `decl.key()` is now **`file:Recv.Method`**, and interface methods are qualified by their declaring interface. The report prints the qualified name too (`Lines.Ready`, `Spinner.loop`). |
| **4** | Figure **Q**'s universe label claims `_test.go excluded` but the count included test files | **Accepted.** `scanTree` counts aliases for production files only, matching the label. |
| **5** | The gate **does not satisfy its own arms** (`newDecl` is n = 8 with adjacent identical-type runs; `figures()` is a table of unkeyed literals) | **Accepted.** `newDecl` takes a named **`declSpec`**; every `figure{…}` is **keyed**; and a new **`self-application` self-test** scans `tools/params` itself and asserts 0 sites — a pin, not a widened walk (the governed scope stays `internal/**`). |
| **6** | Four smaller items: `actionText` is an appeasement; `treeSpec.descend` is a **new** adjacent-string pair; `dispatchDeps.usageReport` is dead; `writeAllowlist` always prints "Empty at round 098"; `exemptKey` treats `error` as a parameter exemption | **All accepted.** `actionText`'s comment now says plainly that the **public** `FormatToolAction` / `ActionLine` keep the pair (n = 3, below the boundary); `treeSpec.descend`'s comment says it is a pair this refactor **introduced** and where it is recorded; the dead field is gone; the allow-list header is conditional; `error` is a **result-only** exemption now (parameter pairs of `error` red like any other), and the `param.exempt` comment matches. |

### Post-fold labels (re-measured from the script, not patched in prose)

| # | Figure | Pre-round | Post-round |
| --- | --- | --- | --- |
| J / J′ | RED arms (a)/(b) | 31 / 4 | **0 / 0** |
| K / K′ | RED arms (a)/(b)/(c) | 36 / 4 | **0 / 0** |
| **L** | RED arm (d) — one row per `file:line:Type` | **6** | **0** |
| **total** | the red set | **46** | **0** |

The pre-round arm-(d) rows the first implementation missed: `internal/cli/deps_boundary_test.go:60`
(`violation`) · `internal/cli/retry_gateway_test.go:127` (`note`) · `…:133` (`[]note{{2,3,0},{3,3,0}}`) ·
`internal/infrastructure/tools/command_test.go:26` (`res`) — plus the two explicit ones
(`agent/callhooks_test.go:38/42`) the first version did catch.

### The lesson this fold adds

**The round's own thesis, one layer up.** It exists to make an unfalsifiable class (the compile-blind swap)
mechanically visible; its fold found **two claims of its own that no witness could fail** — an inherited
`-count=1` rationale and an arm whose definition was wider than its implementation. Both had *quoted
evidence that contradicted them*. The remedy is the same one the gate applies to code: name the claim, then
find the run that can red it.

### Post-fold gates

`make verify` **OK** (twelve members) · gate **0 sites** · self-application **0** · E2E **334 / 2519
unchanged** · `test-race` clean · `modelith-check` no drift · `modelith-lint` 0/0 · `verify-adr-index`
consistent · `golangci-lint` 0 issues · `govulncheck` clean.

---

## Second fold — re-review residuals (2026-10-27; reviewer comment [`6436423322`](https://github.com/gosharplite/tellme/pull/211#issuecomment-6436423322))

The re-review confirmed the first fold (corrected W1 = **46**, all six items disposed) and raised **two
residuals** plus two nits. All four accepted; both residuals were fixed rather than recorded, because both
are *parser-attributable* — recording them would have left the arm's stated coverage wider than its
implementation for a second time.

| # | Finding | Disposition |
| --- | --- | --- |
| **R1** | The **self-application pin is vacuous for arms (a)–(c)**: every line of `tools/params` lives in `params_test.go`, and `redSites()` skips test declarations — so the pin could only ever see arm (d) on the gate's own source, and would **not** have redded on the old positional `newDecl` it was cited as proving | **Accepted, fixed.** `scanSubtree` now takes a named **`scanTarget`** (a `root, sub string` pair would itself be a hazard) with a `pinProduction` flag; the pin sets it, so `_test.go` is treated as **production** there. **Verified with the reviewer's own probe:** appending `func zzSelfProbe(a, b, c int, d string) {}` + an unkeyed `site{…}` literal now reds **both** — `zzSelfProbe arms=ab n=4 run=3 int` and `site arms=d`. |
| **R2** | Arm (d) still misses two **parser-attributable** shapes: a **nested** container (`[][]CostInputs{{{1, 2, 3, 4}}}`) and a **struct-typed map KEY** (`map[CostInputs]int{{1, 2, 3, 4}: 1}`) — neither listed in RF-068-3b | **Accepted, fixed.** The walk now resolves each literal's type expression in **one prepass** (`effectiveLitTypes`) and pushes the **expression**, so containers unwrap one level per elided literal; a second walk (`keyLits`) attributes a `KeyValueExpr` key to the map's **key** type, which the element walk never reaches. **Verified with the reviewer's exact probe:** zz1–zz6 all red (including `zz5` nested and `zz6` map key) and `zz7` (field-keyed) correctly does not. Seven shapes + two negatives are now pinned in `selfTestUnkeyed`, and a new `selfTestUnkeyedLocalStruct` pins the function-local `type note` table shape. |
| **N1** | `specs/truth/techstack.md`: missing full stop after "(`go help test`)" | **Accepted, fixed.** |
| **N2** | The arm-(d) row collapses same-identity literals, so its `shape` printed only the first literal's element count | **Accepted, fixed.** Rows aggregate into `2 literals, 3-3 unkeyed elements` when they stand for more than one; pinned by the local-struct self-test's shape assertion. |

**Re-measured after the second fold** (the script, not prose): pre-round `dev` `0231f64` → **46 sites**
(J 31 · J′ 4 · K 36 · K′ 4 · **L 6** · M 9 · N 6 · O 38 · P 11 · Q 0 · R 0 · I 69) — **unchanged**, which
is itself the useful result: the two newly attributed shapes do not occur in `internal/**`, so the extra
coverage is confirmed free of figure churn. This tree → **0 sites**, allow-list empty.

**RF-068-3b narrowed** to what a parser walk genuinely cannot reach — a literal whose type is implied by a
**func parameter/return** or by a **struct field** — with the measurement that neither occurs here.

**Also recorded in the ADR** (D4 + a new paragraph): the elision class and its two follow-on shapes, and
the pin-scope defect, because "the witness for the refactor was itself vacuous" is the same class the
round exists to eliminate.

---

## Third fold — re-review of the second fold (2026-10-27; reviewer comment [`6436583950`](https://github.com/gosharplite/tellme/pull/211#issuecomment-6436583950))

Three findings plus a nit; all accepted and fixed. Two are code, one is the **forward item**, and the
forward item is the one that matters: it was a claim with no compiler behind it.

| # | Finding | Disposition |
| --- | --- | --- |
| **1** | **RF-068-3b records shapes that are not valid Go.** It listed a literal elided through a **func parameter/return** (`f({1, 2})`) and through a **struct field** (`Outer{Inner: {…}}`) as "needs a type checker" gaps | **Accepted — the item described a limit that does not exist.** Verified with the compiler: `syntax error: unexpected {` and `missing type in composite literal`. Go permits elision **only** in an array/slice/map literal's element or key position. RF-068-3b is **rewritten** to the one real limit — a **cross-package** named container (`pkg.Table{{…}}`), which needs another package's AST — **with the measurement** that `internal/**` declares no same-package named array/slice/map type at all today. The ADR also records the correction, because a forward item is a claim like any other. |
| **2** | A **same-package named container** is opaque: `elemType` unwraps `ArrayType`/`MapType` only, so `type costTable []CostInputs; costTable{{1,2,3,4}}` was missed | **Accepted, fixed** — and widened: `declaredContainers` collects the whole file's named array/slice/map types, and `resolveContainer` follows them **and binds generic instantiations** (`type gen[T any] []T`, used as `gen[pair]{{…}}`) by substituting each type parameter. Verified on the reviewer's probe (`zz8`) and on the generic form; the negative (`other.Table{{…}}`) stays silent. |
| **3** | **Double count** when a map's key and value are the same struct: `map[T]T{{…}: {…}}` reported `3 literals` for two | **Accepted, fixed at the root cause.** `effectiveLitTypes` is now a **worklist** that assigns a map's **key** position the key type and its value position the element type — one attribution per literal — so the separate `keyLits` walk is **deleted**. The probe now reports `2 literals, 4-4 unkeyed elements`, pinned by an assertion. |
| **N** | `elementLits`/`keyLits` iterated a map and `redSites` sorted by `(file, line)` only, so two types on one line were nondeterministic (report **and** allow-list file) | **Accepted, fixed.** The sort gains a `name` tiebreak; with `keyLits` gone there is one producer anyway. |

**Re-measured after the third fold** (`dev` `0231f64`, this tree): pre-round **46 sites** (L 6) —
**unchanged**, and this tree **0 sites**. The three newly attributed shapes have zero instances in
`internal/**`, so the widened coverage costs no figure churn — the same result as the second fold, now
for the last elision position Go allows.

**Self-tests now pin twelve shapes and three negatives** (`selfTestUnkeyed`) plus the function-local
`type note` table (`selfTestUnkeyedLocalStruct`), and the self-application pin remains live for arms
(a)–(c).

---

## Fourth fold — re-review of the third fold (2026-10-27; reviewer comment [`6436725202`](https://github.com/gosharplite/tellme/pull/211#issuecomment-6436725202))

The re-review recommended merging and raised one wording point and one over-claim. Both accepted; the
over-claim was fixed rather than softened, because the three shapes are parser-attributable.

| # | Finding | Disposition |
| --- | --- | --- |
| **1** | D4 and Consequences say **both** refuted shapes are *syntax errors*. Only `f({1, 2})` is — `Outer{Inner: {1, 2}}` **parses** and is rejected by the **type checker** | **Accepted.** Verified with `go/parser` directly: `f({1,2})` → `expected operand, found '{'` (parser); `Outer{Inner: {1,2}}` → parse **OK**, `missing type in composite literal` (types2). The ADR now says "a syntax error and a type-check error", because its own point is precision about what refuted what. |
| **2** | RF-068-3b says "the ONE shape", but three same-package forms are still unattributed — `[]*table{{{…}}}` (pointer to a named container), `costGrid[T] [][]T` (generic body is itself a container), `costAlias[T] costTableGen[T]` (chained generic) | **Accepted, fixed.** `resolveContainer` strips `StarExpr` (the `&T` elision for a pointer element type); `substitute` now walks type **expressions** recursively (so `[][]T` becomes `[][]pair`); the substitution is **composed** across hops instead of overwritten; and `declaredContainers` collects a named type whose **body is itself a named/instantiated type**, which is what makes a chained generic resolvable. All three probes now red; the cross-package negative still stays silent. Pinned in `selfTestUnkeyed` (which now covers fifteen shapes and three negatives). RF-068-3b was corrected — again — to the genuinely single limit, with its history recorded. |

**Re-measured after the fourth fold** (`dev` `0231f64`): **46 sites**, L 6, arms `a=23 ab=10 ac=1 b=1 c=5 d=6`
— **unchanged**, as expected: `internal/**` declares no same-package named container type at all, so the
three newly attributed forms have zero instances on this tree.

**The lesson, now recorded twice in this round:** §Forward is the last place anyone looks for an
unwitnessed claim, and it has needed a compiler behind it twice. Every remaining limit in RF-068-3b now
carries a measurement, and the shapes it excludes are pinned as **negatives** in the gate's own tests.

---

## Fifth fold — re-review of the fourth fold (2026-10-27; reviewer comment [`6436859675`](https://github.com/gosharplite/tellme/pull/211#issuecomment-6436859675))

One finding, accepted — and it is the round's own thesis turned on the **witness** side, which is where it
had not been pointed yet.

| # | Finding | Disposition |
| --- | --- | --- |
| **1** | **The pin for the container-bodied generic was vacuous.** `selfTestUnkeyed` declared `type grid[T any] [][]pair` — a body that never mentions `T`, so the recursive `substitute` the fold had just added was never on the path. Reverting `substitute`'s `*ast.ArrayType` case left the self-test **green** | **Accepted, fixed.** The synthetic body is `[][]T`. Reproduced the reviewer's mutation table exactly, then went one step further and mutation-tested **all three** new pins individually — each now reds only its own row: revert the `ArrayType` case ⇒ row **35** (grid) missing; remove `stripType`'s `StarExpr` case ⇒ row **34** (pointer) missing; overwrite instead of compose the substitution ⇒ row **36** (chain) missing. |
| **N** | `declaredContainers` records every non-struct, non-interface `TypeSpec`, so `type n int` / `type F func()` / aliases land in the map too | **Accepted as a comment, not a change.** Verified the reviewer's reasoning holds (resolution stops at a non-container body and `elemType` yields nil; struct bodies are skipped so the two maps are disjoint) and documented *why* the map is deliberately wider — narrowing it would buy nothing and would need a second pass for a chained generic. |

**The lesson, recorded in ADR 0068 §Consequences:** *a pin must be mutated against the code it claims to
pin.* "The self-test covers this shape" is itself a claim — and this one was false while the code was
right, which is the **first** time in five folds that the code was correct and the evidence was not.

**Figures unchanged** (`dev` `0231f64`: 46 sites, L 6; this tree: 0) — the fold is evidence-quality only.
