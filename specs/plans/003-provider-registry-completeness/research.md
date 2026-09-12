# Phase 0 Research: tellme Provider Registry Completeness (Round 003)

Topic: the round-003 schema completeness slice — (1) **grow `PROVIDERS` entry** to the real typed provider schema (`TYPE`, `MODEL`, `URL`, `API_KEY`, `MAX_TOKENS`, `HEADERS`, `THINKING_BUDGET`, `THINKING_LEVEL`) needed for LLM requests, (2) **`${VAR}` and `${VAR:-default}` environment variable expansion** in credentials, URLs, and headers, (3) **deterministic offline validation and error reporting** (exit code `3`, class phrase `tellme: the provider configuration is invalid`), and (4) **unit tests** covering schema parsing, expansion, and validation.

Scope note: The language (`Go 1.26`), module, testing framework (`godog` + stdlib `testing`), and base tooling (`golangci-lint` + `govulncheck`) were locked in rounds 001 and 002. This research establishes the technical decisions for the provider configuration model, variable expansion engine, and validation pipeline.

---

## Decision 1: Provider Entry Schema & Struct Representation

- **Decision**: Expand `internal/config.Provider` to a fully typed Go struct representing the core request attributes:
  ```go
  type Provider struct {
      Type           string            `yaml:"TYPE"`
      Model          string            `yaml:"MODEL"`
      URL            string            `yaml:"URL"`
      APIKey         string            `yaml:"API_KEY"`
      MaxTokens      int               `yaml:"MAX_TOKENS"`
      Headers        map[string]string `yaml:"HEADERS"`
      ThinkingBudget int               `yaml:"THINKING_BUDGET"`
      ThinkingLevel  string            `yaml:"THINKING_LEVEL"`
  }
  ```
  Top-level YAML decoding of `Config` remains tolerant (no `Decoder.KnownFields(true)` on `Config`) so that real-world `tell-me-go` configurations containing future blocks (e.g. `MODELS:`, `MCP_SERVERS:`, `MEMORY:`) continue to load cleanly.
- **Rationale**: Exactly fulfills Clarify Round 1 Q1 (Option 1). These eight fields represent all parameters required by Slice 004 to construct authentic HTTP requests to Gemini, OpenAI, DeepSeek, Anthropic, and Moonshot APIs without taking on premature pricing, accounting, or agent memory complexity.
- **Alternatives considered**:
  - **`map[string]interface{}` (untyped bag)**: Rejected because it provides zero compile-time safety and requires brittle runtime type assertions.
  - **Immediate adoption of `Decoder.KnownFields(true)`**: Rejected because it would break compatibility with standard `tell-me-go` configuration files that declare top-level keys intended for future slices.
  - **Inclusion of `USER_ID` and `THINKING_ENABLED`**: Deferred; `USER_ID` is a vendor-specific DeepSeek header, and `THINKING_ENABLED` is an optional toggle that defaults cleanly at call time in Slice 004.

---

## Decision 2: Environment Variable Expansion Engine (Zero External Dependencies)

- **Decision**: Implement a hand-crafted, deterministic string expansion helper in `internal/config/expand.go` using a compiled regular expression and Go standard library `os.LookupEnv`:
  - Pattern: `\$\{([a-zA-Z_][a-zA-Z0-9_]*)(?::-([^}]*))?\}`
  - Targeted application: Applied exclusively to `APIKey`, `URL`, and the string values of `Headers` map entries for the resolved provider entry.
  - Error semantics:
    - If a `${VAR}` expression references a variable that is unset or empty, and no default is provided (`:-default`), expansion returns a typed `ErrUnsetVariable` error.
    - If a `${VAR:-default}` expression references an unset or empty variable, `default` is substituted.
    - If `${VAR:-}` has an empty default, `""` is substituted.
    - If an unclosed `${` or malformed expression is detected, expansion returns an `ErrMalformedVariable` error.
- **Rationale**: Fulfills Clarify Round 1 Q2 (Option 1). Using Go's standard library (`regexp` + `os.LookupEnv`) ensures zero third-party dependencies, total determinism, and instant execution. Failing on unset variables without defaults protects operators from silent auth or routing failures.
- **Alternatives considered**:
  - **Third-party libraries (e.g., `github.com/a8m/envsubst` or `drone/envsubst`)**: Rejected; adds external supply-chain dependencies for a simple, closed syntactic grammar easily handled by stdlib regex.
  - **`os.ExpandEnv` stdlib function**: Rejected because `os.ExpandEnv` does not support `${VAR:-default}` syntax and silently replaces missing variables with empty strings rather than failing.
  - **Universal expansion across all YAML values**: Rejected; expanding `TYPE` or `MODEL` creates accidental ambiguity and departs from standard CLI conventions where model identifiers are explicit.

---

## Decision 3: Provider Validation Pipeline and Failure Reporting

- **Decision**: Validate the effective selected provider entry during `config.Resolve()`:
  - **Mandatory presence**: `Type`, `Model`, and `URL` must be non-empty strings.
  - **Numeric constraints**: `MaxTokens >= 0` and `ThinkingBudget >= 0`.
  - **Expansion execution**: Execute targeted variable expansion on the selected provider.
  - **Failure contract**: If any check fails, return a wrapped error resulting in exit code `3` and emitting the frozen class phrase:
    `tellme: the provider configuration is invalid: <actionable detail>`
  - **Scope of validation**: Validation is performed on the *effective selected provider*. Syntactically valid YAML entries for unselected providers in the registry do not fail configuration loading if their unreferenced environment variables are absent.
- **Rationale**: Fulfills Clarify Round 1 Q3 (Option 1) and preserves the exit code hierarchy (`3` for configuration errors) and class-phrase structure established in Round 002. Scoping semantic variable validation to the active provider allows multi-provider configuration files (e.g. containing Gemini, DeepSeek, and OpenAI definitions) to run without requiring dummy API keys for providers that are not currently selected.
- **Alternatives considered**:
  - **Validating every registry entry on startup**: Rejected because it would require the operator to export API keys for every provider defined in `PROVIDERS`, even when only running a single provider.
  - **Reusing `tellme: the configuration could not be parsed`**: Rejected; conflates YAML syntax errors with provider-level schema and variable resolution errors.

---

## Decision 4: Pure-Helper Unit Testing Architecture

- **Decision**: Complement E2E acceptance tests with fast, table-driven unit tests in `internal/config/`:
  - `expand_test.go`: Tests `${VAR}`, `${VAR:-default}`, `${VAR:-}`, multiple substitutions in a single string, missing variable errors, and malformed syntax.
  - `config_test.go`: Tests full struct unmarshaling, optional field defaults, missing mandatory fields, negative bounds, and active vs. inactive provider validation.
- **Rationale**: Conforms to Round 002 Decision 1 (E2E for acceptance, unit tests for pure helpers per F9/D5). Expansion and validation logic consists of pure string and struct transformations that execute in sub-millisecond time without subprocess overhead.
- **Alternatives considered**:
  - **E2E-only verification**: Leaves fine-grained edge cases (such as exotic regex combinations or negative integers) slow and cumbersome to verify exclusively via subprocess CLI invocations.

---

## Residual Risks / Forward Links

- **Slice 004 prerequisite**: Slice 004 will consume the populated `Provider` struct (`Type`, `Model`, `URL`, `APIKey`, `Headers`, `ThinkingBudget`, `ThinkingLevel`) to configure the HTTP client transport.
- **Model Context Window**: Model context window lookup is deferred to Slice 004 where token accounting and window limits will be introduced.
