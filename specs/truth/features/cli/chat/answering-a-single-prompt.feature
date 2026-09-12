Feature: Answering a single prompt

  # Interface truth (CLI end, `chat` module) — atomic rules for one reasoning turn and its offline
  # guards. The acceptance journeys live in the plan package (`features/acceptance/**`).

  Rule: A prompt sends exactly one request to the selected provider and prints the answer

    Example: The selected provider answers the prompt
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "List packages with go list"
      When the operator starts tellme with the prompt "How do I list Go packages?"
      Then tellme sends exactly one request to the provider "test-model"
      And tellme prints the provider's answer "List packages with go list"
      And tellme exits successfully

    Example: The provider override selects a different configured provider
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And configured providers "alpha" and "beta" whose endpoints answer
      And the selected provider override is "beta"
      When the operator starts tellme with the prompt "Hello"
      Then tellme sends exactly one request to the provider "beta"
      And tellme exits successfully

  Rule: A run outside the normal prompt turn makes no provider request

    Example: A prompt-less run does not contact a provider
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "unused"
      When the operator starts tellme
      Then tellme performs no network access
      And tellme exits successfully

    Example: An explicit diagnostic run with a prompt does not contact a provider
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "unused"
      When the operator runs tellme's diagnostic with the prompt "Hello"
      Then tellme performs no network access
      And tellme exits successfully
