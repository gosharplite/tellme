Feature: Answering a single prompt

  # Acceptance only: what the operator observes at the command line when tellme performs one reasoning turn.
  # This round proves a single prompt produces exactly one provider answer; the request shape is the contract for later slices.
  # This round implements the OpenAI-compatible provider family (Clarify Round 1, Q2); the acceptance journey below stays operator-facing.

  Rule: A prompt argument must be sent to the selected provider and the answer printed

    Example: The operator asks one question and receives the provider's answer
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers with "Use go list ./... to list packages."
      When the operator starts tellme with the prompt "How do I list Go packages?"
      Then tellme sends exactly one request to the provider "test-model"
      And tellme prints the provider's answer "Use go list ./... to list packages."
      And tellme exits successfully

    Example: The operator switches the selected provider through the environment
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with two reachable providers "alpha" and "beta"
      And the operator selects the provider "beta" through the environment
      When the operator starts tellme with the prompt "Hello"
      Then tellme sends exactly one request to the provider "beta"
      And tellme prints the provider's answer
      And tellme exits successfully

  Rule: A run without a prompt must keep the existing offline behavior

    Example: The operator starts tellme without a prompt and no provider is contacted
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model"
      When the operator starts tellme without a prompt
      Then tellme reports that the configuration is ready
      And tellme performs no network access
      And tellme exits successfully

    Example: An explicit diagnostic flag takes precedence over a prompt
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model"
      When the operator runs tellme's diagnostic "-d" with the prompt "Hello"
      Then tellme reports that the configuration resolved
      And tellme performs no network access
      And tellme exits successfully
