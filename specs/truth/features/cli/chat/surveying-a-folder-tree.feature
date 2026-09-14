Feature: Surveying a folder tree

  # Interface truth (CLI end, `chat` module) — the `get_tree` tool: a connector tree of a directory,
  # stopping at `max_depth` (default 2). Acceptance journey:
  # features/acceptance/surveying-a-folder-tree.feature.

  Rule: A folder tree is shown

    Example: A nested folder structure is drawn as a tree
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a sub-folder "src"
      And the working directory contains a file "src/main.go" whose text is "package main"
      And a configured provider "test-model" whose endpoint shows the folder tree and then answers with "done"
      When the operator starts tellme with the prompt "Show me the project tree."
      Then tellme showed the folder tree using its get_tree tool
      And the tree shows "src"
      And the tree shows "main.go"
      And tellme exits successfully
