Feature: Presenting the progress spinner

  # Interface truth (CLI end, `chat` module) — while a prompt-bearing turn on the non-TUI surfaces is
  # waiting (awaiting the model, or running tools) the run shows a live progress spinner on the
  # diagnostic stream (`stderr`): a braille frame, a phase label naming the model or the tool(s), and a
  # whole-seconds elapsed counter; while tools run the line also reports the machine's CPU/memory. The
  # spinner is drawn only when the diagnostic stream (`stderr`) is a terminal and `-r` is off (`the operator
  # is watching a terminal` arranges the round-019 `stderr` terminal seam); it yields the line to the answer. The
  # negatives are carried by the interface root (`the run shows no progress spinner`) on the `diagnostics`
  # and `history` modules (round 023 removes the `-i` surface from the negatives — the `-i` submit now
  # draws the spinner). Acceptance journeys:
  # features/acceptance/showing-a-progress-indicator.feature, labelling-the-progress-indicator.feature,
  # and keeping-the-progress-indicator-bounded.feature.

  Rule: A prompt-bearing run shows a live progress spinner while it waits

    Example: The operator runs a prompt at a terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "summarise the changelog"
      Then the run shows the progress spinner while it waits
      And tellme exits successfully

    Example: The operator pipes the prompt in at a terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator pipes "list the open issues" into tellme
      Then the run shows the progress spinner while it waits
      And tellme exits successfully

  Rule: The spinner names the model while the run waits for it

    Example: The operator waits for the model
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "hi"
      Then the progress spinner names the model it is waiting for
      And the progress spinner shows how long it has waited
      And tellme exits successfully

  Rule: The spinner names the tools and the resources while tools run

    Example: The turn runs a single tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "all good"
      When the operator starts tellme with the prompt "read the notes"
      Then the progress spinner names the tool it is running
      And the progress spinner reports the machine's resource usage
      And tellme exits successfully

    Example: The turn runs several tools at once
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and "todo.txt" and then answers with "all good"
      When the operator starts tellme with the prompt "read the notes and the todo"
      Then the progress spinner names the first tool and counts the remaining tools
      And the progress spinner reports the machine's resource usage
      And tellme exits successfully

  Rule: The spinner leaves no residue on a terminal narrower than its line

    Example: The turn runs several tools on a narrow terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal narrower than the indicator line
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and "todo.txt" and then answers with "all good"
      When the operator starts tellme with the prompt "read the notes and the todo"
      Then the progress indicator is cleared from every row it occupied
      And the progress spinner no longer appears once the answer is written
      And tellme exits successfully

  Rule: The spinner yields the line to the answer

    Example: The turn runs a tool before answering
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "all good"
      When the operator starts tellme with the prompt "read the notes"
      Then the progress spinner no longer appears once the answer is written
      And tellme exits successfully

  Rule: The spinner is drawn only when the diagnostics are shown at a terminal

    Example: The diagnostics are not shown at a terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "hi"
      Then the run shows no progress spinner
      And tellme exits successfully

    Example: The operator asks for the raw answer at a terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "hi" and the raw flag
      Then the run shows no progress spinner
      And tellme exits successfully

  Rule: The spinner is paused while a command's output streams

    Example: A command streams its output without the spinner flipping
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint runs a command and then answers with "all good"
      When the operator starts tellme with the prompt "run it"
      Then the progress spinner does not appear while the command's output streams
      And tellme exits successfully
