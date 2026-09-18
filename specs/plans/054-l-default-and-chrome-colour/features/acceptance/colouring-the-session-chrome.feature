Feature: Colouring the session chrome

  # Acceptance only (round 054). On a terminal, the operator's turn chrome accents
  # four elements in green so a long session is easier to scan. Redirected output
  # and the session's saved turn log stay plain. Business language only.

  Rule: On a terminal the chrome accents four elements in green

    Example: A tool-using turn on a terminal shows the green accents
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And the working directory contains a file "notes.txt" whose text is "alpha"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with the reason "inspect notes" and then answers with "done"
      When the operator starts tellme with the prompt "Read the notes."
      Then the session chrome accents the tool reason, the mode, the measured tokens, and the session cost in green
      And tellme exits successfully

  Rule: Redirected output stays plain

    Example: Piped diagnostics carry no colour
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator starts tellme with the prompt "Say hi."
      Then the session chrome carries no colour
      And tellme exits successfully

  Rule: The saved turn log stays free of decoration

    Example: A terminal turn keeps its saved turn log plain
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator starts tellme with the prompt "Say hi."
      Then the session turn log carries no decoration
      And tellme exits successfully
