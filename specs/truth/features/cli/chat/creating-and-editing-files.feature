Feature: Creating and editing files

  # Interface truth (CLI end, `chat` module) — the two write tools: `write_file` creates a file with
  # exactly the given content, creating missing parent folders, and **refuses to overwrite** an
  # existing file (create-only; the write is atomic). `replace_text` replaces a block only when it is
  # **uniquely** present, and refuses (leaving the file untouched) on 0 or >1 matches. There is no
  # security/consent gate and no undo. Acceptance journeys:
  # features/acceptance/creating-a-file.feature and features/acceptance/editing-a-file.feature.
  #
  # Atomicity (both tools): every write goes to a temp file in the target folder and is moved into
  # place — `write_file` via an atomic create-only move (`os.Link`/`EEXIST`; an existing file is never
  # clobbered), `replace_text` via `rename` — so a destination is never partial. The created file is
  # mode 0644. This invariant is verified at the **unit** tier (no E2E fault injection exists); the
  # Examples below assert the observable outcomes only.

  Rule: A new file is created with exactly the requested content

    Example: A new file is created
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains no file "notes.txt"
      And a configured provider "test-model" whose endpoint creates the file "notes.txt" with the content "hello" and then answers with "done"
      When the operator starts tellme with the prompt "Create notes.txt containing 'hello'."
      Then tellme created the file "notes.txt"
      And the content of "notes.txt" is exactly "hello"
      And tellme exits successfully

    Example: A new file inside a folder that does not exist yet
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains no folder "drafts"
      And a configured provider "test-model" whose endpoint creates the file "drafts/todo.txt" with the content "buy milk" and then answers with "done"
      When the operator starts tellme with the prompt "Create drafts/todo.txt containing 'buy milk'."
      Then the folder "drafts" was created
      And tellme created the file "drafts/todo.txt"
      And the content of "drafts/todo.txt" is exactly "buy milk"
      And tellme exits successfully

  Rule: An existing file is never overwritten

    Example: Creating over an existing file is refused
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "original"
      And a configured provider "test-model" whose endpoint creates the file "notes.txt" with the content "replacement" and then answers with "done"
      When the operator starts tellme with the prompt "Create notes.txt containing 'replacement'."
      Then the creation is refused
      And the content of "notes.txt" is still exactly "original"
      And tellme exits successfully

  Rule: A uniquely identified block is replaced, and nothing else changes

    Example: The single matching block is replaced
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "config.txt" whose lines are:
        | alpha |
        | BETA  |
        | gamma |
      And a configured provider "test-model" whose endpoint edits the file "config.txt" replacing "BETA" with "beta" and then answers with "done"
      When the operator starts tellme with the prompt "In config.txt, replace BETA with beta."
      Then the lines of "config.txt" are:
        | alpha |
        | beta  |
        | gamma |
      And tellme exits successfully

  Rule: An edit that cannot be uniquely placed is refused, leaving the file untouched

    Example: A block that is not present is refused
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "config.txt" whose lines are:
        | alpha |
        | BETA  |
        | gamma |
      And a configured provider "test-model" whose endpoint edits the file "config.txt" replacing "DELTA" with "delta" and then answers with "done"
      When the operator starts tellme with the prompt "In config.txt, replace DELTA with delta."
      Then the edit is refused because the block is not present
      And the lines of "config.txt" are still:
        | alpha |
        | BETA  |
        | gamma |
      And tellme exits successfully

    Example: A block that appears more than once is refused
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "log.txt" that contains the line "note" twice
      And a configured provider "test-model" whose endpoint edits the file "log.txt" replacing "note" with "done" and then answers with "done"
      When the operator starts tellme with the prompt "In log.txt, replace note with done."
      Then the edit is refused because the block is not unique
      And the file "log.txt" still contains the line "note" twice
      And tellme exits successfully
