Feature: Keeping the progress indicator visible while a command streams

  # Acceptance journey (business language). While a shell command streams its live output block, the
  # progress indicator presently stays hidden for the whole block, so a command that prints a line and
  # then works quietly for a while leaves the operator staring at a static header. The operator wants
  # the indicator to come back during a quiet stretch, and to leave no trace once the command finishes.

  Rule: A quiet stretch of a command's output shows the progress indicator again

    Example: The indicator returns while a long-running command is quiet
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint runs a command that prints a line and then stays quiet and then answers with "all good"
      When the operator starts tellme with the prompt "run it"
      Then the run shows the progress indicator again while the command stays quiet
      And tellme exits successfully

    Example: The indicator leaves no trace once the command finishes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint runs a command that prints a line and then stays quiet and then answers with "all good"
      When the operator starts tellme with the prompt "run it"
      Then the run shows no leftover indicator once the command finishes
      And tellme exits successfully

  Rule: The indicator is not shown when the operator is not watching a terminal

    # The negative is carried by the interface root (`the run shows no progress spinner`); no dedicated
    # acceptance journey is needed here.

    Example: A quiet command's output is unaffected when the diagnostics are not a terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a command that prints a line and then stays quiet and then answers with "all good"
      When the operator starts tellme with the prompt "run it"
      Then the run streamed the command's output on its diagnostic output
      And the run shows no progress spinner
      And tellme exits successfully
