Feature: Accounting for how the tools are used

  # Interface truth (CLI end, `chat` module) — every EXECUTED agent-tool invocation is recorded to a
  # user-global append-only log (`~/.tellme/tools-count.jsonl`, resolved via the CHILD's `HOME`) with its
  # outcome (`ok` / `error` / `timeout`; error = a non-nil tool error, timeout = a nil-error result at
  # the per-call deadline, ok otherwise, incl. a bounded/truncated result), and the operator can review
  # the per-tool roll-up offline via the `--tool-usage` flag (no provider, no stdin). Acceptance
  # journeys: features/acceptance/keeping-a-record-of-tool-use.feature and
  # features/acceptance/reviewing-how-the-tools-have-been-used.feature.
  #
  # The E2E harness points the CHILD's `HOME` at a per-scenario temporary directory, so the global log
  # inspected here is the scenario's own `$HOME/.tellme/tools-count.jsonl` — never the operator's real
  # `~/.tellme`.

  Rule: Each executed tool call is recorded with its outcome

    Example: A tool that succeeds is recorded as a success
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the tool usage shows the tool "read_files" with 1 successes, 0 failures, and 0 timeouts
      And tellme exits successfully

    Example: A tool that fails is recorded as a failure
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint lists a folder that does not exist and then answers with "done"
      When the operator starts tellme with the prompt "List the folder that does not exist."
      Then the tool usage shows the tool "list_files" with 0 successes, 1 failures, and 0 timeouts
      And tellme exits successfully

    Example: A tool that runs out of time is recorded as having run out of time
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a command that never returns within a short limit and then answers with "done"
      When the operator starts tellme with the prompt "Run a command that never finishes."
      Then the tool usage shows the tool "execute_command" with 0 successes, 0 failures, and 1 timeouts
      And tellme exits successfully

    Example: A turn that uses no tool adds no record
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "Four"
      When the operator starts tellme with the prompt "What is two plus two?"
      Then the tool usage shows no tool has been used
      And tellme exits successfully

  Rule: The record is kept across sessions

    Example: The record survives a fresh session
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the tool usage already records that the tool "read_files" was used 3 times with the outcome "succeeded"
      And a configured provider "test-model" whose endpoint answers with "done"
      When the operator starts a fresh session with "--new" and the prompt "hi"
      Then the tool usage shows the tool "read_files" with 3 successes, 0 failures, and 0 timeouts
      And tellme exits successfully

  Rule: The operator reviews the per-tool roll-up offline

    Example: The operator reviews the tools after some use
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the tool usage already records that the tool "read_files" was used 2 times with the outcome "succeeded"
      And the tool usage already records that the tool "read_files" was used 1 times with the outcome "failed"
      And the tool usage already records that the tool "list_files" was used 1 times with the outcome "succeeded"
      And a configured provider "test-model" whose endpoint answers with "done"
      When the operator reviews how the tools have been used
      Then the review shows the tool "read_files" with 2 successes, 1 failures, and 0 timeouts
      And the review shows the tool "list_files" with 1 successes, 0 failures, and 0 timeouts
      And tellme sends no request to the provider "test-model"
      And tellme exits successfully

    Example: A tool that has never been used appears with no uses
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the tool usage already records that the tool "read_files" was used 1 times with the outcome "succeeded"
      When the operator reviews how the tools have been used
      Then the review shows the tool "get_tree" was never used
      And tellme exits successfully

    Example: The operator reviews the tools before any use
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      When the operator reviews how the tools have been used
      Then the review shows every tool with no uses
      And tellme exits successfully

    Example: The report works without a runtime home
      # The report is `--version`-class: it must NOT require `TELL_ME_HOME` (it reads only
      # `os.UserHomeDir()` + the live registry). A bare `tellme --tool-usage` in a shell without
      # `TELL_ME_HOME` must succeed, not exit 4 (the `-l`-class trap).
      Given the operator has a runnable tellme installation
      And the runtime home is not set
      And the tool usage already records that the tool "read_files" was used 1 times with the outcome "succeeded"
      When the operator reviews how the tools have been used
      Then the review shows the tool "read_files" with 1 successes, 0 failures, and 0 timeouts
      And tellme exits successfully
