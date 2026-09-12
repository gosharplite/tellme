# Feature Specification: tellme Provider Registry Completeness (round 003)

**Feature Branch**: `003-provider-registry-completeness`

**Created**: 2026-09-11

**Status**: Draft

**Input**: User request and GitHub Issue #9: "003 — Provider-registry completeness (config input contract)" — refined by Clarify Round 1 (2026-09-11):
- **Q1 -> Option 1**: Core request set (`TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL`); defer `USER_ID`, `THINKING_ENABLED`, and top-level `MODELS` pricing tables.
- **Q2 -> Option 1**: Targeted `${VAR}` and `${VAR:-default}` expansion in `API_KEY`, `URL`, and `HEADERS`; missing variables without defaults fail deterministically at configuration resolution time.
- **Q3 -> Option 1**: Reuse Exit Code `3` (configuration error) with dedicated frozen class phrase `tellme: the provider configuration is invalid`.

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Parse and load complete provider configurations (Priority: P1)

As an operator configuring tellme, I want to declare full, typed provider configurations in `PROVIDERS` (including credentials, custom headers, and reasoning thinking budgets), so that subsequent reasoning slices can construct authentic provider requests.

**Why this priority**: It establishes the foundational schema required for Slice 004 to dispatch real LLM reasoning requests. Without these fields, tellme cannot supply API keys, custom headers, or thinking parameters to provider APIs.

**Independent verification**: Provide a configuration file containing a provider with all core fields populated and verify via tellme boot/diagnostic that the configuration loads cleanly and the fields are recognized.

**Acceptance Scenarios**:

1. **Given** a configuration file with a provider entry containing valid `TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, and `THINKING_LEVEL`, **When** tellme boots or runs diagnostics, **Then** it accepts the configuration as valid and ready.
2. **Given** a configuration file where a provider entry specifies only the mandatory fields (`TYPE`, `MODEL`, `URL`) and omits optional fields (`API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL`), **When** tellme resolves the provider, **Then** it accepts the configuration with default empty/zero values.

**Functional Requirements**:

- **FR-001**: The system MUST recognize and parse the following typed fields within each provider entry under `PROVIDERS`:
  - `TYPE` (string, mandatory)
  - `MODEL` (string, mandatory)
  - `URL` (string, mandatory)
  - `API_KEY` (string, optional)
  - `MAX_TOKENS` (integer, optional)
  - `HEADERS` (map of string to string, optional)
  - `THINKING_BUDGET` (integer, optional)
  - `THINKING_LEVEL` (string, optional)
- **FR-002**: The fields `TYPE`, `MODEL`, and `URL` MUST be present and non-empty strings for any resolved provider entry.
- **FR-003**: The fields `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, and `THINKING_LEVEL` MUST be optional; when omitted, `API_KEY` and `THINKING_LEVEL` default to `""`, `MAX_TOKENS` and `THINKING_BUDGET` default to `0`, and `HEADERS` defaults to an empty map.

---

### User Story 2 - Environment variable expansion in provider settings (Priority: P2)

As an operator managing secrets and dynamic endpoints, I want `${VAR}` and `${VAR:-default}` syntax in provider credentials, URLs, and headers to expand from environment variables, so that I do not hardcode secrets or environment-specific hosts in committed configuration files.

**Why this priority**: Production `tell-me-go` configurations routinely store credentials as `${VAR}` (e.g. `API_KEY: "${DEEPSEEK_API_KEY}"` or `URL: "https://.../${PROJECT_ID}"`). Expanding these deterministically at load time prevents secret leakage and allows container/shell runtime injection.

**Independent verification**: Configure a provider entry with `${KEY}` and `${URL_VAR}` in `API_KEY` and `URL`; run tellme with the environment variables exported and verify successful resolution. Run with an unset variable and verify deterministic failure.

**Acceptance Scenarios**:

1. **Given** a provider entry where `API_KEY`, `URL`, or a `HEADERS` value contains `${VAR}`, **When** the environment variable `VAR` is set and non-empty, **Then** tellme expands `${VAR}` to its environment value during configuration loading.
2. **Given** a provider entry where a field contains `${VAR:-default_value}`, **When** `VAR` is unset or empty in the environment, **Then** tellme substitutes `default_value`.
3. **Given** a provider entry where a field contains `${VAR}` without a default, **When** `VAR` is unset or empty in the environment, **Then** tellme refuses to proceed and reports a provider configuration error.

**Functional Requirements**:

- **FR-004**: The system MUST expand `${VAR}` and `${VAR:-default}` syntax in string values of `API_KEY`, `URL`, and each header value in `HEADERS` for the resolved provider entry.
- **FR-005**: If an environment variable referenced in `${VAR}` has no default and is unset or empty in the environment, configuration resolution MUST fail deterministically.
- **FR-006**: Environment variable expansion MUST NOT alter string literals outside `${...}` syntax and MUST NOT expand header keys, `TYPE`, or `MODEL`.

---

### User Story 3 - Deterministic validation & operator-facing failure contract (Priority: P3)

