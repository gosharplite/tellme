Feature: Remembering the conversation across runs

  # Acceptance only: what the operator experiences once tellme keeps a durable session.
  # By default a new prompt is answered with the earlier conversation remembered.

  Rule: A completed exchange is remembered for the next prompt

    Example: A follow-up question is answered with the earlier exchange remembered
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers using the conversation it receives
      When the operator asks tellme "My name is Alice."
      And the operator asks tellme "What is my name?"
      Then the second answer reflects the earlier exchange
      And tellme exits successfully

  Rule: A conversation with no history starts empty

    Example: The first prompt carries no earlier exchange
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers using the conversation it receives
      And the runtime home has no earlier conversation
      When the operator asks tellme "What is my name?"
      Then the answer reflects no earlier exchange
      And tellme exits successfully
