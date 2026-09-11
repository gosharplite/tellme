# Feature Specification: tellme Follow-up Cleanups (round 002)

**Feature Branch**: `002-followup-cleanups`

**Created**: 2026-09-11

**Status**: Draft

**Input**: User description: "Let's have a very small 002-* slice and clean up F4, F9 and all other small things." — refined by a clarify round (2026-09-11): **Q1 → remove the `--json` flag entirely** (no machine-readable diagnostic mode at all); **Q2 → also (1) adopt quality-gate hardening, (2) pin the NFR-004 error-message wording, and (3) pin the FR-014 exit-code numeric values.** F9 (pure-helper unit tests) is in scope as stated.

*(Revision note: this round **reverses round-001 FR-013** by removing the `--json` diagnostic flag — a user-ratified `DELETE`. The frozen round-001 package `specs/plans/001-cli-bootstrap-and-config/**` is left untouched as history; this round supersedes the behavior in `specs/truth/**`.)*

*(Revision — **grill round #6** (issue [#8](https://github.com/gosharplite/tellme/issues/8)), 2026-09-11: the failure-message freeze is a **class phrase** (`tellme: {phrase}`), **not** a verbatim line — the trailing detail (a path, a provider name, an OS error string) is contract-free. FR-004/FR-006 and SC-002/SC-003 are reworded accordingly, and the class list is reconciled to the eight emitted `tellme:` forms (seven classes — the two "runtime home" forms share one phrase; the "workspace path is not a directory" form is added). SC-003's "fails on a deliberately introduced ignored error" is now a **recorded manual witness**. Grill verdict: *proceed with changes*.)*

## User Scenarios & Testing *(mandatory)*

### User Story 1 - A single, well-defined diagnostic output (remove `--json`) (Priority: P1)

As an operator running tellme's diagnostic, I want exactly one output form and no flag that silently does nothing, so that what I ask for is what I get.

**Why this priority**: It removes a shipped but half-supported flag (round-001 review finding **F4** — `--json` was silently ignored when given without `-d`). Removing it is the smallest change that eliminates the silent surprise and simplifies the diagnostic contract to a single, honest path.

**Independent verification**: Run `tellme --json` and `tellme -d --json` and confirm both are rejected as usage errors; run `tellme -d` on a resolved and on an unresolved setup and confirm the plain report plus the retained exit behavior.

**Acceptance Scenarios**:

1. **Given** a resolved setup, **When** I run tellme with `--json` alone, **Then** it refuses to proceed as a usage error and never boots.
2. **Given** a resolved setup, **When** I run the diagnostic with `--json`, **Then** it refuses to proceed as a usage error.
3. **Given** a resolved setup, **When** I run the diagnostic without `--json`, **Then** it reports the resolution in plain form and exits successfully.
4. **Given** an unresolved setup, **When** I run the diagnostic without `--json`, **Then** it reports that the setup did not resolve and exits with the dedicated diagnostic error code.

**Functional Requirements**:

- **FR-001**: The system MUST NOT accept a `--json` flag — the flag is removed, and there is no machine-readable diagnostic output mode.
- **FR-002**: Any occurrence of `--json` (alone, or combined with `-d` or `--version`) MUST be treated as an unrecognized-flag usage error — never silently ignored.
- **FR-003**: The plain diagnostic (`-d`) MUST retain its round-001 behavior: a plain report; success when the setup resolves; the dedicated non-zero "unresolved" code when it does not.

**Non-Functional Requirements**:

- **NFR-001**: Every diagnostic path MUST remain offline and deterministic — no provider contact and no network access (carried from round-001 NFR-003).

---

### User Story 2 - A stable, documented operator-facing failure contract (Priority: P2)

As an operator or script author, I want tellme's failure messages and numeric exit codes to be fixed and documented, so my automation can depend on them across releases.

**Why this priority**: Round 001 deliberately left the exact message wording (`NFR-004`) and the exit-code numeric values (`FR-014`) unfixed. Once the CLI is a released contract, callers need stability; this freezes both.

**Independent verification**: Trigger each documented failure class, then diff stderr and the process exit code against the documented message/code table.

**Acceptance Scenarios**:

1. **Given** a setup that fails in a documented way, **When** I run tellme, **Then** stderr carries the documented **class phrase** for that failure class — its `tellme: {phrase}` line — and emits no other `tellme: `-prefixed line.
2. **Given** a setup that fails in a documented way, **When** I run tellme, **Then** the process exits the documented numeric code for that class.
3. **Given** the documented operator-facing contract, **When** the acceptance tests run, **Then** each documented **class phrase** and each exit code is asserted — the class phrase as the `tellme: {phrase}` prefix of exactly one stderr line, and the code as the pinned numeric value.

**Functional Requirements**:

- **FR-004**: The system MUST emit a fixed, exact operator-facing stderr **class phrase** for each failure class — the line beginning `tellme: ` that names the class. Any class-specific trailing detail (a path, a provider name, an OS error string) is **not part of the frozen contract**. The frozen class phrases are: `the configuration could not be found` (missing via `-c`); `no configuration could be found` (missing default); `the configuration could not be parsed`; `the selected provider is not in the registry`; `the workspace path is not a directory`; `the runtime home is not usable` (a single class covering both the *not usable* and the *unset* paths); and `the command-line usage is invalid`.
- **FR-005**: The system MUST use fixed numeric exit codes for: success, usage error, configuration error, environment error, and diagnostic "unresolved".
- **FR-006**: The fixed **class phrases** and the exit-code values MUST be recorded as the operator-facing contract and asserted at **class-phrase granularity** by the acceptance set — each failure class's stderr line MUST begin `tellme: {phrase}`; the trailing detail is not asserted.

**Non-Functional Requirements**:

- **NFR-002**: Failure messages MUST remain actionable — stating what was expected and where — and MUST be written to stderr.

---

### Edge Cases

- When `--json` appears together with any other flag (including `-d` and `--version`), it MUST be a usage error.
- When an operator asks for machine-readable output by any means, the system MUST NOT emit JSON on the diagnostic path (the capability no longer exists) and MUST NOT silently fall back to another path.
- When more than one failure class could apply, the documented round-001 resolver precedence MUST determine the single message and code emitted.
- When a documented **class phrase** or exit code changes in a later round, the contract and its class-phrase assertions MUST change together.

## Requirements *(mandatory)*

> Story-specific FR/NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Non-Functional Requirements

- **NFR-003**: The pure resolution rules — the effective mode, the effective selected provider, provider-registry membership, and workspace creation/reuse — MUST be covered by fast, isolated, table-driven unit tests that run offline and deterministically, complementing (never replacing) the end-to-end acceptance path (round-001 review finding **F9**).
- **NFR-004**: The bootstrap/diagnostic code paths MUST handle their error returns; the project's verification MUST fail when an error return is ignored.
- **NFR-005**: The project's verification MUST include a dependency-vulnerability check that fails on a known vulnerability.
- **NFR-006**: The round-001 end-to-end acceptance MUST remain green after the `--json` removal and the contract freeze (no regression).

### Key Entities *(include if feature involves data)*

- **Diagnostic contract**: the single plain output form of `-d` (resolved / unresolved) and its retained exit behavior; there is no machine-readable form.
- **Exit-code contract**: the fixed numeric code table — success, usage error, configuration error, environment error, diagnostic "unresolved".
- **Operator-facing message catalog**: the fixed stderr **class phrases**, one per failure class (the frozen prefix `tellme: {phrase}`; the trailing detail is contract-free).
- Round 002 adds no `contracts/**` (no API surface) and no `data/**` model — it is CLI-end behavior owned by `/axb-dsl-refine` under `specs/truth/features/cli/**`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the acceptance set, 100% of `--json` invocations (alone or with `-d`) exit with the usage-error code, and `-d` without `--json` still reports resolved and unresolved correctly.
- **SC-002**: In the acceptance set, 100% of documented failure **class phrases** are asserted as the `tellme: {phrase}` prefix of the failure's stderr line (with **exactly one** `tellme: `-prefixed line emitted per failing run), and 100% of documented exit codes match the running binary.
- **SC-003**: The pure resolution rules are covered by unit tests; `make verify` passes on the clean tree; and the ignored-error gate (`errcheck`, via `make lint`) exits non-zero on a deliberately introduced unchecked error — **witnessed once, manually, and recorded in the round's evidence** (round-001 SC-004 shape).

## Assumptions

- The `--json` removal reverses round-001 **FR-013** (`-d --json` machine-readable output); it is ratified by the user in the round-002 clarify round (Q1). Round-001's plan package stays frozen history.
- Once `--json` is not a flag, its rejection is covered by the existing unrecognized-flag usage rule (the same class as `--wibble`); no new rejection mechanism is introduced.
- The exact message strings and the numeric exit-code values are operator-facing contract detail; this round fixes them, and the concrete values are recorded by `/axb-dsl-refine` in the interface truth (`specs/truth/features/cli/**`).
- The choice of tooling for the ignored-error and vulnerability gates (for example a linter with unchecked-error checking, or a vulnerability scanner) is an RD decision recorded in plan-side `research.md` and `specs/truth/techstack.md`, not a spec requirement.
- Round 002 adds no API surface, no data model, and no new external dependency on the product path; provider calls, the reasoning model, tools, MCP, memory, TUI, and history persistence remain out of scope (as in round 001).
