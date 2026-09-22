Feature: Preserving an interrupted turn's completed work

  # Acceptance journey (plan-side) — round 079, anchor issue #159. A long tool turn
  # gathers context across many steps (searches, multi-file analysis, tests/builds).
  # Today a Ctrl+C during that turn throws ALL of it away — the prompt and every
  # completed step — because the turn never reached its final answer. tellme should
  # instead keep the work already done and close the turn cleanly, so the operator can
  # carry on from where it stopped rather than start over.

  Rule: An interrupted turn keeps the tool work it had already done
    When the operator interrupts a long tool turn with Ctrl+C, tellme keeps the tool
    steps that had already finished and closes the turn with an answer that records the
    interruption — so the completed work survives and the session stays readable.

    Example: A turn interrupted after some steps keeps them
      Given a tool turn in which some tool steps have already completed
      When the operator interrupts the turn with Ctrl+C
      Then the session history keeps the completed tool steps for that prompt
      And the interrupted turn is closed with an answer that records the interruption

    Example: The session can be continued from the kept work
      Given a tool turn that was interrupted after some steps and its work was kept
      When the operator sends the next prompt
      Then tellme accepts the resumed conversation and answers it

  Rule: An interruption before any tool step completes leaves nothing behind
    When the operator interrupts a turn before a single tool step has finished, tellme
    records nothing — the clean abort the operator already knows is preserved.

    Example: An interruption before the first tool step writes no history entry
      Given a turn in which no tool step has completed
      When the operator interrupts the turn with Ctrl+C
      Then the session history is unchanged
