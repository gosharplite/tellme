Feature: Using tools from a remote MCP server

  # Interface truth (CLI end, `chat` module) — round 032: on a prompt-bearing turn, tellme discovers the
  # tools of each configured, enabled **remote (Streamable HTTP)** MCP server and offers them to the
  # model alongside its native tools under a deterministic namespaced name; a prompt may then call one,
  # and its result is fed back like a native tool result. Discovery never stalls the run: a server that
  # does not answer within the fixed fast-fail bound is skipped with a warning, and a server marked off
  # is never contacted. Acceptance journeys: features/acceptance/using-a-tool-from-a-remote-mcp-server.feature,
  # …/keeping-the-run-responsive-when-an-mcp-server-is-down.feature, …/turning-off-an-mcp-server.feature.

  Rule: A configured remote MCP server's tools are offered to the model

    Example: The offered tool set includes the server's tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the request offered the tool "lookup_price" from the MCP server "shop"
      And tellme exits successfully

  Rule: A prompt that needs an MCP tool is answered using it

    Example: A question answered by calling the server's tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" and then answers with "The gadget costs $42."
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then tellme called the tool "lookup_price" on the MCP server "shop"
      And tellme prints the provider's answer "The gadget costs $42."
      And tellme exits successfully

  Rule: An unreachable MCP server is skipped without stalling the run

    Example: A server that never answers is skipped and the run still answers
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "hf" that never answers
      And a configured provider "test-model" whose endpoint answers with "done"
      When the operator starts tellme with the prompt "hello"
      Then tellme reported on stderr that the MCP server "hf" could not be reached
      And tellme prints the provider's answer "done"
      And tellme exits successfully

  Rule: A server marked off is never contacted

    Example: A disabled server is skipped and its tools are not offered
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "hf" that is marked off
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "hello"
      Then tellme never contacted the MCP server "hf"
      And the request offered no tool from the MCP server "hf"
      And tellme exits successfully

  Rule: tellme sends the resolved token to the server

    Example: A token-protected server receives the token
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that requires the token "s3cr3t" and offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" and then answers with "The gadget costs $42."
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the MCP server "shop" received the token "s3cr3t"
      And tellme exits successfully
