Feature: Answering with a declared tool

  # Acceptance only: what the operator experiences when tellme can use read-only
  # filesystem tools to answer a prompt. No writes and no process execution.

  Rule: A prompt that needs a tool is answered using it

    Example: A question answered by reading a file
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers using the tool results it receives
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What is the launch code? Read notes.txt to find out."
      Then tellme reads "notes.txt" using its read_files tool
      And the final answer states that the launch code is "ORANGE"
      And tellme exits successfully

  Rule: A prompt that needs no tool stays a single request

    Example: A plain question answered without a tool
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator asks tellme "What is two plus two?"
      Then tellme makes exactly one request to the provider
      And tellme uses no tool
      And tellme exits successfully
