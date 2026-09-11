Feature: Invalid provider configurations fail with a documented failure contract

  # Acceptance only: the operator-facing failure surface for invalid provider configurations.
  # Validation failures must exit with configuration error code 3 and emit the frozen class phrase.

  Rule: Provider entries missing mandatory fields or containing invalid values must be rejected

    Example: The operator configures a provider with missing mandatory fields or negative limits
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration where the selected provider "broken-prov" is missing the "MODEL" field
      When the operator starts tellme
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider configuration is invalid
      And tellme exits with a configuration error code distinct from the success code

      When the runtime home holds a configuration where the selected provider "negative-prov" has a negative "MAX_TOKENS" value
      And the operator starts tellme
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider configuration is invalid
      And tellme exits with a configuration error code distinct from the success code

  Rule: Unresolved environment variables without defaults must fail provider configuration loading

    Example: The operator configures a provider referencing an unset environment variable
      Given the operator has a runnable tellme installation
      And the environment variable "MISSING_API_KEY" is unset
      And the runtime home holds a configuration with a provider "unresolved-prov" specifying:
        | field   | value                     |
        | TYPE    | openai                    |
        | MODEL   | gpt-5.5                   |
        | URL     | https://api.openai.com/v1 |
        | API_KEY | ${MISSING_API_KEY}        |
      And the selected provider is "unresolved-prov"
      When the operator starts tellme
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider configuration is invalid
      And tellme exits with a configuration error code distinct from the success code
