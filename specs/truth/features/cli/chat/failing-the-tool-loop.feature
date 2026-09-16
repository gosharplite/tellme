Feature: Bounding and failing the tool loop

  # Interface truth (CLI end, `chat` module) — the loop always ends: a recoverable tool error is fed back
  # (non-terminal), and an unrecoverable loop failure is reported. Acceptance journey:
  # features/acceptance/bounding-and-failing-the-tool-loop.feature.

  Rule: A tool that returns an error is fed back, not fatal

    Example: Reading a missing file is reported back and the run still finishes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint asks tellme to read "missing.txt" and then answers with "No such file"
      When the operator starts tellme with the prompt "Read missing.txt."
      Then the request carried the read-tool error for "missing.txt"
      And tellme prints the provider's answer "No such file"
      And tellme exits successfully

  Rule: A loop that cannot finish is stopped and reported

    Example: A provider that keeps asking for a tool is stopped at the loop limit
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And the tool-loop limit is "3"
      And a configured provider "test-model" whose endpoint always asks tellme to read "notes.txt"
      When the operator starts tellme with the prompt "Keep reading until you find it."
      Then tellme stopped after 3 tool iterations
      And tellme explains on stderr that "the tool request failed"
      And tellme exits with the tool error code

    Example: A request for a tool that is not available is reported
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint asks for a tool that is not available
      When the operator starts tellme with the prompt "Use the time-travel tool."
      Then tellme explains on stderr that "the tool request failed"
      And tellme exits with the tool error code

  Rule: The loop reports a step marker only for executed rounds

    Example: A provider that exceeds the loop limit reports no marker for the final request
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And the tool-loop limit is "1"
      And a configured provider "test-model" whose endpoint always asks tellme to read "notes.txt"
      When the operator starts tellme with the prompt "Keep reading until you find it."
      Then the run reported a tool step marker for each of the 1 executed rounds
      And the final model request reports no tool step marker
      And tellme exits with the tool error code
