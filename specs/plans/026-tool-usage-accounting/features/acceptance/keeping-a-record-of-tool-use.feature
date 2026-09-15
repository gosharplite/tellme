Feature: Keeping a record of how the tools are used

  # Acceptance only: every time the agent uses a tool, tellme keeps a durable
  # record of whether the tool succeeded, failed, or ran out of time — so the
  # operator can later see which tools the agent actually uses and which give it
  # trouble. The record is kept across runs and sessions.

  Rule: Every tool the agent uses is recorded with its outcome

    Example: A tool that succeeds is recorded as a success
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads a file before answering
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What does notes.txt say?"
      Then the tool usage shows the read_files tool was used once and succeeded
      And the operator sees the answer
      And tellme exits successfully

    Example: A tool that fails is recorded as a failure
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that lists a folder that does not exist before answering
      When the operator asks tellme "List the folder that does not exist."
      Then the tool usage shows the list_files tool was used once and failed
      And the operator sees the answer
      And tellme exits successfully

    Example: A tool that runs out of time is recorded as having run out of time
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that runs a command which never returns before answering
      When the operator asks tellme "Run a command that never finishes, with a short time limit."
      Then the tool usage shows the execute_command tool was used once and ran out of time
      And the operator sees the answer
      And tellme exits successfully

    Example: A turn that uses no tool adds no record
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers without using any tool
      When the operator asks tellme "What is two plus two?"
      Then the tool usage shows no tool has been used

  Rule: The record is kept across sessions

    Example: The record survives starting a fresh session
      Given the operator has a runnable tellme installation
      And the tool usage already records that the read_files tool was used three times
      When the operator starts a fresh session with "--new" and the prompt "hi"
      Then the tool usage still records that the read_files tool was used three times
      And tellme exits successfully
