Feature: Remembering the conversation across runs

  # Interface truth (CLI end, `chat` module) — the prompt turn persists each completed exchange and
  # carries the persisted conversation to the provider. The session store and its lifecycle live in the
  # `history` module. Acceptance journeys: features/acceptance/remembering-the-conversation.feature.

  Rule: A completed prompt turn is stored in the session history

    Example: The exchange is stored after the run
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "Paris"
      When the operator starts tellme with the prompt "What is the capital of France?"
      Then tellme stored the exchange "What is the capital of France?" and "Paris" in the session history
      And tellme exits successfully

  Rule: A prompt run carries the persisted conversation to the provider

    Example: A follow-up prompt carries the earlier exchange
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
      And a configured provider "test-model" whose endpoint answers with "Alice"
      When the operator starts tellme with the prompt "What is my name?"
      Then the request carried the earlier exchange "My name is Alice." and "Noted."
      And tellme exits successfully

  Rule: A prompt run with no persisted conversation carries no earlier exchange

    Example: The first prompt carries nothing
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "unused"
      When the operator starts tellme with the prompt "What is my name?"
      Then the request carried no earlier exchange
      And tellme exits successfully

  Rule: An operator-interrupted turn keeps the tool work it had already done (round 079)

    # Round 079 (ADR 0051; closes #159): when the operator interrupts a tool turn with
    # Ctrl+C, tellme keeps the completed tool steps and closes the partial turn with a
    # synthetic answer — so the stored history replays as a valid `user … assistant`
    # sequence on both provider families and the session can continue. A turn interrupted
    # BEFORE any tool step completed writes nothing (today's clean abort; the zero-step
    # negative is carried by unit pins — see the round-079 note in chat/dsl.md).
    # The live interruption is produced hermetically: a scripted `execute_command` runs
    # `kill -INT $PPID`, a REAL SIGINT to the turn process delivered from inside the tool call (no pty).

    Example: A turn interrupted after a tool step keeps it
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the tool-loop limit is "5"
      And a configured provider "test-model" whose endpoint runs the command "kill -INT $PPID; sleep 30" and then answers with "unused"
      When the operator starts tellme with the prompt "Explore the repository."
      Then tellme stored an interrupted turn in the session history with 1 tool step
      And the interrupted turn was recorded with 1 completed provider call
      And the session history's last turn was closed with the operator-interruption answer
      And tellme reports on stderr that the turn was interrupted by the operator
      And tellme does not explain on stderr that "the provider request failed"
      And the run wrote nothing to standard output
      And tellme sent the request exactly once
      And tellme exits successfully

  Rule: A session whose interrupted turn was kept continues from it (round 079)

    Example: The resumed request carries the kept tool step and the interruption answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds an interrupted exchange with 1 tool step
      And a configured provider "test-model" whose endpoint answers with "continued"
      When the operator starts tellme with the prompt "continue"
      Then the request replayed the earlier tool step "execute_command"
      And the request closes the earlier turn with the assistant answer "[Turn interrupted by operator via Ctrl+C]"
      And tellme exits successfully

  Rule: A failed turn keeps the tool work it had already done (round 080)

    # Round 080 (ADR 0052; closes #161): when a turn FAILS after some tool steps already
    # completed — the provider keeps failing (the bounded retry exhausts) or rejects the
    # request outright — tellme keeps those steps and closes the turn with a class-specific
    # synthetic answer, then reports the failure EXACTLY as before (the frozen phrase + exit 6).
    # A failure with zero completed steps writes nothing (today's clean abort).
    # The failure is produced hermetically by the in-process fake provider (an always-drop
    # transport; an outright rejection); the retry delay seam is collapsed to 0.

    Example: A turn whose provider keeps failing keeps its completed steps
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then always drops the connection
      When the operator starts tellme with the prompt "Read the notes, then keep going."
      Then tellme stored the failed turn in the session history with 1 tool step
      And the failed turn was closed with the turn-failure answer
      And tellme reports on stderr that it kept the completed tool step
      And the run wrote nothing to standard output
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

    Example: A turn whose request is rejected outright keeps its completed steps
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then rejects the request outright
      When the operator starts tellme with the prompt "Read the notes, then keep going."
      Then tellme stored the failed turn in the session history with 1 tool step
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code

  Rule: A session whose failed turn was kept continues from it (round 080)

    Example: The resumed request carries the kept step and the failure answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds a failed exchange with 1 tool step
      And a configured provider "test-model" whose endpoint answers with "continued"
      When the operator starts tellme with the prompt "continue"
      Then the request replayed the earlier tool step "read_files"
      And the request closes the earlier turn with the assistant answer "[Turn ended early: the provider request failed]"
      And tellme exits successfully

  Rule: A failure before any tool step completes leaves nothing behind (round 080)

    Example: A turn that fails before the first tool step writes no history entry
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "dead-model" whose endpoint is unreachable
      When the operator starts tellme with the prompt "Hello"
      Then the session history is empty
      And tellme explains on stderr that "the provider request failed"
      And tellme exits with the provider error code
