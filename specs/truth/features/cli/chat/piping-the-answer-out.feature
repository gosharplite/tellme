Feature: Piping the answer out

  # Interface truth (CLI end, `chat` module) — atomic rules for the output contract of a piped run.
  # Adopts the reference's TTY-aware posture (round-005 Clarify Q3): presentation is suppressed when
  # standard output is not a terminal. The E2E suite always captures stdout as a non-terminal pipe,
  # so the assertions below pin the redirected-output behaviour: the answer bytes verbatim, plus a
  # single CLI-appended terminating newline — no system-added decoration.
  # Quoted params support the escapes `\n`, `\t`, `\\`, and `\xHH` (decoded by the step definitions).

  Rule: A redirected answer is the answer text alone

    Example: The operator captures the answer from a redirected stream
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "List packages with go list"
      When the operator starts tellme with the prompt "How do I list Go packages?"
      Then the captured standard output is exactly "List packages with go list"
      And the captured standard output carries no terminal decoration
      And tellme exits successfully

    Example: A newline-terminated answer is passed through verbatim with only the terminating newline appended
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "line one\nline two\n"
      When the operator starts tellme with the prompt "print two lines"
      Then the captured standard output is exactly "line one\nline two\n"
      And the captured standard output carries no terminal decoration
      And tellme exits successfully

    Example: An answer carrying control bytes is passed through unchanged and is not flagged as decoration
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "red \x1b[31mtext"
      When the operator starts tellme with the prompt "colourise"
      Then the captured standard output is exactly "red \x1b[31mtext"
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
