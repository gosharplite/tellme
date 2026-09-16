Feature: Offering the write tools

  # Acceptance only: the run offers the agent its file-writing capability — a
  # tool for creating files and a tool for editing files — alongside the readers
  # and the command tool, and nothing the operator removed.

  Rule: The run offers a file-creation tool and a file-editing tool

    Example: The offered tool set includes the write tools
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reports the offered tools before answering
      When the operator asks tellme "Which tools can you use?"
      Then the run offers a tool for creating files and a tool for editing files
      And the run also offers the filesystem readers and the command tool
      And the run offers no other tool
      And tellme exits successfully
