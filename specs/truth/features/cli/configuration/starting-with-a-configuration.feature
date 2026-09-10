Feature: Starting tellme with a configuration

  Background:
    Given the operator has a runnable tellme installation

  Rule: A run with a well-formed configuration reports readiness

    Example: A well-formed configuration at the pointed path is ready
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme reports the configuration is ready
      And tellme exits successfully

  Rule: A run pointing at a configuration that does not exist stops

    Example: The pointed configuration file is missing
      Given the runtime home is "ait-tmg"
      And no configuration exists at "configs/butler.yaml"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the configuration could not be found"
      And tellme exits with the configuration error code

  Rule: A run pointing at a malformed configuration stops

    Example: The pointed configuration contains malformed YAML
      Given the runtime home is "ait-tmg"
      And a malformed configuration "configs/butler.yaml"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the configuration could not be parsed"
      And tellme exits with the configuration error code

  Rule: With no configuration flag and no default found, the run stops

    Example: No default configuration exists for the effective mode
      Given the runtime home is "ait-tmg"
      And no configuration exists at the default location
      When the operator starts tellme
      Then tellme refuses to proceed
      And tellme explains on stderr that "no configuration could be found"
      And tellme exits with the configuration error code

  Rule: With no configuration flag, the default for the effective mode is used

    Example: The default configuration for the effective mode is found
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/coder.yaml"
      And the effective mode is "coder"
      When the operator starts tellme
      Then tellme reports the configuration is ready
      And tellme exits successfully

  Rule: A run proceeds when the effective selected provider is in the registry

    Example: The file's selected provider is in the provider registry
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml" whose selected provider "deepseek-flash" is in its registry
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme reports the configuration is ready
      And tellme exits successfully

  Rule: A run stops when an environment override names a provider not in the registry

    Example: The environment override points at a provider absent from the registry
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml" whose selected provider "deepseek-flash" is in its registry
      And the selected provider override is "ghost"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the selected provider is not in the registry"
      And tellme exits with the configuration error code

  Rule: A run stops when the provider registry is empty

    Example: The configuration declares no providers
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml" whose provider registry is empty
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the selected provider is not in the registry"
      And tellme exits with the configuration error code
