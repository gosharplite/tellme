# Round 098 — `098-param-swap-guard`

**Theme**: give tellme a **mechanical guard for the compile-blind argument swap** — two adjacent
parameters of an **identical type** whose swap at a call site **still compiles** — and land the refactor
the guard forces: **42** red sites cleared (40 signatures + 2 call-site literals) and
`internal/cli/cli.go` split from **1682** lines into cohesive same-package files. The guard **reds on
today's tree** and ships green only because the tree was fixed; it then joins `make verify` as
`verify-params`.

**Branch base**: off `dev` `0231f64` (session-89 closeout head).

**Anchor issue**: [#210](https://github.com/gosharplite/tellme/issues/210) — *"Code quality: guard for
compile-blind argument swaps (35–40 sites, arms a–d) + split `internal/cli/cli.go`"*, which supersedes
[#206](https://github.com/gosharplite/tellme/issues/206)–[#209](https://github.com/gosharplite/tellme/issues/209).

**Input (operator directive, 2026-10-27)**: *"Create a new branch from dev branch. Start a new aixbdd
round, the goal is to close issue 210. Create a PR for review."*

---

## 1. Why this round

tellme's quality machine is dense — eleven `make verify` members, a **334-scenario / 2519-step**
executable E2E contract, ~16.9k test LOC against ~15.0k production LOC, a green advisory
`modelith-drift` and a clean advisory `dead-code`. The classic remedies have little left to buy.

What **no** gate can see is the **compile-blind argument swap**. Three examples from the pre-round tree:

```go
got := ComputeCost(p, 4, 6, 3, 2) // miss 4, hit 6, completion 3, thinking 2
                                  // ^ the comment exists BECAUSE the call site cannot say what it means
func renderTurn(homeDir, configPath, prompt string, …)          // three adjacent strings
func NewRemoteClient(ctx, url, authorization string, …)         // swap ⇒ connects to a URL made of a token
```

The round is **guard-first**: a refactor-only round has **no falsifiable carrier** — nothing would red if
the hazard returned. The obligation is carried by **ADR 0041 D2** (*a check must demonstrate a catch*) and
by upstream `aixbdd-tmg` ADR 0006 (*no normative clause reaches truth unwitnessed*).

## 2. The change (surfaces)

1. **`tools/params/`** — a new build-tagged (`//go:build quality`) gate + its `allowlist.txt` (empty at
   delivery) + `doc.go`. The scan is a `go/parser` walk over `internal/**` with a **structural** type key.
2. **`Makefile`** — `verify-params` (three commands, `-count=1`) joins `make verify`;
   `verify-params-report` prints every baseline figure with its label and universe; `verify-params-update`
   regenerates the allow-list and **refuses to grow** it.
3. **The refactor** — named seam structs at every red site; `internal/cli/cli.go` split into
   `boot.go` · `turn.go` · `reporting.go` · `workspace.go` · `diagnostics.go` · `interactive.go`
   (same package ⇒ **zero import-graph movement**).
4. **Truth** — `specs/truth/techstack.md` gains the *Parameter-shape gate* row (the gate's own truth
   home) and names `verify-params` in the task-runner aggregate.
5. **Decision record** — a **new ADR 0068** (the two hazards, arms a–d, the interface scope, the `n ≥ 4`
   boundary with its measurement, results out of scope, syntactic type identity, the allow-list design,
   and the measurement-bug lesson).

**Not a behaviour change**: no provider request changes, no rendered byte changes, no persisted-state
change, no new dependency (`go.mod`/`go.sum` unchanged), no new import edge.

## 3. User stories

### Story 1 — a swappable signature is refused by a machine, not by a reader (P1)

As a tellme maintainer, I want a **gate** that reds when a production signature or call site can have two
same-typed values exchanged without a compile error, so that the class of defect that produced
`ComputeCost(p, 4, 6, 3, 2)` cannot return silently in a green build.

**Why this priority**: it is the round's whole value — the guard is what makes the refactor durable.

**Independent verification method**: run the scan on the pre-round tree (`dev` `0231f64`) → it reds with
exactly the committed script's **42** sites; run it on the round's tree → **0** sites, allow-list empty.

**Acceptance scenarios**:

1. **Given** a production function with `n ≥ 4` parameters and two adjacent ones of the same type,
   **When** the gate runs, **Then** it fails and names `file:line`, the function, and the arm that fired.
2. **Given** a tree in which every such signature has been reshaped or keyed, **When** the gate runs,
   **Then** it passes with an empty allow-list.

**Functional Requirements (FR)**:

- **FR-001**: the gate MUST red a production `FuncDecl` or **top-level interface method** under
  `internal/**` (excluding `_test.go`) when `n ≥ 4` **and** an adjacent run of ≥ 2 parameters shares an
  identical structural type key. [Verification Intent: unobservable → the gate's own first red run (W1)
  + mutation witness W3]
- **FR-002**: the gate MUST red an adjacent identical-type run of **≥ 3** regardless of arity (so a
  post-fix accidental re-introduction cannot pass vacuously at `n = 3`). [Verification Intent:
  unobservable → W3/W4/W5, each a run-of-3 mutation]
- **FR-003**: the gate MUST red an adjacent run of ≥ 2 **`func`-typed** parameters, and the ADR MUST
  state that this arm is the **DI idiom**, not a swap rule. [Verification Intent: unobservable → the
  `cliDeps` fold + the arm-(c) rows of the report]
- **FR-004**: the gate MUST red a **same-package named struct** composite literal with ≥ 2 unkeyed
  elements, scanning `_test.go` for this arm. [Verification Intent: unobservable → mutation witness W10]
- **FR-005**: the gate MUST scan **top-level interface methods** (the two domain ports are contracts, not
  conveniences). [Verification Intent: unobservable → W5]
- **FR-006**: a violation is cleared only by fixing it or by an allow-list entry carrying a recorded
  reason; the allow-list MUST be **fail-on-stale** and MUST refuse to grow. [Verification Intent:
  unobservable → W6's synthetic stale entry + the refuse-to-grow branch]
- **FR-007**: the scan MUST emit, per site, `file:line · name · arms · the type string it keyed on · n ·
  longest run`, and MUST print every baseline figure with its label and universe. [Verification Intent:
  unobservable → `verify-params-report` output; its figures ARE the plan's figures]

**Non-Functional Requirements (NFR)**:

- **NFR-001**: `-count=1` MUST be used — the verdict reads the whole tree, which the Go test cache cannot
  see (`tools/arch/doc.go` is the rationale's origin; ADR 0068 D9). [Verification Intent: unobservable →
  W7, a violation added against a warm cache]
- **NFR-002**: the gate MUST honour ADR 0012's hermetic invocation boundary and MUST NOT add a
  dependency (stdlib `go/parser` only). [Verification Intent: unobservable → `go.mod`/`go.sum` unchanged;
  the Makefile recipe shape]
- **NFR-003**: the gate MUST NOT contribute to the default `go test ./...` run (build tag) and MUST NOT
  be an advisory-with-no-catch (ADR 0064's distinction). [Verification Intent: unobservable → the
  `//go:build quality` tag + the red-first witness]

---

### Story 2 — the tree itself is free of the hazard (P1)

As a tellme maintainer, I want every red site fixed rather than allow-listed, so that the invariant is
true of **both** signatures and call sites on the day the gate ships.

**Why this priority**: an allow-list on day one would make the guard's green meaningless at the sites
that matter most (the wire layer, the two domain ports).

**Independent verification method**: the post-fix report shows **0** sites with an **empty** allow-list;
the E2E contract is unchanged.

**Acceptance scenarios**:

1. **Given** the round's tree, **When** the gate runs, **Then** it passes and the allow-list file carries
   no entry.
2. **Given** the round's tree, **When** the full E2E contract runs, **Then** it reports the same
   **334 scenarios · 2519 steps**, all green.

**Functional Requirements (FR)**:

- **FR-008**: every red site MUST be fixed by a **named seam struct** (the house idiom) or by keying the
  literal — never by inserting a differently-typed filler parameter. [Verification Intent: observable →
  the gate's 0-site report + the diff]
- **FR-009**: **both** domain port contracts MUST change (`render.Lines` and
  `suggestions.WorkspaceSource`), with the **import graph unchanged**. [Verification Intent: observable
  → `verify-architecture`'s header-only baseline]
- **FR-010**: unit-test edits MUST be signature-only — **no assertion changes**. [Verification Intent:
  observable → the test diff + the unchanged E2E counts]

**Non-Functional Requirements (NFR)**:

- **NFR-004**: no rendered output, exit code, or persisted state may change. [Verification Intent:
  observable → W8, the full E2E contract unchanged at 334 / 2519]

---

### Story 3 — the CLI's largest file is a set of cohesive files (P2)

As a tellme maintainer, I want `internal/cli/cli.go` split into cohesive same-package files, so that the
package's entry surface, its turn path, its offline reporters, and its diagnostics are readable and
reviewable in isolation.

**Why this priority**: it is the second half of issue #210's mandate, and it is what makes the CLI-side
seam structs land in a file a reviewer can hold in their head.

**Independent verification method**: every production file in `internal/cli` is under ~600 lines;
`verify-architecture`'s baseline is still **header-only** (same package ⇒ no import edge moved); the
package's tests pass unchanged.

**Acceptance scenarios**:

1. **Given** the split, **When** `verify-architecture` runs, **Then** the baseline is still header-only.
2. **Given** the split, **When** the CLI's test suite runs, **Then** it passes with no behavioural edit.

**Functional Requirements (FR)**:

- **FR-011**: `internal/cli/cli.go` MUST be split into same-package files, and the pre-existing
  `internal/cli/prompt.go` name collision MUST be resolved explicitly (the new interactive file is
  `interactive.go`). [Verification Intent: observable → the file set + `wc -l`]
- **FR-012**: no production file in `internal/cli` may exceed ~600 lines. [Verification Intent:
  observable → `wc -l`]

**Non-Functional Requirements (NFR)**:

- **NFR-005**: the split MUST NOT move an import edge or add a dependency. [Verification Intent:
  observable → W9 (`verify-architecture` + `modelith-check` + `test-race`)]

---

## 4. Requirements (global)

> No cross-story global FR/NFR; Stories 1–3 carry the round.

### Key entities

- **Swap hazard** — a signature or literal position where two same-typed values can be exchanged and the
  program still compiles; arms (a)/(b)/(d).
- **DI idiom (arm c)** — an adjacent run of ≥ 2 `func`-typed parameters: an injected-closure list that
  belongs in a named spec struct. **Not** a swap rule (0 identical-signature func runs measured).
- **Allow-list** — the committed, empty-at-delivery ratchet of reviewed exemptions; fail-on-stale,
  refuse-to-grow.

## 5. Success criteria

### Measurable outcomes

- **SC-001**: the committed scan, run on `dev` `0231f64`, reds with exactly the **42** sites its report
  prints (W1). [Verification Intent: unobservable → W1's quoted output]
- **SC-002**: on the round's tree the gate reports **0** sites and the allow-list is empty (W2).
  [Verification Intent: unobservable → W2]
- **SC-003**: `make verify` OK (now twelve members), E2E **334 / 2519 unchanged and green**, `test-race`
  clean, `modelith-check` no drift, `modelith-lint` 0/0, `verify-adr-index` consistent.
  [Verification Intent: observable → `make check-full`]
- **SC-004**: witness W3/W4/W5 each red **exactly one line** and are **arity-independent** (run-of-3);
  W6 (a synthetic stale allow-list entry) reds; W7 (a violation against a warm cache) reds; W10 (an
  unkeyed same-package literal) reds. [Verification Intent: unobservable → the recorded mutation runs]
- **SC-005**: `internal/cli`'s largest production file is **≤ 600** lines (1682 → 474).
  [Verification Intent: observable → `wc -l`]
- **SC-006**: `go.mod`/`go.sum` unchanged; no new `make verify` member beyond `verify-params` itself.
  [Verification Intent: observable → the diff]

## 6. Edge cases

- **EC-001**: a run of `context.Context` parameters MUST NOT red (`ctx, cancel`-style or `(ctx1, ctx2)`).
  [Verification Intent: unobservable → the exemption unit pins in the gate + the pre/post figures]
- **EC-002**: a run of `interface{}` / `any` MUST NOT red. [Verification Intent: unobservable → the same]
- **EC-003**: a lone trailing variadic MUST NOT extend a run. [Verification Intent: unobservable → the
  same]
- **EC-004**: a **non-scalar** adjacent identical-type run (`io.ReadCloser`, `chan struct{}`) MUST red —
  a "scalar-only" rule would miss the two measured non-scalar sites. [Verification Intent: unobservable →
  W4]
- **EC-005**: an **empty/absent** allow-list MUST NOT be treated as "nothing allowed" or as "no check
  configured" — a missing or unreadable allow-list is a hard failure. [Verification Intent: unobservable →
  the gate's reader]
- **EC-006**: a cross-package unkeyed literal is **out of scope** for arm (d) — `go vet`'s `composites`
  analyser already covers it, and duplicating it would red on every `time.Time{…}`-class literal.
  [Verification Intent: unobservable → the arm-(d) pre/post figures]

## 7. Invariants

- **I-1** — the round is **behaviour-neutral**: identical rendered bytes, exit codes and persisted state.
- **I-2** — the guard is **red-first**: it is never green before the tree is fixed.
- **I-3** — the numbers in this plan come from the **script**, never from prose.
- **I-4** — no accepted ADR **body** is edited (ADR 0068 is new; ADR 0011/0020/0041/0012 are **cited**).
- **I-5** — frozen `specs/plans/NNN-*/**` packages are never modified.

## 8. Assumptions

- **A1** — **option A** (fix every site) is chosen over option B (hot seams only) and over dropping arm
  (c): an allow-list on day one would hollow out the guard, and the five injected-closure lists are the
  repo's own named-spec-struct idiom in waiting.
- **A2** — the `n ≥ 4` arity boundary is **kept** (not widened): the measured remainder (6 interface
  methods + 33 `FuncDecl`s, plus 11 result-run functions) is recorded in ADR 0068 §Forward rather than
  absorbed.
- **A3** — the guard is a **build-tagged test**, not a `go vet` analyser: the parser walk needs no
  `go list` and no host build context, so it is cheaper and hermetic by construction. The earlier
  build-tag *justification* for this choice was **false** and is corrected in ADR 0068 D7.
- **A4** — the domain model moves only in `quality.modelith.yaml` (the `QualityPipeline` list + the
  `QualityGate` examples): the product model's `usage-summed-for-the-ready-summary` concerns values, not
  parameter lists (ADR 0041 requires the question be **resolved**, not assumed).
