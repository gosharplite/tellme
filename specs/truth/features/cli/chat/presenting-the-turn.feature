Feature: Presenting the turn

  # Interface truth (CLI end, `chat` module) — a prompt-bearing turn on the non-TUI surfaces opens the
  # way tell-me-go does: an input-capture acknowledgement, an 80-column rule, a `╭─⠿ Turn N - <mode>`
  # header over the pre-flight payload line, and a blank gap before the answer. The chrome is emitted on
  # the positional/piped prompt turn, the round-012 plain reader, and — round 023 — the `-i` submit
  # surface (which first clears its editor frame). The non-prompt paths (`--version`, a
  # prompt-less `--new`) are carried by the `diagnostics` and `history` modules respectively. `stdout`
  # stays byte-exact. Acceptance journeys:
  # features/acceptance/announcing-the-captured-input.feature, framing-the-turn.feature, and
  # keeping-the-new-chrome-off-the-other-surfaces.feature.

  Rule: A prompt-bearing run announces the captured input

    Example: The operator passes the prompt as an argument
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "summarise the changelog"
      Then the input capture is announced for the turn
      And the input capture is announced before the turn frame
      And tellme exits successfully

    Example: The operator pipes the prompt in
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator pipes "list the open issues" into tellme
      Then the input capture is announced for the turn
      And the input capture is announced before the turn frame
      And tellme exits successfully

    Example: A prompt supplied at the interactive terminal is read and captured
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator pipes "carry on" into tellme
      Then the reading announcement is reported on the diagnostic output
      And the input capture is announced for the turn
      And tellme exits successfully

  Rule: A prompt-bearing run opens the turn with a rule and a turn header

    Example: The operator runs the first turn of a fresh session
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "hi"
      Then the turn opens with a horizontal rule
      And the turn is headed "Turn 1" for the active mode
      And tellme reports the estimated payload status for the turn
      And tellme exits successfully

    Example: The frame still opens the turn under the raw flag
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "hi" and the raw flag
      Then the input capture is announced for the turn
      And the turn opens with a horizontal rule
      And the turn is headed "Turn 1" for the active mode
      And the captured standard output is exactly "all good"
      And tellme exits successfully

  Rule: The turn header counts the model requests made so far

    Example: The operator continues a session that already made two plain requests
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
        | I use Go.         | Noted. |
      When the operator starts tellme with the prompt "carry on"
      Then the turn is headed "Turn 3" for the active mode
      And tellme exits successfully

    Example: The operator continues a session whose single earlier turn consulted a tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      And the session history already holds a tool-using exchange with no provider token
      When the operator starts tellme with the prompt "carry on"
      Then the turn is headed "Turn 3" for the active mode
      And tellme exits successfully

    Example: The operator starts a fresh session after an earlier tool-using turn
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      And the session history already holds a tool-using exchange with no provider token
      When the operator starts a fresh session with "--new" and the prompt "hi"
      Then the turn is headed "Turn 1" for the active mode
      And tellme exits successfully

  Rule: The turn frame is separated from the answer

    Example: The operator runs a turn on a fresh session
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "hi"
      Then the turn chrome is shown before the answer
      And the turn frame is separated from the answer
      And tellme exits successfully

