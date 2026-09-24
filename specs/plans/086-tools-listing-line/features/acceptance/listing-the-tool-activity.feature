Feature: Reading a turn's tool activity from the listing

  # Plan-side acceptance journey — round 086, operator request (no anchor issue).
  # The `-l`/`--list` listing (round 073 / ADR 0045, extended by round 082 /
  # ADR 0054) presents each turn as a role header line plus a body and
  # deliberately omits the tool activity. This round surfaces ONE derived figure
  # — how many tool calls the turn made — as a single line between the operator's
  # prompt and the model's answer, without leaking any tool content.
  #
  # This layer is PM-readable business language; the executable interface truth
  # lives under specs/truth/features/cli/**.

  Rule: Each listed turn reports how much tool work it did

    Example: A tool-using turn is marked with its tool count
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history holds a turn that used a tool
      When the operator lists the last messages
      Then the listed turn reports that it used 1 tool call
      And the tool-activity report sits between the operator's prompt and the model's answer
      And tellme exits successfully

    Example: A turn that used no tools is marked with zero
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history holds a turn that used no tools
      When the operator lists the last messages
      Then the listed turn reports that it used 0 tool calls
      And tellme exits successfully

  Rule: The tool-activity report shows a count, never the tool's contents

    Example: The listing shows the count but not what the tool read or returned
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history holds a turn that used a tool
      When the operator lists the last messages
      Then the listed turn reports that it used 1 tool call
      And the listing reveals none of the tool's arguments or results
      And tellme exits successfully

  Rule: The tool-activity line is accented only when the listing goes to a terminal

    Example: A terminal listing accents the tool-activity line
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the output is shown at a terminal
      And the session history holds a turn that used a tool
      When the operator lists the last messages
      Then the tool-activity line is accented in yellow
      And tellme exits successfully

    Example: A redirected listing shows the tool-activity line plainly
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history holds a turn that used a tool
      When the operator lists the last messages
      Then the tool-activity line carries no accent
      And tellme exits successfully
