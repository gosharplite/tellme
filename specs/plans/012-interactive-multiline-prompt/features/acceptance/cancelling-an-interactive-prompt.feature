Feature: Cancelling an interactive prompt

  # Acceptance only: cancelling the reader or sending nothing makes no request and
  # prints no answer.

  Rule: Cancelling the reader makes no request

    Example: The operator cancels with Ctrl+C
      Given the operator has a runnable tellme installation
      And the operator is working at an interactive terminal
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator starts tellme with no prompt and cancels the input
      Then no request is made to the model
      And tellme exits without printing an answer

  Rule: Sending nothing makes no request

    Example: The operator sends the input immediately with no content
      Given the operator has a runnable tellme installation
      And the operator is working at an interactive terminal
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator starts tellme with no prompt and sends the input immediately
      Then no request is made to the model
      And tellme exits without printing an answer
