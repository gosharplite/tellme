Feature: Recording the shared prompt log

  # Interface truth (CLI end, `chat` module). An `-i` interactive submission is recorded in the
  # shared, append-only prompt log at ~/.tellme/global_prompts.jsonl (the per-user log, shared across
  # every environment/repository/mode — round 028 moved it out of $TELL_ME_HOME/output/); a one-shot
  # or piped run reads and writes nothing. Acceptance journey:
  # features/acceptance/sharing-prompts-across-environments.feature.

  Rule: An interactive submission records the prompt in the shared prompt log

    Example: A submitted prompt is appended to the shared log
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator submits the prompt "summarize the last release" at the interactive prompt
      Then the shared prompt log records the prompt "summarize the last release"
      And tellme exits successfully

  Rule: A one-shot run leaves the shared prompt log untouched

    Example: A prompt argument does not write the shared log
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      And the shared prompt log already holds "review the last two commits"
      When the operator starts tellme with the prompt "What is the capital of France?"
      Then the shared prompt log still holds only "review the last two commits"
      And tellme exits successfully

  Rule: A piped run leaves the shared prompt log untouched

    Example: A piped prompt does not write the shared log
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      And the shared prompt log already holds "review the last two commits"
      When the operator pipes "summarize the changelog" into tellme
      Then the shared prompt log still holds only "review the last two commits"
      And tellme exits successfully
