Feature: Rejecting unsupported command-line usage

  Background:
    Given the operator has a runnable tellme installation

  Rule: An unrecognized flag stops the run with a usage error

    Example: A mistyped flag is a usage error, not a configuration error
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml" with the unrecognized flag "--wibble"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the command-line usage is invalid"
      And tellme exits with the usage error code
