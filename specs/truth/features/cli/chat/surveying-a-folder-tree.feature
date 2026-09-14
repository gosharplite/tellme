Feature: Surveying a folder tree

  # Interface truth (CLI end, `chat` module) — the `get_tree` tool: a connector tree of a directory,
  # stopping at `max_depth` (default 2) and never descending into `.git`. Acceptance journey:
  # features/acceptance/surveying-a-folder-tree.feature.

  Rule: A folder tree is shown at a bounded depth

    Example: A nested folder structure is drawn as a tree, stopping at the default depth
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a sub-folder "src"
      And the working directory contains a file "src/main.go" whose text is "package main"
      And the working directory contains a sub-folder "src/pkg"
      And the working directory contains a file "src/pkg/deep/leaf.go" whose text is "package deep"
      And a configured provider "test-model" whose endpoint shows the folder tree and then answers with "done"
      When the operator starts tellme with the prompt "Show me the project tree."
      Then tellme showed the folder tree using its get_tree tool
      And the tree shows "src"
      And the tree shows "main.go"
      And the tree shows "pkg"
      And the tree does not show "leaf.go"
      And tellme exits successfully

  Rule: The tree never descends into the repository metadata directory

    Example: A ".git" directory is listed but not recursed into
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a sub-folder ".git"
      And the working directory contains a file ".git/config" whose text is "gitconfig"
      And a configured provider "test-model" whose endpoint shows the folder tree and then answers with "done"
      When the operator starts tellme with the prompt "Show me the project tree."
      Then tellme showed the folder tree using its get_tree tool
      And the tree shows ".git"
      And the tree does not descend into ".git"
      And tellme exits successfully
