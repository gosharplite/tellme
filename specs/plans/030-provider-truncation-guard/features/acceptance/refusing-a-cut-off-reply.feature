Feature: Refusing a reply cut off at the output limit

  # Acceptance only: when the provider cuts the model's reply off at its output
  # limit before the reply is finished, tellme refuses the run loudly — it never
  # acts on a half-finished instruction and never presents a half-finished answer
  # as if it were complete. A cut-off file change must never be written, and a
  # cut-off answer must never be shown as the answer.

  Rule: A reply cut off while asking tellme to change a file is refused, and the file is untouched

    Example: A cut-off file creation writes nothing
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that is cut off at the output limit while creating a file before answering
      And the working directory contains no file "notes.txt"
      When the operator asks tellme "Create notes.txt containing exactly the text 'hello'."
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider request failed
      And tellme creates no file "notes.txt"
      And tellme exits with the provider error code

    Example: A cut-off file edit leaves the file as it was
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that is cut off at the output limit while editing a file before answering
      And the working directory contains a file "config.txt" whose lines are:
        | alpha |
        | BETA  |
        | gamma |
      When the operator asks tellme "In config.txt, replace the line 'BETA' with 'beta'."
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider request failed
      And the lines of "config.txt" are still:
        | alpha |
        | BETA  |
        | gamma |
      And tellme exits with the provider error code

  Rule: A reply cut off before the answer is finished is refused, not shown as complete

    Example: A cut-off answer is not presented as the answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that is cut off at the output limit before finishing its answer
      When the operator asks tellme "Summarise the release notes in detail."
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider request failed
      And tellme prints no answer
      And tellme exits with the provider error code
