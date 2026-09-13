Feature: Reporting tool activity before the answer

  # Acceptance only: when tellme works with a tool before answering, the operator
  # must see that activity reported before the answer — in the order it happened,
  # not reporting work before it has visibly "happened".

  Rule: Tool activity is reported before the answer

    Example: The operator sees the tool activity ahead of the answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that asks tellme to read "notes.txt" and then answers
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      When the operator asks tellme "What is the launch code?"
      Then the run reports, before the answer, the tool activity for the turn
      And the final answer is printed on standard output
