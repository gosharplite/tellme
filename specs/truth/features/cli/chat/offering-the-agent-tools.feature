Feature: Offering the agent tools

  # Interface truth (CLI end, `chat` module) — a prompt run offers exactly the **eight** agent tools:
  # the three filesystem readers (`list_files`, `read_files`, `get_tree`), the in-file content search
  # (`search_files` — round 071), the two write tools (`write_file`, `replace_text` — round 029), the
  # command tool (`execute_command` — round 024), and the skills listing tool (`list_skills` — round 033).
  # The round-008 summarisation tool is removed. **Round 076 (issue #154):** a request for a tool tellme
  # does not offer is now a **recoverable fold-back** (the loop feeds back `no tool named …` and continues),
  # bounded per turn by `maxUnknownToolFolds`; a run that *only* ever asks for the removed tool is stopped at
  # the per-turn cap with the frozen phrase + exit 7 (see `using-an-unknown-tool-name.feature`). The Example
  # below uses the same always-unknown fixture, so it holds either way — its failure is now cap-mediated.
  # Round 062 (ADR 0032): the set is a **function of the selected provider's capability** — a provider
  # whose entry declares `VISION: true` is additionally offered `read_image` (nine tools); a provider
  # without that key (the default) is offered exactly the eight below. The offered list tells the model
  # the truth about what it can do.
  # Acceptance journey: features/acceptance/offering-the-agent-tools.feature.

  Rule: tellme offers exactly its agent tools

    Example: The offered tool set is the readers, the search tool, the write tools, the command tool, and the skills tool
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

    Example: A provider that only asks for a tool tellme does not offer is stopped at the per-turn cap
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint asks for a tool that is not available
      When the operator starts tellme with the prompt "Summarise our conversation so far."
      Then the run reported the unavailable tool "time_travel"
      And tellme explains on stderr that "the tool request failed"
      And tellme exits with the tool error code
