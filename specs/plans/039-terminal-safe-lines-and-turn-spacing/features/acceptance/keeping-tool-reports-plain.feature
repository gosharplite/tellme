Feature: Keeping the tool reports plain

  # Acceptance journey (business language). Everything the run prints about a tool call — the
  # reason the model gave, the result the tool returned, and the arguments the model supplied —
  # must be shown as plain text. Text that carries terminal control data (colour codes, cursor
  # moves, window titles, …) must not be able to change how the operator's terminal renders the
  # rest of the run. (Round 038 already keeps the live command-output block plain; this extends
  # the same guarantee to the reason, the result, and the arguments.)

  Rule: A tool report is shown as plain text

    Example: A reason that carries control data does not tint the terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with a reason that carries terminal control data and then answers with "done"
      When the operator starts tellme with the prompt "Read notes.txt."
      Then the run's reason report is shown as plain text
      And the operator still reads the reason's visible text
      And tellme exits successfully

    Example: A result that carries control data is shown as plain text
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "colour.txt" whose text carries terminal control data
      And a configured provider "test-model" whose endpoint asks tellme to read "colour.txt" and then answers with "done"
      When the operator starts tellme with the prompt "Read colour.txt."
      Then the run's result report is shown as plain text
      And tellme exits successfully

    Example: An argument that carries control data is shown as plain text
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint creates the file "out.txt" with a name that carries terminal control data and then answers with "done"
      When the operator starts tellme with the prompt "Create the file."
      Then the run's action report is shown as plain text
      And tellme exits successfully
