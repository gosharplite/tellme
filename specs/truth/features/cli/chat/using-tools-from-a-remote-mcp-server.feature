Feature: Using tools from a remote MCP server

  # Round 056 (ADR 0025; folds #121): an MCP tool call states its reason (tellme's own field) and the
  # server receives ONLY its own arguments; the server's advertised definition is relayed verbatim
  # (positioned as `MCP_PAYLOAD`'s subschema). The call JSON is `{reason, MCP_PAYLOAD}`; an
  # envelope-less call is refused (see `requiring-a-reason-to-call-a-tool.feature`).

  # Interface truth (CLI end, `chat` module) — round 032: on a prompt-bearing turn, tellme discovers the
  # tools of each configured, enabled **remote (Streamable HTTP)** MCP server and offers them to the
  # model alongside its native tools under a deterministic namespaced name; a prompt may then call one,
  # and its result is fed back like a native tool result. Discovery never stalls the run: a server that
  # does not answer within the fixed fast-fail bound is skipped with a warning, and a server marked off
  # is never contacted. Acceptance journeys: features/acceptance/using-a-tool-from-a-remote-mcp-server.feature,
  # …/keeping-the-run-responsive-when-an-mcp-server-is-down.feature, …/turning-off-an-mcp-server.feature.

  Rule: A server's tools are offered alongside the native tools

    Example: The offered tool set includes the server's tool alongside the native tools
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the request offered the tool "lookup_price" from the MCP server "shop" alongside the agent tools
      And tellme exits successfully

  Rule: A prompt that needs an MCP tool is answered using it

    Example: A question answered by calling the server's tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the gadget price" and then answers with "The gadget costs $42."
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
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the gadget price" and then answers with "The gadget costs $42."
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the MCP server "shop" received the token "s3cr3t"
      And tellme exits successfully

  Rule: A reachable server still works while another is unavailable

    Example: A reachable server's tool is used while another server is down
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a remote MCP server "hf" that never answers
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the gadget price" and then answers with "The gadget costs $42."
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then tellme called the tool "lookup_price" on the MCP server "shop"
      And tellme reported on stderr that the MCP server "hf" could not be reached
      And tellme exits successfully

  Rule: Several unresponsive servers still do not add up to a long wait

    # FR-010's *time* bound is witnessed by the single-server case (tasks.md T028 falsifiability
    # witness); this Rule asserts the observable behaviour for N servers (each warned+skipped, still answers).

    Example: Three unresponsive servers are each skipped and the run still answers
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "hf1" that never answers
      And a remote MCP server "hf2" that never answers
      And a remote MCP server "hf3" that never answers
      And a configured provider "test-model" whose endpoint answers with "done"
      When the operator starts tellme with the prompt "hello"
      Then tellme reported on stderr that the MCP server "hf1" could not be reached
      And tellme reported on stderr that the MCP server "hf2" could not be reached
      And tellme reported on stderr that the MCP server "hf3" could not be reached
      And tellme exits successfully

  Rule: A tool that advertises a malformed schema is not offered

    Example: A malformed schema is skipped and the run still succeeds
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "bad" that advertises a tool "broken" with a malformed schema
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the request offered no tool from the MCP server "bad"
      And tellme exits successfully

  Rule: A failed MCP tool call does not abort the run

    Example: A tool that reports a tool error does not abort the run
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" whose tool "lookup_price" reports a tool error
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the gadget price" and then answers with "I could not check the price."
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the run continued past the failed MCP tool call
      And tellme prints the provider's answer "I could not check the price."
      And tellme exits successfully

    Example: A transport failure does not abort the run
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that fails the tool call
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the gadget price" and then answers with "I could not check the price."
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the run continued past the failed MCP tool call
      And tellme prints the provider's answer "I could not check the price."
      And tellme exits successfully

  Rule: A server with no off switch is used as before

    # FR-012 (default-true): the server omits the off switch, so its tool is offered AND callable.
    Example: A server that omits the off switch is used
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the gadget price" and then answers with "The gadget costs $42."
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the request offered the tool "lookup_price" from the MCP server "shop" alongside the agent tools
      And tellme called the tool "lookup_price" on the MCP server "shop"
      And tellme exits successfully

  Rule: An MCP tool call states its reason, and the server gets only its own arguments

    # Round 056 (ADR 0025): the call is `{reason, MCP_PAYLOAD}`; tellme renders the reason as
    # `[Tool Reason]` and forwards ONLY the `MCP_PAYLOAD` object to the server. The server's declared
    # input schema is relayed verbatim as `MCP_PAYLOAD`'s subschema (never mutated).

    Example: The reason is shown and the server receives only its own arguments
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the gadget price" and then answers with "The gadget costs $42."
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the run reported the reason "check the gadget price" for the tool call "lookup_price"
      And the MCP server "shop" received only the arguments its tool expects
      And tellme exits successfully

    Example: The server's tool definition is offered unchanged
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the request offered the tool "lookup_price" from the MCP server "shop" with a reason and the server's declared schema
      And tellme exits successfully
