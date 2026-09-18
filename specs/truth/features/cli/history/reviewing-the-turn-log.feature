Feature: Reviewing the session turn log

  # Interface truth (CLI end, `history` module) — `-t` prints the session's turn
  # log (the rendered turn chrome) and exits, strictly offline (round 053; closes
  # #103; ADR 0022). Acceptance journeys:
  # features/acceptance/reviewing-a-session-turns-log.feature.

  Rule: The turn log reviewed is the one the named configuration belongs to

    Example: Reviewing a session's turn log
      Given the operator has a runnable tellme installation
      And the runtime home holds the configuration "a.yaml" in mode "alpha" whose turn log holds "the run showed a frame"
      When the operator reviews the turn log of the configuration "a.yaml"
      Then tellme prints exactly the turn log line "the run showed a frame"
      And tellme exits successfully

    Example: A session with no turn log yet reviews as empty
      Given the operator has a runnable tellme installation
      And the runtime home holds the configuration "a.yaml" in mode "alpha" holding the answer "from alpha"
      When the operator reviews the turn log of the configuration "a.yaml"
      Then tellme prints nothing
      And tellme exits successfully
