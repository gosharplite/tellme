Feature: Offering the reader tools

  # Interface truth (CLI end, `chat` module) — a prompt run offers exactly the filesystem reader tools
  # (`list_files`, `read_files`, `get_tree`); the round-008 summarisation tool is removed, so a request
  # for it fails under the existing unknown-tool contract. Acceptance journey:
  # features/acceptance/offering-only-the-reader-tools.feature.

  Rule: tellme offers exactly its reader tools

    Example: The offered tool set is the three readers
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use to look at my files?"
      Then the request offered exactly the reader tools
      And tellme exits successfully

  Rule: A removed tool is not offered

    Example: A provider that asks for the summarisation tool fails
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint asks for a tool that is not available
      When the operator starts tellme with the prompt "Summarise our conversation so far."
      Then tellme explains on stderr that "the tool request failed"
      And tellme exits with the tool error code
