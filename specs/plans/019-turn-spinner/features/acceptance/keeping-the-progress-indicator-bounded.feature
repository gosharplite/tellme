Feature: The progress indicator stays bounded

  # Acceptance only: the indicator appears only when the operator is watching a
  # terminal, only on a prompt-bearing run, and it never changes the answer or
  # the other commands.

  Rule: The indicator is shown only when the diagnostics reach a terminal

    Example: The operator is not watching the diagnostics at a terminal
      # the diagnostics are not attached to a terminal in this scenario
      Given the operator has a runnable tellme installation
      When the operator runs tellme with the prompt "hi"
      Then the run shows no progress indicator
      And the answer is unchanged by the progress indicator

    Example: The operator asks for the raw answer at a terminal
      Given the operator has a runnable tellme installation
      And the operator is watching a terminal
      And the operator asks for the raw answer
      When the operator runs tellme with the prompt "hi"
      Then the run shows no progress indicator

  Rule: The other commands keep their existing output

    Example: The operator asks for the version
      Given the operator has a runnable tellme installation
      When the operator runs tellme with the version flag
      Then the run prints the version without a progress indicator

    Example: The operator lists the recent conversation
      Given the operator has a runnable tellme installation
      When the operator lists the recent conversation
      Then the run prints the conversation without a progress indicator

  Rule: The interactive prompt is unchanged

    Example: The operator composes the prompt interactively
      Given the operator has a runnable tellme installation
      When the operator composes the prompt with the interactive prompt
      Then the interactive prompt shows no progress indicator

  Rule: A failed run leaves no indicator

    Example: The provider request fails
      Given the operator has a runnable tellme installation
      And the operator is watching a terminal
      And the provider is unavailable
      When the operator runs tellme with the prompt "hi"
      Then the run reports the failure
      And no progress indicator remains
