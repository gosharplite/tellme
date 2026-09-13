Feature: Reading a multi-line prompt

  # Interface truth (CLI end, `chat` module). The interactive multi-line reader is a
  # POSIX-terminal capability that the pty-less E2E harness CANNOT drive (the round-005
  # named pin). The POSITIVE read — print the hint to `stderr`, read the prompt from a
  # terminal to EOF (`Ctrl+D`), bounded 1 MiB, one reasoning turn — and the empty/cancel
  # contract (no request) are verified at the UNIT layer through the injected terminal
  # seam (`internal/cli`, `runtimeEnv.isTTY`), not by a Gherkin step. This feature pins
  # the one E2E-observable fact: the reader never engages on a non-terminal input, so a
  # piped run prints NO reading announcement.

  Rule: A piped prompt never enters the interactive reader

    Example: A piped prompt is handled without the interactive announcement
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator pipes "What is two plus two?" into tellme
      Then no reading announcement is reported
      And tellme prints the provider's answer "ok"
      And tellme exits successfully
