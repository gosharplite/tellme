Feature: Refusing providers tellme cannot drive

  # Acceptance only: adding the Gemini family must not change how tellme refuses
  # a provider family it cannot drive (for example, an Anthropic provider).

  Rule: A provider family tellme cannot drive keeps its existing refusal

    Example: The operator selects an Anthropic provider
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose selected provider "vertex-claude" is an Anthropic provider
      When the operator asks tellme "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider request failed
      And tellme exits with the provider error code
