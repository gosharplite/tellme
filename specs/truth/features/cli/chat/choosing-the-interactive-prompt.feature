Feature: Choosing the interactive prompt

  # Interface truth (CLI end, `chat` module). The rich interactive prompt is opt-in: a terminal
  # invocation without the flag keeps the round-012 plain reader, and a non-terminal input never
  # opens it. Acceptance journey:
  # features/acceptance/choosing-between-the-interactive-prompt-and-plain-input.feature.

  Rule: A terminal invocation without the flag keeps the plain reader

    Example: The plain multi-line reader engages by default
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator pipes "What is two plus two?" into tellme
      Then the reading announcement is reported on the diagnostic output
      And the interactive prompt is not shown
      And tellme prints the provider's answer "ok"
      And tellme exits successfully

  Rule: A non-terminal input never opens the interactive prompt

    Example: Enabling the prompt with piped input falls back
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator pipes "What is two plus two?" into tellme with the interactive prompt enabled
      Then the interactive prompt is not shown
      And tellme prints the provider's answer "ok"
      And tellme exits successfully
