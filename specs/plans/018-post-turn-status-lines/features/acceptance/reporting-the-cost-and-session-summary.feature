Feature: Reporting the cost and the session summary

  # Acceptance only: after the metrics, a prompt-bearing run reports the cost
  # of the last request, of the turn, and of the session, plus the session's
  # token totals and how much of the prompt was served from cache.

  Rule: The run reports the cost of the last request, the turn, and the session

    Example: The operator runs a single-request turn on a fresh session
      Given the operator has a runnable tellme installation
      And the configuration prices the active model
      When the operator runs tellme with the prompt "hi"
      Then the run reports the cost of the request, the turn, and the session
      And the three costs are equal on a fresh session

    Example: The operator runs a turn that uses a tool
      Given the operator has a runnable tellme installation
      And the configuration prices the active model
      And the model asks for a tool before answering
      When the operator runs tellme with the prompt "read the notes"
      Then the turn's cost is greater than the last request's cost

  Rule: The session summary accumulates across a session and resets on a fresh one

    Example: The operator continues a session that already has turns
      Given the operator has a runnable tellme installation
      And the operator's session already has completed turns
      When the operator runs tellme with the prompt "carry on"
      Then the session summary includes the earlier turns

    Example: The operator starts a fresh session
      Given the operator has a runnable tellme installation
      And the operator is starting a fresh session
      When the operator runs tellme with the prompt "hi"
      Then the session summary starts empty

  Rule: The session summary reports the session's tokens and cache share

    Example: The operator runs a turn on a fresh session
      Given the operator has a runnable tellme installation
      When the operator runs tellme with the prompt "hi"
      Then the run reports the session's missed, cached, and output tokens
      And the run reports how much of the prompt was served from cache

  Rule: A model the configuration does not price is reported as free

    Example: The operator uses a model the configuration does not price
      Given the operator has a runnable tellme installation
      And the configuration does not price the active model
      When the operator runs tellme with the prompt "hi"
      Then the run reports a zero cost
      And the run still reports the session's tokens
