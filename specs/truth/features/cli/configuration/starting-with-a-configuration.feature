Feature: Starting tellme with a configuration

  Background:
    Given the operator has a runnable tellme installation

  Rule: A run with a well-formed configuration reports readiness

    Example: A well-formed configuration at the pointed path is ready
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme reports the configuration is ready
      And tellme exits successfully

  Rule: A run pointing at a configuration that does not exist stops

    Example: The pointed configuration file is missing
      Given the runtime home is "ait-tmg"
      And no configuration exists at "configs/butler.yaml"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the configuration could not be found"
      And tellme exits with the configuration error code

  Rule: A run pointing at a malformed configuration stops

    Example: The pointed configuration contains malformed YAML
      Given the runtime home is "ait-tmg"
      And a malformed configuration "configs/butler.yaml"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the configuration could not be parsed"
      And tellme exits with the configuration error code

  Rule: With no configuration flag and no default found, the run stops

    Example: No default configuration exists for the effective mode
      Given the runtime home is "ait-tmg"
      And no configuration exists at the default location
      When the operator starts tellme
      Then tellme refuses to proceed
      And tellme explains on stderr that "no configuration could be found"
      And tellme exits with the configuration error code

  Rule: With no configuration flag, the default for the effective mode is used

    Example: The default configuration for the effective mode is found
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/coder.yaml"
      And the effective mode is "coder"
      When the operator starts tellme
      Then tellme reports the configuration is ready
      And tellme exits successfully

  Rule: A run proceeds when the effective selected provider is in the registry

    Example: The file's selected provider is in the provider registry
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml" whose selected provider "deepseek-flash" is in its registry
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme reports the configuration is ready
      And tellme exits successfully

  Rule: A run stops when an environment override names a provider not in the registry

    Example: The environment override points at a provider absent from the registry
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml" whose selected provider "deepseek-flash" is in its registry
      And the selected provider override is "ghost"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the selected provider is not in the registry"
      And tellme exits with the configuration error code

  Rule: A run stops when the provider registry is empty

    Example: The configuration declares no providers
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml" whose provider registry is empty
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the selected provider is not in the registry"
      And tellme exits with the configuration error code

  Rule: A run proceeds when the selected provider declares complete typed attributes

    Example: The selected provider declares all core attributes
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml" where provider "deepseek-custom" specifies:
        | field           | value                    |
        | TYPE            | deepseek                 |
        | MODEL           | deepseek-v4-pro          |
        | URL             | https://api.deepseek.com |
        | API_KEY         | secret-token-123         |
        | MAX_TOKENS      | 32768                    |
        | THINKING_BUDGET | 16384                    |
        | THINKING_LEVEL  | HIGH                     |
      And the provider "deepseek-custom" in configuration "configs/butler.yaml" includes custom headers:
        | header         | value         |
        | X-Custom-Trace | trace-abc-001 |
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme reports the configuration is ready
      And tellme exits successfully

    Example: The selected provider omits optional attributes
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml" where provider "minimal-prov" specifies only mandatory fields:
        | field | value                  |
        | TYPE  | gemini                 |
        | MODEL | gemini-3-flash         |
        | URL   | https://api.google.com |
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme reports the configuration is ready
      And tellme exits successfully

  Rule: Dynamic environment variable expressions in the selected provider must expand at load time

    Example: Environment variable expansion resolves configured variables and defaults
      Given the runtime home is "ait-tmg"
      And the environment variable "TEST_PROVIDER_KEY" is set to "env-secret-xyz"
      And the environment variable "TEST_GATEWAY_HOST" is set to "gateway.internal.net"
      And the environment variable "TEST_UNSET_VAR" is unset
      And a well-formed configuration "configs/butler.yaml" where provider "expanded-prov" specifies:
        | field   | value                           |
        | TYPE    | openai                          |
        | MODEL   | gpt-5.5                         |
        | URL     | https://${TEST_GATEWAY_HOST}/v1 |
        | API_KEY | ${TEST_PROVIDER_KEY}            |
      And the provider "expanded-prov" in configuration "configs/butler.yaml" includes custom headers:
        | header        | value                            |
        | X-Routing-Tag | ${TEST_UNSET_VAR:-fallback-zone} |
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme reports the configuration is ready
      And tellme exits successfully

  Rule: A run stops when the selected provider entry is malformed or missing mandatory fields

    Example: The selected provider is missing a mandatory field
      Given the runtime home is "ait-tmg"
      And a configuration "configs/butler.yaml" where selected provider "broken-prov" is missing "MODEL"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider configuration is invalid"
      And tellme exits with the configuration error code

    Example: The selected provider has a negative token limit
      Given the runtime home is "ait-tmg"
      And a configuration "configs/butler.yaml" where selected provider "negative-prov" has negative "MAX_TOKENS"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider configuration is invalid"
      And tellme exits with the configuration error code

  Rule: A run stops when an environment variable referenced without a default is unset

    Example: An environment variable in the selected provider has no value and no default
      Given the runtime home is "ait-tmg"
      And the environment variable "MISSING_API_KEY" is unset
      And a well-formed configuration "configs/butler.yaml" where provider "unresolved-prov" references unset variable "MISSING_API_KEY"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider configuration is invalid"
      And tellme exits with the configuration error code
