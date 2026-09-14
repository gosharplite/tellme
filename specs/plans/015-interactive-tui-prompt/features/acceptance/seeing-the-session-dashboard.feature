Feature: Seeing the session dashboard at the prompt

  # Acceptance only: the interactive prompt tells the operator, before they
  # send, which provider/model is active and how much of the token budget and
  # how many turns the session has used.

  Rule: The interactive prompt shows the current session's provider, token usage, and turn count

    Example: The dashboard reflects the active session
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose active provider is "vertex-flash"
      When the operator opens the interactive prompt
      Then the interactive prompt shows the active provider "vertex-flash"
      And the interactive prompt shows the session's token usage
      And the interactive prompt shows the session's turn count

    Example: The dashboard updates after a turn completes
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose active provider is "vertex-flash"
      When the operator opens the interactive prompt
      And the operator submits the prompt "hello"
      And the operator opens the interactive prompt again
      Then the interactive prompt shows at least one turn
