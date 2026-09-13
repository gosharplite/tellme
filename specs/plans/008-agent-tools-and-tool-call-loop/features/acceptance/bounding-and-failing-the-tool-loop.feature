Feature: Bounding and failing the tool loop

  # Acceptance only: the loop always ends; a recoverable tool error is fed back,
  # and an unrecoverable loop failure is reported clearly.

  Rule: A tool that returns an error is fed back, not fatal

    Example: A read of a missing file is reported back and the run still finishes
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads a missing file, then answers using the result it received
      When the operator asks tellme "Read missing.txt and tell me what it says."
      Then the run feeds the read error back to the model
      And tellme produces a final answer
      And tellme exits successfully

  Rule: A loop that cannot finish is stopped and reported

    Example: A model that keeps requesting tools is stopped at the loop limit
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that always asks for another tool call
      When the operator asks tellme "Keep searching until you find it."
      Then tellme stops after the configured number of tool iterations
      And the operator sees the message "tellme: the tool request failed"
      And tellme exits with code 7

    Example: A request for a tool tellme does not provide is reported
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that asks for a tool that is not available
      When the operator asks tellme "Use the time-travel tool to check yesterday."
      Then the operator sees the message "tellme: the tool request failed"
      And tellme exits with code 7
