Feature: Piping the answer out

  # Interface truth (CLI end, `chat` module) — atomic rules for the output contract of a piped run.
  # Adopts the reference's TTY-aware posture (round-005 Clarify Q3): presentation is suppressed when
  # standard output is not a terminal. The E2E suite always captures stdout as a non-terminal pipe,
  # so the assertions below pin the redirected-output behaviour.

  Rule: A redirected answer is the answer text alone

    Example: The operator captures the answer from a redirected stream
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "List packages with go list"
      When the operator starts tellme with the prompt "How do I list Go packages?"
      Then the captured standard output is exactly "List packages with go list"
      And the captured standard output carries no terminal decoration
      And tellme exits successfully

  Rule: A piped run does not wait for interactive terminal input

    Example: The operator pipes a prompt in a non-interactive environment
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "Hello back"
      When the operator pipes "Hello" into tellme
      Then tellme completes the turn without waiting for terminal input
      And tellme prints the provider's answer "Hello back"
      And tellme exits successfully
