Feature: Handing the terminal back after the interactive prompt

  # Acceptance only: what the operator sees once they finish at tellme's
  # interactive prompt — the editor gives the terminal back instead of lingering
  # on screen, whether they send or abandon the prompt.

  Rule: The interactive editor releases the terminal when the operator is done

    Example: The operator sends a composed prompt
      Given the operator has a runnable tellme installation
      And the operator is working at an interactive terminal
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator submits the prompt "hi" at the interactive prompt
      Then the editor no longer occupies the terminal
      And tellme exits successfully

    Example: The operator abandons the prompt
      Given the operator has a runnable tellme installation
      And the operator is working at an interactive terminal
      When the operator abandons the interactive prompt
      Then the editor no longer occupies the terminal
      And tellme sends no request to the provider
      And tellme exits successfully
