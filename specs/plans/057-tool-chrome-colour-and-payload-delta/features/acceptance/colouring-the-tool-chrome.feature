Feature: Colouring the tool chrome

  # Acceptance only (round 057). Business language: the operator wants the tool
  # stream to be easier to scan — the block that shows a command's output framed
  # in grey, and the line that announces an action in yellow. Colour is a
  # terminal nicety: it never appears when the output is piped or raw.

  Rule: The block that frames a command's output is grey

    Example: A command that prints is shown with a grey frame
      Given the operator has a runnable tellme installation
      And colours are shown in the terminal
      And the assistant runs a command that prints a few lines
      When the operator asks tellme to run it
      Then the line announcing the command's output is shown in grey
      And the opening and closing rules around the output are shown in grey
      And tellme exits successfully

  Rule: The line announcing an action is yellow

    Example: An action line is shown in yellow
      Given the operator has a runnable tellme installation
      And colours are shown in the terminal
      And the assistant runs a command that prints a few lines
      When the operator asks tellme to run it
      Then the line announcing the action is shown in yellow
      And tellme exits successfully

  Rule: No colour is used when the output is not a terminal or raw output is asked for

    Example: Piped output carries no colour
      Given the operator has a runnable tellme installation
      And the diagnostic output is piped rather than shown in a terminal
      And the assistant runs a command that prints a few lines
      When the operator asks tellme to run it
      Then the announcing lines carry no colour at all
      And tellme exits successfully

    Example: Raw output carries no colour
      Given the operator has a runnable tellme installation
      And the operator asks for raw output
      And the assistant runs a command that prints a few lines
      When the operator asks tellme to run it
      Then the announcing lines carry no colour at all
      And tellme exits successfully

  Rule: The saved copy of the turn carries the same lines but no colour

    Example: The saved turn log is plain
      Given the operator has a runnable tellme installation
      And colours are shown in the terminal
      And the assistant runs a command that prints a few lines
      When the operator asks tellme to run it
      Then the saved turn log carries the same lines without any colour
      And tellme exits successfully
