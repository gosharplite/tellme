Feature: Watching the tool loop work

  # Interface truth (CLI end, `chat` module) — the tool loop reports its activity on the diagnostic
  # output in the reference's decomposed shape (round 034). Acceptance journeys:
  # features/acceptance/showing-the-tool-calls.feature, framing-the-turn-per-call.feature,
  # streaming-the-command-output.feature.

  Rule: A tool-using run reports each call's step, reason, action, and result

    Example: The run reports the step marker, the reason, the action, and the result
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with the reason "checking the launch code" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the run reported the tool step marker
      And the run reported the reason "checking the launch code" for the tool call "read_files"
      And the run reported the action for the tool call "read_files" without the reason
      And the run reported the result for the tool call "read_files"
      And tellme exits successfully

  Rule: A tool-using run reports the round's reasons again at the call's end

    Example: The reasons reappear grouped before the measured payload status
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with the reason "checking the launch code" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the run reported the reasons again before the measured payload status
      And tellme exits successfully

  Rule: A tool-using run reports every call in the order it was made

    Example: Two tools used in sequence are reported in order
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a sub-folder "src"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint shows the folder tree and then reads "notes.txt" and then answers with "The launch code is ORANGE."
      When the operator starts tellme with the prompt "Survey the project, then read notes.txt."
      Then the run reported the action for the tool call "get_tree" before the action for the tool call "read_files"
      And tellme exits successfully

  Rule: A tool-using run streams a command's output live

    Example: A shell command's output is streamed as it runs
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a command and then answers with "done"
      When the operator starts tellme with the prompt "Run the command."
      Then the run streamed the command's output on its diagnostic output
      And tellme exits successfully

    Example: A command that writes to a file streams no output block
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a command that writes its output to a file and then reads it and then answers with "done"
      When the operator starts tellme with the prompt "Run the command."
      Then the run streamed no command output block for the writing command
      And tellme exits successfully

  Rule: An over-long rendered value is shortened to its rune cap

    Example: A very long file result is shortened to 200 runes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "long.txt" whose text is "The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog."
      And a configured provider "test-model" whose endpoint asks tellme to read "long.txt" and then answers with "done"
      When the operator starts tellme with the prompt "Read long.txt."
      Then the run reported the result for the tool call "read_files" shortened to at most 200 runes
      And tellme exits successfully

    Example: A very long argument value is shortened to 189 runes
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint creates the file "out.txt" with the content "The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog." and then answers with "done"
      When the operator starts tellme with the prompt "Create out.txt."
      Then the run reported the action value for the tool call "write_file" shortened to at most 189 runes
      And tellme exits successfully

  Rule: A shell command's streamed output is free of terminal control sequences

    # Round 038 (issue #78): the streamed `[Tool Output]` block must not be able to change how the
    # operator's terminal renders the rest of the run. The content lines are sanitized (ANSI escape
    # sequences + stray control bytes removed) and the block always closes in a default state.

    Example: A colouring command's output is shown as plain text
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a colouring command and then answers with "done"
      When the operator starts tellme with the prompt "Run the colouring command."
      Then the run streamed the command's output on its diagnostic output
      And the run streamed the command's output free of terminal control sequences
      And tellme exits successfully

    Example: A command stopped mid-output leaves the terminal in its default state
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint runs a colouring command that is stopped at its time limit and then answers with "done"
      When the operator starts tellme with the prompt "Run the long colouring command."
      Then the run streamed the command's output free of terminal control sequences
      And the terminal is left in its default state
      And tellme exits successfully

  Rule: Every tool-loop line is free of terminal control sequences

    # Round 039 (issue #80): the control-sequence sanitization is a single-owned `internal/ui` policy applied
    # by EVERY `[Tool …]` formatter, not just `[Tool Output]`. A model-authored reason, a result snippet, or
    # an argument key/value that carries terminal control data is shown as plain text.

    Example: A reason that carries control data is shown as plain text
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with a reason that carries terminal control data and then answers with "done"
      When the operator starts tellme with the prompt "Read notes.txt."
      Then the run reported a reason line free of terminal control sequences
      And tellme exits successfully

    Example: A result that carries control data is shown as plain text
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "colour.txt" whose text is "\x1b[31mthe launch code is ORANGE\x1b[0m"
      And a configured provider "test-model" whose endpoint asks tellme to read "colour.txt" and then answers with "done"
      When the operator starts tellme with the prompt "Read colour.txt."
      Then the run reported the result for the tool call "read_files" free of terminal control sequences
      And tellme exits successfully

    Example: An argument that carries control data is shown as plain text
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint creates the file "out.txt" with the content "\x1b[31mhello\x1b[0m" and then answers with "done"
      When the operator starts tellme with the prompt "Create out.txt."
      Then the run reported the action for the tool call "write_file" free of terminal control sequences
      And tellme exits successfully

  Rule: The live turn output is grouped by blank lines

    # Round 039 (operator spacing request): the live turn output is grouped — a blank line precedes EACH call's
    # begin block (the `[Tool Reason]` line, else the `[Tool Action]` line), a blank line precedes the trailing
    # grouped `[Tool Reason]` block (with no blanks inside it), and a blank line precedes the post-status group
    # (the measured payload line + metrics line + `Ready`). A recorded divergence from the reference, which
    # writes these lines densely.

    Example: A round of two calls separates each call's report
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a sub-folder "src"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint shows the folder tree and then reads "notes.txt" and then answers with "the launch code is ORANGE"
      When the operator starts tellme with the prompt "Survey the project, then read notes.txt."
      Then each tool call's report begins after a blank line
      And tellme exits successfully

    Example: A call that states no reason still starts a fresh block
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "done"
      When the operator starts tellme with the prompt "Read notes.txt."
      Then the action of the call without a reason begins after a blank line
      And tellme exits successfully

    Example: The round's grouped reasons follow a blank line
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" with the reason "checking the launch code" and then answers with "The launch code is ORANGE"
      When the operator starts tellme with the prompt "Read notes.txt and summarise it."
      Then the trailing reason summary follows a blank line
      And tellme exits successfully

    Example: The turn's closing status follows a blank line
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "the launch code is ORANGE"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "done" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      When the operator starts tellme with the prompt "Read notes.txt."
      Then the turn's closing status follows a blank line
      And tellme exits successfully
