Feature: Reporting the metrics of the request

  # Acceptance only: after answering, a prompt-bearing run reports the token
  # usage of the request that just completed — the missed, cached, completed,
  # and reasoning tokens.

  Rule: A prompt-bearing run reports the metrics of the request that just completed

    Example: The operator runs a single-request turn
      Given the operator has a runnable tellme installation
      When the operator runs tellme with the prompt "hi"
      Then the run reports the token usage of the request that just completed
      And the report names the missed, cached, completed, and reasoning tokens

    Example: The provider reports no reasoning tokens
      Given the operator has a runnable tellme installation
      And the provider reports no reasoning tokens
      When the operator runs tellme with the prompt "hi"
      Then the report shows the reasoning tokens as zero

  Rule: The report reflects the last request of a turn that used a tool

    Example: The operator runs a turn that uses a tool before answering
      Given the operator has a runnable tellme installation
      And the model asks for a tool before answering
      When the operator runs tellme with the prompt "read the notes"
      Then the run reports the token usage of the answering request
