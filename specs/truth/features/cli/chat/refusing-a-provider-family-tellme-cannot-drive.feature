Feature: Refusing a provider family tellme cannot drive

  # Interface truth (CLI end, `chat` module) — atomic rule for the provider-family boundary. The
  # acceptance journey lives in the plan package
  # (`features/acceptance/refusing-providers-tellme-cannot-drive.feature`). Adding the Gemini family
  # must not change how tellme refuses a family it cannot drive (for example, an Anthropic provider).

  Rule: A provider family tellme cannot drive is refused with the frozen phrase

    Example: An Anthropic provider is selected
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "vertex-claude" of a family tellme cannot drive
      When the operator starts tellme with the prompt "Hello"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code
