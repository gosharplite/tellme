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
      *Verification*: the gate reds with exactly the **42** sites its report prints (W1).

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
| **W1** | The guard is **red first** (the pre-round tree) | the committed gate in a `git worktree` at `dev` `0231f64` | **FAIL — 42 sites**, every pre-fix figure reproduced: J 31 · J′ 4 · K 36 · K′ 4 · L 2 · M 9 · N 6 · O 38 · P 11 · Q 0 · R 0 · I 69 |
| **W2** | The round's tree is **clean** | the gate on this tree | **ok** — 0 sites, allow-list empty |
| **W3** | Arm (a)/(b) fire on a **leaf**, arity-independent | `a, b, c float64` added to `formatReadyColour` | **1 site**: `internal/ui/metrics.go:43 formatReadyColour arms=ab n=5 run=3 float64` |
| **W4** | The **non-scalar** run reds | `a, b, c io.ReadCloser` added to `abortCapture` | **1 site**: `…command.go:292 abortCapture arms=ab n=6 run=3 io.ReadCloser` |
| **W5** | **Interface** scanning is live | `a, b, c int` added to `render.Lines.Ready` | **1 site**: `…render/ports.go:43 Ready arms=ab n=4 run=3 int` |
| **W6** | The allow-list is **fail-on-stale** | a synthetic entry for the already-fixed `ui/metrics.go:FormatReady` | **red**: `1 stale allow-list entr(ies) … remove them` |
| **W7** | `-count=1` is required (the cache cannot see the tree) | a violation added after a cacheable green run | **Not reproducible on this host, recorded not smoothed** — Go re-ran the gate on every mutation shape tried (a signature mutation, a comment-only edit in another package, a *committed* violation, with and without `-buildvcs=false`), so a stale green was not served. The arm is **retained** because the gate's verdict reads files **outside its own package**, which the test cache's key does not model — the property is about what the cache is keyed on, and this host's invalidation is incidental (VCS stamping / timing), not a guarantee. `-count=1` therefore stays in the recipe (DoD 1) as the specified discipline, with this negative result recorded. |
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
| L | **RED** arm (d) | `internal/**` incl. `_test.go` | 2 | **0** |
| M | remainder — `_test.go` arms (a)/(b)/(c) | `*_test.go` (out of scope) | 9 | 5 |
| N | remainder — interface below the arity | `n ≤ 3`, run ≥ 2 | 6 | 6 |
| O | remainder — `FuncDecl` `n ≤ 3`, run = 2 | `FuncDecl` | 38 | 33 |
| P | remainder — RESULT run (G2) | `FuncDecl`, `error` exempt | 11 | 11 |
| Q | type aliases (G3) | `internal/**` | 0 | 0 |
| R | identical-signature func runs | `FuncDecl` | 0 | 0 |

**42 = 40 signature sites (36 `FuncDecl` + 4 interface) + 2 arm-(d) literals.** The `FuncDecl` total
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
| 9 | W1–W10 recorded with the runs quoted | ✅ (this table; **W7 recorded as not-reproducible, not smoothed**) |
| 10 | The scan script is committed first, emits arms + type string per site, prints every figure with its label and universe; its output matches every figure in the plan | ✅ (T001 / T002 / the figure table) |
| 11 | One ADR recording the two hazards, the arm-(c) decision, arm (d), the interface scope, the `n ≥ 4` boundary with 6 + 33, results out of scope (11), syntactic identity (0 aliases), `n` counts exempt params, the allow-list design, the measurement-bug lesson; index row; techstack row (incl. `-count=1` + its `doc.go` origin); the quality model names the gate | ✅ (T009 / T010 / ADR 0068) |
| 12 | The two dead `0012-deterministic-hermetic-tests.md` links fixed or a forward item recorded | ✅ **fixed** (ADR 0056 / 0058) |
