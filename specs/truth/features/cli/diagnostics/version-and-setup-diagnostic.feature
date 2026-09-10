Feature: Checking the build version and diagnosing setup

  Background:
    Given the operator has a runnable tellme installation

  Rule: tellme reports its build version

    Example: The operator asks which build is running
      When the operator runs tellme with "--version"
      Then tellme prints the build version
      And tellme exits successfully

  Rule: The diagnostic reports a resolved setup, offline

    Example: Resolution is reported plainly
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml"
      When the operator runs tellme's diagnostic
      Then tellme reports the configuration resolved
      And tellme reports the runtime home resolved to "ait-tmg"
      And tellme reports the session workspace resolved to "ait-tmg/output/butler"
      And tellme performs no network access
      And tellme exits successfully

    Example: Resolution is reported as structured output
      Given the runtime home is "ait-tmg"
      And a well-formed configuration "configs/butler.yaml"
      When the operator runs tellme's diagnostic with "--json"
      Then tellme emits the resolution status as structured output
      And tellme performs no network access
      And tellme exits successfully

  Rule: The diagnostic reports an unresolved setup, offline

    Example: An unresolved setup is reported plainly with a dedicated exit code
      Given the runtime home is "ait-tmg"
      And a configuration "configs/butler.yaml" that does not resolve
      When the operator runs tellme's diagnostic
      Then tellme reports the configuration did not resolve
      And tellme reports the reason the configuration did not resolve
      And tellme performs no network access
      And tellme exits with the diagnostic error code

    Example: An unresolved setup is reported as structured output with a dedicated exit code
      Given the runtime home is "ait-tmg"
      And a configuration "configs/butler.yaml" that does not resolve
      When the operator runs tellme's diagnostic with "--json"
      Then tellme emits the unresolved status as structured output
      And tellme performs no network access
      And tellme exits with the diagnostic error code
