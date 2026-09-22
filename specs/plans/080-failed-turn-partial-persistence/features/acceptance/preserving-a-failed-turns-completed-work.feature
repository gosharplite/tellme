Feature: Preserving a failed turn's completed work

  # Acceptance journey (plan-side) — round 080, anchor issue #161. Round 079 made a
  # Ctrl+C keep the tool work already done. The SAME loss still happens on a genuine
  # turn FAILURE: the provider keeps failing (the bounded retry exhausts) or rejects
  # the request outright, and tellme reports the failure and throws away every completed
  # tool step. This journey keeps that work too — while reporting the failure exactly as
  # it does today (the same phrase, the same exit code) so nothing about the failure
  # surface changes.

  Rule: A failed turn keeps the tool work it had already done
    When a turn fails after some tool steps had already finished — because the provider
    kept failing or rejected the request outright — tellme keeps those steps and closes
    the turn with an answer that records the failure, so the work survives, the session
    stays readable, and the failure is still reported the same way.

    Example: A turn whose provider keeps failing keeps its completed steps
      Given a tool turn in which a tool step has already completed and the provider keeps failing
      When the operator sends the prompt
      Then the session history keeps the completed tool steps for that prompt
      And the failed turn is closed with an answer that records the failure
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

    Example: A turn whose request is rejected outright keeps its completed steps
      Given a tool turn in which a tool step has already completed and the provider rejects the request outright
      When the operator sends the prompt
      Then the session history keeps the completed tool steps for that prompt
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

  Rule: The session can be continued from the kept work
    When a failed turn's work was kept, the next prompt resumes the conversation as if
    the turn had ended normally — the kept turn is closed with a model answer.

    Example: The next prompt is accepted after a failed turn kept its work
      Given a tool turn that failed after some steps and its work was kept
      When the operator sends the next prompt
      Then tellme accepts the resumed conversation and answers it

  Rule: A failure before any tool step completes leaves nothing behind
    When the very first request fails, there is no completed work to keep — the clean
    abort the operator already knows is preserved.

    Example: A turn that fails before the first tool step writes no history entry
      Given a turn in which no tool step has completed
      When the operator sends the prompt
      Then the session history is unchanged
