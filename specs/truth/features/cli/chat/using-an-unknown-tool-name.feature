Feature: Recovering from an unknown tool name

  # Interface truth (CLI end, `chat` module) — round 076 (issue #154). An unknown
  # tool name (a name the registry does not hold) is a recoverable model slip: the
  # loop folds back a `tool`-role result naming the unknown tool and a bounded list
  # of the available wire names, and CONTINUES — never the terminal request-level
  # failure. The recoverable path is BOUNDED per turn: after `maxUnknownToolFolds`
  # fold-backs in one turn the loop stops folding back and reports the frozen
  # `the tool request failed` phrase with the tool error code. An unknown call runs
  # no tool and records nothing (no history step, no usage record).
  # Acceptance journey: features/acceptance/recovering-from-an-unknown-tool-name.feature.

  Rule: An unknown tool name is fed back and the turn continues

    Example: A one-off unknown tool name is recovered and the turn finishes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint first asks for a tool that is not available and then asks tellme to read "notes.txt" and then answers with "found it"
      When the operator starts tellme with the prompt "Use the time-travel tool, then read the note."
      Then the run reported the unavailable tool "time_travel"
      And tellme prints the provider's answer "found it"
      And tellme exits successfully

  Rule: A provider that keeps asking for an unknown tool is stopped at the per-turn cap

    Example: A provider that only ever asks for an unavailable tool is stopped and reported
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint asks for a tool that is not available
      When the operator starts tellme with the prompt "Use the time-travel tool."
      Then the run reported the unavailable tool "time_travel"
      And tellme explains on stderr that "the tool request failed"
      And tellme exits with the tool error code
