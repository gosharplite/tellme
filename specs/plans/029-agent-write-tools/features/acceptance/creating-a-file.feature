Feature: Creating a file with exact content

  # Acceptance only: tellme can produce a new file whose content is exactly what
  # the task asked for — creating the folders it needs — and it never destroys a
  # file that already exists.

  Rule: A task can create a new file with exactly the requested content

    Example: A new file is created with exactly what was asked
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that creates a file before answering
      And the working directory contains no file "notes.txt"
      When the operator asks tellme "Create notes.txt containing exactly the text 'hello'."
      Then tellme creates "notes.txt"
      And the content of "notes.txt" is exactly "hello"
      And tellme exits successfully

    Example: A new file inside a folder that does not exist yet
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that creates a file in a new folder before answering
      And the working directory contains no folder "drafts"
      When the operator asks tellme "Create drafts/todo.txt containing exactly the text 'buy milk'."
      Then tellme creates the folder "drafts"
      And tellme creates the file "drafts/todo.txt"
      And the content of "drafts/todo.txt" is exactly "buy milk"
      And tellme exits successfully

  Rule: A file that already exists is never overwritten

    Example: Creating over an existing file is refused and the file is untouched
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that tries to create a file before answering
      And the working directory contains a file "notes.txt" whose content is exactly "original"
      When the operator asks tellme "Create notes.txt containing exactly the text 'replacement'."
      Then the creation is refused
      And the content of "notes.txt" is still exactly "original"
      And tellme exits successfully

    Example: A failed or interrupted creation never leaves a half-written file
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that creates a large file before answering
      And the working directory contains no file "big.txt"
      When the operator asks tellme "Create big.txt with a large amount of content."
      Then "big.txt", if it exists, holds the complete content and is never a partial file
      And tellme exits successfully
