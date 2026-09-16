Feature: Refusing a reply that stops before it is finished

  # Acceptance only: when the model's reply stops before it is finished, tellme
  # refuses the run loudly — it never acts on a half-finished instruction and never
  # presents a half-finished answer as if it were complete. A half-finished file
  # change must never be written, and a half-finished answer must never be shown.

  Rule: A reply that stops while asking tellme to change a file is refused, and the file is untouched

    Example: A half-finished file creation writes nothing
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" whose reply stops before it finishes asking tellme to create a file
      And the working directory contains no file "notes.txt"
      When the operator asks tellme "Create notes.txt containing exactly the text 'hello'."
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider request failed
      And tellme creates no file "notes.txt"
      And tellme exits with the provider error code

    Example: A half-finished file edit leaves the file as it was
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" whose reply stops before it finishes asking tellme to edit a file
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

  Rule: A reply that stops before the answer is finished is refused, not shown as complete

    Example: A half-finished answer is not presented as the answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" whose reply stops before it finishes its answer
      When the operator asks tellme "Summarise the release notes in detail."
      Then tellme refuses to proceed
      And tellme explains on stderr that the provider request failed
      And tellme prints no answer
      And tellme exits with the provider error code
