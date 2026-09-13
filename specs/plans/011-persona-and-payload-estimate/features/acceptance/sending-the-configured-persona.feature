Feature: Sending the configured persona to the model

  # Acceptance only: the operator's configured persona instruction must actually
  # reach the model. Today tellme parses PERSON but never sends it, so every run
  # behaves as a generic assistant regardless of the configuration.

  Rule: The configured persona is sent as the model's leading instruction

    Example: The operator's persona instruction reaches the model
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose persona is "You are a terse assistant. Answer in one sentence."
      And the configuration selects a reachable provider "test-model" that answers directly
      When the operator asks tellme "What is two plus two?"
      Then what tellme sends to the model begins with the configured persona instruction
      And the persona instruction precedes the operator's question
      And the final answer is printed on standard output

  Rule: A configuration without a persona sends no persona instruction

    Example: An empty persona adds no instruction to the model
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose persona is empty
      And the configuration selects a reachable provider "test-model" that answers directly
      When the operator asks tellme "What is two plus two?"
      Then what tellme sends to the model carries no persona instruction
      And the final answer is printed on standard output

  Rule: The persona never leaks into the answer stream

    Example: The persona stays out of the printed answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose persona is "You are a terse assistant."
      And the configuration selects a reachable provider "test-model" that answers directly
      When the operator asks tellme for a raw answer with "-r"
      Then the answer on standard output is the model's text unchanged
      And the persona instruction is not printed on standard output
