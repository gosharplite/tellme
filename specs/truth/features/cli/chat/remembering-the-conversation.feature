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
    # `kill -INT $PPID`, a REAL SIGINT to the turn process (no pty, no stall harness).

    Example: A turn interrupted after a tool step keeps it
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the tool-loop limit is "5"
      And a configured provider "test-model" whose endpoint runs the command "kill -INT $PPID; sleep 30" and then answers with "unused"
      When the operator starts tellme with the prompt "Explore the repository."
      Then tellme stored an interrupted turn in the session history with 1 tool step
      And the session history's last turn was closed with the operator-interruption answer
      And tellme reports on stderr that the turn was interrupted by the operator
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
