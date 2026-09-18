Feature: Listing the last message by default

  # Acceptance only (round 054). The list flag takes a count; when the count is
  # omitted, listing the most recent message is the intended default — the same
  # shorthand the reference offers. Business language only.

  Rule: Omitting the count lists the most recent message

    Example: The operator lists without a count
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
        | What is my name?  | Alice  |
      When the operator asks tellme to list the last messages
      Then tellme lists the last 1 messages
      And tellme exits successfully

    Example: An explicit count is still honoured
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
        | What is my name?  | Alice  |
      When the operator asks tellme to list the last 2 messages
      Then tellme lists the last 2 messages
      And tellme exits successfully
