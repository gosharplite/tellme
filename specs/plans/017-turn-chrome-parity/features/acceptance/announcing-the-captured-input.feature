Feature: Announcing the captured input

  # Acceptance only: the moment a prompt-bearing run has taken the operator's
  # prompt, it announces that capture before the turn is framed — reproducing
  # tell-me-go's opening acknowledgement.

  Rule: A prompt-bearing run announces the captured input before the turn

    Example: The operator passes the prompt as an argument
      Given the operator has a runnable tellme installation
      When the operator runs tellme with the prompt "summarise the changelog"
      Then the run announces that the input was captured
      And the announcement names the time the input was captured
      And the announcement is shown before the turn is framed

    Example: The operator pipes the prompt in
      Given the operator has a runnable tellme installation
      When the operator pipes "list the open issues" into tellme
      Then the run announces that the input was captured
      And the announcement is shown before the turn is framed

    Example: The operator composes the prompt at the terminal
      Given the operator is working at an interactive terminal
      And the operator starts a fresh session without a prompt
      When the operator sends a prompt across several lines
      Then the run announces that the input was captured
      And the announcement is shown before the turn is framed
