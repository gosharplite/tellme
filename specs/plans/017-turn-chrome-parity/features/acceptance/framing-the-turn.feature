Feature: The turn is framed like tell-me-go

  # Acceptance only: each prompt-bearing turn opens with the reference's frame —
  # a horizontal rule, a "Turn N" heading for the active persona, and the
  # estimated payload for the request — separated from the answer by a blank line.
  # (The post-turn lines are out of scope this round.)

  Rule: The turn opens with a rule, a turn heading, and the payload estimate

    Example: The operator runs the first turn of a fresh session
      Given the operator has a runnable tellme installation
      And the operator is starting a fresh session
      When the operator runs tellme with the prompt "hi"
      Then the turn opens with a horizontal rule
      And the turn is headed "Turn 1" for the active persona
      And the turn shows the estimated payload for the request

  Rule: The turn heading counts the session's turns

    Example: The operator continues a session that already has two turns
      Given the operator has a runnable tellme installation
      And the operator's session already has two completed turns
      When the operator runs tellme with the prompt "carry on"
      Then the turn is headed "Turn 3" for the active persona

  Rule: The frame is separated from the answer

    Example: The operator runs a turn on a fresh session
      Given the operator has a runnable tellme installation
      And the operator is starting a fresh session
      When the operator runs tellme with the prompt "hi"
      Then the frame is followed by a blank line
      And the answer follows the blank line
