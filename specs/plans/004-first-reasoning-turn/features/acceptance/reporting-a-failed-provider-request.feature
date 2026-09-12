Feature: Reporting a failed provider request

  # Acceptance only: the operator-facing failure surface when a reasoning turn cannot complete.
  # A failed request must name the frozen class phrase and exit with a distinct provider error code,
  # keeping provider failures separable from the existing boot, configuration, and usage classes.

  Rule: An unreachable provider or an error response must fail with the frozen class phrase

    Example: The operator's provider cannot be reached
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose selected provider "dead-model" cannot be reached
      When the operator starts tellme with the prompt "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider request failed
      And tellme exits with the provider error code

    Example: The provider answers with an error status
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose selected provider "error-model" answers with an error status
      When the operator starts tellme with the prompt "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider request failed
      And tellme exits with the provider error code

  Rule: A response that cannot be understood as an answer must fail the same way

    Example: The provider returns no usable answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose selected provider "garbled-model" answers with a response that contains no usable answer
      When the operator starts tellme with the prompt "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider request failed
      And tellme exits with the provider error code
