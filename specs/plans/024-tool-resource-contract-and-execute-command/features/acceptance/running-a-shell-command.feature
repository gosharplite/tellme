Feature: Running a shell command

  # Acceptance only: tellme can carry out a task that needs a shell command, and
  # the run survives whatever the command does — it succeeds, it fails, it
  # produces a great deal of output, or it never returns.

  Rule: A command the task needs is run and its outcome is reported

    Example: A command that succeeds
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that runs a command before answering
      And the working directory contains a file "hello.txt" whose text is "hi"
      When the operator asks tellme "Show me the files here with a shell command."
      Then tellme runs the requested command
      And the answer reflects the command's output
      And tellme exits successfully

    Example: A command that fails is reported without breaking the run
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that runs a command which fails and then answers with "the command failed"
      When the operator asks tellme "Run a command that fails."
      Then tellme reports the command's failure
      And the operator sees the answer
      And tellme exits successfully

  Rule: What a command returns cannot overwhelm the run

    Example: A command with a great deal of output is trimmed
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that runs a command producing a great deal of output before answering
      When the operator asks tellme "Produce a lot of output and summarise it."
      Then what tellme keeps of the command's output is limited to what the run can hold
      And tellme notes that the output was trimmed
      And the operator sees the answer
      And tellme exits successfully

    Example: A command that writes its output to a file
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that runs a command writing its output to a file before answering
      When the operator asks tellme "Write the output to a file and tell me where it went."
      Then tellme reports where the output was written
      And tellme reads that file back
      And tellme exits successfully

  Rule: A command that never returns is stopped

    Example: A command that hangs is stopped so the run finishes
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that runs a command which never returns and then answers
      When the operator asks tellme "Run a command that never finishes."
      Then tellme stops waiting for the command
      And the operator sees the answer
      And tellme exits successfully
