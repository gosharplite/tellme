Feature: Piping the answer out

  # Interface truth (CLI end, `chat` module) — atomic rules for the output contract of a raw run and a
  # piped run. Under round-006 reference parity the byte-exact/plain output is the "-r"/"--raw" contract
  # (the default output is rendered — see rendering-the-answer.feature); the round-005 FR-007 amendment
  # moved the byte-fidelity assertion to the raw path. Presentation introduced by tellme's own chrome is
  # suppressed when stdout is not a terminal. The E2E suite always captures stdout as a non-terminal pipe.
  # Quoted params support the escapes `\n`, `\t`, `\\`, and `\xHH` (decoded by the step definitions).

  Rule: A raw ("-r") answer is the answer text alone

    Example: The operator captures the raw answer from a redirected stream
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "List packages with go list"
      When the operator starts tellme with the prompt "How do I list Go packages?" and the raw flag
      Then the captured standard output is exactly "List packages with go list"
      And the captured standard output carries no terminal decoration
      And tellme exits successfully

    Example: A newline-terminated raw answer is passed through verbatim with only the terminating newline appended
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "line one\nline two\n"
      When the operator starts tellme with the prompt "print two lines" and the raw flag
      Then the captured standard output is exactly "line one\nline two\n"
      And the captured standard output carries no terminal decoration
      And tellme exits successfully

    Example: A raw answer carrying control bytes is passed through unchanged and is not flagged as decoration
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "red \x1b[31mtext"
      When the operator starts tellme with the prompt "colourise" and the raw flag
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
