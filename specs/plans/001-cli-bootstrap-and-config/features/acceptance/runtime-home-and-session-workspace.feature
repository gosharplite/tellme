Feature: Resolving the runtime home and session workspace

  Rule: tellme must prepare a per-mode session workspace under the runtime home and reuse it across runs

    Example: The operator boots once to create the workspace and again to reuse it
      Given the runtime home "TELL_ME_HOME" is set to "ait-tmg"
      And the configuration's mode is "butler"
      And no session workspace exists yet under "ait-tmg"
      When the operator starts tellme for the first time
      Then tellme creates the session workspace "ait-tmg/output/butler"
      And tellme reports the resolved workspace path "ait-tmg/output/butler"
      And tellme exits successfully

      When the operator starts tellme again with nothing changed
      Then tellme reuses the session workspace "ait-tmg/output/butler"
      And anything already stored in that workspace is preserved
      And tellme exits successfully

    Example: The effective mode is taken from the environment first, then from the file
      Given the runtime home "TELL_ME_HOME" is set to "ait-tmg"
      And the configuration "configs/butler.yaml" declares the mode "butler"
      When the operator sets "TELL_ME_MODE" to "architect"
      And the operator starts tellme pointing at "configs/butler.yaml"
      Then the effective mode is "architect"
      And tellme creates or reuses the session workspace "ait-tmg/output/architect"
      And tellme exits successfully

  Rule: An unusable runtime home or a conflicting workspace path must fail actionably

    Example: The operator runs tellme with no usable home, then with a workspace path that is a regular file
      Given the runtime home "TELL_ME_HOME" is not set
      When the operator starts tellme
      Then tellme refuses to proceed
      And tellme explains on stderr that the runtime home is not usable
      And tellme exits with an environment error code distinct from the success code

      When the runtime home "TELL_ME_HOME" is set to "ait-tmg"
      And "ait-tmg/output/butler" already exists as a regular file
      And the operator starts tellme
      Then tellme refuses to proceed
      And tellme explains on stderr that the workspace path is not a directory
      And tellme exits with an environment error code distinct from the success code
