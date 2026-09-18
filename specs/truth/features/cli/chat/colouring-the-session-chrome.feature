Feature: Colouring the session chrome

  # Interface truth (CLI end, `chat` module) — on a terminal the diagnostic turn
  # chrome accents four elements in green (round 054; ADR 0023). Acceptance
  # journeys: features/acceptance/colouring-the-session-chrome.feature.

  Rule: The chrome accents four elements in green at a terminal

    Example: A tool-using turn at a terminal
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And the working directory contains a file "notes.txt" whose text is "alpha"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with the reason "inspect notes" and then answers with "done" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      When the operator starts tellme with the prompt "Read the notes."
      Then the session chrome accents the tool reason, the mode, the measured tokens, and the session cost in green
      And tellme exits successfully

  Rule: The chrome is plain when the diagnostic stream is not a terminal

    Example: Piped diagnostics carry no colour
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator starts tellme with the prompt "Say hi."
      Then the session chrome carries no colour
      And tellme exits successfully
