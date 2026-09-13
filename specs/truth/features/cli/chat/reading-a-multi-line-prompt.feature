Feature: Reading a multi-line prompt

  # Interface truth (CLI end, `chat` module). The interactive multi-line reader is a
  # POSIX-terminal capability. The POSITIVE read is driven end-to-end through the
  # `TELL_ME_FORCE_STDIN_TTY` diagnostic seam (`the operator is working at an interactive
  # terminal`, round-012 review RF1): print the hint to `stderr`, read the prompt from the
  # "terminal" to EOF, bounded 1 MiB, one reasoning turn. The empty/cancel contract (no
  # request, exit `0`) stays a unit pin (`internal/cli`). The reader never engages on a
  # non-terminal input: neither a pipe nor the null device (a character device that is not
  # a terminal — the round-012 BLOCKER B1 fix, review RF2) prints a reading announcement.

  Rule: At a terminal, tellme reads the multi-line prompt and announces it

    Example: A prompt supplied at the interactive terminal is read and answered
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator pipes "What is two plus two?" into tellme
      Then the reading announcement is reported on the diagnostic output
      And tellme prints the provider's answer "ok"
      And tellme exits successfully

  Rule: A piped prompt never enters the interactive reader

    Example: A piped prompt is handled without the interactive announcement
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator pipes "What is two plus two?" into tellme
      Then no reading announcement is reported
      And tellme prints the provider's answer "ok"
      And tellme exits successfully

  Rule: A character-device input that is not a terminal never enters the interactive reader

    Example: The null device on standard input takes the boot path
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml"
      When the operator starts tellme with the null device on standard input
      Then no reading announcement is reported
      And tellme exits successfully
