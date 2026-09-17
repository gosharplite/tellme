Feature: Presenting the post-turn status

  # Interface truth (CLI end, `chat` module) — after the answer, a prompt-bearing turn reports the
  # reference's post-turn status on the diagnostic stream: the metrics of the API call that just
  # returned (missed/cached/completed/reasoning tokens) and a `╰─⠿ Ready` session summary (three costs,
  # the session's token totals, and the cache-hit rate). It is suppressed when the provider reports no
  # usage, and `stdout` stays byte-exact. Acceptance journeys:
  # features/acceptance/reporting-the-metrics-of-the-request.feature,
  # reporting-the-cost-and-session-summary.feature, keeping-the-post-turn-lines-bounded.feature.
  #
  # The token counts are deliberately large (a real turn's magnitude) so the `$%.4f` costs are
  # observable at four decimals against the per-million-token `MODELS` rates.

  Rule: The run reports the metrics of the request that just completed

    Example: The operator runs a single-request turn
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      When the operator starts tellme with the prompt "hi"
      Then the run reports the token metrics of the request that just completed
      And the reported metrics line shows 40000 missed, 60000 cached, 3000 completed, and 2000 reasoning tokens
      And tellme exits successfully

    Example: A tool-less turn's closing status is not blank-separated
      # Round 039 (review B1): the blank-line grouping applies only to a tool-using turn — a tool-less
      # turn is unchanged, so its post-status group gains no leading blank.
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      When the operator starts tellme with the prompt "hi"
      Then the run reports the token metrics of the request that just completed
      And the turn shows no doubled blank line
      And tellme exits successfully

    Example: The provider reports no reasoning tokens
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 0        |
      When the operator starts tellme with the prompt "hi"
      Then the run reports the token metrics of the request that just completed
      And the reported metrics line shows 40000 missed, 60000 cached, 3000 completed, and 0 reasoning tokens
      And tellme exits successfully

  Rule: The metrics reflect the last request of a tool-using turn

    Example: The operator runs a turn that used a tool before answering
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "hello"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      When the operator starts tellme with the prompt "read the notes"
      Then the run reports the token metrics of the request that just completed
      And the reported metrics line shows 40000 missed, 60000 cached, 3000 completed, and 2000 reasoning tokens
      And tellme exits successfully

  Rule: The run reports the cost of the last request, the turn, and the session

    Example: The operator runs a single-request turn on a fresh session
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      And the configuration prices the active model with hit "0.0028", miss "0.14", and completion "0.28" per million tokens
      When the operator starts tellme with the prompt "hi"
      Then the run reports the cost of the request, the turn, and the session
      And the reported request, turn, and session costs are equal
      And tellme exits successfully

  Rule: A tool-using turn costs more than its last request

    Example: The operator runs a turn that used a tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "hello"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      And the configuration prices the active model with hit "0.0028", miss "0.14", and completion "0.28" per million tokens
      When the operator starts tellme with the prompt "read the notes"
      Then the run reports the cost of the request, the turn, and the session
      And the reported turn cost is greater than the request cost
      And tellme exits successfully

  Rule: The session summary accumulates across a session

    Example: The operator continues a session that already has usage
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session's usage log already records a prior call
      And a configured provider "test-model" whose endpoint answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      And the configuration prices the active model with hit "0.0028", miss "0.14", and completion "0.28" per million tokens
      When the operator starts tellme with the prompt "carry on"
      Then the run reports the cost of the request, the turn, and the session
      And the reported session summary includes the earlier call
      And tellme exits successfully

  Rule: The session summary resets on a fresh session

    Example: The operator starts a fresh session
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the session's usage log already records a prior call
      And a configured provider "test-model" whose endpoint answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      And the configuration prices the active model with hit "0.0028", miss "0.14", and completion "0.28" per million tokens
      When the operator starts a fresh session with "--new" and the prompt "hi"
      Then the run reports the cost of the request, the turn, and the session
      And the reported request, turn, and session costs are equal
      And tellme exits successfully

  Rule: The session summary reports the session's tokens and cache share

    Example: The operator runs a turn on a fresh session
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      And the configuration prices the active model with hit "0.0028", miss "0.14", and completion "0.28" per million tokens
      When the operator starts tellme with the prompt "hi"
      Then the run reports the session's missed, cached, and output tokens
      And the run reports the share of the prompt served from cache
      And tellme exits successfully

  Rule: A model the configuration does not price is reported as free

    Example: The operator uses a model the configuration does not price
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      When the operator starts tellme with the prompt "hi"
      Then the run reports the cost of the request, the turn, and the session
      And the run reports a zero cost
      And the run reports the session's missed, cached, and output tokens
      And tellme exits successfully

  Rule: The run reports no post-turn status when the provider reports no usage

    Example: The provider reports no usage
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good" and reports no usage
      When the operator starts tellme with the prompt "hi"
      Then the run reports no post-turn status
      And tellme prints the provider's answer "all good"
      And tellme exits successfully

  Rule: The post-turn status trails the answer

    Example: The operator pipes the prompt in and reads the answer
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "test-model" whose endpoint answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      When the operator pipes "summarise the notes" into tellme
      Then the post-turn status trails the answer
      And tellme prints the provider's answer "all good"
      And tellme exits successfully

  Rule: The post-turn status is reported once per model request

    Example: A tool-using turn reports a status tail for each request
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt" whose text is "all good"
      And a configured provider "test-model" whose endpoint asks tellme to read "notes.txt" and then answers with "all good" and reports the token usage:
        | prompt | cached | completion | thinking |
        | 100000 | 60000  | 3000       | 2000     |
      When the operator starts tellme with the prompt "read the notes"
      Then the run reports the post-turn status once per model request
      And the last post-turn status trails the answer
      And tellme exits successfully
