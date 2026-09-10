# Feature Specification: tellme CLI Bootstrap & Configuration

**Feature Branch**: `001-cli-bootstrap-and-config`

**Created**: 2026-09-10

**Status**: Draft

**Input**: User description: "Create the first iteration of tellme — a narrow foundation slice. The CLI must boot, locate and validate its YAML configuration and runtime home, initialize a per-mode session workspace, and expose its build version plus an offline setup diagnostic. No provider calls, tools, MCP, memory, TUI, or history persistence this round."

*(Revision — clarify Q2 **released** the "treated as a slice-local input, not a system truth artifact" scope lock and authorized a **minimal** data truth; that minimal data truth was then **withdrawn** — grill round #3 argued its content out of `data/**`, and the user ratified reversing clarify Q2 on 2026-09-11. Round 001 therefore owes **no** `data/**` model: the config input contract and the workspace lifecycle are executable CLI-contract behaviour owned by `/axb-dsl-refine` under `specs/truth/features/cli/**`.)*

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Boot with a valid configuration (Priority: P1)

As a user launching tellme, I want the tool to locate and load my YAML configuration and tell me clearly whether it is usable, so I can start the tool with confidence and fix misconfigurations without guessing.

**Why this priority**: Nothing runs without a resolved and validated configuration. Every later capability (provider calls, sessions, tools) depends on this boot path, so it is the first value to prove.

**Independent verification**: Run the CLI against a valid configuration and against several invalid ones; confirm a ready outcome (success exit) for the valid case and a distinct, actionable failure (non-zero exit, stderr message) for each invalid case.

**Acceptance Scenarios**:

1. **Given** a valid YAML configuration file, **When** I start tellme pointing at it, **Then** the tool loads and validates it and reports readiness with a success exit code.
2. **Given** a configuration whose selected provider is not present in the provider registry, **When** I start tellme, **Then** it refuses to proceed, explains the mismatch on stderr, and exits with a non-zero code.
3. **Given** a configuration path that does not exist or contains malformed YAML, **When** I start tellme, **Then** it reports an actionable error on stderr and exits with a non-zero code.

**Functional Requirements**:

- **FR-001**: The system MUST accept a configuration file path via a `-c` / `--config` flag.
- **FR-002**: The system MUST load its configuration from a YAML file as the boot-time input.
- **FR-003**: The system MUST resolve the *effective* selected provider — the `TELL_ME_SELECTED_PROVIDER` environment variable when set, otherwise the configuration's `SELECTED_PROVIDER` — and MUST validate that it references an entry in the provider registry.
- **FR-004**: When the configuration is missing, unreadable, or fails validation, the system MUST emit an actionable message on stderr and exit with a non-zero code distinct from the success code.
- **FR-005**: The system MUST support, at minimum, the configuration keys `MODE`, `PERSON`, `SELECTED_PROVIDER`, and a `PROVIDERS` registry that starts with a single provider.

**Non-Functional Requirements**:

- **NFR-001**: Configuration loading and validation MUST be offline and deterministic — no network access is required or attempted.

---

### User Story 2 - Resolve the runtime home and session workspace (Priority: P2)

As a returning user, I want tellme to resolve its runtime home and prepare a per-mode session workspace, so that my state has a stable, predictable location that persists across runs.

**Why this priority**: It establishes where all future persisted state will live. It has no standalone value without a successful boot, but every stateful slice after this one depends on it.

**Independent verification**: Run once with a runtime home set and confirm the per-mode workspace is created; run again and confirm the same workspace is reused rather than recreated or lost.

**Acceptance Scenarios**:

1. **Given** a runtime home directory, **When** I run tellme for the first time, **Then** it creates the per-mode session workspace and reports where it is.
2. **Given** a workspace created by a previous run, **When** I run tellme again, **Then** it reuses the same workspace without losing or recreating it.

**Functional Requirements**:

- **FR-006**: The system MUST resolve its runtime home from the `TELL_ME_HOME` environment variable.
- **FR-007**: The system MUST derive the session workspace path as `output/<mode>/` under the runtime home, where `<mode>` is the *effective* mode — the `TELL_ME_MODE` environment variable when set, otherwise the configuration's `MODE` value.
- **FR-008**: On first run the system MUST create the session workspace; on subsequent runs it MUST reuse the existing one.
- **FR-009**: The system MUST report the resolved workspace path in a human-readable form.

**Non-Functional Requirements**:

- **NFR-002**: Workspace initialization MUST be idempotent — repeated runs MUST NOT error, and MUST NOT destroy existing workspace state.

---

### User Story 3 - Inspect the build version and diagnose setup (Priority: P3)

As an operator, I want to see the running build version and run a setup diagnostic that reports configuration and home resolution, so I can confirm what I am running and pinpoint where a problem lies.

**Why this priority**: Observability is only useful once boot and home resolution work; it is a convenience layer over the P1/P2 behavior.

**Independent verification**: Run the version flag; run the diagnostic command and its machine-readable form; confirm the reported version and resolution status.

**Acceptance Scenarios**:

1. **Given** any environment, **When** I run tellme with the version flag, **Then** it prints the build version and exits successfully.
2. **Given** a resolved configuration and runtime home, **When** I run tellme's diagnostic command, **Then** it reports configuration-resolution and home-resolution status without contacting any provider or network.
3. **Given** machine-readable output is requested, **When** I run the diagnostic command with the JSON flag, **Then** it emits structured output.

**Functional Requirements**:

- **FR-010**: The system MUST expose its build version via a version flag (`--version`).
- **FR-011**: The system MUST expose a diagnostic command (`-d`) that reports configuration- and home-resolution status.
- **FR-012**: The diagnostic command MUST NOT perform any network access.
- **FR-013**: The diagnostic command MUST support a machine-readable (`--json`) output mode.

---

### Edge Cases

- When no configuration path is given and no default configuration can be found, the system MUST report an actionable error and exit non-zero rather than proceeding with an empty configuration.
- When a configuration parses but its `PROVIDERS` registry is empty, the system MUST treat the selected-provider validation as failed.
- When an unrecognized flag is supplied, the system MUST report a usage error and exit with a code distinct from both success and configuration failure.
- When `TELL_ME_HOME` is unset or points to a non-writable location, the system MUST report an actionable environment error and exit non-zero.
- When `TELL_ME_MODE` / `TELL_ME_SELECTED_PROVIDER` are set, they MUST take precedence over the configuration's `MODE` / `SELECTED_PROVIDER`; when an override names an unknown provider, the same actionable validation failure MUST occur as for an in-file mismatch.
- When the resolved workspace path already exists as a regular file rather than a directory, the system MUST fail with an actionable error instead of overwriting it.
- When the setup does not fully resolve, the diagnostic command MUST still produce a report (resolved or unresolved) and MUST exit with a dedicated, distinct non-zero "diagnostic: unresolved" code; the `FR-014` success/usage/configuration/environment codes bind the boot path only.

## Requirements *(mandatory)*

> Story-specific FR/NFR are listed under each story above; this section holds only requirements that constrain multiple stories or cannot be reasonably attributed to a single story.

### Global Requirements

#### Functional Requirements

- **FR-014**: The system MUST return distinct, deterministic exit codes for at least: success, usage error, configuration error, and environment error.
- **FR-015**: The system MUST give `TELL_ME_*` environment variables precedence over the YAML configuration when resolving the effective mode (`TELL_ME_MODE`) and the effective selected provider (`TELL_ME_SELECTED_PROVIDER`); this constrains both the configuration story and the workspace story.

#### Non-Functional Requirements

- **NFR-003**: All bootstrap, configuration-resolution, and diagnostic behavior MUST be offline and deterministic — no provider contact or network access occurs during these paths.
- **NFR-004**: Failure messages MUST be actionable — stating what was expected and where — and MUST be written to stderr. [NEEDS CLARIFICATION: exact message wording/format is not yet fixed]

### Key Entities *(include if feature involves data)*

- **Configuration**: The boot-time YAML input that bootstraps a run. Carries at least `MODE`, `PERSON`, `SELECTED_PROVIDER`, and a `PROVIDERS` registry. No `contracts/**` (no API surface) and no `data/**` model this round — the config input contract and the workspace lifecycle are executable CLI-contract behaviour owned by `/axb-dsl-refine` under `specs/truth/features/cli/**` (grill round #3 + user ratification).
- **Runtime Home**: The `TELL_ME_HOME` root directory under which all tellme state is namespaced.
- **Session Workspace**: The per-mode directory `output/<mode>/` under the runtime home — the future home of session state.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In the acceptance set, 100% of valid configurations launch to readiness, while 100% of missing, invalid, or misconfigured configurations exit non-zero with an actionable stderr message.
- **SC-002**: In the acceptance set, after the first run `output/<mode>/` exists, and every subsequent run reuses it with no state loss.
- **SC-003**: In the acceptance set, the version flag and the diagnostic command each produce their expected output and exit successfully.
- **SC-004**: The diagnostic command performs zero network calls in the acceptance environment.

## Assumptions

- Round 1 is invoked as a single binary/command named `tellme`.
- When `-c`/`--config` is omitted, the default configuration is sought at `$TELL_ME_HOME/configs/<mode>.yaml`, where `<mode>` is the effective mode; when no effective mode is set, `<mode>` defaults to `butler` (`$TELL_ME_HOME/configs/butler.yaml`), matching the tell-me-go reference default.
- Runtime-home and mode/provider resolution is aligned with the tell-me-go Niffler shell environment: Niffler exports `TELL_ME_HOME=ait-<tag>/`, `TELL_ME_MODE=<role>`, and `TELL_ME_SELECTED_PROVIDER=<provider>`, and always passes an explicit `-c "$TELL_ME_HOME/configs/<mode>.yaml"`. tellme therefore resolves home from `TELL_ME_HOME`, honours the `TELL_ME_*` overrides with precedence over the file, and uses the `output/<mode>/` layout that Niffler already provisions under the tag workspace.
- Integration note (binary name): Niffler currently invokes a binary named `tell-me-go` (`$NIFFLER_GOBIN/tell-me-go`, and `tell-me-go completion bash`), whereas tellme's binary is named `tellme`. Running tellme under an unmodified Niffler requires a matching binary/alias or a Niffler-side adaptation; reconciling this is out of scope for this round.
- Round 1's provider registry contains exactly one provider; multi-provider support belongs to a later slice.
- The implementation language is Go, mirroring the tell-me-go reference; the exact toolchain and test stack are fixed later by technical research (techstack truth).
- Configuration is a boot-time input; this round adds no `contracts/**` (no API surface) and **no** `data/**` model — the config input contract and the workspace lifecycle are executable CLI-contract behaviour owned by `/axb-dsl-refine` (`specs/truth/features/cli/**`). (Clarify Q2 released the earlier "no `data/**`" lock; grill round #3 + user ratification withdrew the minimal data truth that release had authorized.)
- Out of scope this round: provider API calls, the reasoning/`Thought` model, `Turn`/`History` persistence, cost/metrics, tools, MCP, memory, TUI prompts, browsing, retry/edit, and callback delivery.
