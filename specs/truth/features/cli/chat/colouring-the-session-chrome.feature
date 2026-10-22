Feature: Colouring the session chrome

  # Interface truth (CLI end, `chat` module) — on a terminal the diagnostic turn
  # chrome accents four elements in green (round 054; ADR 0023), and — since
  # round 097 (ADR 0067) — a fifth element, the input-capture acknowledgement, in
  # bright blue. Acceptance journeys: features/acceptance/colouring-the-session-chrome.feature
  # (round 054) and features/acceptance/colouring-the-prompt-capture.feature (round 097).

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

    # Round 058 (ADR 0028; F-058-1): the negative must exercise an ACTUAL
    # `[Tool Output]` block — the tool-less Example above builds none, so it cannot
    # witness "the block is plain when the gate is off". This Example runs a
    # command (so the block exists) off a terminal; the first Then is the
    # anti-vacuity clause.
    Example: A piped command run keeps its whole output block plain
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a colouring command and then answers with "done"
      When the operator starts tellme with the prompt "Run the colouring command."
      Then the run streamed the command's output on its diagnostic output
      And the session chrome carries no colour
      And tellme exits successfully

  Rule: The saved turn log carries the same content, plain

    # Round 057 (ADR 0027; F-057-2): the operator's Q3 → 1 lock is CONTENT parity —
    # the file carries the same lines as the terminal (including the new estimated
    # `+<delta> ~<n>` payload shape) but no colour.

    Example: A coloured terminal turn keeps a plain turn log
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator starts tellme with the prompt "Say hi."
      Then the session turn log carries no decoration
      And the saved turn log carries the pre-flight payload with its increment and no allowance
      And tellme exits successfully

  Rule: Every tool output line is grey and the action line is yellow at a terminal

    # Round 057 (ADR 0027) greyed the `[Tool Output]` header + BOTH separators and
    # the yellow `[Tool Action]` line. Round 058 (ADR 0028) EXTENDS the grey to the
    # streamed CONTENT lines too, so the whole `[Tool Output]` block reads as one
    # grey region (superseding round 057's A3). The `turns.log` artifact stays plain.

    Example: A command run at a terminal colours the whole output block and the action line
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint runs a colouring command and then answers with "done"
      When the operator starts tellme with the prompt "Run the colouring command."
      Then the tool output frame is shown in grey
      And the action line is shown in yellow
      And tellme exits successfully

  Rule: The input-capture acknowledgement is blue at a terminal

    # Round 097 (ADR 0067; closes #204). The acknowledgement is the one
    # tellme-owned line that marks the moment the operator's prompt enters the
    # session. It is accented in bright blue because blue's only other meaning —
    # the `[USER]` header of the `-l` listing — is the SAME concept (the
    # operator's side) on a DIFFERENT surface (`stdout`), so the two never share a
    # block; green's dominant instance is the `$` session cost on the `Ready` line,
    # three lines below the ack. The two Examples below are the cross-surface
    # witness (the accent must not be a per-surface accident).

    Example: A prompt given as an argument is acknowledged in blue
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator starts tellme with the prompt "Say hi."
      Then the input capture acknowledgement is shown in blue
      And tellme exits successfully

    Example: A prompt typed at the interactive prompt is acknowledged in blue
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator submits the prompt "hi" at the interactive prompt
      Then the input capture acknowledgement is shown in blue
      And tellme exits successfully

  Rule: The acknowledgement keeps its line but loses its accent under raw output

    # Round 097 (ADR 0067) D5: the ack gates on the prompt-bearing surface, NOT on
    # `-r` — so `-r` must suppress the COLOUR without suppressing the LINE. A
    # "skip the ack when raw" refactor would delete a real line and red the
    # round-017 announced row.

    Example: Raw output still announces the capture, without the accent
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator starts tellme with the prompt "Say hi." and the raw flag
      Then the input capture is announced for the turn
      And the session chrome carries no colour
      And tellme exits successfully
