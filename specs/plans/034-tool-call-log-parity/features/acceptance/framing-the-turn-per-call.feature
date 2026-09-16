Feature: The turn is framed for each model call, and the closing status follows the answer
  # Acceptance journey (PM language) for round 034 — the per-call status frame cadence.

  Rule: Each model call gets its own frame
    Example: A tool-using turn shows a frame for every model call it makes
      Given the operator has a runnable tellme installation
      And a configured provider whose reply asks tellme to read a file and then answers
      When the operator asks tellme a question that makes it use a tool
      Then the operator sees more than one turn frame
      And each frame names the current turn number
      And the turn number advances from one frame to the next
      And tellme exits successfully

    Example: A plain question shows exactly one frame
      Given the operator has a runnable tellme installation
      And a configured provider whose reply answers directly
      When the operator asks tellme a plain question
      Then the operator sees exactly one turn frame
      And tellme exits successfully

  Rule: The closing status comes after the answer
    Example: The measured usage and the ready summary follow the answer
      Given the operator has a runnable tellme installation
      And a configured provider whose reply asks tellme to read a file and then answers with a final answer
      When the operator asks tellme a question that makes it use a tool
      Then the operator sees the final answer
      And the measured usage summary appears after the answer
      And the ready summary appears after the answer
      And tellme exits successfully

  Rule: The estimated payload is shown for every model call
    Example: Each frame reports the payload it is about to send
      Given the operator has a runnable tellme installation
      And a configured provider whose reply asks tellme to read a file and then answers
      When the operator asks tellme a question that makes it use a tool
      Then each turn frame reports an estimated payload
      And the later estimates are not smaller than the earlier ones
      And tellme exits successfully
