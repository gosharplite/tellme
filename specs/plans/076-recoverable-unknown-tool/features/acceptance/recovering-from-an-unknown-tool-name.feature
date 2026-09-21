Feature: Recovering from an unknown tool name

  # Acceptance journey (plan-side) — round 076, anchor issue #154. An off-list tool
  # name is a recoverable model slip: tellme folds back a "no such tool" result and
  # keeps the turn going, instead of aborting the whole run. The recoverable path is
  # bounded per turn, so a provider that fixates is stopped and reported.

  Rule: A one-off unknown tool name does not end the turn
    The model sometimes asks for a tool tellme does not offer (a typo, an alias, an
    upstream tool named by its bare name). tellme tells the model the name is unknown
    and carries on, so a single slip can be corrected within the same turn.

    Example: An unavailable name is recovered and the turn finishes
      Given the model first asks for a tool tellme does not offer, then asks to read a note, then answers
      When the operator sends the prompt
      Then the model is told the requested tool is unavailable, naming it
      And the turn finishes with the note's answer
      And no tool ran for the unavailable name

  Rule: A model that will not stop asking for an unknown tool is stopped and reported
    tellme bounds the recoverable retries per turn, so a model that keeps asking for a
    tool tellme cannot offer is stopped (rather than spending the whole tool-round budget)
    and the failure is reported.

    Example: A model that only ever asks for an unavailable tool is stopped
      Given the model always asks for a tool tellme does not offer
      When the operator sends the prompt
      Then tellme reports the tool request failed
      And tellme exits with the tool error code
