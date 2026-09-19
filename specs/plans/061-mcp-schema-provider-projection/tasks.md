# Tasks — Provider-wire projection of MCP-relayed tool schemas (round 061)

**Plan Package**: `specs/plans/061-mcp-schema-provider-projection`
**Execution**: `/axb-implement` — Red → Green → Refactor, one task at a time, marked `[X]` after verification.

## Phase 1 — Setup

- [X] **T001** — Branch `061-mcp-schema-provider-projection` off `dev` `768dbec`; plan package created (spec + checklist + truth-delta skeleton); `STATUS.md` phase-gate update.
- [X] **T002** — `/axb-clarify` closed (CQ-1 → C · CQ-2 → ii · CQ-3 → i) + CQ-4…CQ-7 (probe approved · gate = unit pin · floor cross-family · silent) recorded in `spec.md` (S-1…S-7).

## Phase 2 — Foundational

- [X] **T003** — `/axb-spec-by-example`: acceptance journey written (`features/acceptance/keeping-a-gemini-turn-usable-with-an-annotated-mcp-server.feature`; 3 Rules; no residual `# [need clarification]`).
- [X] **T004** — `/axb-technical-research`: the **live probe** (declaration-only `generateContent`, `ait-comment` → `websc-dev-433809`, `gemini-3.8-flash`); the measured accept/reject table recorded in **ADR 0031**; `techstack.md` rows updated (the floor row, the qualified envelope row, the projection row, the gate row); ADR 0025 gains the amendment pointer; index rows added.
- [X] **T005** — `/axb-system-analysis`: `plan.md` (1 interface, 1 wave; api/data NOOP, ui skipped).

## Phase 3 — Test Alignment & Implementation

- [X] **T006** (`/axb-dsl-refine`) — truth rows: the round-056 declaration row gained the **carve-out** qualification; `chat/using-tools-from-a-remote-mcp-server.feature` gained the new Rule + 3 Examples; `chat/dsl.md` gained the round-061 Given (2) + Then (3) rows. Topology audit: no new findings.
- [X] **T007 (RED)** — Unit pins written first: `internal/infrastructure/mcp/schema_test.go::TestNormalizeMCPSchema_StripsVendorExtensions` and `internal/infrastructure/llm/gemini/schema_test.go` (3 pins incl. the envelope-shaped gate). Observed RED (no floor / no projection).
- [X] **T008 (GREEN)** — Floor implemented (`isVendorExtension`, recursive `stripVendorExtensions`, wired into `NormalizeMCPSchema`) and projection implemented (`supportedSchemaKeys`, `projectSchema` recursive + fail-closed, wired into `buildToolDeclarations`). Both packages green.
- [X] **T009 (RED → GREEN)** — E2E: `mcptest.AnnotatedSchema()`, the five new steps (`step_r061_mcp_schema.go`), `writeGeminiConfig` carrying the MCP block; the truth feature's new Rule drives both the **gemini** leg and the **tolerant-family** leg; the E2E contract ran green (the three new scenarios executed).
- [X] **T010 (REFACTOR)** — `projectSchema` failed closed on a non-object schema (caught by `TestProjectSchema_FailsClosed`), documented at the site.

## Phase 4 — Verification

- [X] **T011** — **Witness A**: removing the projection call turned `TestProjectToolDeclarations_NoUnsupportedKeywordReachesTheWire` **RED** (unsupported keywords on the wire) → reverted.
- [X] **T012** — **Witness B**: removing `stripVendorExtensions(obj)` turned `TestNormalizeMCPSchema_StripsVendorExtensions` **RED** (vendor keys kept at root/nested) → reverted.
- [X] **T013** — Gates: `gofmt` clean · `go vet ./...` clean · `go build ./...` clean · `go test -count=1 ./...` green (incl. the godog E2E) · `make verify` (see `STATUS.md`/the round record for the run).
- [X] **T014** — Docs/ledger: `research.md`, `plan.md`, this `tasks.md`, `truth-delta.md` owner rows.

## Notes / ledger conventions

- Witness executions are recorded here (TD-6 convention); the red output of both witnesses is reproducible from the two call sites (`buildToolDeclarations`'s projection call, `NormalizeMCPSchema`'s floor call).
- No product file outside the five sites in `plan.md` was touched; `go.mod`/`go.sum` unchanged.

## Phase 5 — Review folds (PR #128, review `5740291616`)

