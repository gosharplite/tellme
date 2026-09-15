Feature: Watching the tool loop work

  # Interface truth (CLI end, `chat` module) — the tool loop reports its activity on the diagnostic
  # output while it runs (discrete loop-step lines, not token streaming). Acceptance journeys:
  # features/acceptance/reporting-each-tool-use.feature and
  # features/acceptance/separating-the-tools-from-the-answer.feature.

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

    # The fixture deliberately scripts a `read_files` call WITHOUT `reason` — a schema-nonconforming
    # call the three filesystem tools' schema forbids (round-021 D2) — to cover the generic renderer
    # for a tool that states no reason (round-022 research Decision 4).

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

  Rule: A tool-using run reports the tools in the order they were used

    Example: Two tools used in sequence are reported in order
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a sub-folder "src"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint shows the folder tree and then reads "notes.txt" and then answers with "The launch code is ORANGE."
      When the operator starts tellme with the prompt "Survey the project, then read notes.txt."
      Then the run reported the tool calls in order "get_tree" and "read_files"
      And tellme exits successfully

  Rule: A tool-using run separates the tool report from the answer

    Example: A positional run separates the report from the answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the tool report is separated from the answer
      And tellme exits successfully

    Example: An interactive-prompt run also separates the report from the answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "The launch code is ORANGE"
      When the operator submits the prompt "Read notes.txt and summarise it." at the interactive prompt
      Then the tool report is separated from the answer
      And tellme exits successfully

  Rule: A run that used no tool adds no separating blank line

    # Round 023: the `-i` submit surface is now a CHROME surface, so the round-017 frame gap supplies the
    # single blank before the answer. The negative is re-anchored (PR #50 review B1): on a non-tool turn
    # exactly ONE blank line separates the pre-flight payload line from the answer (the frame gap,
    # undoubled) — a stray round-022 blank would make it two.

    Example: An interactive-prompt run that used no tool adds no second separating blank
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the operator is working at an interactive terminal
      And a configured provider "test-model" whose endpoint answers with "The launch code is ORANGE"
      When the operator submits the prompt "Say the launch code." at the interactive prompt
      Then the pre-flight payload line is separated from the answer by a single blank line
      And tellme exits successfully
