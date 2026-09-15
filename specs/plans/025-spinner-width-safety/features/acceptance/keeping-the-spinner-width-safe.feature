Feature: Keeping the spinner width-safe

  # Acceptance only: the tool-phase indicator stays short however many tools run
  # at once, and the indicator leaves nothing behind on a terminal narrower than
  # its line — so a long tool batch neither clips the resource segment nor
  # survives the answer.

  Rule: The tool label stays short however many tools run at once

    Example: The turn runs several tools at once
      Given the operator has a runnable tellme installation
      And the diagnostics are shown at a terminal
      And the model asks for the tools "read the notes", "list the folder", "run the tests", and "show the tree" before answering
      When the operator runs tellme with the prompt "tidy the repository"
      Then the progress indicator names the first tool and counts the remaining tools
      And the progress indicator reports the machine's resource usage

    Example: The turn runs a single tool
      Given the operator has a runnable tellme installation
      And the diagnostics are shown at a terminal
      And the model asks for the tool "read the notes" before answering
      When the operator runs tellme with the prompt "read the notes"
      Then the progress indicator names the tool it is running

  Rule: The indicator leaves no residue on a terminal narrower than its line

    Example: The operator's terminal is narrower than the indicator line
      Given the operator has a runnable tellme installation
      And the diagnostics are shown at a terminal narrower than the indicator line
      And the model asks for the tools "read the notes", "list the folder", "run the tests", and "show the tree" before answering
      When the operator runs tellme with the prompt "tidy the repository"
      Then the progress indicator line is wider than the terminal
      And no progress indicator remains once the answer is written
