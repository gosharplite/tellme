Feature: Colouring every line of a command's output block

  # Acceptance only (round 058). Business language: when a command's output is
  # shown at a terminal, the operator wants the WHOLE block — the announcing line,
  # every line of output, and the rules around it — to read as one grey region,
  # not a grey frame around plain text. Colour still never appears when the output
  # is piped or raw, and the saved copy stays plain.

  Rule: The whole output block is grey

    Example: A command that prints is shown as one grey block
      Given the operator has a runnable tellme installation
      And colours are shown in the terminal
      And the assistant runs a command that prints a few lines
      When the operator asks tellme to run it
      Then the line announcing the command's output is shown in grey
      And every line of the command's output is shown in grey
      And the opening and closing rules around the output are shown in grey
      And tellme exits successfully

  Rule: The block is plain when the output is not a terminal or raw output is asked for

    Example: Piped output carries no colour
      Given the operator has a runnable tellme installation
      And the diagnostic output is piped rather than shown in a terminal
      And the assistant runs a command that prints a few lines
      When the operator asks tellme to run it
      Then every line of the command's output carries no colour at all
      And tellme exits successfully

    Example: Raw output carries no colour
      Given the operator has a runnable tellme installation
      And the operator asks for raw output
      And the assistant runs a command that prints a few lines
      When the operator asks tellme to run it
      Then every line of the command's output carries no colour at all
      And tellme exits successfully

  Rule: The saved copy of the turn stays plain

    Example: The saved turn log carries no colour
      Given the operator has a runnable tellme installation
      And colours are shown in the terminal
      And the assistant runs a command that prints a few lines
      When the operator asks tellme to run it
      Then the saved turn log carries the same lines without any colour
      And tellme exits successfully
