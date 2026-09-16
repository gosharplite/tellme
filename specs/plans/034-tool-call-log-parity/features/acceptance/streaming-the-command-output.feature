Feature: A running command's output is shown live
  # Acceptance journey (PM language) for round 034 — the live [Tool Output] block.

  Rule: A command's output appears while it runs
    Example: The operator sees the command's output as it is produced
      Given the operator has a runnable tellme installation
      And a configured provider whose reply asks tellme to run a command that prints output
      When the operator asks tellme a question that makes it run that command
      Then the operator sees an announcement that the command is running and its output is shown below
      And the operator sees the command's printed output
      And tellme exits successfully

    Example: A command that produces a great deal of output is stopped rather than filling the screen forever
      Given the operator has a runnable tellme installation
      And a configured provider whose reply asks tellme to run a command that prints far more output than the run allows
      When the operator asks tellme a question that makes it run that command
      Then the operator sees the command's output up to the allowed amount
      And the command's result reports that the output was cut off
      And tellme exits successfully

  Rule: A command that writes to a file does not print its output inline
    Example: An output-to-file command shows no streamed output block
      Given the operator has a runnable tellme installation
      And a configured provider whose reply asks tellme to run a command that writes its output to a file
      When the operator asks tellme a question that makes it run that command
      Then the operator does not see a streamed output block for that command
      And tellme exits successfully
