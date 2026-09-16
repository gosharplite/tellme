Feature: Turning an MCP server off without removing it

  # Acceptance only: the operator can keep a server's definition in the
  # configuration and simply switch it off, so a known-troublesome server is
  # never contacted — no need to delete or comment out its entry, and no need
  # to remember how to restore it later.

  Rule: A server marked off is never contacted

    Example: A server switched off is ignored, and the definition stays
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "hf" that is switched off
      When the operator asks tellme "hello"
      Then tellme never contacts the server "hf"
      And tellme prints an answer
      And tellme exits successfully

  Rule: A server is on unless the operator switches it off

    Example: A server with no off switch is used as before
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "shop" that is reachable, offers a tool "lookup_price", and is not switched off
      When the operator asks tellme "What does the gadget cost?"
      Then the assistant uses the tool "lookup_price" on the server "shop"
      And tellme exits successfully
