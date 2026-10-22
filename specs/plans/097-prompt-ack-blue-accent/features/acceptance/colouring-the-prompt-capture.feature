Feature: Colouring the prompt-capture acknowledgement

  # Acceptance only (round 097; issue #204). Business language: when the operator
  # hands tellme a prompt, tellme acknowledges it with one short line — "Input
  # captured. Processing...". That line is the moment the operator's words enter
  # the session, so it should be visually distinct from the plain rule and header
  # that follow it. The accent is a terminal nicety: piped output, raw output, and
  # the saved turn log stay exactly as they are.

  Rule: The acknowledgement is marked in blue when the operator is watching a terminal

    Example: A prompt turn at a terminal marks the acknowledgement
      Given the operator has a runnable tellme installation
      And the operator is watching the diagnostics at a terminal
      When the operator asks tellme with the prompt "What is the current datetime?"
      Then the acknowledgement of the captured prompt is marked in blue
      And tellme exits successfully

    Example: The accent covers the whole line, timestamp included
      Given the operator has a runnable tellme installation
      And the operator is watching the diagnostics at a terminal
      When the operator asks tellme with the prompt "Say hi."
      Then the whole acknowledgement line, its timestamp included, is marked in blue
      And tellme exits successfully

  Rule: The accent disappears the moment the operator is not watching a terminal

    Example: Piped diagnostics keep the acknowledgement plain
      Given the operator has a runnable tellme installation
      When the operator asks tellme with the prompt "Say hi."
      Then the acknowledgement of the captured prompt is shown plain
      And tellme exits successfully

    Example: Raw output keeps the acknowledgement, without the accent
      Given the operator has a runnable tellme installation
      And the operator is watching the diagnostics at a terminal
      When the operator asks tellme with raw output and the prompt "Say hi."
      Then the acknowledgement of the captured prompt is still shown
      And the acknowledgement of the captured prompt is shown plain
      And tellme exits successfully

    Example: A coloured terminal turn still saves a plain turn log
      Given the operator has a runnable tellme installation
      And the operator is watching the diagnostics at a terminal
      When the operator asks tellme with the prompt "Say hi."
      Then the saved turn log carries no decoration
      And tellme exits successfully

  Rule: Every way of giving tellme a prompt gets the same accent

    # The point of choosing this line over the operator's prompt text: the prompt
    # text is echoed by the terminal itself on one surface, styled by the editor on
    # another, and not echoed at all on a third — so only the acknowledgement can
    # look the same everywhere.

    Example: A prompt given as an argument gets the accent
      Given the operator has a runnable tellme installation
      And the operator is watching the diagnostics at a terminal
      When the operator asks tellme with the prompt "Say hi."
      Then the acknowledgement of the captured prompt is marked in blue
      And tellme exits successfully

    Example: A prompt typed at the interactive prompt gets the accent
      Given the operator has a runnable tellme installation
      And the operator is watching the diagnostics at a terminal
      And the operator is working at an interactive terminal
      When the operator submits the prompt "hi" at the interactive prompt
      Then the acknowledgement of the captured prompt is marked in blue
      And tellme exits successfully
