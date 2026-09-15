Feature: Continuing the interactive prompt

  # Interface truth (CLI end, `chat` module). Round 023: on submit (`Ctrl+S`/`Alt+Enter`) or abort
  # (`Esc`/`Ctrl+C`) the `-i` interactive prompt CLEARS its editor frame (reference parity — the box
  # does not linger); a submitted prompt then continues on the standard turn surface — the submitted
  # prompt is ECHOED (the box that held it is gone), then the input-capture acknowledgement, the
  # 80-column `─` rule, the `╭─⠿ Turn N - <mode>` header, and (at a terminal diagnostic stream) the live
  # spinner. The other prompt surfaces are unchanged and do NOT echo the prompt. Acceptance journeys:
  # features/acceptance/handing-back-the-terminal.feature and
  # features/acceptance/continuing-on-the-standard-surface.feature. Driven end-to-end through the
  # `TELL_ME_FORCE_STDIN_TTY` seam with a scripted key sequence.

  Rule: The editor clears when the operator finishes

    Example: The operator sends the prompt
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator submits the prompt "hi" at the interactive prompt
      Then the interactive prompt is cleared
      And tellme exits successfully

    Example: The operator abandons the prompt
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      When the operator aborts the interactive prompt
      Then the interactive prompt is cleared
      And tellme exits successfully

  Rule: A submitted interactive prompt continues on the standard turn surface

    Example: The submitted prompt is echoed and the turn is framed
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator submits the prompt "hi" at the interactive prompt
      Then the submitted prompt "hi" is echoed on the diagnostic output
      And the input capture is announced for the turn
      And the turn opens with a horizontal rule
      And the turn is headed "Turn 1" for the active mode
      And tellme prints the provider's answer "all good"
      And tellme exits successfully

    Example: A prompt entered positionally is not echoed back
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "hi"
      Then the turn opens with a horizontal rule
      And the prompt is not echoed on the diagnostic output
      And tellme exits successfully

  Rule: The interactive prompt shows the progress spinner while it waits

    Example: An interactive run waits at a terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator submits the prompt "hi" at the interactive prompt
      Then the run shows the progress spinner while it waits
      And tellme exits successfully
