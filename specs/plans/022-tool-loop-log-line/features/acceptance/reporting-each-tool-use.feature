Feature: Reporting each tool use

  # Acceptance only: what the operator sees while tellme uses a tool during a
  # run — a single line naming the tool and, when the tool states one, its reason.

  Rule: Each tool the agent uses is reported on one line

    Example: A tool that states a reason
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that surveys the folder tree before answering
      When the operator asks tellme "Show me the folder tree."
      Then the run reports the tool it used and the reason the tool was given
      And the report names the tool and its reason on a single line
      And the report does not repeat the tool's raw request or its raw result
      And tellme exits successfully

    Example: A tool that states no reason
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that uses a tool which states no reason before answering
      When the operator asks tellme "Do the thing."
      Then the run reports the tool it used
      And the report names only the tool, with no reason
      And tellme exits successfully

    Example: Several tools in one run
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a reachable provider "test-model" that surveys the folder tree and then reads a file before answering
      When the operator asks tellme "Inspect the project."
      Then the run reports every tool it used, in the order they were used
      And each report is on its own line
      And tellme exits successfully
