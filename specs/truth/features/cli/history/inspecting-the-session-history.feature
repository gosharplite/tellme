Feature: Inspecting the session history

  # Interface truth (CLI end, `history` module) — `-l N` lists the most recent messages without a
  # provider request. Acceptance journeys: features/acceptance/inspecting-the-session-history.feature.

  Rule: The operator can list the most recent messages

    Example: Listing the last two messages
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
        | What is my name?  | Alice  |
      When the operator asks tellme to list the last 2 messages
      Then tellme lists the last 2 messages
      And tellme sends no request to any provider
      And tellme exits successfully

  Rule: Listing a session with no persisted history lists nothing

    Example: Listing an empty session
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      When the operator asks tellme to list the last 5 messages
      Then tellme lists no messages
      And tellme exits successfully
