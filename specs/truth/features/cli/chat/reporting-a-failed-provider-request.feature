Feature: Reporting a failed provider request

  # Interface truth (CLI end, `chat` module) — the provider/transport failure surface. A failed
  # request must name the frozen class phrase `the provider request failed` and exit with the pinned
  # provider error code `6`.

  Rule: An unreachable provider or an error response is reported with the frozen class phrase

    Example: The provider cannot be reached
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "dead-model" whose endpoint is unreachable
      When the operator starts tellme with the prompt "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

    Example: The provider answers with an error status
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "error-model" whose endpoint answers with an error status
      When the operator starts tellme with the prompt "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

  Rule: A response that cannot be understood as an answer fails the same way

    Example: The provider returns no usable answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "garbled-model" whose endpoint answers with no usable answer
      When the operator starts tellme with the prompt "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code
