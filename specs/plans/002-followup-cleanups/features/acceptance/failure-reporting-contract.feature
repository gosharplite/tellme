Feature: Failures are reported with a stable, documented contract

  # Acceptance only: the operator-facing failure surface. Each class of failure is
  # expected to carry its own documented message and its own documented exit code.

  Rule: Every documented failure must report its documented message and exit code

    Example: The operator hits a configuration failure, then a usage failure
      Given the operator has a runnable tellme installation
      And the runtime home holds a well-formed configuration "configs/butler.yaml"
      When the operator points tellme at "configs/missing.yaml", which does not exist
      Then tellme refuses to proceed
      And tellme explains on stderr that the configuration could not be found
      And tellme exits with a configuration error code distinct from the success code

      When the configuration "configs/butler.yaml" contains malformed YAML
      And the operator starts tellme pointing at "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that the configuration could not be parsed
      And tellme exits with a configuration error code distinct from the success code

      When the operator starts tellme with the unrecognized flag "--wibble"
      Then tellme refuses to proceed
      And tellme explains on stderr how tellme is meant to be invoked
      And tellme exits with a usage error code distinct from the success, configuration, and environment error codes

    Example: The operator hits an environment failure
      Given the operator has a runnable tellme installation
      And the runtime home "TELL_ME_HOME" is not set
      When the operator starts tellme
      Then tellme refuses to proceed
      And tellme explains on stderr that the runtime home is not usable
      And tellme exits with an environment error code distinct from the success and configuration error codes
