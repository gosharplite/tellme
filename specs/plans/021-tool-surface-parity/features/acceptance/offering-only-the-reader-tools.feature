Feature: Offering only the reader tools

  # Acceptance only: the operator can rely on a stable, minimal set of read-only
  # filesystem tools; the summarisation tool is gone.

  Rule: tellme offers exactly its filesystem reader tools

    Example: The tools tellme makes available to the model
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reports which tools it was offered
      When the operator asks tellme "Which tools can you use to look at my files?"
      Then tellme offers exactly its "list_files", "read_files", and "get_tree" tools
      And tellme offers no summarisation tool
      And tellme exits successfully

  Rule: A request for the removed tool is refused

    Example: A provider that asks for the summarisation tool
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that asks for a "summarize_history" tool
      When the operator asks tellme "Summarise our conversation so far."
      Then the operator sees the message "tellme: the tool request failed"
      And tellme exits with code 7
