Feature: Reviewing a session's turn log

  # Acceptance only (round 053, closes #103). A session keeps a turn log — the plain-text
  # progress it showed while it ran. An operator or an orchestrating agent can review a
  # session's turn log, and can name which configuration's session to review. Business
  # language only: no terminal, stream, or storage detail.

  Rule: A session keeps a turn log of what it showed while running

    Example: A run leaves a turn log for its session
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      And no session mode is forced by the environment
      When the operator runs a turn under that configuration
      Then the session keeps a turn log of the progress the run showed
      And tellme exits successfully

  Rule: The reviewed turn log is the one the named configuration belongs to

    Example: Reviewing the turn log of a named configuration reads that session
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration named "first" whose session ran one or more turns
      And the runtime home holds a configuration named "second" whose session ran different turns
      And no session mode is forced by the environment
      When the operator reviews the turn log of the session belonging to the configuration "first"
      Then tellme prints the "first" session's turn log
      And the printing is not the "second" session's turn log
      And tellme exits successfully

  Rule: Reviewing a turn log that does not exist yet is not an error

    Example: A session with no turn log yet reviews as empty
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose session has run no turns yet
      And no session mode is forced by the environment
      When the operator reviews the turn log of the session belonging to that configuration
      Then tellme prints nothing
      And tellme exits successfully
