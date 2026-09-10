Feature: Starting tellme with a configuration

  # Acceptance only: business-level behavior the operator sees at the command
  # line. The flags, environment variables, and configuration keys named here
  # are the operator-facing interface, not implementation detail.

  Rule: A run must not proceed unless its configuration is present and well-formed

    Example: An operator starts tellme with a usable file, then with a broken one
      Given the operator has a runnable tellme installation
      And the runtime home holds a well-formed configuration "configs/butler.yaml"
      When the operator starts tellme pointing at "configs/butler.yaml" with "-c"
      Then tellme reports that the configuration is ready
      And tellme exits successfully

      When the operator points tellme at "configs/missing.yaml", which does not exist
      Then tellme refuses to proceed
      And tellme explains on stderr that the configuration could not be found
      And tellme exits with a configuration error code distinct from the success code

      When the configuration "configs/butler.yaml" contains malformed YAML
      And the operator starts tellme pointing at "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that the configuration could not be parsed
      And tellme exits with a configuration error code distinct from the success code

    Example: The operator omits "-c" and no default configuration can be found
      Given the operator has a runnable tellme installation
      And the runtime home holds no configuration at the default location
      When the operator starts tellme without any "-c" flag
      Then tellme refuses to proceed
      And tellme explains on stderr that no configuration could be found
      And tellme exits with a configuration error code distinct from the success code

    Example: The operator omits "-c" and the default configuration for the effective mode is found
      Given the operator has a runnable tellme installation
      And the runtime home holds a well-formed configuration "configs/coder.yaml"
      And the effective mode is "coder"
      When the operator starts tellme without any "-c" flag
      Then tellme reports that the configuration is ready
      And tellme exits successfully

  Rule: A run must not proceed unless its effective selected provider exists in the provider registry

    Example: The selected provider is taken from the environment first, then from the file
      Given the runtime home holds a configuration "configs/butler.yaml" whose selected provider "deepseek-flash" is in its provider registry
      When the operator starts tellme pointing at "configs/butler.yaml"
      Then tellme reports that the configuration is ready
      And tellme exits successfully

      When the operator sets "TELL_ME_SELECTED_PROVIDER" to "ghost", which is not in the registry
      And the operator starts tellme pointing at "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that the selected provider is not in the registry
      And tellme exits with a configuration error code distinct from the success code

    Example: A configuration whose provider registry is empty cannot start
      Given the runtime home holds a configuration "configs/butler.yaml" whose provider registry is empty
      When the operator starts tellme pointing at "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that the selected provider is not in the registry
      And tellme exits with a configuration error code distinct from the success code
