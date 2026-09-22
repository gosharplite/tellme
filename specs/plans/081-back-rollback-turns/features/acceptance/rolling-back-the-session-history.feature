Feature: Rolling back the session history

  # Plan-side acceptance journey (round 081) — `tellme -b`/`--back [N]` rolls back the last N
  # complete turns of the session history. Acceptance journeys for the CLI end; `-l` (list last N)
  # and `--new` (archive everything, start fresh) already exist. `-b` is the "undo" verb.
  #
  # A turn is ONE persisted exchange (prompt + answer + any tool steps) — so rolling back a turn
  # removes all of it at once.

  Rule: The operator undoes the most recent turn without contacting any provider

    Example: Rolling back the last turn
      Given a session whose active history holds the exchanges:
        | prompt    | answer   |
        | first Q   | first A  |
        | second Q  | second A |
      When the operator rolls back the last turn
      Then the active history holds only the first exchange
      And tellme reports how many turns were rolled back
      And tellme exits successfully
      And tellme sends no request to any provider

  Rule: The operator undoes several turns with an explicit count

    Example: Rolling back the last two turns
      Given a session whose active history holds the exchanges:
        | prompt  | answer |
        | one     | a1     |
        | two     | a2     |
        | three   | a3     |
      When the operator rolls back the last 2 turns
      Then the active history holds only the first exchange
      And tellme reports that 2 turns were rolled back
      And tellme exits successfully

  Rule: Rolling back the last turn removes its tool activity with it

    Example: A tool-using turn is undone whole
      Given a session whose active history holds a tool-using exchange followed by a plain exchange
      When the operator rolls back the last turn
      Then the active history holds only the tool-using exchange
      And the tool-using exchange still carries its recorded tool step

  Rule: Undoing more turns than exist clears the session without failing

    Example: Rolling back more turns than the session holds
      Given a session whose active history holds the exchanges:
        | prompt | answer |
        | only Q | only A |
      When the operator rolls back the last 5 turns
      Then the active history holds no exchanges
      And tellme exits successfully

  Rule: Undoing a turn does not disturb the archive

    Example: The archive is left untouched by a rollback
      Given a session whose active history holds the exchanges:
        | prompt | answer |
        | a Q    | an A   |
      When the operator rolls back the last turn
      Then the session archive holds no exchanges
      And tellme exits successfully

  Rule: The operator undoes the last turn and immediately asks again

    Example: Undo the last turn and re-prompt
      Given a session whose active history holds the exchanges:
        | prompt    | answer   |
        | first Q   | first A  |
        | second Q  | second A |
      When the operator rolls back the last turn and asks "third Q"
      Then the active history holds the first exchange followed by the answer to "third Q"
      And tellme exits successfully

  Rule: An invalid rollback count is refused

    Example: Rolling back zero turns is a usage error
      Given a session whose active history holds the exchanges:
        | prompt | answer |
        | a Q    | an A   |
      When the operator rolls back a count of 0 turns
      Then tellme refuses the command-line usage
      And the active history still holds the exchange

  Rule: A rollback combined with starting a fresh session is refused

    Example: Rolling back and starting fresh at once is refused
      Given a session whose active history holds the exchanges:
        | prompt | answer |
        | a Q    | an A   |
      When the operator rolls back the last turn and starts a fresh session
      Then tellme refuses the command-line usage
      And the active history still holds the exchange
