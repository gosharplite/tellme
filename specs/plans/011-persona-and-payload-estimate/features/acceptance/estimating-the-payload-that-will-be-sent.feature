Feature: Estimating the payload that will be sent

  # Acceptance only: the pre-flight estimate must measure what the request
  # actually carries — the persona instruction, the tools offered to the model,
  # and the conversation — so the figure shown before the request is comparable
  # to the figure the provider measures after it.

  Rule: The pre-flight estimate measures the whole payload that will be sent

    Example: The estimate reflects the persona, the offered tools, and the conversation
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose persona is "You are a terse assistant. Answer in one sentence."
      And the configuration selects a reachable provider "test-model" that reports its usage
      When the operator asks tellme "What is two plus two?"
      Then the run reports, before the answer, an estimated payload size for the turn
      And the estimated size accounts for the persona instruction, the tools offered to the model, and the conversation
      And the final answer is printed on standard output

  Rule: The estimate grows when the payload that will be sent grows

    Example: A longer persona instruction raises the reported estimate
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose persona is "Be brief."
      And the configuration selects a reachable provider "test-model" that answers directly
      When the operator asks tellme "Hi"
      Then the run reports an estimated payload size for the turn
      When the operator uses a much longer persona instruction and asks "Hi" again
      Then the second run reports a strictly larger estimated payload size

  Rule: The estimate is stable for identical inputs

    Example: The same request yields the same estimate every time
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration whose persona is "You are a terse assistant."
      And the configuration selects a reachable provider "test-model" that answers directly
      When the operator asks tellme "What is two plus two?" twice
      Then both runs report the same estimated payload size
