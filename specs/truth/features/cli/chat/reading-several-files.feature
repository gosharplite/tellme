Feature: Reading several files in one request

  # Interface truth (CLI end, `chat` module) — the multi-file `read_files` tool: one request may read
  # several files, each framed by a header; oversize files are truncated, binary/directory/too-many are
  # reported inside the result. Acceptance journey:
  # features/acceptance/answering-from-several-files.feature.

  Rule: One request reads several files, each framed by a header

    Example: Two files are read in a single framed request
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "left.txt" whose text is "the code is ORANGE"
      And the working directory contains a file "right.txt" whose text is "the code is not BLUE"
      And a configured provider "test-model" whose endpoint asks tellme to read "left.txt" and "right.txt" in one request and then answers with "the code is ORANGE"
      When the operator starts tellme with the prompt "Compare left.txt and right.txt."
      Then the run made a single read_files request carrying "left.txt" and "right.txt"
      And the read result frames "left.txt"
      And the read result frames "right.txt"
      And tellme prints the provider's answer "the code is ORANGE"
      And tellme exits successfully

  Rule: A file larger than the read limit is truncated

    Example: A large file is trimmed with a truncation marker
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "big.txt" whose text is longer than the read limit
      And a configured provider "test-model" whose endpoint asks tellme to read "big.txt" and then answers with "done"
      When the operator starts tellme with the prompt "Read big.txt and tell me how it starts."
      Then the part of "big.txt" that tellme read ends with a truncation marker
      And tellme exits successfully

  Rule: A path that cannot be shown is reported, not fatal

    Example: A binary file is reported inside the result
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a binary file "logo.png"
      And a configured provider "test-model" whose endpoint asks tellme to read "logo.png" and then answers with "done"
      When the operator starts tellme with the prompt "Read logo.png and describe it."
      Then tellme reports that "logo.png" is a binary file that cannot be shown as text
      And tellme exits successfully

    Example: A directory is reported, not read
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a sub-folder "src"
      And a configured provider "test-model" whose endpoint asks tellme to read "src" and then answers with "done"
      When the operator starts tellme with the prompt "Read src."
      Then tellme reports that "src" is a directory
      And tellme exits successfully

    Example: Too many files in one request are reported
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains more files than tellme reads in one request
      And a configured provider "test-model" whose endpoint asks tellme to read more files than one request allows and then answers with "done"
      When the operator starts tellme with the prompt "Read all of the files in this directory."
      Then tellme reports that too many files were requested
      And tellme exits successfully
