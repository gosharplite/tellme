Feature: Starting a fresh session

  # Interface truth (CLI end, `history` module) — `--new` begins a fresh session and retains the
  # previous one. Acceptance journeys: features/acceptance/starting-a-fresh-conversation.feature.

  Rule: A fresh session begins with no earlier conversation

    Example: Starting fresh empties the active session
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
      When the operator starts a fresh session with "--new"
      Then the active session history holds no exchanges
      And tellme exits successfully

  Rule: The previous conversation is retained when a fresh session starts

    Example: The earlier exchange is kept
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
      When the operator starts a fresh session with "--new"
      Then the archived session history holds the exchange "My name is Alice." and "Noted."
      And tellme exits successfully
