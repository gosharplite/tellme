Feature: Explaining a remote tool call

  # Acceptance only (round 056). When the assistant calls a tool a remote MCP
  # server offers, the operator must be able to see *why* — the same "[Tool
  # Reason]" line the operator already gets for tellme's own tools. The reason
  # belongs to tellme's call, not to the server. Business language only.

  Rule: A remote tool call states its reason

    Example: The operator sees why the assistant called a remote tool
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "shop" that offers a tool "lookup_price"
      And the MCP server "shop" is reachable
      And the assistant answers by calling the tool "lookup_price" on the server "shop" and saying why
      When the operator asks tellme "What does the gadget cost?"
      Then tellme shows the reason the assistant gave for calling "lookup_price"
      And tellme prints an answer that reflects the value the server returned
      And tellme exits successfully

  Rule: The reason belongs to the operator's session, not to the server

    Example: The server is not given the assistant's reason
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "shop" that offers a tool "lookup_price"
      And the MCP server "shop" is reachable
      And the assistant answers by calling the tool "lookup_price" on the server "shop" and saying why
      When the operator asks tellme "What does the gadget cost?"
      Then tellme shows the reason the assistant gave for calling "lookup_price"
      And the server "shop" was never given the assistant's reason
      And tellme exits successfully
