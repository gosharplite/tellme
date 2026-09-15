Feature: Reading a large file without a hard cap

  # Acceptance only: how much of a file tellme reads is governed by the run's
  # budget rather than a fixed small limit — so a big file is actually read, and
  # a request for more files than fit reports what was not shown.

  Rule: A large file is read as far as the run's budget allows

    Example: A file larger than the old fixed limit is still read
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads the file it is asked for before answering
      And the working directory contains a file "big.txt" that is larger than tellme would previously show
      When the operator asks tellme "Read big.txt and tell me how it starts."
      Then tellme reads "big.txt" beyond its previous fixed limit
      And tellme exits successfully

  Rule: A request for more than fits reports what was not shown

    Example: Reading more content than one result can carry shows what fits and names the rest
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reads every file it is asked for before answering
      # The bound here is the result budget (not the ≤50-files-per-call cap): the content overflows a single read's result.
      And the working directory holds more content than a single read's result can carry
      When the operator asks tellme "Read all of the files here."
      Then tellme reads the files that fit
      And tellme reports that some files were not shown
      And tellme exits successfully
