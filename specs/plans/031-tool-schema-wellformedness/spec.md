# Feature Specification: tool-schema well-formedness (round 031)

**Feature Branch**: `031-tool-schema-wellformedness`

**Created**: 2026-09-16

**Status**: Draft — scope resolved from issue [#64](https://github.com/gosharplite/tellme/issues/64). Clarify Round 1 locked **Q1 → 1** (hermetic tool-schema well-formedness gate + a **manual** live closeout check) and **Q2 → 1** (minimal fix + gate).

**Input**: Operator request: *"kick off 031 for #64."*

**Scope note**: this is a **schema-correctness + verification-gate** round (like round 020 in shape). It corrects the **advertised argument schema** every tool sends to the model so that `required ⊆ properties` holds for all of them, and adds an automated well-formedness check so the class cannot silently recur. It does **not** change any tool's behaviour, the offered tool set, the request/response wire, or the frozen class-phrase / exit-code vocabulary. It does **not** add a security/consent gate.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Every tool tellme advertises is schema self-consistent (Priority: P1)

As an operator who drives tellme against a **strict provider** (Vertex/Gemini), I want every tool tellme advertises to the model to describe its arguments consistently — every argument it marks **mandatory** is one it actually **offers** — so the provider accepts my requests instead of rejecting the whole call.

**Why this priority**: This is the defect itself. When any advertised tool marks an argument mandatory without offering it, a strict provider rejects **every** request — including prompts that use no tool, because tellme attaches all tool declarations on every request. It currently blocks the entire Vertex/Gemini capability (round 013) and the `#60` dogfooding track. It is the round's core value; US2 can ship without it, not the reverse.

**Independent verification**: render every registered tool's advertised schema and assert that each tool's mandatory-argument set is a subset of its offered-argument set (and that the shared `reason` argument is offered by every tool). Optionally confirm against a real Vertex/Gemini endpoint.

**Acceptance Scenarios**:

1. **Given** an operator drives tellme with a strict provider (Vertex/Gemini), **When** they send any prompt — with or without a tool — **Then** tellme completes the turn instead of failing with the provider-failure phrase because a tool argument it marked mandatory was undefined.
2. **Given** tellme advertises its tool set to the model, **When** each tool's advertised argument description is inspected, **Then** every argument the tool marks mandatory is also offered among the tool's arguments.

**Functional Requirements (FR)**:

- **FR-001**: Every tool tellme advertises MUST carry a well-formed argument schema — a JSON object that declares its **offered** arguments (properties) together with its **mandatory** arguments (required).
- **FR-002**: For every advertised tool, the mandatory-argument set MUST be a **subset** of the offered-argument set (`required ⊆ properties`).
- **FR-003**: The shared `reason` argument — mandatory for every tool (FR-012) — MUST be **offered** (declared) by every tool, not merely marked mandatory.
- **FR-004**: Correcting the schemas MUST NOT change any tool's name, description, the offered-tool set, or its execution semantics. `reason` remains **mandatory-by-schema** and **unvalidated at execution** (round-024/029 FR-012 unchanged).

**Non-Functional Requirements (NFR)**:

- **NFR-001**: The round MUST add no new third-party module (Go standard library only) and MUST stay POSIX-only (tellme's standing scope).

---

### User Story 2 - A malformed tool schema cannot land silently (Priority: P2)

As a maintainer, I want an automated check that fails whenever any offered tool marks an argument mandatory without offering it, so a new or edited tool cannot reintroduce this class of defect unnoticed.

**Why this priority**: This is the durable protection. US1 fixes the present defect; US2 ensures a future tool cannot silently regress the same way — which is exactly how the defect shipped (round-024 `cfa005c`) and survived every gate (the fakes never schema-validate, and non-strict providers tolerate the mismatch).

**Independent verification**: run the check against the current toolset (passes); introduce a temporary tool whose schema marks an argument mandatory but never offers it (the check fails, naming the tool); revert.

**Acceptance Scenarios**:

1. **Given** the set of tools tellme offers, **When** the well-formedness check runs, **Then** it passes only if every tool's mandatory-argument set is a subset of its offered-argument set **and** every schema parses.
2. **Given** a tool that marks an argument mandatory but never offers it, **When** the check runs, **Then** the check fails and names the offending tool.

**Functional Requirements (FR)**:

- **FR-005**: The project's automated verification MUST include a well-formedness check that asserts FR-002 over the **complete** set of registered tools — **not a subset** (today's blind-spot test covers only three of the six tools).
- **FR-006**: The check MUST cover **every** registered tool, including the command tool, with no exemptions.
- **FR-007**: The check MUST reject a schema that does not parse, or that is not a JSON object carrying a declared-properties section.

**Non-Functional Requirements (NFR)**:

- **NFR-002**: The check MUST run **hermetically** (offline, no network) as part of the ordinary test suite.

---

### Edge cases

- **A tool with zero mandatory arguments** → vacuously well-formed (the check MUST NOT fail).
- **A schema that is not a JSON object, or omits its declared-properties or mandatory sections** → the check MUST fail (FR-007).
- **A newly added mandatory argument whose offered property is never declared** → the check MUST fail and name the tool (US2).
- **A new tool added to the registry** → automatically covered by the check, with no per-tool registration required (FR-005/FR-006).

## Requirements *(mandatory)*

> The per-story FR / NFR are attached under each story above; this section holds only requirements that constrain both stories or cannot be reasonably attributed to a single story.

### Global requirements

#### Functional Requirements

- **FR-008**: The round MUST NOT change the frozen class-phrase vocabulary or the exit-code set; the defect is corrected at the schema source, so **no** new error surface is introduced.
- **FR-009**: Each corrected schema MUST retain the resource parameters it already advertises (`max_output_tokens`, `timeout`) exactly as before, adding only the missing `reason` property; the offered-tool set remains exactly the six agent tools.

#### Non-Functional Requirements

- **NFR-003**: Verification is **except** for the manual live confirmation, hermetic (Q1 → 1): the well-formedness check runs offline in the ordinary test suite, and the real Vertex/Gemini confirmation is a **manual closeout step**, not part of the automated gate.

### Key entities

- **Advertised tool schema**: the JSON description tellme sends to the model for each tool, listing the arguments it **offers** and the ones it marks **mandatory**. This round corrects its well-formedness (`required ⊆ properties`) and adds the missing `reason` property. No persisted data changes; the round introduces no new file format, table, or record.

## Success criteria *(mandatory)*

### Measurable outcomes

- **SC-001**: 100% of the registered tools (all six) render a schema whose mandatory set is a subset of its offered set, as asserted by the automated check (covers FR-001, FR-002, FR-005, FR-006).
- **SC-002**: A live Vertex/Gemini call — a plain prompt **and** a tool-using prompt — completes without a `400` "required fields … are not defined in the schema properties" (manual closeout confirmation; covers the operator-visible outcome of US1).
- **SC-003**: `make verify`, the E2E suite, and the Gherkin/DSL topology audit are green, with a **falsifiability witness**: introducing a tool with a mandatory-but-undeclared argument fails the check (then reverted) (covers US2).
- **SC-004**: The offered-tool set is unchanged (exactly the six tools) with unchanged names/descriptions/execution semantics, and the class-phrase / exit-code vocabulary is unchanged (covers FR-004, FR-008, FR-009).

## Assumptions

- The strict provider's rejection is caused **solely** by `required ⊄ properties`; a well-formed schema makes the request acceptable (confirmed by the manual live call, SC-002).
- `reason` stays **mandatory-by-schema** for every tool (round-024/029 FR-012) and is advertised only — never validated at execution.
- Per Q1 → 1, the live Vertex/Gemini confirmation is a **manual closeout check**, not part of `make verify`; no network enters the gate.
- Broader schema-conformance / provider-strictness modelling (e.g. teaching the hermetic E2E fake to schema-validate) is **out of scope** this round; it is recorded as a forward item and coordinated with the `#60` dogfooding track.
- Per Q2 → 1, the fix is **minimal**: the shared schema construction declares the `reason` property so every tool it builds is well-formed; `execute_command`, which builds its schema inline and already complies, is left unchanged. A structural "impossible-by-construction" rework is a recorded forward item.
- This is a **tooling/gate-shaped** round (round-020 precedent): `/axb-spec-by-example` is likely **skipped** (no new business journey), and `/axb-api-plan` / `/axb-data-plan` / `/axb-dsl-refine` are expected **NOOP**; the truth change is limited to `specs/truth/techstack.md` (the tool-schema rows), made by the truth owner (`fresh-package-per-round`).
