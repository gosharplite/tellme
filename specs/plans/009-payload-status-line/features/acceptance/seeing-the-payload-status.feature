Feature: Seeing the payload status of a turn

  # Acceptance only: the operator can see, per conversation turn, how large the
  # payload is relative to a budget. Diagnostics only — the answer stream is
  # untouched.

  Rule: A prompt run reports the payload's estimated size before it is sent

    Example: The operator sees the estimated payload and the budget before the answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      And the conversation already remembers earlier turns
      When the operator asks tellme "Summarise where we are."
      Then the run reports, before the answer, an estimated payload size for the turn
      And the reported estimate is measured against the configured payload budget
      And the report names the active mode and model
      And the final answer is printed on standard output

  Rule: A prompt run reports the provider's actual payload size after the turn

    Example: The operator sees the actual payload the provider measured
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reports its usage
      When the operator asks tellme "What is two plus two?"
      Then the run reports, after the answer, the payload size the provider actually measured
      And the measured size is measured against the configured payload budget

    Example: A provider that reports no usage shows no measured figure
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reports no usage
      When the operator asks tellme "What is two plus two?"
      Then the run reports the estimated payload size for the turn
      And the run reports no measured payload size
      And tellme exits successfully

  Rule: The payload status never disturbs the answer

    Example: The answer stays unchanged while the status is shown separately
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme for a raw answer with "-r"
      Then the answer on standard output is the model's text unchanged
      And the payload status for the turn is still reported, separately from the answer

  Rule: Only conversation turns report a payload status

    Example: An offline listing command shows no payload status
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model"
      When the operator asks tellme to list the last 2 messages with "-l 2"
      Then no payload status is reported
      And tellme exits successfully
