# ADR 0068 — A gate for compile-blind argument swaps (and the refactor it forces)

- **Status:** Accepted
- **Date:** 2026-10-27
- **Deciders:** tellme owner (operator directive, 2026-10-27: *"Start a new aixbdd round, the goal is to close issue 210"*)
- **Related:** **ADR 0011** (the layer-discipline gate — the `tools/arch` shape this gate mirrors, incl. its
  **D3 fail-on-stale** ratchet) · **ADR 0020 fold R-51-1** (refuse-to-grow) · **ADR 0012** (the hermetic
  `make` Go environment this gate inherits) · **ADR 0041 D2** (a check must demonstrate a catch) ·
  **ADR 0064** (the advisory-vs-gate distinction this round keeps out of the new gate) ·
  **ADR 0039**/**ADR 0019** (the named-spec-struct house idiom: `ToolSetSpec`, `LoopSpec`) ·
  [issue #210](https://github.com/gosharplite/tellme/issues/210) (which supersedes #206–#209) ·
  upstream `aixbdd-tmg` ADR 0006 (*no normative clause reaches truth unwitnessed*) ·
  `tools/params/` · `internal/cli/` · `internal/ui/` · `internal/domain/render/`

## Context

`tellme`'s quality machine is dense: eleven `make verify` members, a 334-scenario / 2519-step
executable E2E contract, ~16.9k test LOC against ~15.0k production LOC, a green advisory
`modelith-drift` and a clean advisory `dead-code`. What **no** gate could see is the **compile-blind
argument swap**: two adjacent parameters of an **identical type**, where swapping them at a call site
**still compiles**. `ComputeCost(p, 4, 6, 3, 2)` carried a comment to say what it meant; `renderTurn`'s
head was three adjacent `string`s; `NewRemoteClient(ctx, url, authorization string, …)` would happily
connect to a URL made of a bearer token.

`internal/cli/cli.go` had also accreted to **1682 lines · 58 functions · 69.6 %** of its package.

**The five drafts behind issue #210 are the reason this ADR exists in this shape.** Every wrong number
came from *memory*; every right one came from a *script*. The coarse type normaliser (flattening every
`*ast.FuncType` to `"func"`) surfaced **four** times — once creating reds, three times in
re-verification. So the round is **guard-first**: a mechanical scan that **reds on today's tree** and
goes green only because the tree was fixed. A refactor-only round has no falsifiable carrier — the
obligation carried by **ADR 0041 D2** ("a check must demonstrate a catch") and by upstream
`aixbdd-tmg` ADR 0006.

## Decision

### D1 — Four arms, two hazards

**RED** when a named production func/method under `internal/**` (`_test.go` excluded) satisfies **any** of:

| Arm | Rule | Hazard |
| --- | --- | --- |
| **(a)** | `params ≥ 4` **AND** an adjacent run of **≥ 2** parameters of an **identical type** | swap |
| **(b)** | an adjacent run of **≥ 3** parameters of an identical type (arity-independent) | swap |
| **(c)** | an adjacent run of **≥ 2 `func`-typed** parameters, **regardless of signature** | **DI idiom** |
| **(d)** | a same-package **named struct** `CompositeLit` with **≥ 2 unkeyed** elements | swap **at the call site** |

**Exempt:** runs of `context.Context`; runs of `interface{}` / `any`; a lone trailing variadic (the
idiomatic collector, never a swappable pair). **No bare parameter-count arm** — a count is a smell, not a
hazard, and the earlier `≥6 params` draft that claimed one was wrong twice (today's tree had 22–23 at
`n ≥ 6`, and the `discover` it named is `n = 5`).

### D2 — Arm (c) is NOT a swap rule (option 2, kept and re-derived)

**Verified:** there is **no** identical-signature `func` run anywhere in `internal/**` (figure R = 0) —
every adjacent closure pair in the five injected-closure lists differs in signature, so swapping them
does **not** compile. Arm (c) therefore cannot claim the swap rationale.

It is kept on its **own** rationale: *an injected-closure list is a DI surface that belongs in a spec
struct — the `turnOptions` / `LoopSpec` / `ListingSpec` / `ToolSetSpec` house idiom*. Option 1 (drop the
arm) would have left the five lists in place; option 2 reds them and the round folds all five into **one**
`cliDeps` bundle. Red set: **40** (36 `FuncDecl` + 4 interface) with option 2, **35** with option 1.

### D3 — The guard MUST scan top-level interface methods

**4 of the 40** signature reds are **interface methods**, and they are the *contracts* that force the
implementations to change: `render.Lines.{PayloadMeasured, PayloadEstimate, Ready}` and
`suggestions.WorkspaceSource.Entries`. Scope is precisely **top-level `TypeSpec` interfaces** — one inline
interface type exists (`internal/infrastructure/mcp/client.go`: `interface{ UseCallTimeout() }`, zero
params) and is out of scope by construction.

### D4 — Arm (d): the call site is half the hazard, and `go vet` cannot see it

`govet`'s `composites` analyser — in the `standard` lint set, with **no** override in `.golangci.yml` —
flags unkeyed fields **only for struct types from another package** (verified empirically on a
throwaway two-package module). So a fix that only changes signatures could leave
`ComputeCost(costInputs{4, 6, 3, 2})` **compiling and swapping silently**, re-creating the hazard one
layer down. Hence arm (d) in the same walk, and it scans `_test.go` too (the same-package construction
sites `vet` misses are largely tests).

**Arm (d) must attribute ELIDED literal types.** A literal's `Type` is `nil` when the type is implied by
its container — `[]T{{1, 2}}`, `map[K]T{"k": {1, 2}}` — which is the **table shape**, the most common
way a same-package unkeyed literal appears in a test. A first implementation matched only
`Type.(*ast.Ident)` and silently skipped those; the review caught it with a live instance in the tree.

The walk resolves the type expression every literal denotes (its own `Type`, else the type it inherits
from its position) in a **worklist from the explicitly-typed literals outward**, and it carries the
**expression** rather than a name — so a **nested** container resolves one level per elided literal
(`[][]T{{{1, 2}}}`). It assigns a map's **key** position the map's **key** type and its value position the
element type, which is **one attribution per literal**: an earlier version gave every child the value type
(so a key was attributed twice once key and value were the same struct) and needed a second walk to
recover map keys at all.

Named structs are collected by walking the **whole file**, so a **function-local** `type note struct{…}`
declared beside its literal counts. **Same-package named container types** are collected the same way and
resolved through — `type table []pair` makes `table{{1, 2}}` attributable, and a **generic** container
(`type gen[T any] []T`, used as `gen[pair]{{1, 2}}`) is bound by substituting each type parameter with its
argument. Generic *struct* instantiations (`pair[int]{1, 2}`) resolve to their base identifier.

Go permits type elision **only** in an array/slice/map literal's element or key position, so those are the
only positions the walk has to reach. A first version of §Forward listed two others as gaps, and **neither
can occur in a compiled tree** — but they are refuted by *different* stages, and the distinction is the
point: `f({1, 2})` is a **syntax error** (`expected operand, found '{'` — `go/parser` rejects it), while
`Outer{Inner: {1, 2}}` **parses fine** and is rejected by the **type checker**
(`missing type in composite literal`). Every position that *is* valid is pinned in the arm-(d) self-tests:
the elided element/key/value shapes, the named and generic container forms, a **pointer to** a named
container (`[]*table{{…}}` — the `&T` elision for a pointer element type), a generic whose **body is itself
a container** (`type grid[T any] [][]T` — the body must mention the parameter, or the substitution the pin
exists for never runs), and a **chained** generic (`type chain[T any] gen[T]`); plus the
negatives — a field-keyed literal, a cross-package named struct (`time.Time{}`) and a cross-package named
container (`other.Table{{…}}`).

**The self-application pin must see arms (a)–(c).** The gate is scanned by itself, and every line of
`tools/params` lives in a `_test.go` file, which `redSites` skips — so the first pin could only ever
observe arm (d) there, and would **not** have redded on the old positional `newDecl` it was cited as
proving (review residual 1). The pin therefore scans its own subtree with the test file treated as
**production**, which makes it a real carrier for the very refactor it justifies.

### D5 — The `n ≥ 4` boundary is deliberate and its remainder is measured

Below the boundary sit **6** interface methods (`ToolLineRenderer.{EngineLine,ActionLine,ResultLine}` ·
`durableFS.Rename` · `cacheFS.{CreateTemp,Rename}`) and **38** `FuncDecl`s at `n ≤ 3, run = 2`
(post-refactor: 6 and 33). `ActionLine(t, tool, arguments string)` is the **same** hazard as
`Entries(ctx, dir, prefix string, limit int)`; only arity differs. **Not widened here** — recorded so the
next reader does not reopen the threshold from scratch. Note that **`n` counts exempt params**:
`Entries` is `n = 4` only because `ctx` counts toward arity while being run-exempt — that is exactly why
`ActionLine` (n=3) and `Entries` (n=4) sit either side of the boundary.

Two **honesty notes** about this round's own output: `actionText` folds only the private colour helper —
the pair `tool, arguments string` **remains** on the public `FormatToolAction` and
`ToolLineRenderer.ActionLine`, which are `n = 3` and therefore below the boundary (so the hazard is
*recorded*, not removed, at those two surfaces); and `treeSpec.descend(name, extraIndent string)` is an
adjacent identical-type pair the refactor **introduced**, kept because it is `n = 2` and therefore in the
measured remainder.

### D6 — Result lists are the same class and are out of scope

`return mem, cpu` compiles just as silently. Measured: **11** `FuncDecl`s return an adjacent
identical-type run (excl. `error`) — `splitQuery` · `expandFieldWarning` · `ResolveAuthorization` ·
`fromTokenSource` · `parseFrontmatter` · `Sample` ×3 · `machCPUTicks` · `treeConnectors` ·
`sampleResources`. Recorded, not fixed.

### D7 — Type identity is syntactic (`go/parser`), and that is a design property

`a string, b LocalAlias` (where `type LocalAlias = string`) would not red although it swaps. Measured:
**0 type aliases** in `internal/**` production code — no false negative on this tree. The parser-walk
choice itself is also a **design property, not a catch**: it needs no `go list`, no host build context,
and is cheaper than `verify-architecture`. *(A prior draft justified it by claiming an analyser "would
silently skip two reds" because of the `system_metrics_*` build tags — **factually wrong**: both telemetry
reds live in `system_metrics_math.go`, which has no build constraint. The four tagged files hold 13
functions and zero reds.)*

### D8 — The allow-list is a fail-on-stale ratchet that refuses to grow

A violation's allow-list identity is `file:name`, or **`file:Recv.Method`** when the declaration is
qualified — the qualifier is **required**, not cosmetic: `writer.go` declares `Execute` on two receivers
and `mcp/client.go` declares `Close` on two, so an unqualified key would let one entry silently allow the
other. (Interface methods are qualified by their declaring interface for the same reason.) Arm-(d) sites
have no name, so they key on `file:line:Type`, and two literals of one type on one line collapse to one
row.

A violation is cleared **only** by fixing it or by an allow-list entry carrying a **recorded reason**.
The list is **fail-on-stale** (an entry that no longer matches a violation fails the gate) and
`-update-allowlist` **refuses to GROW** it (**ADR 0020 fold R-51-1**). At delivery the list is **empty**
— option A: every site fixed. There is **no generated baseline**.

### D9 — Enforcement shape: a build-tagged test, `-count=1`, three commands

`tools/params` mirrors `tools/arch`'s **shape**: a `//go:build quality` test (never in the default
`go test ./...`), a `verify-params` member of `make verify` running **three** commands (`go vet
-tags=quality` · `golangci-lint run --build-tags=quality` · `go test -count=1 -tags=quality -run …`),
honouring ADR 0012's hermetic invocation boundary like every other member.

**`-count=1` is kept, but its rationale is NOT inherited — and the distinction is the point.** The
architecture gate shells a `go list` **child process**, which the Go test cache genuinely cannot see, so
there `-count=1` is the difference between a witness and a vacuous pass. **This** gate reads the tree
**in-process** (`filepath.WalkDir` → `os.ReadDir`; `parser.ParseFile` → `os.ReadFile`), and `go help
test` is explicit: *"Tests that open files within the package's module or that consult environment
variables only match future runs in which the files and environment variables are unchanged."* The
opened files and directories are therefore **part of the cache key**, and a violation reds on a plain
`go test` run. `-count=1` here is belt-and-braces + recipe parity.

This clause is recorded in this form because its **first** form was wrong: the rationale was inherited
verbatim from `tools/arch/doc.go`, landed in an Accepted ADR and in `techstack.md`, and **contradicted
its own witness** — W7's mutation runs re-ran on every attempt, which is the documented behaviour, not a
host quirk. The round-098 review caught it. An inherited rationale is not a verified one; that is the
same class as the round's own thesis (a claim needs a carrier that can fail).

### D10 — The script owns the numbers

Per red site the scan emits `file:line · name · arms-fired (a/b/c/d) · the type string it keyed on · n ·
longest run`, and it prints **every** baseline figure **with its label and its universe**. `make
verify-params-report` is the reporting entry (`TestParameterShapeReport`, never fails); `verify-params`
is the gate. This is what makes "fix all X sites" a `diff` rather than a re-argument.

### D11 — The refactor the guard forces (all 40 signature sites + 6 call-site literals, 46 total)

Named seam structs, following the house idiom: `render.ReadyStats` · `render.PayloadStatus` ·
`turnOptions` (pre-existing) · `sessionRef` (`homeDir, configPath`) · `cliDeps` (the five injected-closure
lists) · `spinnerLine` · `loopSignals` · `actionText` · `tokenBound` · `WorkspaceQuery` · `CostInputs` ·
`usageCall` · `retrySpec` · `mcpFieldRef` · `treeSpec` · `captureTarget` · `capturedPipes` ·
`darwinMachVM` / `darwinSysctlVM` · `mcp.RemoteSpec` (both `NewRemoteClient`s and the `ClientFactory`
seam) · gemini/openai `requestSpec`. Plus `internal/cli/cli.go` split into cohesive **same-package**
files (`boot.go` · `turn.go` · `reporting.go` · `workspace.go` · `diagnostics.go` · `interactive.go`;
the pre-existing `internal/cli/prompt.go` name collision resolved by naming the new file
`interactive.go`). Same package ⇒ **zero import-graph movement** ⇒ `verify-architecture`'s header-only
baseline is untouched, which is what **proves** the slice legal.

## Alternatives considered

- **A refactor-only round.** Rejected: no falsifiable carrier — nothing would red if the hazard returned.
- **Guard-only first, fix later.** Rejected: a guard green before it means anything is not a witness.
- **Hot seams only** (`internal/{cli,ui}`, allow-list the rest). Rejected: ships an allow-list debt on day
  one and leaves the two `requestBody` hazards in the wire layer.
- **A `go vet` analyser instead of a parser walk.** Rejected: needs `go list` + host build context; the
  parser walk is cheaper and the build-tag worry that once justified this was **false** (D7).
- **A bare parameter-count arm.** Rejected: a count is not a hazard (D1).
- **A generator baseline** (the `tools/arch` shape). Rejected for this gate: the endpoint is **zero**, and
  a generator would launder exactly the growth the round exists to prevent.

## Consequences

- **Positive.** The hazard class becomes mechanically visible: a new adjacent identical-type run, a new
  injected-closure list, or a new unkeyed same-package literal **reds the gate**. `internal/cli`'s largest
  production file drops from **1682** to **474** lines. Two domain port contracts (`render.Lines`,
  `suggestions.WorkspaceSource`) got honest shapes, which is also what let `CostInputs` replace
  `ComputeCost(p, 4, 6, 3, 2) // miss 4, hit 6, …`.
- **Negative / accepted.** Syntactic type identity (D7) can miss an alias-mediated swap; result lists
  (D6) stay unscanned; the `n ≥ 4` boundary (D5) leaves a measured remainder. Each is recorded with its
  measurement rather than smoothed. The allow-list mechanism exists and is empty — a future round may use
  it, at the cost of a recorded reason.
- **The PR review corrected two things the round had recorded wrongly, and both corrections are the
  ADR's own thesis applied to itself.** (1) The `-count=1` rationale was **inherited** from
  `tools/arch/doc.go` and is **false here** (D9): the architecture gate shells a `go list` child process
  the cache cannot see; this gate reads the tree in-process, which the cache **does** key. The clause
  had reached an Accepted ADR and `techstack.md` while **contradicting its own witness**, whose mutation
  runs re-ran every time. (2) Arm (d)'s first implementation **matched only explicitly-typed literals**
  (D4), so `[]T{{1, 2}}` — the test-table shape — was skipped, and the reported `L = 0` was not what the
  arm's definition says it measures; the corrected walk found **6** live sites on the pre-round tree.
  Both were caught by review, not by the round's own witnesses: a lesson one layer up from the one the
  round exists to enforce.
- **Positive, corrected figures.** The pre-round red set is **46** = 40 signature sites + **6** arm-(d)
  call-site literals (the round first reported 42, on the buggy arm (d)). The re-review then found **two
  more** elided shapes (`[][]T{{{…}}}`, `map[T]int{{…}: 1}`) and a pin that could not see arms (a)–(c) on
  the gate's own source; all three are fixed, and the re-measured figures did not move — a second
  confirmation that this tree's arm-(d) population is exactly those six.
- **A note on the arm-(d) row shape.** Literals sharing one allow-list identity (`file:line:Type`)
  collapse to one row, but the row **says how many** it stands for (`2 literals, 3-3 unkeyed elements`),
  so `[]T{{…}, {…}}` cannot read as a single literal. The red set sorts by `(file, line, name)` — the
  `name` tiebreak keeps two different struct types on one line out of map-iteration order, so the report
  and the allow-list file are deterministic.
- **The fifth review found a VACUOUS PIN.** `selfTestUnkeyed` declared its container-bodied generic as
  `type grid[T any] [][]pair` — a body that never mentions `T`, so the recursive substitution the fold had
  just added was never on the path. Reverting `substitute`'s `*ast.ArrayType` case left the self-test
  **green**; the same mutant reds the moment the body is `[][]T`. The shape is fixed, and all three new pins
  are now mutation-verified individually (removing `stripType`'s `StarExpr` case reds the pointer row;
  overwriting instead of composing the substitution reds the chained row). The lesson is the round's own,
  one turn further in: **a pin must be mutated against the code it claims to pin** — "the test covers this
  shape" is itself a claim, and this one was wrong while the code was right.
- **The third and fourth reviews corrected §Forward itself — twice.** RF-068-3b v1 listed two elision
  shapes that cannot occur in a compiled tree (`f({1, 2})` is a **syntax error**; `Outer{Inner: {…}}` parses
  and is a **type-check error**) and omitted the real limit; v2 said "the ONE shape" while three
  same-package forms (`[]*table{{…}}`, a container-bodied generic, a chained generic) were still
  unattributed. **All three are now fixed and pinned.** A forward item is a claim like any other, and this
  one has needed a compiler behind it twice — the same class the ADR opens with, in the last place one looks
  for it.
- **Not modelled.** The gate is a quality-process instrument; it is modelled only as far as the
  `QualityPipeline`/`QualityGate` entities reach (ADR 0041). The product model does **not** move: D11
  changes parameter **shapes**, not behaviour, and `usage-summed-for-the-ready-summary` concerns values,
  not parameter lists.

## Forward

- **RF-068-1** — Result lists (D6): **11** measured; a future round may widen the scan to result runs. The
  trigger is a delivered defect traced to a swapped return; without such a trigger this stays inert.
- **RF-068-2** — The `n ≥ 4` boundary (D5): 6 interface methods + 33 `FuncDecl`s sit below it. Widening is
  a new ADR, not a knob.
- **RF-068-3** — Arm (d) keys an allow-list entry by `file:line:TypeName` (a literal has no name), so an
  allow-listed literal stales when unrelated lines move above it. The list is empty, so this is a
  property of the mechanism, not a live cost.
- **RF-068-3b** — The ONE shape arm (d) cannot attribute: an elided literal reached through a
  **cross-package** named container (`pkg.Table{{1, 2}}`, or a chained/aliased generic declared in another
  package). Resolving it needs the other package's AST — the same boundary that leaves `time.Time{}` out of
  scope. **Within** a package, D4 now covers every form: element, map value, map key, nested, named
  container, pointer-to-named-container, generic body, and chained generic. Measured: `internal/**` declares
  **no** same-package named array/slice/map type at all today (`grep -rnE '^\s*type \w+ (\[\]|\[[0-9]+\]|map\[)' internal`
  is empty), so the shape this limit guards has zero instances on this tree — recorded with its
  measurement, not fixed.
  *(This item has now been wrong twice, both times caught in review: v1 named `f({1, 2})` and
  `Outer{Inner: {…}}` as gaps — `f({1,2})` is a **syntax error** and the other a **type-check error**, so
  neither can occur — and v2 said "the ONE shape" while `[]*table{{{…}}}`, a container-bodied generic and a
  chained generic were still unattributed (all three fixed, and pinned). A forward item is a claim like any
  other, and this one has needed a compiler behind it both times.)*
- **RF-068-4** — The `@quality` build tag is not evaluated against build-tag-gated files' union: the scan
  parses files **as written**. Two darwin-tagged files declare the same symbol today and both are
  scanned; a future tag-gated divergence would be measured twice, not missed.
- **RF-068-5** — `requestSpec` carries `MaxTokens` as a **named field with ownership undecided**, so the
  per-(provider, model) budget-field selection of [#195](https://github.com/gosharplite/tellme/issues/195)
  has a home. This round deliberately **pre-empts nothing** there.

## Related

- [issue #210](https://github.com/gosharplite/tellme/issues/210) — the durable home of this round's
  candidate, absorbing drafts #206 #207 #208 #209 and their four reviews.
- [ADR 0011](0011-layer-discipline-gate.md) — the gate shape, the fail-on-stale ratchet, the `tools/arch`
  recipe this one mirrors.
- [ADR 0020](0020-cli-ui-decoupling.md) fold **R-51-1** — refuse-to-grow (the allow-list's guard).
- [ADR 0041](0041-domain-model-drift-guard.md) **D2** — "a check must demonstrate a catch".
- `specs/truth/techstack.md` — the *Parameter-shape gate* row (the gate's truth home).
- `docs/domain-model/quality.modelith.yaml` — the `QualityPipeline`/`QualityGate` entry.