- [X] **T015** (`B-061-1`) — the floor is **structure-aware** (keyword positions only; `properties` names opaque; data values never walked) **+** `NormalizeMCPSchema` re-asserts `required ⊆ properties` on the post-floor object. Three pins: a property **named** `x-…` survives with the postcondition holding; an `x-`-keyed member inside a `default` **value** is preserved; an undeclared `required` entry is still refused.
- [X] **T016** (`F-061-1`, also `TD-061-2`) — the gate now reads the **owner** (`gemini.SupportedSchemaKeys()`): a **containment** pin walks the projected declaration structure-aware and fails on any keyword not in the set, and a **golden-set** pin asserts the owner equals the probe's measured set **bidirectionally** (delete a key → RED; add an unmeasured key → RED), and a **coverage** pin proves every allowlisted key survives. The former hardcoded deny-lists are gone; the E2E containment check reads the same owner.
- [X] **T017** (`F-061-2`) — value-shape normalization (`type` arrays, `enum` coercion) implemented + pinned; the shape probe (2 rejected shapes, 2 accepted controls) recorded in `research.md` and ADR 0031 D2/D6a.
- [X] **T018** (`F-061-3`) — (1) the invariance claim is narrowed to **semantic** invariance (spec US1 AS3 → the plan-side journey); **corrected after fold-verification R-4**: at `e67954c` the control Example existed only **plan-side** (godog loads no plan path), so the interface carrier was added in the R-4 fold — one `Then … carries no server-side mark` line on the pre-existing no-mark Example (`using-tools-from-a-remote-mcp-server.feature:18`), now **executed** by the E2E; (2) the args Then asserts **declared properties of the `MCP_PAYLOAD` subschema with descriptions**, matching its DSL row; (3) all three Thens are **declaration-scoped** (the body-wide `x-` substring scan is gone).
- [X] **T019** (`TD-061-1`, `RF-061-4`) — ADR §Forward records the **unclosed class** for the OpenAI-compatible wire (RF-061-7) and the `$ref`+`$defs` dangling-subschema loosening; a new `cmd/tellme` pin runs every **native** declaration through the projection and asserts semantic identity (RF-061-4 closed durably, not by inspection).
- [X] **T020** (nits 1–4) — the redundant `json.RawMessage(...)` conversion is gone (`ProjectSchema` is exported and returns the type); the two freeform literals + the `ProjectSchema` empty-input doc are named in ADR §Forward (RF-061-9); the value-shape fixtures are consistent across the pins.

### Fold witnesses (reproduced RED, then reverted)

