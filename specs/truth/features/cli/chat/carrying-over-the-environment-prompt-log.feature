Feature: Carrying over the environment prompt log

  # Interface truth (CLI end, `chat` module). On the first interactive run after the shared prompt
  # log moved to the user's home (~/.tellme/global_prompts.jsonl, round 028), tellme copies the
  # environment's existing prompt log ($TELL_ME_HOME/output/global_prompts.jsonl) into it, so the
  # operator's earlier prompts are offered. The copy is skipped when the shared log already exists.
  # Acceptance journey: features/acceptance/carrying-over-the-existing-prompt-history.feature.

  Rule: On first use, the environment prompt log is carried into the shared prompt log

    Example: An existing environment prompt log is copied into the absent shared prompt log
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log has not been created yet
      And the environment prompt log already holds "review the last two commits"
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator submits the prompt "summarize the last release" at the interactive prompt
      Then the shared prompt log records the prompt "summarize the last release"
      And the shared prompt log also holds the earlier prompt "review the last two commits"
      And tellme exits successfully

  Rule: An existing shared prompt log is not overwritten

    Example: A present shared prompt log is left as-is (no carry-over)
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log already holds "deploy to staging"
      And the environment prompt log already holds "review the last two commits"
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator submits the prompt "ship it" at the interactive prompt
      Then the shared prompt log records the prompt "ship it"
      And the shared prompt log does not hold "review the last two commits"
      And tellme exits successfully

  Rule: With no environment prompt log, the shared prompt log starts empty

    Example: A first-ever run records only the submitted prompt
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the shared prompt log has not been created yet
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator submits the prompt "first prompt" at the interactive prompt
      Then the shared prompt log records the prompt "first prompt"
      And tellme exits successfully
