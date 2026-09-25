Feature: Keeping remembered MCP tools working

  # Plan-side acceptance journey — round 088, anchor issue #182.
  #
  # Round 087 cached a server's tool list across invocations, but a warm cache hit
  # broke any server whose tools route an argument as an HTTP header (the GitHub MCP
  # server): the lazy client connected without issuing `tools/list`, so the client
  # SDK never learned the header annotations. The cache also lived at $TELL_ME_HOME
  # root instead of the per-mode workspace. This round repairs both.
  #
  # This layer is PM-readable business language; the executable interface truth
  # lives under specs/truth/features/cli/**.

  Rule: A remembered tool that needs its arguments routed as headers still works

    Example: A remembered header-routed tool is actually run
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "gh" that offers a tool "create_issue" whose "owner" argument is routed as a header
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "create_issue" from the server "gh" with the reason "file a bug" and then answers with "Filed."
      And the tools of the MCP server "gh" have already been discovered
      When the operator starts tellme with the prompt "File a bug."
      Then tellme called the tool "create_issue" on the MCP server "gh"
      And tellme prints the provider's answer "Filed."
      And tellme exits successfully

  Rule: A remembered tool list is kept per session mode

    Example: The remembered tool list lives in the session's own workspace
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then tellme remembered the tools of the MCP server "shop" in the session workspace
      And tellme exits successfully
