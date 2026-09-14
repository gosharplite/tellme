Feature: Using the interactive prompt

  # Interface truth (CLI end, `chat` module). Submitting at the `-i` interactive prompt runs exactly
  # one reasoning turn; aborting sends nothing. (Strict parity, round 016: the prompt carries NO
  # session dashboard header.) Acceptance journey:
  # features/acceptance/seeing-a-prompt-that-matches-the-reference.feature. Driven end-to-end through
  # the `TELL_ME_FORCE_STDIN_TTY` seam with a scripted key sequence.

  Rule: Submitting at the interactive prompt runs exactly one reasoning turn

    Example: A composed prompt is submitted
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator submits the prompt "What is two plus two?" at the interactive prompt
      Then the interactive prompt is shown
      And tellme sends exactly one request to the provider "test-model"
      And tellme prints the provider's answer "ok"
      And tellme exits successfully

  Rule: Aborting the interactive prompt sends no request

    Example: The operator aborts before sending
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator aborts the interactive prompt
      Then tellme sends no request to the provider "test-model"
      And tellme exits successfully
