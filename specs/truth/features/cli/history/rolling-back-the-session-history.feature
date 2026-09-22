Feature: Rolling back the session history

  # Interface truth (CLI end, `history` module) — round 081 (ADR 0053): `tellme -b`/`--back [N]`
  # rolls back the last N complete turns (default 1) of the session history OFFLINE (no provider
  # request); `tellme -b [N] "prompt"` rolls back THEN runs the prompt against the trimmed history.
  # A turn is ONE persisted exchange, so a rollback removes the prompt, the answer, and every tool
  # step of each removed turn atomically. The archive is never touched (rollback ≠ `--new`).

  Rule: The operator rolls back the most recent turn offline

    Example: Rolling back the last turn
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt   | answer   |
        | first Q  | first A  |
        | second Q | second A |
      When the operator rolls back the last turn
      Then the active history holds exactly 1 exchange
      And tellme reports that 1 turn was rolled back
      And tellme sends no request to any provider
      And tellme exits successfully

  Rule: The operator rolls back several turns with an explicit count

    Example: Rolling back the last two turns
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt | answer |
        | one    | a1     |
        | two    | a2     |
        | three  | a3     |
      When the operator rolls back the last 2 turns
      Then the active history holds exactly 1 exchange
      And tellme reports that 2 turns were rolled back
      And tellme exits successfully

  Rule: A rolled-back turn takes its tool activity with it

    Example: A tool-using exchange survives while a later plain exchange is undone
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds a tool-using exchange
      And the session history then holds a plain exchange
      When the operator rolls back the last turn
      Then the active history holds exactly 1 exchange
      And the remaining exchange carries its recorded tool step
      And tellme exits successfully

  Rule: Rolling back more turns than exist clears the session without failing

    Example: Rolling back more turns than the session holds
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt | answer |
        | only Q | only A |
      When the operator rolls back the last 5 turns
      Then the active session history holds no exchanges
      And tellme exits successfully

  Rule: Undoing a turn leaves the archive untouched

    Example: A rollback does not archive the removed turn
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt | answer |
        | a Q    | an A   |
      When the operator rolls back the last turn
      Then the session archive holds no exchanges
      And the active session history holds no exchanges
      And tellme exits successfully

  Rule: The operator undoes the last turn and immediately asks again

    Example: Undo the last turn and re-prompt
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "fake" whose endpoint answers with "third A"
      And the session history already holds the exchanges:
        | prompt   | answer   |
        | first Q  | first A  |
        | second Q | second A |
      When the operator rolls back the last turn and asks "third Q"
      Then tellme reports that 1 turn was rolled back
      And the active history holds exactly 2 exchanges
      And the last persisted exchange asks "third Q" and answers "third A"
      And the provider was asked against the trimmed history
      And tellme exits successfully

  Rule: An invalid rollback count is refused

    Example: Rolling back zero turns is refused
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt | answer |
        | a Q    | an A   |
      When the operator rolls back a count of 0 turns
      Then tellme explains on stderr that "the command-line usage is invalid"
      And the active history holds exactly 1 exchange

  Rule: Rolling back while starting a fresh session is refused

    Example: Rolling back and starting fresh at once is refused
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt | answer |
        | a Q    | an A   |
      When the operator rolls back the last turn and starts a fresh session
      Then tellme explains on stderr that "the command-line usage is invalid"
      And the active history holds exactly 1 exchange
