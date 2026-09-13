Feature: Reporting the payload status

  # Interface truth (CLI end, `chat` module) — a prompt run reports the payload's estimated size
  # before it is sent and the provider's measured size after it, measured against a configured
  # payload budget. Diagnostic only: the answer stream (stdout) is unchanged, and the line carries
  # no `tellme:` prefix. Acceptance journeys: features/acceptance/seeing-the-payload-status.feature
  # and features/acceptance/choosing-the-payload-budget.feature.

  Rule: A prompt run reports the estimated payload size before it is sent

    Example: The operator sees the estimated payload and the budget
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
      When the operator starts tellme with the prompt "Summarise where we are."
      Then tellme reports the estimated payload status for the turn
      And the payload status measures against a budget of 1000000 tokens
      And the payload status names the active mode and model
      And tellme exits successfully

  Rule: A prompt run reports the provider's measured payload size after the turn

    Example: The provider reports its usage
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good" and reports its usage
      When the operator starts tellme with the prompt "What is two plus two?"
      Then tellme reports the measured payload status for the turn
      And the payload status measures against a budget of 1000000 tokens
      And tellme exits successfully

    Example: The provider reports no usage
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good" and reports no usage
      When the operator starts tellme with the prompt "What is two plus two?"
      Then tellme reports the estimated payload status for the turn
      And tellme reports no measured payload status
      And tellme exits successfully

  Rule: The payload status never disturbs the answer

    Example: The answer stays unchanged while the status is reported
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "the launch code is ORANGE"
      When the operator starts tellme with the prompt "What is the launch code?" and the raw flag
      Then the captured standard output is exactly "the launch code is ORANGE"
      And tellme reports the estimated payload status for the turn

  Rule: The payload budget defaults to one million tokens

    Example: A run with no configured budget measures against one million
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      When the operator starts tellme with the prompt "What is two plus two?"
      Then the payload status measures against a budget of 1000000 tokens

  Rule: The payload budget can be overridden

    Example: The environment overrides the budget
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      And the payload budget is "2000000"
      When the operator starts tellme with the prompt "What is two plus two?"
      Then the payload status measures against a budget of 2000000 tokens

  Rule: An invalid payload budget stops the run

    Example: A negative payload budget is rejected
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good"
      And the payload budget is "-1"
      When the operator starts tellme with the prompt "What is two plus two?"
      Then tellme refuses to proceed
      And tellme explains on stderr that "the configuration is invalid"
