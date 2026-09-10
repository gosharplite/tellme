Feature: Checking the build version and diagnosing setup

  Rule: tellme must report its build version

    Example: The operator asks tellme which build is running
      Given any environment
      When the operator runs tellme with "--version"
      Then tellme prints the build version
      And tellme exits successfully

  Rule: tellme must report how it resolved its setup, offline, without contacting any provider

    Example: The operator diagnoses configuration and home resolution, plainly and machine-readably
      Given the runtime home "TELL_ME_HOME" is set to "ait-tmg"
      And a configuration "configs/butler.yaml" resolves to a ready state
      When the operator runs tellme's diagnostic "-d"
      Then tellme reports that the configuration resolved
      And tellme reports that the runtime home resolved to "ait-tmg"
      And tellme reports that the session workspace resolved to "ait-tmg/output/butler"
      And tellme performs no network access
      And tellme exits successfully

      When the operator runs the diagnostic once more with "--json"
      Then tellme emits the same resolution status as structured output
      And tellme performs no network access
      And tellme exits successfully

    Example: The operator diagnoses a setup that does not resolve, plainly and machine-readably
      Given the runtime home "TELL_ME_HOME" is set to "ait-tmg"
      And a configuration "configs/butler.yaml" does not resolve to a ready state
      When the operator runs tellme's diagnostic "-d"
      Then tellme reports that the configuration did not resolve
      And tellme reports the reason the configuration did not resolve
      And tellme performs no network access
      And tellme exits with a diagnostic error code distinct from the success code

      When the operator runs the diagnostic once more with "--json"
      Then tellme emits the same unresolved status as structured output
      And tellme performs no network access
      And tellme exits with a diagnostic error code distinct from the success code
