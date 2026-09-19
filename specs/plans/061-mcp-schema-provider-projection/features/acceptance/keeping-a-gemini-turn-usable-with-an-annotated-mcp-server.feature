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
      And the runtime home is "ait-tmg"
      And a remote MCP server "github" that offers a tool "add_issue_comment" answering "commented" whose arguments carry a server-side mark
      And a configured gemini provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "comment on the issue"
      Then the offered tool "add_issue_comment" from the MCP server "github" carries no server-side mark
      And the offered tool "add_issue_comment" from the MCP server "github" carries no keyword the provider cannot read
      And tellme prints the provider's answer "done"
      And tellme exits successfully

    Example: The same question on a tolerant provider still loses the server-side marks
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "github" that offers a tool "add_issue_comment" answering "commented" whose arguments carry a server-side mark
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "comment on the issue"
      Then the offered tool "add_issue_comment" from the MCP server "github" carries no server-side mark
      And tellme exits successfully

    Example: A server with no argument marks still offers its tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured gemini provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the request offered the tool "lookup_price" from the MCP server "shop" alongside the agent tools
      And the offered tool "lookup_price" from the MCP server "shop" carries no server-side mark
      And tellme exits successfully

  Rule: The offered tool still describes the server's real arguments

    # The mark is server-side plumbing and means nothing to the model; its removal
    # must not cost the model the arguments it actually has to fill in.

    Example: The model is offered the server's own arguments, without the mark
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "github" that offers a tool "add_issue_comment" answering "commented" whose arguments carry a server-side mark
      And a configured gemini provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "comment on the issue"
      Then the offered tool "add_issue_comment" from the MCP server "github" still describes the arguments "owner,repo,body"
      And tellme exits successfully

  Rule: The protection is part of every delivery

    # Documented narrowing (the round-059 second-Rule precedent): the projection is
    # not fully observable through the built binary alone — a faithful capture
    # cannot ask a real provider whether it would reject a payload. Its carriers
    # are therefore (a) the live probe recorded in ADR 0031 (the accepted/rejected
    # keyword table) and (b) the hermetic unit pin over the declaration path
    # (S-5), designed to fail if the projection is removed — plus the Examples
    # above, which observe the wire bytes the fake provider recorded. No
    # `# [need clarification]` gap remains: the clarify round closed all three
    # decisions (CQ-1 → C, CQ-2 → ii, CQ-3 → i).
