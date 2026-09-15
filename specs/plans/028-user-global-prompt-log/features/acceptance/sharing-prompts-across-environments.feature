Feature: Sharing prompts across tellme environments

  # Acceptance only: the interactive prompt history is kept once per user, so a
  # prompt typed in one tellme environment is offered in another; a one-shot
  # (non-interactive) run never writes the shared prompt log.

  Rule: A prompt submitted interactively is remembered across environments

    Example: A prompt recorded in one environment is offered in another
      Given the operator has a runnable tellme installation
      And the shared prompt log is empty
      When the operator opens the interactive prompt in the "work" environment
      And the operator types "summarize the last release"
      And the operator submits the prompt
      Then the shared prompt log records exactly the prompt "summarize the last release"
      When the operator opens the interactive prompt in the "personal" environment
      Then the interactive prompt offers the recent prompt "summarize the last release"

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
