Feature: Retrieving the session named by a configuration

  # Acceptance only (round 053, closes #103). An operator or an orchestrating agent
  # who asks tellme for the last messages of a session and names WHICH configuration
  # that session belongs to must get that configuration's session — the same session a
  # run under that configuration writes. Business language only: the request is
  # "review the session belonging to this configuration", never an environment setting.

  Rule: The session reviewed is the one the named configuration belongs to

    Example: Each configuration reviews its own session
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration named "first" whose session was seeded with an exchange
      And the runtime home holds a configuration named "second" whose session was seeded with a different exchange
      And no session mode is forced by the environment
      When the operator asks tellme for the last messages of the session belonging to the configuration "first"
      Then tellme reports the last exchange of the "first" session
      And the report is not the exchange of any other configuration's session
      And tellme exits successfully

    Example: Reviewing a second configuration reads a different session
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration named "first" whose session was seeded with an exchange
      And the runtime home holds a configuration named "second" whose session was seeded with a different exchange
      And no session mode is forced by the environment
      When the operator asks tellme for the last messages of the session belonging to the configuration "second"
      Then tellme reports the last exchange of the "second" session
      And tellme exits successfully

    Example: A forced session mode outranks the named configuration
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration named "first" whose session was seeded with an exchange
      And the runtime home holds a session whose mode is forced by the environment and seeded with a different exchange
      When the operator asks tellme for the last messages of the session belonging to the configuration "first"
      Then tellme reports the exchange of the environment-forced session
      And tellme exits successfully

  Rule: Starting fresh applies to the session the named configuration belongs to

    Example: Starting a fresh session archives the named configuration's session
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration named "first" whose session was seeded with an exchange
      And no session mode is forced by the environment
      When the operator starts a fresh session under the configuration "first"
      Then the "first" session is archived and begins empty
      And the session of any other configuration is left untouched
      And tellme exits successfully

  Rule: A named configuration that cannot be honoured is refused

    Example: Naming a configuration that does not exist is refused
      Given the operator has a runnable tellme installation
      And the runtime home holds no configuration named "missing"
      When the operator asks tellme for the last messages of the session belonging to the configuration "missing"
      Then tellme refuses the request with a configuration error
      And tellme does not fall back to any default session
      And tellme exits with a failure status

  # Reviewing a session must stay a purely local act: no request is made to any
  # provider, and the command works even when the run would otherwise be unreasonable.