- **C** — making the floor context-blind again (walking every map) turns `TestNormalizeMCPSchema_PreservesPropertyNamedLikeAnExtension` RED.
- **D** — deleting a measured key from the owner (`items` removed from `supportedSchemaKeys`) turns the **golden-set** pin RED (the review's F-061-1 failure mode: under the former deny-list this change was invisible); the owner is pinned **bidirectionally** to the probe's measured set, so adding an unmeasured key fails too.
- **E** — removing the `type`-array normalization turns `TestProjectSchema_NormalizesValueShapes` RED.

## Phase 6 — Fold-verification folds (PR #128, verification `5740338531`)

- [X] **T021** (`R-1`) — `$defs`/`definitions`/`dependencies` are name→schema maps and are back in `schemaNodeChildMaps`, so the floor's contract ("recursively, for every family") holds again for that position. Pin: `TestNormalizeMCPSchema_StripsMarksInsideDefinitionMaps`.
- [X] **T022** (`R-2`) — the shapes the coercion **emits** were probed (2 rejected: `nullable` without a `type`; `enum` beside a non-scalar/absent `type`; controls accepted). The normalization now records `nullable` only when a `type` remains and **drops** `enum` unless the type is scalar. Pin: `TestProjectSchema_CoercedShapesAreTheMeasuredOnes`. ADR D2/D6a′.
- [X] **T023** (`R-3`) — a non-object subschema at `properties` values / `items` / `additionalProperties` (a `bool` there is legal and kept) degrades to the accepted empty `{}`. Pinned in the same test. ADR D6c.
- [X] **T024** (`R-4` — record honesty) — the no-mark control now has an **executed** interface carrier, T018 is corrected, and the fold comment's "4 Examples" claim is corrected in the fold-response record. `truth-current` restored.
- [X] **T025** (nits) — (1) `SupportedSchemaKeys()` returns a **copy** (no cross-package aliasing); (2) the triplicated recursion discipline is recorded as RF-061-10 for a future single-walker round; (3) a `null` enum member is dropped; (4) the redundant 14-key deny-list inside the containment pin is gone.

### Fold-verification witnesses (reproduced RED, then reverted)

- **F** — restoring a context-blind `$defs` (i.e. removing it from the child maps) turns `TestNormalizeMCPSchema_StripsMarksInsideDefinitionMaps` RED.
- **G** — re-emitting `nullable` without a `type` (the pre-R-2 behaviour) turns `TestProjectSchema_CoercedShapesAreTheMeasuredOnes` RED.
- **H** — passing a non-object subschema through (the pre-R-3 behaviour) turns the same pin RED.

## Phase 7 — Re-verification folds (PR #128, verification `5740370401`)

- [X] **T026** (`R-5`) — `projectList` now routes its elements through `projectSubschema`, so a boolean/scalar element inside `oneOf`/`allOf` (allowlisted *because* their elements are schema nodes) degrades to the accepted `{}` instead of reaching the wire. Pin rows added (`oneOf:[true]`, `allOf:[false]`); ADR D6c widened to "every position that *is* a schema node".
- [X] **T027** (nit 1) — the stray duplicate `D6b` heading (an empty stub left by an earlier fold) is deleted from ADR 0031.
- [X] **T028** (nit 2) — the `additionalProperties` bool form was **probed** at both root and property level (both accepted) and the rows + the site comment now cite the measurement rather than an assumption.
- [X] **T029** (nit 3) — recorded as **RF-061-11** (an `enum` with no/non-scalar `type` loses the constraint; the string-enum shape is the measured-accepted alternative) — a forward item, not adopted.

### Re-verification witness (reproduced RED, then reverted)

- **I** — routing `oneOf`/`allOf` elements back through `projectValue` turns `TestProjectSchema_CoercedShapesAreTheMeasuredOnes` RED (a `true`/`false` element on the wire).

## Phase 8 — Verification-2 fold (PR #128, verification `5740392315`)

- [X] **T030** (`V-061-1`, FOLD path chosen) — the surface's single owner is now the **key + value-kind table** `supportedSchemaValueKinds` (`supportedSchemaKeys` derived); the projection applies each key's kind, dropping a mismatch and applying only the **measured** coercions. Wrong-kind values were probed (7 rejected, 3 tolerated-but-dropped) and the rows recorded (ADR D2/D8, `research.md`). Pins: `TestProjectSchema_ValueKindsAreEnforced`, `TestSchemaValueKindOK_MirrorsTheTable`; the containment walker now asserts **shape** as well as key membership. **Witness J** (kind enforcement bypassed) reproduced RED.
- [X] **T031** (the review's enum-object nit) — a non-scalar `enum` member is **dropped** rather than JSON-stringified into a fake value; pinned in `TestProjectSchema_ValueKindsAreEnforced`.

### Witness (reproduced RED, then reverted)

- **J** — `applyValueKind` returning every value unvalidated turns `TestProjectSchema_ValueKindsAreEnforced` RED (`description: 123` on the wire).

## Phase 9 — Verification-3 fold (PR #128, verification `5740423585`)

- [X] **T032** (`W-061-1`) — `normalizeType` re-checks the **substituted member** (a `type` array's lone member must be a string, else no `type` is emitted). Pin rows (`type:[5]`, `[true]`, `[{}]`, `[["string"]]` → no `type`) + a **round-trip pin** (`TestProjectSchema_OutputSatisfiesTheGate`) asserting the projection's own output passes the gate's predicate for ten adversarial inputs. **Witness K** reproduced RED.

### Witness (reproduced RED, then reverted)

- **K** — returning `kept[0]` without the member-kind check turns `TestProjectSchema_ValueKindsAreEnforced` RED (`type:[5]` → `type:5`).

## Certification

- **T033** — final verification `5740450136` at **`d6b2786`**: **CERTIFIED MERGE-READY · review loop CLOSED** (gates green; witness K re-reproduced by the reviewer — it reds *both* the value-kind pin and the round-trip pin, proving the round-trip assertion non-vacuous; an independent cross-check of 260 (key, wrong-kind value) combinations showed 0 disagreements). Review trail: **B-061-1 · F-061-1…3 · R-1…R-5 · V-061-1 · W-061-1**, each an instance of *"a third-party dialect reaching a closed proto"*, all closed and pinned.
- **Post-merge (operator)**: verify with the GitHub MCP server enabled + `TYPE: gemini` (the one leg not carriable hermetically — S-4/CQ-4), close [#127](https://github.com/gosharplite/tellme/issues/127), then run `SESSION-CLOSEOUT.md` (including the `dev → main` propagation and the `round-061` tag on the operator's approval).
