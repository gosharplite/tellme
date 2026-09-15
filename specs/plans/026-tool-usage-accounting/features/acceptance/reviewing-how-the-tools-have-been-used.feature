Feature: Reviewing how the tools have been used

  # Acceptance only: the operator can ask tellme to show how each of its agent
  # tools has been used — how many times, and how it fared — so they can spot the
  # tools the agent never uses and the tools that often fail. The review is
  # offline: it needs no provider and no prompt.

  Rule: The operator can review each tool's use

    Example: The operator reviews the tools after some use
      Given the operator has a runnable tellme installation
      And the tool usage records that the read_files tool was used three times, succeeding twice and failing once
      And the tool usage records that the list_files tool was used once and succeeded
      When the operator reviews how the tools have been used
      Then the review shows the read_files tool was used three times
      And the review shows the read_files tool succeeded twice and failed once
      And the review shows the list_files tool was used once and succeeded
      And the operator is not asked for a prompt
      And tellme exits successfully

    Example: A tool that has never been used appears with no uses
      Given the operator has a runnable tellme installation
      And the tool usage records that the read_files tool was used once and succeeded
      When the operator reviews how the tools have been used
      Then the review shows the get_tree tool was never used
      And tellme exits successfully

    Example: The operator reviews the tools before any use
      Given the operator has a runnable tellme installation
      When the operator reviews how the tools have been used
      Then the review shows every tool with no uses
      And tellme exits successfully
