Feature: Answering from several files at once

  # Acceptance only: what the operator experiences when tellme can read more than
  # one file in a single request. No writes and no process execution.

  Rule: A question that spans files is answered in one read

    Example: A question needing two files
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads every file it is asked for before answering
      And the working directory contains a file "left.txt" whose text is "the code is ORANGE"
      And the working directory contains a file "right.txt" whose text is "the code is not BLUE"
      When the operator asks tellme "Compare left.txt and right.txt and tell me the code."
      Then tellme reads both files in a single read_files request
      And the final answer reflects both files
      And tellme exits successfully

  Rule: A file that is too large to show in full is trimmed

    Example: A very large file is truncated
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads the file it is asked for before answering
      And the working directory contains a file "big.txt" that is larger than tellme will show in one read
      When the operator asks tellme "Read big.txt and tell me how it starts."
      Then the part of "big.txt" that tellme reads ends with a note that it was truncated
      And tellme exits successfully

  Rule: A file that cannot be shown is reported, not fatal

    Example: A file that is not text
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads the file it is asked for before answering
      And the working directory contains a binary file "logo.png"
      When the operator asks tellme "Read logo.png and describe it."
      Then tellme reports that "logo.png" is a binary file that cannot be shown as text
      And tellme exits successfully

    Example: Too many files in one request
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads every file it is asked for before answering
      And the working directory contains more files than tellme will read in a single request
      When the operator asks tellme "Read all of the files in this directory."
      Then tellme reports that too many files were requested
      And tellme exits successfully
