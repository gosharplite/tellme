# Phase 0 Research: tool-schema well-formedness (round 031)

**Topic**: how tellme guarantees every agent tool's **advertised argument schema** is well-formed — declaring the mandatory `reason` property so `required ⊆ properties` holds for all six tools, and adding an automated check so the defect (issue [#64](https://github.com/gosharplite/tellme/issues/64)) cannot silently recur. Each decision supports `spec.md` (US1/US2 · FR-001..FR-009) and the operator-locked clarifications (**Q1 → hermetic registry gate + manual live check**; **Q2 → minimal fix + gate**).

**Must-ask questions (settled).** Per the AIxBDD three must-asks, all three are already written in the existing `specs/truth/techstack.md` and are **unchanged** this round: the system has a **single CLI end** (no web frontend, no HTTP server); the BDD techstack is **`godog`** (Cucumber for Go) driving the interface Gherkin E2E against the built binary; the test strategy is **E2E** for the acceptance path plus fast unit tests for pure helpers. No re-ask is warranted (`Rule 2` of the must-ask rule).

## Decision 1: Fix at the **shared schema builder** — declare the `reason` property there

- **Decision**: correct the defect at its single source — the shared schema builder (`resourceSchema`, `internal/infrastructure/tools/filesystem.go`) — by declaring the `reason` property in the `properties` object it emits. Because **every** caller already passes `"reason"` in `required` and `reason` is mandatory for every tool (round-021/FR-012), the builder declares the `reason` property **unconditionally**; each caller keeps passing only its bespoke props + the required list. Result: `required ⊆ properties` holds for every tool the builder backs (`list_files`, `read_files`, `get_tree`, `write_file`, `replace_text`). `execute_command` builds its schema inline and **already declares `reason`**, so it is the lone compliant tool and needs no change (Decision 3).
- **Rationale**: the regression was introduced **by** the shared builder — round-024 fold `cfa005c` replaced inline schemas (which correctly declared `reason`) with a builder that emits only `{<extraProps>, max_output_tokens, timeout}` while callers kept passing `"reason"` as required; round-029 fold `eb0367c` renamed it and extended it to the write pair, carrying the omission forward. Fixing the builder fixes all five affected tools at once and prevents the same omission from being reintroduced by a future caller — the smallest change that restores `required ⊆ properties` for the whole tool set.
- **Alternatives considered**:
  - **Re-add `reason` in each caller's `extraProps`** — rejected: it would re-declare the same property in five call sites and let a future caller forget it again (the exact failure mode); the builder already owns the resource params, so it should own `reason` too.
  - **Derive `required` from the declared properties** so the two cannot diverge structurally — rejected for this round (Q2 → 1): a larger rework of the schema layer than the defect warrants; recorded as a forward item (Decision 6).

## Decision 2: The well-formedness gate is a **unit test over the non-overridable production assembler**

