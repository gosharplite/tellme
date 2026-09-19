Feature: Keeping a Gemini turn usable when a remote MCP server annotates its tools

  # Acceptance only (round 061, issue #127). Business language only.
  #
  # The operator's report: with the GitHub MCP server enabled and a Gemini model
  # selected, EVERY turn died — the model was never reached. The server's tool
  # definitions carry server-side marks (they tell the *server* to route an
  # argument as a request header); tellme passed those marks on to Gemini
  # untouched, and Gemini's tool-definition reader has no slot for them, so it
  # rejected the whole request. The fix keeps the operator able to run a turn
  # while still offering the server's real arguments to the model.

  Rule: A turn completes while a remote MCP server with server-side argument marks is enabled

    Example: The operator asks a question with the GitHub tools enabled on a Gemini model
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-comment"
      And a remote MCP server "github" that advertises a tool "add_issue_comment" whose arguments carry a server-side routing mark
      And a configured provider "vertex-gemini" whose endpoint records the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "comment on the issue"
      Then the offered tool definitions carry no argument mark the provider cannot read
      And tellme answers with "done" and exits successfully

    Example: The same question with a server that carries no argument marks is unchanged
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-comment"
      And a remote MCP server "github" that advertises a tool "add_issue_comment" whose arguments carry no server-side routing mark
      And a configured provider "vertex-gemini" whose endpoint records the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "comment on the issue"
      Then tellme answers with "done" and exits successfully

  Rule: The offered tool still describes the server's real arguments

    # The mark is server-side plumbing and means nothing to the model; its removal
    # must not cost the model the arguments it actually has to fill in.

    Example: The model is offered the server's own arguments, without the mark
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-comment"
      And a remote MCP server "github" that advertises a tool "add_issue_comment" whose arguments "owner", "repo" and "body" carry a server-side routing mark
      And a configured provider "vertex-gemini" whose endpoint records the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "comment on the issue"
      Then the offered tool for "add_issue_comment" still describes the arguments "owner", "repo" and "body"
      And tellme answers with "done" and exits successfully

  Rule: The protection is part of every delivery

    # Documented narrowing (the 059 second-Rule precedent): the projection itself
    # is not observable through the built binary — a faithful capture cannot see
    # which keys were dropped before the wire. Its carrier is therefore the
    # hermetic unit pin over the declaration path (S-5), designed to fail if the
    # projection is removed; the Examples above remain the operator-visible half.
    # Recorded so a future PM pass may mirror it into the journey if a
    # runner-owned form appears. No `# [need clarification]` gap remains: the
    # clarify round closed all three decisions (CQ-1 → C, CQ-2 → ii, CQ-3 → i).
