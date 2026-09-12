Feature: Using a declared tool

  # Interface truth (CLI end, `chat` module) — a prompt run may call the read-only filesystem tools and
  # fold their result into the answer. Acceptance journey: features/acceptance/answering-with-a-declared-tool.feature.

  Rule: A prompt that needs a tool is answered using it

    Example: A question answered by reading a file
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "What is the launch code? Read notes.txt."
      Then tellme read "notes.txt" using its read-files tool
      And tellme prints the provider's answer "The launch code is ORANGE"
      And tellme exits successfully

  Rule: A prompt that needs no tool invokes no tool

    Example: A plain question uses no tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "Four"
      When the operator starts tellme with the prompt "What is two plus two?"
      Then tellme used no tool
      And tellme sends exactly one request to the provider "test-model"
      And tellme exits successfully
