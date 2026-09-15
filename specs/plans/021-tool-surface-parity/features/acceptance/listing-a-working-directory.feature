Feature: Listing a working directory

  # Acceptance only: what the operator experiences when tellme lists a directory
  # in the familiar shape.

  Rule: A directory is listed in the familiar shape

    Example: Listing a folder that holds a file and a sub-folder
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that lists the directory it is asked about before answering
      And the working directory contains a file "notes.txt"
      And the working directory contains a sub-folder "src"
      When the operator asks tellme "What is in the current directory?"
      Then tellme lists the current directory using its list_files tool
      And the listing shows "notes.txt" as a file and "src" as a folder
      And tellme exits successfully

    Example: Listing the current directory by default
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that lists the directory it is asked about before answering
      And the working directory contains a file "notes.txt"
      When the operator asks tellme "List the files here."
      Then tellme lists the current directory using its list_files tool
      And tellme exits successfully
