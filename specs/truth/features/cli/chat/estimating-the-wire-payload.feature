Feature: Estimating the wire payload

  # Interface truth (CLI end, `chat` module) — the pre-flight estimate measures the
  # payload actually sent (the persona message + the tool declarations + the
  # conversation), is responsive to it, and is deterministic. Acceptance journey:
  # features/acceptance/estimating-the-payload-that-will-be-sent.feature.

  Rule: The pre-flight estimate measures more than the conversation messages alone

    Example: The estimate exceeds the conversation-message-only size
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok" and reports its usage
      And the runtime home holds a configuration whose persona is "be terse"
      When the operator starts tellme with the prompt "Hello"
      Then tellme reports the estimated payload status for the turn
      And the estimated payload exceeds the conversation messages alone
      And tellme exits successfully

  Rule: The estimate grows when the wired payload grows

    Example: A longer persona raises the estimate above a shorter persona's
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      And a previous run with the persona "be brief" and the prompt "Hi" reported an estimated payload status
      When the operator uses the persona "You are a terse assistant that always answers in exactly one short sentence and never adds extra words." and starts tellme with the prompt "Hi"
      Then the estimated payload is larger than the previous run's

  Rule: The estimate is stable for identical inputs

    Example: The same request yields the same estimate
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "ok"
      And a previous run with the persona "be terse" and the prompt "Hi" reported an estimated payload status
      When the operator uses the persona "be terse" and starts tellme with the prompt "Hi"
      Then the estimated payload matches the previous run's
