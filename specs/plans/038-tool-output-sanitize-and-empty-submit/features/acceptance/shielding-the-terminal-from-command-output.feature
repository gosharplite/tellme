Feature: Shielding the terminal from a command's output

  # Acceptance journey (business language). The live `[Tool Output]` block that the operator sees
  # while a shell command runs must not be able to change how the terminal renders the rest of the
  # run. Command output that carries terminal control data (colour codes, cursor moves, window
  # titles, …) is shown as plain text, and the block leaves the terminal in its default state when it
  # closes — whatever way it closes.

  Rule: A command's streamed output is shown without terminal control data

    Example: A colouring command does not tint the terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a colouring command and then answers with "done"
      When the operator starts tellme with the prompt "Run the colouring command."
      Then the run streamed the command's output on its diagnostic output
      And the run streamed the command's output free of terminal control sequences
      And tellme exits successfully

  Rule: The streamed output block closes with the terminal in its default state

    Example: A colouring command stopped mid-output leaves the terminal neutral
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a colouring command that is stopped at its time limit and then answers with "done"
      When the operator starts tellme with the prompt "Run the long colouring command."
      Then the run streamed the command's output free of terminal control sequences
      And the terminal is left in its default state
      And tellme exits successfully
