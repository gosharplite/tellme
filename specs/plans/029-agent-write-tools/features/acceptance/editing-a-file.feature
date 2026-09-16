Feature: Editing a file with an exact block

  # Acceptance only: tellme can replace one exactly-identified block of a file,
  # and it refuses — leaving the file untouched — when the block is missing or
  # appears more than once.

  Rule: A uniquely identified block is replaced, and nothing else changes

    Example: The single matching block is replaced
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that edits a file before answering
      And the working directory contains a file "config.txt" whose lines are:
        | alpha |
        | BETA  |
        | gamma |
      When the operator asks tellme "In config.txt, replace the line 'BETA' with 'beta'."
      Then the lines of "config.txt" are:
        | alpha |
        | beta  |
        | gamma |
      And tellme exits successfully

  Rule: An edit that cannot be uniquely placed is refused, leaving the file untouched

    Example: A block that is not present is refused
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that edits a file before answering
      And the working directory contains a file "config.txt" whose lines are:
        | alpha |
        | BETA  |
        | gamma |
      When the operator asks tellme "In config.txt, replace the line 'DELTA' with 'delta'."
      Then the edit is refused because the block is not present
      And the lines of "config.txt" are still:
        | alpha |
        | BETA  |
        | gamma |
      And tellme exits successfully

    Example: A block that appears more than once is refused
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that edits a file before answering
      And the working directory contains a file "log.txt" that contains the line "note" twice
      When the operator asks tellme "In log.txt, replace the line 'note' with 'done'."
      Then the edit is refused because the block is not unique
      And the file "log.txt" still contains the line "note" twice
      And tellme exits successfully
