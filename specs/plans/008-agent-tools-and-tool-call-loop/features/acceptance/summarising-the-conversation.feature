Feature: Summarising the conversation with an agent tool

  # Acceptance only: the deferred summarisation capability arrives as an
  # on-demand tool, not as automatic context pruning.

  Rule: The model can summarise earlier history through a tool

    Example: A long conversation is condensed on demand
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model"
      And the runtime home already remembers an earlier conversation between the operator and tellme
      When the operator asks tellme "Summarise our conversation so far."
      Then tellme produces a summary of the earlier conversation
      And the earlier conversation records are left unchanged
      And tellme exits successfully
