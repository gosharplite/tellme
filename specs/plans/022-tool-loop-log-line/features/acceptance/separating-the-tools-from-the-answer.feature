Feature: Separating the tool report from the answer

  # Acceptance only: the answer is visually set apart from the tool report that
  # precedes it in a tool-using run, and a plain run is unaffected.

  Rule: The answer is set apart from the tool report

    Example: A run that used a tool
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that lists the directory before answering
      When the operator asks tellme "What is in the current directory?"
      Then the run reports the tool it used
      And a blank line separates the last tool report from the answer
      And tellme exits successfully

    Example: A run that used no tool adds no separating blank line
      Given the operator has a runnable tellme installation
      And the operator is working at an interactive terminal
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator submits the prompt "Say hello." at the interactive prompt
      Then tellme answers directly with no tool report
      And no separating blank line is added before the answer
      And tellme exits successfully
