Feature: Offering the agent tools

  # Acceptance only: what the operator experiences when tellme offers its agent
  # tools — the filesystem readers and the command tool, and nothing the operator
  # removed.

  Rule: The run offers its agent tools, and only those

    Example: The offered tool set is the readers plus the command tool
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reports the offered tools before answering
      When the operator asks tellme "Which tools can you use?"
      Then the run offers the filesystem readers and the command tool
      And the run offers no other tool
      And tellme exits successfully

  Rule: A tool the operator removed is not offered

    Example: A request for the removed summarisation tool fails
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that asks for a tool that is not available
      When the operator asks tellme "Summarise our conversation so far."
      Then tellme reports that the tool request failed
      And tellme exits with the tool error code
