Feature: Starting a fresh session

  # Interface truth (CLI end, `history` module) — `--new` begins a fresh session and retains the
  # previous one. Acceptance journeys: features/acceptance/starting-a-fresh-conversation.feature.

  Rule: A fresh session begins with no earlier conversation

    Example: Starting fresh empties the active session
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
      When the operator starts a fresh session with "--new"
      Then the active session history holds no exchanges
      And the run shows no turn chrome
      And the run reports no post-turn status
      And the run shows no progress spinner
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

  # Round 053 (closes #103; ADR 0022): a prompt-less `--new` selects the session
  # named by the `-c` configuration's MODE when the environment forces no mode.
  Rule: Starting fresh applies to the session the named configuration belongs to

    Example: Starting fresh for a named configuration archives its session
      Given the operator has a runnable tellme installation
      And the runtime home holds the configuration "a.yaml" in mode "alpha" holding the answer "from alpha"
      When the operator starts a fresh session of the configuration "a.yaml"
      Then the "alpha" session holds no active exchanges
      And the "alpha" session archived the exchange "hello" and "from alpha"
      And tellme exits successfully

