Feature: Starting a fresh conversation

  # Acceptance only: the operator can deliberately begin a new conversation,
  # without losing the one that came before.

  Rule: A fresh conversation does not remember the previous one

    Example: After starting fresh, a follow-up question is answered with nothing remembered
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers using the conversation it receives
      When the operator asks tellme "My name is Alice."
      And the operator starts a fresh conversation with "--new"
      And the operator asks tellme "What is my name?"
      Then the answer reflects no earlier exchange
      And tellme exits successfully

  Rule: Starting fresh keeps the previous conversation

    Example: The previous conversation is retained, not discarded
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers using the conversation it receives
      When the operator asks tellme "My name is Alice."
      And the operator starts a fresh conversation with "--new"
      Then the fresh conversation remembers nothing from before
      And the previous conversation is kept in the runtime home, not discarded
      And tellme exits successfully
