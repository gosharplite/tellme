Feature: Listing a directory

  # Interface truth (CLI end, `chat` module) — the `list_files` tool: `Contents of <path>:` plus one
  # `[d]`/`[f]` line per entry, with `path` defaulting to the current directory. Acceptance journey:
  # features/acceptance/listing-a-working-directory.feature.

  Rule: A directory is listed in the reference shape

    Example: A folder that holds a file and a sub-folder
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "hi"
      And the working directory contains a sub-folder "src"
      And a configured provider "test-model" whose endpoint lists the current directory and then answers with "done"
      When the operator starts tellme with the prompt "What is in the current directory?"
      Then tellme listed the directory using its list_files tool
      And the listing shows "notes.txt" as a file
      And the listing shows "src" as a folder
      And tellme exits successfully
