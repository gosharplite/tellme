Feature: Using a tool a remote MCP server offers

  # Acceptance only: when the operator configures a remote MCP server, tellme
  # lets the assistant use that server's tools within a turn — the same way it
  # uses its own built-in tools — and reflects their results in the answer. A
  # server that reports a tool failed must not crash the run.

  Rule: tellme lets the assistant use a tool a configured remote MCP server offers

    Example: The assistant answers using a server's tool
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "shop" that offers a tool "lookup_price"
      And the MCP server "shop" is reachable
      When the operator asks tellme "What does the gadget cost?"
      Then the assistant uses the tool "lookup_price" on the server "shop"
      And tellme prints an answer that reflects the value the server returned
      And tellme exits successfully

    Example: A tool that fails at the server does not crash the run
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "shop" whose tool "lookup_price" reports a failure
      And the MCP server "shop" is reachable
      When the operator asks tellme "What does the gadget cost?"
      Then the assistant attempts the tool "lookup_price" on the server "shop"
      And tellme prints an answer that reflects that the lookup could not be completed
      And tellme exits successfully

  Rule: tellme presents a server's tools to the assistant alongside its own

    Example: The assistant can choose between a server's tool and a built-in tool
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "fs" that offers a tool "find_notes"
      And the MCP server "fs" is reachable
      When the operator asks tellme "Find my notes."
      Then the assistant may use the tool "find_notes" from the server "fs"
      And tellme prints an answer
      And tellme exits successfully

  Rule: tellme authenticates to a server that requires a token

    Example: A token-protected server is used when the operator provides a token
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "shop" that requires a token the operator provides
      And the MCP server "shop" accepts that token and offers a tool "lookup_price"
      When the operator asks tellme "What does the gadget cost?"
      Then the assistant uses the tool "lookup_price" on the server "shop"
      And tellme never writes the token where a reader could see it
      And tellme exits successfully
