Feature: Watching the tool loop work

  # Interface truth (CLI end, `chat` module) — the tool loop reports its activity on the diagnostic
  # output while it runs (discrete loop-step lines, not token streaming). Acceptance journey:
  # features/acceptance/watching-the-tool-loop.feature.

  Rule: A tool-using run reports its tool-loop activity

    Example: The run reports the tool call it made
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the run reported the tool call "read_files" on its diagnostic output
      And the tool activity is reported before the answer
      And tellme exits successfully

  Rule: A tool-using run reports the call's reason

    Example: The reason accompanies the tool call in the loop log
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with the reason "checking the launch code" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the run reported the reason "checking the launch code" for the tool call "read_files"
      And tellme exits successfully

  Rule: A tool-using run reports a call that states no reason

    Example: The call is reported without a reason tail
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the run reported the tool call "read_files" without a reason
      And tellme exits successfully

  Rule: A tool-using run reports each tool call on its own line

    Example: Two calls in one round are each reported
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "left.txt" whose text is "the code is ORANGE"
      And the working directory contains a file "right.txt" whose text is "the code is not BLUE"
      And a configured provider "test-model" whose endpoint asks tellme to read "left.txt" and "right.txt" and then answers with "The code is ORANGE."
      When the operator starts tellme with the prompt "Compare left.txt and right.txt."
      Then the run reported one tool-loop log line for each tool call
      And tellme exits successfully

  Rule: The answer is set apart from the tool report

    Example: A tool-using run separates the report from the answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the tool report is separated from the answer
      And tellme exits successfully
