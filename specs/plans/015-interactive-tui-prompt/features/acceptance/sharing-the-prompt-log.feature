Feature: Sharing the prompt log with the rest of the environment

  # Acceptance only: tellme's interactive prompts are remembered in the same
  # shared, append-only prompt log the other persona tools use, so a prompt
  # typed in one run can be offered again in another; a one-shot (non-interactive)
  # run never writes that log.

  Rule: A prompt submitted interactively is remembered in the shared prompt log

    Example: A prompt submitted interactively is suggested on a later run
      Given the operator has a runnable tellme installation
      And the shared prompt log is empty
      When the operator opens the interactive prompt
      And the operator types "summarize the last release"
      And the operator submits the prompt
      Then the shared prompt log records exactly the prompt "summarize the last release"
      When the operator opens the interactive prompt again
      Then the interactive prompt offers the recent prompt "summarize the last release"

    Example: Recording grows the shared prompt log by exactly one entry
      Given the operator has a runnable tellme installation
      And the shared prompt log already contains "review the last two commits"
      When the operator opens the interactive prompt
      And the operator types "deploy to staging"
      And the operator submits the prompt
      Then the shared prompt log contains both "deploy to staging" and "review the last two commits"

  Rule: A non-interactive run leaves the shared prompt log untouched

    Example: A one-shot prompt does not write the shared prompt log
      Given the operator has a runnable tellme installation
      And the shared prompt log already contains "review the last two commits"
      When the operator asks tellme "What is the capital of France?"
      Then the shared prompt log still contains only "review the last two commits"

    Example: A piped prompt does not write the shared prompt log
      Given the operator has a runnable tellme installation
      And the shared prompt log already contains "review the last two commits"
      When the operator pipes "summarize the changelog" into tellme
      Then the shared prompt log still contains only "review the last two commits"
