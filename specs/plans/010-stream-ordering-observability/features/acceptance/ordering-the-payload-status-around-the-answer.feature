Feature: Ordering the payload status around the answer

  # Acceptance only: the per-turn payload status must appear in the order that
  # matches what it describes — the estimate before the answer, the measurement
  # after it. This is the ordering facet round 009 shipped wrong and that no gate
  # could observe.

  Rule: The estimated payload size is reported before the answer

    Example: The operator sees the estimate ahead of the answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reports its usage
      When the operator asks tellme "What is two plus two?"
      Then the run reports, before the answer, an estimated payload size for the turn
      And the final answer is printed on standard output

    Example: The estimate is still reported ahead of the answer when the provider reports no usage
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reports no usage
      When the operator asks tellme "What is two plus two?"
      Then the run reports, before the answer, an estimated payload size for the turn
      And the run reports no measured payload size
      And the final answer is printed on standard output

  Rule: The measured payload size is reported after the answer

    Example: The operator sees the measurement trailing the answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that reports its usage
      When the operator asks tellme "What is two plus two?"
      Then the final answer is printed on standard output
      And the run reports, after the answer, the payload size the provider actually measured