- **Decision**: add a unit check that iterates the **production tool assembler `agentTools()`** (`internal/cli`; the plain, non-overridable func returning the six tools) — **not** the `newToolRegistry` var (a DI seam that "tests may override") — and asserts, for **every** tool, that each name in the schema's `required` is declared under `properties` **and** that the schema decodes as a JSON object. It sits beside `TestNewToolRegistryOffersAgentTools`. **ARCH-1**: a recurrence gate must not sit on a mutable seam — if a future test overrides `newToolRegistry` without restoring it, a gate reading the var would test a fake registry and pass **vacuously**. (This is also why T002 — which lives in `internal/infrastructure/tools` and **cannot** import `internal/cli` — must **not** re-hand-enumerate the six tools: **completeness is T001's job** — **ARCH-3**.)
- **Rationale**: `agentTools()` is the authoritative "every tool the model is offered" backing (the same set `newToolRegistry` wraps). Running the invariant there covers all six tools (including the write pair and `execute_command`, which today's blind-spot test never touches) and couples directly to what the transport actually sends, so it is the check that would have caught #64.
- **Recorded precondition (ARCH-2)**: the walk is **root-level only** — every schema is flat today (`read_files` nests only an `items` object's **properties**, with no nested `required`), so the invariant holds; a nested object carrying its own `required` would evade a root-only walk. The recursive form (walk `properties.*` / `items`, depth-bounded) is a recorded **`#60` forward item**, not silently assumed.
- **Alternatives considered**:
  - **Read the `newToolRegistry` var** — rejected (**ARCH-1**): it is overridable, so the gate could read a test double and pass vacuously — fatal for a recurrence guard.
  - **A test in `internal/infrastructure/tools`** over `NewFilesystemTools()` + the write pair + the command tool — rejected as the *primary* gate (**ARCH-3**): it enumerates the tools by hand rather than reading the one assembler the CLI wires, so a newly added tool could be omitted; the assembler-backed test cannot miss one. (Its unique `reason`-specific value is kept, scoped to the builder-backed tools, in T002.)
  - **A runtime assertion in `NewRegistry`** (panic/log on a malformed schema) — rejected for this round: the schema is built per call and a panic at construction adds a production failure mode for a defect a hermetic test already blocks; recorded as a forward item (Decision 6).

## Decision 3: Leave `execute_command`'s inline schema in place

- **Decision**: do **not** fold `execute_command` onto the shared builder this round. It already declares `reason` (it is the only tool that complies today) and keeps its bespoke `output_file`/`append` props and its process-tree-specific `timeout` wording.
- **Rationale**: it is already well-formed, so touching it adds risk without fixing anything; the shared builder now single-sources the `reason` property for the five tools it backs. Aligning the command tool to the builder is a cosmetic refactor that would also churn its distinct `timeout` description — out of scope for a defect fix (Q2 → 1).
- **Alternatives considered**:
  - **Align `execute_command` to the shared builder now** — rejected: no correctness gain (it already satisfies the invariant) and it would change a working, reviewed schema's wording; recorded as an optional future tidy.

## Decision 4: Verification posture — **hermetic registry gate + a manual live confirmation**

- **Decision**: the automated verification is **hermetic** (offline): the round relies on the unit gate (Decision 2) plus the existing E2E/unit suites; `make verify` stays network-free. A **real Vertex/Gemini call** — a plain prompt and a tool-using prompt, no `400` — is verified **manually at closeout** (round-013 style, Q1 → 1), not inside the gate. **Recorded divergence (ARCH-5)**: this makes the round's acceptance (US1/US2) carried by the **registry unit gate + the manual live check**, **not** by executable Gherkin (`/axb-spec-by-example` skipped, `/axb-dsl-refine` NOOP) — the tool schemas are not observable through the built binary hermetically. Recorded so `acceptance-coverage`'s intent (PM-defined acceptance is executable) is not silently re-interpreted (round-020 precedent).
- **Rationale**: tellme's `make verify` has never carried network, and the recurrence risk is closed hermetically by the registry gate — the defect was a *schema* defect, and the gate asserts the exact invariant the strict provider enforces. The live call confirms the end-to-end operator outcome but is not a deterministic gate step (it needs credentials and a real endpoint).
- **Alternatives considered**:
  - **A build-tagged live E2E leg** (`//go:build e2e_live`, excluded from `make check-full`) — deferred: adds harness scope for a confirmation the manual closeout check already provides; recorded as a forward item (Decision 6, coordinate with `#60`).
  - **Teach the hermetic E2E fake to schema-validate** — deferred: a broader harness change; the registry gate asserts the same invariant more directly; recorded as a forward item (Decision 6).

## Decision 5: No new dependency; stdlib-only, POSIX-only

- **Decision**: the fix and the gate use only the Go standard library (`encoding/json`, `testing`); no new module; POSIX-only (tellme's standing scope). `go.mod`/`go.sum` are untouched.
- **Rationale**: the change parses JSON tool schemas and asserts a set relationship — nothing beyond the stdlib is needed; the standing scope forbids a Windows branch and a new dependency for a defect fix.
- **Alternatives considered**:
  - **A JSON-Schema validation library** — rejected: heavyweight for one `required ⊆ properties` invariant; the stdlib decode is sufficient and dependency-free.

## Decision 6: Broader recurrence guards are deferred (recorded), coordinated with `#60`

- **Decision**: three stronger guards are **out of scope** this round and recorded as forward items: (a) teaching the hermetic E2E fake provider to schema-validate tool declarations; (b) a live (non-hermetic) Vertex/Gemini leg inside the pipeline; (c) a structural schema builder / runtime registration assertion that makes `required ⊆ properties` impossible by construction. They are noted for the `#60` dogfooding track ("real-endpoint defects hermetic fakes cannot surface").
- **Rationale**: the registry gate closes the recurrence class for this invariant hermetically; the broader guards are larger harness/architecture work the defect does not require (Q1 → 1, Q2 → 1). Recording them keeps the boundary explicit and hands `#60` a concrete follow-on.
- **Alternatives considered**:
  - **Fold (a)/(b)/(c) into this round** — rejected: expands a small correctness round into harness/architecture work; the boundary is a deliberate scope decision.

## Truth impact (for the truth-owner skills)

- `specs/truth/techstack.md` → **MODIFY**: add an **Agent tool schemas** row (the shared schema builder declares the `reason` property so `required ⊆ properties` holds for every tool; `execute_command` inline-compliant) and an **Agent tool-schema gate** row under Testing & Verification (a hermetic registry-level unit check); note the deferred recurrence guards under *Not Introduced Yet*.
- `specs/truth/features/cli/**` → **NOOP** expected (no new business journey; the verification is a registry unit gate — round-020 precedent), confirmed by `/axb-dsl-refine`.
- `specs/truth/contracts/**` and `specs/truth/data/**` → **NOOP**.
