Feature: Watching the tool loop work

  # Acceptance only: the operator can follow the loop as it runs, and the
  # remembered conversation stays operator-facing.

  Rule: A tool-using run reports its tool-loop activity

    Example: The operator sees each tool call and its result
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads a file before answering
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "Read notes.txt and summarise it."
      Then the run reports, on its diagnostic output, the tool call it made and the result it received
      And the final answer is printed on standard output
      And tellme exits successfully

  Rule: The remembered conversation stays operator-facing

    Example: Listing a tool-using conversation shows only the operator's messages
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads a file before answering
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What is the launch code? Read notes.txt to find out."
      And the operator asks tellme to list the last 2 messages with "-l 2"
      Then the displayed list shows the operator's prompt and tellme's answer
      And the displayed list shows no tool activity
      And tellme exits successfully
