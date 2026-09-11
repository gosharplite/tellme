Feature: The setup diagnostic has a single, plain output form

  # Acceptance only: what the operator sees at the command line. The flags named
  # here are the operator-facing interface, not implementation detail.

  Rule: A request for a machine-readable diagnostic must be refused as invalid usage

    Example: The operator asks for machine-readable output, with and without the diagnostic flag
      Given the operator has a runnable tellme installation
      And the runtime home holds a well-formed configuration "configs/butler.yaml"
      When the operator runs tellme's diagnostic with "--json"
      Then tellme refuses to proceed
      And tellme explains on stderr how tellme is meant to be invoked
      And tellme exits with a usage error code distinct from the success, configuration, and environment error codes

      When the operator runs tellme with "--json" and no diagnostic flag
      Then tellme refuses to proceed
      And tellme explains on stderr how tellme is meant to be invoked
      And tellme exits with a usage error code distinct from the success, configuration, and environment error codes

  Rule: The diagnostic must report a resolved and an unresolved setup in one plain form

    Example: The operator runs the plain diagnostic on a resolved setup, then on an unresolved one
      Given the operator has a runnable tellme installation
      And the runtime home holds a well-formed configuration "configs/butler.yaml"
      When the operator runs tellme's diagnostic without "--json"
      Then tellme reports that the configuration resolved
      And tellme performs no network access
      And tellme exits successfully

      When the runtime home holds no configuration at the default location
      And the operator runs tellme's diagnostic without "--json"
      Then tellme reports that the configuration did not resolve
      And tellme reports the reason the configuration did not resolve
      And tellme performs no network access
      And tellme exits with a diagnostic error code distinct from the success code
