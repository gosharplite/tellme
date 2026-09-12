Feature: Inspecting the remembered conversation

  # Acceptance only: the operator can look back at what tellme remembers,
  # without sending a new prompt.

  Rule: The operator can list the most recent messages

    Example: Listing the last two messages of a conversation
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers using the conversation it receives
      When the operator asks tellme "My name is Alice."
      And the operator asks tellme "What is my name?"
      And the operator asks tellme to list the last 2 messages with "-l 2"
      Then the displayed list shows the most recent 2 messages of the conversation
      And tellme sends no new request to the provider
      And tellme exits successfully

  Rule: Listing a conversation with nothing remembered shows nothing

    Example: Listing before any conversation exists
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model"
      And the runtime home has no earlier conversation
      When the operator asks tellme to list the last 5 messages with "-l 5"
      Then the displayed list is empty
      And tellme exits successfully