As an operator or automation script author, I want invalid or incomplete provider definitions to fail with a fixed, documented error code and stderr class phrase, so that configuration errors can be distinguished from file or syntax errors and handled programmatically.

**Why this priority**: Aligns Slice 003 with the frozen failure contract established in Round 002. Automated pipelines require distinct exit codes and stable stderr class phrases to classify errors without brittle string matching.

**Independent verification**: Run tellme against configurations with missing required provider fields, invalid data types, negative token/budget numbers, and unset `${VAR}` references; verify stderr begins with the pinned class phrase and exits with code `3`.

**Acceptance Scenarios**:

1. **Given** a configuration where the selected provider entry is missing `TYPE`, `MODEL`, or `URL`, **When** tellme runs, **Then** stderr emits `tellme: the provider configuration is invalid` and exits with code `3`.
2. **Given** a configuration where the selected provider entry contains negative values for `MAX_TOKENS` or `THINKING_BUDGET`, **When** tellme runs, **Then** stderr emits `tellme: the provider configuration is invalid` and exits with code `3`.
3. **Given** a configuration where the selected provider entry references an unset environment variable without a default, **When** tellme runs, **Then** stderr emits `tellme: the provider configuration is invalid` and exits with code `3`.

**Functional Requirements**:

- **FR-007**: When the resolved provider entry is missing mandatory fields, has invalid field types, contains negative values for `MAX_TOKENS` or `THINKING_BUDGET`, or fails `${VAR}` expansion, the system MUST emit a stderr line starting with the frozen class phrase `tellme: the provider configuration is invalid` and exit with numeric code `3`.
- **FR-008**: Trailing detail following the class phrase MUST be actionable, indicating the specific provider name and validation error reason (e.g. `tellme: the provider configuration is invalid: provider "deepseek": missing required field "model"`).
- **FR-009**: Provider validation MUST validate the effective selected provider entry. Unselected provider entries that are syntactically valid YAML MUST NOT block startup if their unreferenced environment variables are unset.

---

### Edge Cases

- **Multiple variable expansions**: When a field contains multiple expansions (e.g. `https://${HOST}:${PORT}/v1`), all valid `${VAR}` tokens MUST be expanded in order.
- **Empty default**: When syntax `${VAR:-}` is used with an empty default, it MUST resolve to an empty string `""` without failing.
- **Malformed syntax**: An unclosed `${` without a matching `}` MUST be rejected as an invalid provider configuration error.
- **Unknown top-level keys**: Real `tell-me-go` configurations often contain keys for future slices (such as `MODELS:`, `MCP_SERVERS:`, `MEMORY:`). The configuration loader MUST remain tolerant of unknown top-level keys so valid existing configuration files remain compatible.
- **Negative budgets**: `MAX_TOKENS: -1` or `THINKING_BUDGET: -100` MUST be rejected with `tellme: the provider configuration is invalid`.

---

## Requirements *(mandatory)*

### Global Requirements

#### Non-Functional Requirements

- **NFR-001**: All configuration loading, variable expansion, and provider validation MUST remain strictly offline and deterministic — no network calls, DNS lookups, or provider ping endpoints (carried from round-001 NFR-003).
- **NFR-002**: Error messages MUST be written to stderr and carry the frozen class phrase prefix `tellme: the provider configuration is invalid` as the single failure class indicator.
- **NFR-003**: Pure resolution, `${VAR}` expansion, and provider validation logic MUST be covered by fast, table-driven unit tests running in memory (complementing the end-to-end acceptance path per round-002 F9/D5).
- **NFR-004**: The existing Round 001 and Round 002 CLI contracts, exit codes (`0`, `2`, `3`, `4`, `5`), and class phrases MUST remain green with zero regression.

### Key Entities *(include if feature involves data)*

- **ProviderEntry**: The complete provider specification within `PROVIDERS.<name>` containing `TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL`.
- **EnvironmentExpander**: The deterministic substitution engine for `${VAR}` and `${VAR:-default}` string tokens.
- **ProviderValidationContract**: The operator-facing failure specification binding exit code `3` and class phrase `tellme: the provider configuration is invalid`.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of acceptance scenarios for provider entry parsing (all core fields present or default-omitted) pass.
- **SC-002**: 100% of acceptance scenarios for `${VAR}` and `${VAR:-default}` expansion pass, and 100% of unset required variable references fail deterministically.
- **SC-003**: 100% of provider entry validation failures emit the exact frozen class phrase `tellme: the provider configuration is invalid` on stderr and exit with code `3`.
- **SC-004**: Pure resolution and expansion functions achieve 100% unit test coverage in `internal/config/`, and all existing end-to-end BDD tests (20/20 scenarios) remain passing.

---

## Assumptions

- Context-window bounds resolution and model pricing mapping are deferred to Slice 004 and future runtime slices; Slice 003 validates provider configuration schema and credential expansion only.
- `USER_ID`, `THINKING_ENABLED` tri-state toggle, and top-level `MODELS` pricing tables are deferred to future runtime execution slices.
- Validation is enforced on the effective selected provider entry. Unselected entries in the registry that are well-formed YAML do not block booting even if their specific environment variables are unset.
- All operations remain 100% offline.
