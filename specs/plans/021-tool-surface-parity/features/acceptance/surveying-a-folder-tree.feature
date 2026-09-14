Feature: Surveying a folder tree

  # Acceptance only: what the operator experiences when tellme shows a folder
  # tree at a bounded depth.

  Rule: A folder tree is shown at a bounded depth

    Example: Seeing a nested folder structure
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that shows the folder tree it is asked about before answering
      And the working directory contains a sub-folder "src" that holds a file "main.go" and a sub-folder "pkg"
      When the operator asks tellme "Show me the project tree."
      Then tellme shows a folder tree using its get_tree tool
      And the tree shows the folders and files under the directory
      And the tree does not descend past the depth tellme was asked for
      And tellme exits successfully

    Example: The tree keeps to the default depth
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that shows the folder tree it is asked about before answering
      And the working directory contains a deeply nested folder structure
      When the operator asks tellme "Show me the tree."
      Then tellme shows a folder tree using its get_tree tool
      And the tree stops at tellme's default depth
      And tellme exits successfully
