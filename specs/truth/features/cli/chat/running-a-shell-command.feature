Feature: Running a shell command

  # Interface truth (CLI end, `chat` module) — the `execute_command` tool: a prompt run may run a
  # shell command through `bash -c`. A non-zero exit is a reported result (not a failure); a command
  # that exceeds its time is stopped; a command may capture its output to a file. There is no
  # security/consent gate and no `pipe_commands` tool. Acceptance journey:
  # features/acceptance/running-a-shell-command.feature.

  Rule: The command tool is offered and runs

    Example: A command runs and its output is fed back
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a command and then answers with "done"
      When the operator starts tellme with the prompt "Show me the working directory."
      Then the request offered exactly the agent tools
      And tellme ran the command using its execute_command tool
      And tellme prints the provider's answer "done"
      And tellme exits successfully

  Rule: A command that exits non-zero is reported, not fatal

    Example: A command that exits non-zero
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a command that exits non-zero and then answers with "it failed"
      When the operator starts tellme with the prompt "Run a command that fails."
      Then the command result carried the exit status "3"
      And tellme prints the provider's answer "it failed"
      And tellme exits successfully

  Rule: A command that exceeds its time is stopped

    Example: A command that never returns
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a command that never returns and then answers with "stopped"
      When the operator starts tellme with the prompt "Run a command that never finishes."
      Then the command result recorded that the command was stopped
      And tellme exits successfully

  Rule: A command may capture its output to a file

    Example: Output written to a file and read back
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a command that writes its output to a file and then reads it and then answers with "done"
      When the operator starts tellme with the prompt "Write the output to a file and tell me where."
      Then the command result reported that its output went to "out.txt"
      And tellme read "out.txt" using its read_files tool
      And tellme exits successfully
