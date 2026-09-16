Feature: Showing each tool call the way tell-me-go shows it
  # Acceptance journey (PM language) for round 034 — the operator's request:
  # "I want tellme to show tool calls similar to tell-me-go."

  Rule: The operator sees what each tool call did
    Example: A read call is shown with its reason and its arguments
      Given the operator has a runnable tellme installation
      And a configured provider whose reply asks tellme to read a file with a stated reason
      When the operator asks tellme a question that makes it use that tool
      Then the operator sees a step marker for the tool round
      And the operator sees the reason the agent gave for the call
      And the operator sees the tool's name together with the arguments it was called with
      And the operator sees the result of the call
      And every part of the call is shown as plain text
      And tellme exits successfully

    Example: The reason is not repeated inside the argument list
      Given the operator has a runnable tellme installation
      And a configured provider whose reply asks tellme to read a file with a stated reason
      When the operator asks tellme a question that makes it use that tool
      Then the arguments the operator sees do not include the reason
      And tellme exits successfully

    Example: Several calls in one round are each shown
      Given the operator has a runnable tellme installation
      And a configured provider whose reply asks tellme to read two files and then answers
      When the operator asks tellme to compare the two files
      Then the operator sees one step marker for the round
      And the operator sees an action and a result for each call
      And tellme exits successfully

  Rule: Long values are shortened predictably
    Example: A very long argument is shortened so it stays on one line
      Given the operator has a runnable tellme installation
      And a configured provider whose reply asks tellme to read a file with a very long reason
      When the operator asks tellme a question that makes it use that tool
      Then the displayed line stays on a single line
      And the shortened value ends with a single ellipsis character
      And tellme exits successfully
