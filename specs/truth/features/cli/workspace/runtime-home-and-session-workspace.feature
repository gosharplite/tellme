Feature: Resolving the runtime home and session workspace

  Background:
    Given the operator has a runnable tellme installation

  Rule: tellme creates the per-mode session workspace on first run

    Example: The first run prepares the workspace for the effective mode
      Given the runtime home is "ait-tmg"
      And the configuration "configs/butler.yaml" declares the mode "butler"
      And no session workspace exists under "ait-tmg"
      When the operator starts tellme
      Then tellme creates the session workspace "ait-tmg/output/butler"
      And tellme reports the session workspace "ait-tmg/output/butler"
      And tellme exits successfully

  Rule: tellme reuses the session workspace across runs

    Example: A later run reuses the workspace and preserves its contents
      Given the runtime home is "ait-tmg"
      And the configuration "configs/butler.yaml" declares the mode "butler"
      And the session workspace "ait-tmg/output/butler" already exists
      And the workspace "ait-tmg/output/butler" already holds a file "state.txt"
      When the operator starts tellme
      Then tellme reuses the session workspace "ait-tmg/output/butler"
      And the workspace "ait-tmg/output/butler" still holds the file "state.txt"
      And tellme exits successfully

  Rule: The effective mode is taken from the environment first, then from the file

    Example: The mode override selects a different workspace than the file declares
      Given the runtime home is "ait-tmg"
      And the configuration "configs/butler.yaml" declares the mode "butler"
      And the mode override is "architect"
      When the operator starts tellme pointing at the configuration "configs/butler.yaml"
      Then tellme creates the session workspace "ait-tmg/output/architect"
      And tellme exits successfully

  Rule: An unusable runtime home stops the run

    Example: The runtime home is not set
      Given the runtime home is not set
      When the operator starts tellme
      Then tellme refuses to proceed
      And tellme explains on stderr that "the runtime home is not usable"
      And tellme exits with the environment error code

  Rule: A workspace path that is not a directory stops the run

    Example: The workspace path is occupied by a regular file
      Given the runtime home is "ait-tmg"
      And the configuration "configs/butler.yaml" declares the mode "butler"
      And the workspace path "ait-tmg/output/butler" already exists as a regular file
      When the operator starts tellme
      Then tellme refuses to proceed
      And tellme explains on stderr that "the workspace path is not a directory"
      And tellme exits with the environment error code
