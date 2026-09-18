Feature: Reviewing the session turn log

  # Interface truth (CLI end, `history` module) — `-t` prints the session's turn
  # log, and the prompt path records it (round 053; closes #103; ADR 0022).
  # Acceptance journeys: features/acceptance/reviewing-a-session-turns-log.feature.

  Rule: A session keeps a turn log of what it showed while running

    Example: A prompt turn records the turn chrome
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      When the operator starts tellme with the prompt "Say hi."
      Then the session's turn log holds the turn progress
      And tellme exits successfully

  Rule: The turn log reviewed is the one the named configuration belongs to

    Example: Reviewing a session's turn log by its configuration
      Given the operator has a runnable tellme installation
      And the runtime home holds the configuration "a.yaml" in mode "alpha" whose turn log holds "the run showed a frame"
      When the operator reviews the turn log of the configuration "a.yaml"
      Then tellme prints exactly the turn log line "the run showed a frame"
      And tellme performs no network access
      And tellme exits successfully

  Rule: Reviewing a turn log that does not exist yet is not an error

    Example: A session with no turn log yet reviews as empty
      Given the operator has a runnable tellme installation
      And the runtime home holds the configuration "a.yaml" in mode "alpha" holding the answer "from alpha"
      When the operator reviews the turn log of the configuration "a.yaml"
      Then tellme prints nothing
      And tellme exits successfully
