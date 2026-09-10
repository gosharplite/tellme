Feature: Rejecting unsupported command-line usage

  Rule: An unrecognized flag must be reported as a usage error, distinct from configuration and environment failures

    Example: The operator mistypes a flag and gets a usage error instead of a configuration failure
      Given the operator has a runnable tellme installation
      And the runtime home holds a well-formed configuration "configs/butler.yaml"
      When the operator starts tellme with the unrecognized flag "--wibble" and points it at "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr how tellme is meant to be invoked
      And tellme exits with a usage error code distinct from the success, configuration, and environment error codes
