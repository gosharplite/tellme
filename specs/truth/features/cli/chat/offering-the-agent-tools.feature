Feature: Offering the agent tools

  # Interface truth (CLI end, `chat` module) — a prompt run offers exactly the agent tools: the three
  # filesystem readers (`list_files`, `read_files`, `get_tree`) and the command tool
  # (`execute_command`). The round-008 summarisation tool is removed, so a request for it fails under
  # the existing unknown-tool contract. Acceptance journey:
  # features/acceptance/offering-the-agent-tools.feature.

  Rule: tellme offers exactly its agent tools

    Example: The offered tool set is the readers and the command tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the request offered exactly the agent tools
      And tellme exits successfully

  Rule: A removed tool is not offered

    Example: A provider that asks for the summarisation tool fails
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint asks for a tool that is not available
      When the operator starts tellme with the prompt "Summarise our conversation so far."
      Then tellme explains on stderr that "the tool request failed"
      And tellme exits with the tool error code
