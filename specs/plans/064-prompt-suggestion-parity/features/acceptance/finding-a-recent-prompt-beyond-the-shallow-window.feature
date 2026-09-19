Feature: Finding a recent prompt beyond the shallow window

  # Acceptance only (round 064, operator request). Business language only.
  #
  # The operator typed "commit" at the `-i` interactive prompt and saw only three
  # suggestions, while the reference shows ten. The reason is depth: tellme only
  # looks at the ten most recent prompts, so anything older can never be offered.
  # This journey is about the prompt searching a *deep* window of recent prompts
  # (still showing no more than ten), so an older matching prompt is found again.

  Rule: The interactive prompt searches a deep window of recent prompts

    Example: A recent prompt older than the newest ten is offered again
      Given the operator has a runnable tellme installation
      And the shared prompt log already holds "review the last two commits"
      And many newer prompts were recorded after it that do not mention the term
      When the operator opens the interactive prompt and types "commits"
      Then the interactive prompt offers the recent prompt "review the last two commits"

    Example: A shallow log still offers everything it holds
      Given the operator has a runnable tellme installation
      And the shared prompt log already holds "commit and push to the dev branch"
      When the operator opens the interactive prompt and types "commit"
      Then the interactive prompt offers the recent prompt "commit and push to the dev branch"

  Rule: The prompt never shows more than ten suggestions

    Example: Only the ten nearest matches are offered
      Given the operator has a runnable tellme installation
      And the shared prompt log already holds twenty recent prompts that each mention "commit note"
      When the operator opens the interactive prompt and types "commit"
      Then the interactive prompt offers the recent prompt "commit note 20"
      And the interactive prompt does not offer the recent prompt "commit note 10"
