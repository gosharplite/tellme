Feature: Remembering the conversation across runs

  # Interface truth (CLI end, `chat` module) — the prompt turn persists each completed exchange and
  # carries the persisted conversation to the provider. The session store and its lifecycle live in the
  # `history` module. Acceptance journeys: features/acceptance/remembering-the-conversation.feature.

  Rule: A completed prompt turn is stored in the session history

    Example: The exchange is stored after the run
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "Paris"
      When the operator starts tellme with the prompt "What is the capital of France?"
      Then tellme stored the exchange "What is the capital of France?" and "Paris" in the session history
      And tellme exits successfully

  Rule: A prompt run carries the persisted conversation to the provider

    Example: A follow-up prompt carries the earlier exchange
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
      And a configured provider "test-model" whose endpoint answers with "Alice"
      When the operator starts tellme with the prompt "What is my name?"
      Then the request carried the earlier exchange "My name is Alice." and "Noted."
      And tellme exits successfully

  Rule: A prompt run with no persisted conversation carries no earlier exchange

    Example: The first prompt carries nothing
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "unused"
      When the operator starts tellme with the prompt "What is my name?"
      Then the request carried no earlier exchange
      And tellme exits successfully
