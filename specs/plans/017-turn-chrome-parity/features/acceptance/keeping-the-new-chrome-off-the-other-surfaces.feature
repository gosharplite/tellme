Feature: The new chrome stays off the other surfaces

  # Acceptance only: the capture announcement and the turn frame appear only on
  # the non-interactive prompt surfaces; the interactive prompt (round 016) and
  # the non-prompt commands keep their existing output.

  Rule: The interactive prompt keeps its own surface

    Example: The operator sends a prompt through the interactive prompt
      Given the operator is working at an interactive terminal
      When the operator opens the interactive prompt
      And the operator sends a prompt
      Then the run shows the interactive prompt
      And the run shows no capture announcement for the turn
      And the run shows no turn rule or turn heading

  Rule: The non-prompt commands keep their existing output

    Example: The operator asks for the version
      Given the operator has a runnable tellme installation
      When the operator runs tellme with the version flag
      Then the run prints the version without framing a turn

    Example: The operator starts a fresh session with no prompt on a pipe
      Given the operator has a runnable tellme installation
      When the operator starts a fresh session with no prompt and no terminal
      Then the run starts a fresh session without framing a turn
