# Session Summary — 2026-10-27

**Project**: `tellme` — a disciplined BDD re-creation of `tell-me-go`
**Repo**: `github.com/gosharplite/tellme`
**Status file**: [`STATUS.md`](../../../../../STATUS.md) *(back-link — the single live-state source)*
**Workspace**: `$TELL_ME_HOME` on the **`kiosk`** env (a **GKE container managed by ConfigHub**); linux/amd64; Go **1.26.10** (user-scoped).
**Session mode**: `butler`.
**Branch**: `098-param-swap-guard` (off `dev` `0231f64`); `dev` untouched.
**Status at end of session**: **round 098 `098-param-swap-guard` in flight — PR open into `dev`, awaiting human review** (only a human merges). Closes [#210](https://github.com/gosharplite/tellme/issues/210).

---

## Session 90 — bootstrap → issue #210 → round 098

The session began with `SESSION-BOOTSTRAP.md` Steps 1–5 (round 097 delivered/frozen; active branch `dev`,
clean). The operator asked for issue #210 to be read, then directed: *"Create a new branch from dev branch.
Start a new aixbdd round, the goal is to close issue 210. Create a PR for review."*

### At a glance

| Area | Outcome |
| --- | --- |
| Round | **098 `098-param-swap-guard`** (close [#210](https://github.com/gosharplite/tellme/issues/210); **ADR 0068**) |
| The gate | `tools/params` — a `//go:build quality` `go/parser` scan, `verify-params` in `make verify`, allow-list fail-on-stale + refuse-to-grow, **empty** at delivery |
| The refactor | **46** sites cleared (40 signatures + 6 call-site literals) · `internal/cli/cli.go` **1682 → 474** lines |
| Gates | `make verify` **OK** (twelve members) · E2E **334 scenarios · 2519 steps unchanged, green** · `test-race` clean · `modelith-check` no drift · `modelith-lint` 0/0 · `verify-adr-index` consistent · `govulncheck` clean |
| Delivery | branch → **PR open into `dev`** (no Copilot review; only a human merges) |

### The pipeline

| Skill | Artifact |
| --- | --- |
| `/axb-specify` | `spec.md` (3 stories: the guard · the tree · the split) + `checklists/requirements.md` + `truth-delta.md` |
| `/axb-spec-by-example` | `features/acceptance/the-parameter-shape-gate.feature` (6 Rules, 13 Examples) |
| `/axb-technical-research` | `research.md` D1–D11 + **ADR 0068** (answers all four of the issue's open verification questions **by measurement**) |
| `/axb-system-analysis` | `plan.md` (two interfaces: `cli` NOOP · build MODIFY; 5 waves; api/data NOOP; ui skipped) |
| `/axb-dsl-refine` | **NOOP** — behaviour-neutral; recorded on the owner's behalf in `truth-delta.md` |
| `/axb-tasks` | `tasks.md` T001–T011 + the **Claim→Witness ledger** (W1–W10 + the post-fix figure table) + the 12-item DoD checklist |
| `/axb-implement` | one-shot TDD: script first (red), then the fixes (green), then the split |

### The number that mattered

The **script owns the numbers**. Every figure in the plan comes from `make verify-params-report`. The
pre-round red set reproduces issue #210's published **signature** table exactly (J 31 · J′ 4 · K 36 · K′ 4 ·
M 9 · N 6 · O 38 · P 11 · Q 0 · R 0 · I 69 · H 10) and corrects its arm-(d) figure: **L = 6**, not 2. The
round's own red set is therefore **46** = 40 signatures + 6 arm-(d) literals. That is the durable lesson of
the issue's five drafts — the flattened type normaliser produced **four** wrong measurements, all from
prose — and the fold re-taught it: the *first* published `L = 0` came from an arm narrower than its own
definition.

### The two claims the review corrected — the round's thesis, one layer up

**W7** was first recorded as "not reproducible on this host … incidental (VCS stamping / timing)".
**That was backwards.** `go help test` says the cache matches on exactly the files a test opens, and this
gate reads the tree **in-process** — so the re-runs *were* the documented behaviour, and they contradicted
the row that quoted them. The `-count=1` rationale had been **inherited verbatim** from
`tools/arch/doc.go`, where it is true (that gate shells a `go list` **child process** the cache cannot
see). Reworded in ADR 0068 D9, `doc.go`, the Makefile and `techstack.md`; W7 re-recorded as
**confirmed-by-design**; `-count=1` **kept** for parity.

**Arm (d) matched only explicitly-typed literals**, so `[]T{{1, 2}}` — the test-table shape — was skipped
and the published `L = 0` was not what the arm's definition measures. The corrected walk (an
expected-element-name stack + whole-file struct collection, so function-local types count) found **6** live
sites on the pre-round tree. Both defects were found by **review, not by the round's own witnesses** — the
class this round exists to eliminate, appearing one layer up.

### The second review round: the witness for the refactor was itself vacuous

The re-review of the fold found **two residuals**, both fixed rather than recorded:

1. **The self-application pin could not see arms (a)–(c).** Every line of `tools/params` lives in
   `params_test.go`, and `redSites()` skips test declarations — so the pin cited as proving the
   `declSpec` refactor **would not have redded on the old positional `newDecl`**. Fixed with a
   `scanTarget` carrying `pinProduction`; the reviewer's probe now reds both a `func zzSelfProbe(a, b, c
   int, d string)` (arms a/b) and an unkeyed `site{…}` literal (arm d).
2. **Two more elided shapes were still missed** — nested containers (`[][]T{{{…}}}`) and a
   **struct-typed map key** (`map[T]int{{…}: 1}`). Fixed with one `effectiveLitTypes` prepass (the walk now
   pushes the element **expression**, so containers unwrap level by level) plus a key walk; seven shapes
   and two negatives are now pinned in the arm-(d) self-tests.

The re-measured pre-round figure is **unchanged at 46** — the added coverage finds nothing new in
`internal/**`. That is the useful shape of the result: the arm's population is exactly those six, and now
the walk actually says so.

### The third review round: a forward item with no compiler behind it

The re-review of the second fold found three things, and the instructive one is the **forward item**:

- **RF-068-3b listed two elision shapes that are not valid Go** — `f({1, 2})` and `Outer{Inner: {…}}` are
  *syntax errors*; elision is legal only in an array/slice/map literal's element or key position. So the
  item described a limit that cannot occur and **omitted the one that does** (a **cross-package** named
  container). Rewritten, with the measurement that `internal/**` declares no same-package named container
  type at all. A forward item is a claim like any other — and this one had never been compiled.
- A **same-package named container** (`type table []T`) was opaque to the walk; now collected
  (whole-file) and resolved through, **including generic instantiations** (`gen[pair]{{…}}`).
- A **map whose key and value are the same struct** was double-counted; fixed at the root by having the
  preass assign the **key** position the key type, which also let the separate key walk be deleted.

### The fourth review round: the forward item had to be compiled a second time

The last re-review recommended merging and left two points, both fixed:

- The ADR said **both** refuted elision shapes are *syntax* errors. Only `f({1, 2})` is — `go/parser`
  rejects it; `Outer{Inner: {1, 2}}` **parses** and it is the **type checker** that refuses it. Corrected,
  because the ADR's own point is precision about what refuted what.
- **RF-068-3b over-claimed "the ONE shape"** while three same-package forms were still unattributed — a
  **pointer** to a named container (`[]*table{{…}}`), a generic whose **body is a container**
  (`grid[T] [][]T`), and a **chained** generic (`chain[T] gen[T]`). All three now resolve (strip
  `StarExpr`; substitute over type *expressions*; **compose** the substitution across hops; collect a body
  that is itself a named type) and are pinned, so `selfTestUnkeyed` covers **fifteen shapes and three
  negatives**. The figures did not move: `internal/**` has no such types.

So §Forward has now needed a compiler behind it **twice** in one round. That is the durable note: a
deferral is a claim, and the place it is least likely to be checked is the place it is most likely to be
wrong.

### The fifth review round: a vacuous pin — the code was right, the evidence was not

The last re-review kept its merge recommendation and found one thing: the pin for the container-bodied
generic was **vacuous**. `selfTestUnkeyed` declared `type grid[T any] [][]pair` — a body that never
mentions `T` — so the recursive substitution the fourth fold had added was **never on the path**. Reverting
that code left the self-test **green**; the same mutant reds the moment the body is `[][]T`.

Fixed, and then all three new pins were mutation-tested individually rather than assumed:

| Mutation | Row that must go missing |
| --- | --- |
| revert `substitute`'s `*ast.ArrayType` case | 35 (grid) |
| remove `stripType`'s `StarExpr` case | 34 (pointer) |
| overwrite the substitution instead of composing it | 36 (chain) |

This is the first fold in five where the **code** was correct and the **evidence** was not — and it is the
cleanest statement of the round's own lesson, one level further in than the round had aimed: **a pin must
be mutated against the code it claims to pin.** "The self-test covers this shape" is a claim like any
other.

### Decisions locked (round 098)

| # | Decision |
| --- | --- |
| **D1** | Four arms, two hazards (a/b/d swap · c DI idiom); no bare parameter-count arm. |
| **D2** | Arm (c) is **kept** (option 2) on the DI rationale — 0 identical-signature func runs measured, so no swap claim. |
| **D3** | The guard scans **top-level interface methods** (4 of the 40 reds are ports). |
| **D4** | Arm (d) exists because `go vet`'s `composites` analyser is **cross-package only** — verified; W10 proves the blind spot. |
| **D5** | The `n ≥ 4` boundary is **kept**, remainder measured (6 interface + 33 `FuncDecl` post-fix). |
| **D6** | Result lists out of scope (**11** measured); type identity syntactic (**0** aliases). |
| **D7** | `-count=1` mandatory (rationale origin `tools/arch/doc.go`); the parser walk is a **design property, not a catch** (the earlier build-tag justification was false). |
| **D8** | The allow-list is fail-on-stale + refuse-to-grow, and ships **empty** (option A). |

### Environment findings (durable)

- A **tree-reading** gate that opens its inputs **in-process** has those inputs in the test cache's key;
  only a gate that reads the tree through a **child process** (`go list`) is cache-blind. Do not inherit
  `-count=1`'s rationale across that boundary — verify which side a new gate is on.
- The `internal/cli` split was done by **moving whole contiguous concern blocks** and letting `goimports`
  rebuild the import block; the three **aliased** imports (`appsuggestions`, `agentport`, `domaintools`)
  must be re-added by hand (goimports cannot guess aliases).

### Open items (non-blocking)

- **PR for round 098** awaits human review/merge → then the closeout (`SESSION-CLOSEOUT.md`): delete the
  branch, propagate `dev -> main` (**no-ff**), tag **`round-098`**, refresh the installed binary.
- **ADR 0068 §Forward** RF-068-1…5 (result lists · the `n ≥ 4` remainder · arm-(d) line-keyed allow-list
  entries · the build-tag union is parse-as-written · `requestSpec.MaxTokens` ownership left to #195) —
  disclosures, not tasking.

### Next steps

1. Human reviews + merges the round-098 PR; then the closeout (→ tag `round-098`).
2. Re-read `SESSION-BOOTSTRAP.md` next session (active branch `098-param-swap-guard` until merged, then `dev`).
3. Candidate ordering note from the round: **this round before [#195](https://github.com/gosharplite/tellme/issues/195)** — `requestSpec` now carries `MaxTokens` as a named field, which is exactly the shape #195's per-(provider, model) budget selection needs.
