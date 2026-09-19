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

  Rule: The saved turn log stays free of decoration

    Example: A coloured terminal turn keeps a plain turn log
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator starts tellme with the prompt "Say hi."
      Then the session turn log carries no decoration
      And tellme exits successfully

  Rule: The tool output frame is grey and the action line is yellow at a terminal

    # Round 057 (ADR 0027): two more chrome elements gain a terminal-gated accent —
    # the whole `[Tool Output]` header line and BOTH horizontal separators are
    # grey; the whole `[Tool Action]` line is yellow. The streamed content lines
    # and the `turns.log` artifact stay plain.

    Example: A command run at a terminal colours the frame and the action line
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint runs a colouring command and then answers with "done"
      When the operator starts tellme with the prompt "Run the colouring command."
      Then the tool output frame is shown in grey
      And the action line is shown in yellow
      And tellme exits successfully
