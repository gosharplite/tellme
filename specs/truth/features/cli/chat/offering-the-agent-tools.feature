Feature: Offering the agent tools

  # Interface truth (CLI end, `chat` module) — a prompt run offers exactly the **seven** agent tools:
  # the three filesystem readers (`list_files`, `read_files`, `get_tree`), the two write tools
  # (`write_file`, `replace_text` — round 029), the command tool (`execute_command` — round 024), and the
  # skills listing tool (`list_skills` — round 033). The round-008 summarisation tool is removed, so a
  # request for it fails under the existing unknown-tool contract.
  # Round 062 (ADR 0032): the set is a **function of the selected provider's capability** — a provider
  # whose entry declares `VISION: true` is additionally offered `read_image` (eight tools); a provider
  # without that key (the default) is offered exactly the seven below. The offered list tells the model
  # the truth about what it can do.
  # Acceptance journey: features/acceptance/offering-the-agent-tools.feature.

  Rule: tellme offers exactly its agent tools

    Example: The offered tool set is the readers, the write tools, the command tool, and the skills tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the request offered exactly the agent tools
      And tellme exits successfully

    Example: A provider that can take images is also offered the read-image tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "eye" that can take images whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the request offered the read-image tool
      And tellme exits successfully

  Rule: A removed tool is not offered

    Example: A provider that asks for the summarisation tool fails
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint asks for a tool that is not available
      When the operator starts tellme with the prompt "Summarise our conversation so far."
      Then tellme explains on stderr that "the tool request failed"
      And tellme exits with the tool error code
