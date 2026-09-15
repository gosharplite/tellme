Feature: Surveying a folder tree

  # Acceptance only: what the operator experiences when tellme shows a folder
  # tree at a bounded depth.

  Rule: A folder tree is shown at a bounded depth

    Example: Seeing a nested folder structure, bounded at the default depth
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that shows the folder tree it is asked about before answering
      And the working directory contains a sub-folder "src" that holds a file "main.go" and a sub-folder "pkg", which in turn holds a deeply nested file "leaf.go"
      When the operator asks tellme "Show me the project tree."
      Then tellme shows a folder tree using its get_tree tool
      And the tree shows the folders and files under the directory
      And the tree does not show a file that lies deeper than tellme's default depth
      And tellme exits successfully

  Rule: A folder tree does not descend into the repository metadata directory

    Example: A ".git" directory is listed but not opened
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that shows the folder tree it is asked about before answering
      And the working directory contains a ".git" folder holding a file "config"
      When the operator asks tellme "Show me the tree."
      Then tellme shows a folder tree using its get_tree tool
      And the tree shows the ".git" folder
      And the tree does not show the ".git" folder's contents
      And tellme exits successfully
