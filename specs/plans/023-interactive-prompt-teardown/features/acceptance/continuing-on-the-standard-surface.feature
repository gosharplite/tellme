Feature: Continuing a submitted interactive prompt on the standard surface

  # Acceptance only: after the editor releases the terminal, a submitted
  # interactive prompt reads like any other prompt — the operator's own text is
  # reported back, the turn is framed, the wait is shown, and the answer and
  # status follow.

  Rule: A submitted interactive prompt continues like any other prompt

    Example: An interactive submission is reported back and answered
      Given the operator has a runnable tellme installation
      And the operator is working at an interactive terminal
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator submits the prompt "hi" at the interactive prompt
      Then the operator sees the submitted prompt "hi" reported
      And the run acknowledges that the input was captured
      And the run frames the turn
      And the operator sees the answer
      And the run reports the post-turn status
      And tellme exits successfully

    Example: A prompt typed without the interactive prompt is not reported back
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that answers directly
      When the operator asks tellme "hi"
      Then the run frames the turn
      And the run does not report the prompt back
      And the operator sees the answer
      And tellme exits successfully

  Rule: The operator sees that tellme is working while it waits

    Example: An interactive run shows progress while waiting for a slow answer
      Given the operator has a runnable tellme installation
      And the operator is working at an interactive terminal
      And the runtime home holds a configuration with a reachable provider "test-model" that answers slowly
      When the operator submits the prompt "hi" at the interactive prompt
      Then the run shows that tellme is working while it waits
      And the operator sees the answer once it arrives
      And tellme exits successfully
