Feature: Keeping the remote server's definition untouched

  # Acceptance only (round 056). tellme must not alter a remote server's tool
  # definition, and must send the server only the arguments its own tool
  # expects — never anything tellme added. Business language only.

  Rule: A server's tool is offered as the server declared it

    Example: The assistant is offered the server's tool definition unchanged
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "shop" that offers a tool "lookup_price"
      And the MCP server "shop" advertises a definition for the tool "lookup_price"
      And the MCP server "shop" is reachable
      And the assistant is shown the tools it may use
      When the operator asks tellme "Which tools can you use?"
      Then the definition of "lookup_price" the assistant was shown is the one the server advertised, unchanged
      And tellme exits successfully

  Rule: A server receives only the arguments its own tool expects

    Example: The server is not sent anything tellme added
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "shop" that offers a tool "lookup_price"
      And the MCP server "shop" is reachable
      And the assistant answers by calling the tool "lookup_price" on the server "shop" with the arguments its tool expects and saying why
      When the operator asks tellme "What does the gadget cost?"
      Then the server "shop" received exactly the arguments its tool expects
      And the server "shop" received nothing other than those arguments
      And tellme exits successfully
