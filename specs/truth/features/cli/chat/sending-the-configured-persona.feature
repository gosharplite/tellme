Feature: Sending the configured persona

  # Interface truth (CLI end, `chat` module) — the configured `PERSON` is sent to the
  # provider as the leading `system` message of every request the turn makes, and is
  # request-only (never printed). Acceptance journey:
  # features/acceptance/sending-the-configured-persona.feature.

  Rule: The configured persona is sent as the model's leading instruction

    Example: A configured persona reaches the model ahead of the question
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      And the runtime home holds a configuration whose persona is "be terse"
      When the operator starts tellme with the prompt "Hello"
      Then the request carried the persona "be terse"
      And tellme prints the provider's answer "ok"
      And tellme exits successfully

  Rule: A configuration without a persona sends no persona instruction

    Example: An empty persona adds no instruction to the request
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      And the runtime home holds a configuration with no persona
      When the operator starts tellme with the prompt "Hello"
      Then the request carried no persona
      And tellme exits successfully

  Rule: The persona is request-only and never printed

    Example: The persona does not appear in the raw answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      And the runtime home holds a configuration whose persona is "be terse"
      When the operator starts tellme with the prompt "Hello" and the raw flag
      Then the captured standard output is exactly "ok"
