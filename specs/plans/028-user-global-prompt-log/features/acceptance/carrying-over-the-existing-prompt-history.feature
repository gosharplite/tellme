Feature: Carrying existing prompt history into the shared prompt log

  # Acceptance only: when the shared prompt log moves to the operator's home, a
  # first interactive run carries over whatever prompt history the environment
  # already had, so nothing the operator has already typed is lost.

  Rule: On first use, the environment's existing prompt history is carried over

    Example: Prompt history recorded earlier is offered after the move
      Given the operator has a runnable tellme installation
      And the shared prompt log does not yet exist
      And the environment already has prompt history containing "review the last two commits"
      When the operator opens the interactive prompt
      Then the interactive prompt offers the recent prompt "review the last two commits"

  Rule: An existing shared prompt log is never overwritten

    Example: An already-present shared log is left as-is
      Given the operator has a runnable tellme installation
      And the shared prompt log already contains "deploy to staging"
      And the environment already has prompt history containing "review the last two commits"
      When the operator opens the interactive prompt
      Then the shared prompt log still contains only "deploy to staging"
      And the interactive prompt offers the recent prompt "deploy to staging"

  Rule: With no prior history, the operator starts with an empty log

    Example: A first-ever run has nothing to carry over
      Given the operator has a runnable tellme installation
      And the shared prompt log does not yet exist
      And the environment has no prompt history
      When the operator opens the interactive prompt
      Then the interactive prompt offers no recent prompt
      And tellme exits successfully
