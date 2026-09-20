Feature: Inspecting the session history

  # Interface truth (CLI end, `history` module) — `-l N` lists the most recent messages without a
  # provider request. Acceptance journeys: features/acceptance/inspecting-the-session-history.feature.
  #
  # Round 073 (ADR 0045): the listed message is now PRESENTED as a **role header line** (`[USER]` /
  # `[MODEL]`) plus a body — the **model** body rendered as Markdown by the shared glamour renderer,
  # the **operator** body verbatim (a recorded divergence from the reference, which renders both) —
  # with one blank line after every message. The header is accented blue/magenta only when `stdout`
  # is a terminal and `-r` is off; under `-r` the model body is verbatim and no accent is emitted.
  # The count/selection semantics are unchanged; the tool activity stays omitted.

  Rule: The operator can list the most recent messages

    Example: Listing the last two messages
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the diagnostics are shown at a terminal
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
        | What is my name?  | Alice  |
      When the operator asks tellme to list the last 2 messages
      Then tellme lists the last 2 messages
      And tellme sends no request to any provider
      And the run reports no post-turn status
      And the run shows no progress spinner
      And tellme exits successfully

  Rule: Listing a session with no persisted history lists nothing

    Example: Listing an empty session
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      When the operator asks tellme to list the last 5 messages
      Then tellme lists no messages
      And tellme exits successfully

  Rule: Listing after a tool-using turn shows only the operator's messages

    Example: The listing omits the tool activity
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds a tool-using exchange
      When the operator asks tellme to list the last 2 messages
      Then tellme lists only the operator's messages
      And tellme exits successfully

  Rule: Listing reports no payload status

    Example: A listing shows no payload status
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
      When the operator asks tellme to list the last 2 messages
      Then no payload status is reported
      And tellme exits successfully

  # Round 053 (closes #103; ADR 0022): the offline session commands select the
  # session named by the `-c` configuration when the environment forces no mode.
  Rule: The session listed is the one the named configuration belongs to

    Example: Two configurations list their own sessions
      Given the operator has a runnable tellme installation
      And the runtime home holds the configuration "a.yaml" in mode "alpha" holding the answer "from alpha"
      And the runtime home holds the configuration "b.yaml" in mode "beta" holding the answer "from beta"
      When the operator asks tellme to list the last 1 messages of the configuration "a.yaml"
      Then tellme lists the assistant message "from alpha"
      And tellme exits successfully

    Example: The other configuration lists a different session
      Given the operator has a runnable tellme installation
      And the runtime home holds the configuration "a.yaml" in mode "alpha" holding the answer "from alpha"
      And the runtime home holds the configuration "b.yaml" in mode "beta" holding the answer "from beta"
      When the operator asks tellme to list the last 1 messages of the configuration "b.yaml"
      Then tellme lists the assistant message "from beta"
      And tellme exits successfully

  Rule: A named configuration that cannot be read is refused

    Example: Naming a configuration that does not exist
      Given the operator has a runnable tellme installation
      When the operator asks tellme to list the last 1 messages of the configuration "missing.yaml"
      Then tellme exits with the configuration error code

  Rule: The list length defaults to one when the count is omitted

    Example: Listing with no count
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
        | What is my name?  | Alice  |
      When the operator asks tellme to list the last messages without a count
      Then tellme lists the last 1 messages
      And tellme exits successfully

  Rule: A forced session mode outranks the named configuration

    Example: The environment mode wins over the named configuration
      Given the operator has a runnable tellme installation
      And the runtime home holds the configuration "a.yaml" in mode "alpha" holding the answer "from alpha"
      And the runtime home holds the configuration "g.yaml" in mode "gamma" holding the answer "from gamma"
      And the effective mode is "gamma"
      When the operator asks tellme to list the last 1 messages of the configuration "a.yaml"
      Then tellme lists the assistant message "from gamma"
      And tellme exits successfully

  # Round 073 (ADR 0045): the listing's per-message presentation.
  Rule: A listed message is introduced by the role that produced it

    Example: A listed exchange names the operator and the model
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt            | answer |
        | My name is Alice. | Noted. |
        | What is my name?  | Alice  |
      When the operator asks tellme to list the last 2 messages
      Then tellme heads each listed message with its role
      And the listed model answer is presented as formatted prose
      And tellme exits successfully

  Rule: The listed operator prompt is shown exactly as it was written

    Example: A listed prompt is echoed verbatim, never reformatted
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt             | answer |
        | What is **Alice**? | Noted. |
      When the operator asks tellme to list the last 2 messages
      Then the listed operator prompt is shown verbatim
      And the listing carries no accents
      And tellme exits successfully

  Rule: The listed model answer is shown as its raw source under raw output

    Example: The raw listing shows the answer's source
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt | answer                   |
        | hi     | Hello **Alice**, welcome |
      When the operator asks tellme to list the last 2 messages as raw output
      Then the listed model answer is shown as its raw source
      And the listing carries no accents
      And tellme exits successfully

  Rule: The listed messages are separated so the listing reads as a conversation

    Example: A blank line separates the messages
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt | answer |
        | hi     | Noted. |
      When the operator asks tellme to list the last 2 messages
      Then the listed messages are separated by a blank line
      And tellme exits successfully

  Rule: The role lines are accented only when the listing goes to a terminal

    Example: A terminal listing accents the role lines
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the output is shown at a terminal
      And the session history already holds the exchanges:
        | prompt | answer |
        | hi     | Noted. |
      When the operator asks tellme to list the last 2 messages
      Then the listing accents the operator role in blue and the model role in magenta
      And tellme exits successfully

    Example: A redirected listing carries no accents
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session history already holds the exchanges:
        | prompt | answer |
        | hi     | Noted. |
      When the operator asks tellme to list the last 2 messages
      Then the listing carries no accents
      And tellme exits successfully
