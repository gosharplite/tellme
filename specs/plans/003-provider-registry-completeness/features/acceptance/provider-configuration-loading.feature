Feature: Provider configurations support complete request attributes and dynamic environment variable expansion

  # Acceptance only: what the operator configures and observes at the command line.
  # The provider attributes and expansion rules define the contract for subsequent reasoning turns.

  Rule: A provider entry must recognize complete typed attributes for reasoning requests

    Example: The operator configures a provider with all core fields, then with only mandatory fields
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a provider "deepseek-custom" specifying:
        | field           | value                    |
        | TYPE            | deepseek                 |
        | MODEL           | deepseek-v4-pro          |
        | URL             | https://api.deepseek.com |
        | API_KEY         | secret-token-123         |
        | MAX_TOKENS      | 32768                    |
        | THINKING_BUDGET | 16384                    |
        | THINKING_LEVEL  | HIGH                     |
      And the provider "deepseek-custom" includes custom headers:
        | header         | value         |
        | X-Custom-Trace | trace-abc-001 |
      And the selected provider is "deepseek-custom"
      When the operator starts tellme
      Then tellme accepts the provider configuration
      And tellme reports that the configuration is ready
      And tellme performs no network access
      And tellme exits successfully

      When the runtime home holds a configuration where provider "minimal-prov" specifies only mandatory fields:
        | field | value                  |
        | TYPE  | gemini                 |
        | MODEL | gemini-3-flash         |
        | URL   | https://api.google.com |
      And the selected provider is "minimal-prov"
      And the operator starts tellme
      Then tellme accepts the provider configuration
      And tellme reports that the configuration is ready
      And tellme exits successfully

  Rule: Dynamic environment variable expressions in provider entries must expand at load time

    Example: The operator uses environment variable expansion with existing variables and fallback defaults
      Given the operator has a runnable tellme installation
      And the environment variable "TEST_PROVIDER_KEY" is set to "env-secret-xyz"
      And the environment variable "TEST_GATEWAY_HOST" is set to "gateway.internal.net"
      And the environment variable "TEST_UNSET_VAR" is unset
      And the runtime home holds a configuration with a provider "expanded-prov" specifying:
        | field   | value                           |
        | TYPE    | openai                          |
        | MODEL   | gpt-5.5                         |
        | URL     | https://${TEST_GATEWAY_HOST}/v1 |
        | API_KEY | ${TEST_PROVIDER_KEY}            |
      And the provider "expanded-prov" includes custom headers:
        | header        | value                             |
        | X-Routing-Tag | ${TEST_UNSET_VAR:-fallback-zone}  |
      And the selected provider is "expanded-prov"
      When the operator starts tellme
      Then tellme expands all environment variables in the provider configuration
      And tellme accepts the provider configuration
      And tellme reports that the configuration is ready
      And tellme exits successfully
