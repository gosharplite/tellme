Feature: Asking tellme for help

  # Interface truth (CLI end, `usage` module) — round 074 (ADR 0046): `-h`/`--help` prints tellme's flag
  # list to standard output and exits 0 (a successful, offline, prompt-less action). Acceptance
  # journeys: features/acceptance/asking-for-help-and-version.feature.

  Background:
    Given the operator has a runnable tellme installation

  Rule: The operator can ask for help with the short flag

    Example: The operator asks for help with "-h"
      And the diagnostics are shown at a terminal
      When the operator runs tellme with "-h"
      Then tellme prints its flag list
      And the help is reported as a success
      And the run shows no turn chrome
      And tellme performs no network access
      And tellme exits successfully

  Rule: The operator can ask for help with the long flag

    Example: The operator asks for help with "--help"
      And the diagnostics are shown at a terminal
      When the operator runs tellme with "--help"
      Then tellme prints its flag list
      And the help is reported as a success
      And the run shows no turn chrome
      And tellme performs no network access
      And tellme exits successfully
