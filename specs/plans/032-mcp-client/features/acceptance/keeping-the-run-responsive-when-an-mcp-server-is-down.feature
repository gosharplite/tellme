Feature: Keeping a run responsive when an MCP server is unavailable

  # Acceptance only: a configured MCP server that is offline or hanging must
  # never make the operator wait — every run stays fast, and the servers that
  # are reachable keep working. This is the behaviour that makes an off-line
  # server a non-event instead of the reason to comment it out.

  Rule: A configured MCP server that does not answer must not stall a run

    Example: An unresponsive server is skipped and the answer still comes
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "hf" that never responds
      When the operator asks tellme "hello"
      Then tellme answers promptly instead of waiting for the server "hf" to give up
      And tellme warns on stderr that the MCP server "hf" could not be reached
      And tellme exits successfully

  Rule: One unavailable server must not affect the servers that are reachable

    Example: A reachable server still works while another one is down
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "shop" that is reachable and offers a tool "lookup_price"
      And the runtime home holds a configuration with a remote MCP server "hf" that never responds
      When the operator asks tellme "What does the gadget cost?"
      Then the assistant uses the tool "lookup_price" on the server "shop"
      And tellme answers promptly without waiting for the server "hf"
      And tellme exits successfully

  Rule: The wait for MCP servers stays bounded no matter how many are configured

    Example: Many unresponsive servers still do not add up to a long wait
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with several remote MCP servers that never respond
      When the operator asks tellme "hello"
      Then tellme answers promptly within the same small bound
      And tellme exits successfully
