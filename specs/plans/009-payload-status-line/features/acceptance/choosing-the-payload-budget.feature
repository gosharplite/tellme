Feature: Choosing the payload budget

  # Acceptance only: the budget the payload status measures against is the
  # operator's to choose.

  Rule: The payload budget defaults to one million tokens

    Example: A run without a configured budget measures against one million
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      And no payload budget is configured
      When the operator asks tellme "What is two plus two?"
      Then the payload status measures against a budget of one million tokens

  Rule: The payload budget can be overridden

    Example: The operator overrides the budget through the environment
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator asks tellme "What is two plus two?" with the payload budget set to "2000000"
      Then the payload status measures against a budget of two million tokens

  Rule: An invalid payload budget is rejected

    Example: A negative budget is reported as a configuration error
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator asks tellme "What is two plus two?" with the payload budget set to "-1"
      Then tellme reports a configuration error
      And tellme exits without contacting the provider
