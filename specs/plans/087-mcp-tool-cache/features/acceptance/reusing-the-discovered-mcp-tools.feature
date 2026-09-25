Feature: Reusing discovered MCP tools across prompts

  # Plan-side acceptance journey — round 087, anchor issue #180.
  #
  # Every prompt-bearing tellme run dials each enabled remote MCP server before the
  # first provider request, purely to learn which tools to offer the model. With
  # servers declared, that is a network round-trip on the begin path of every
  # command — even a prompt that never uses a tool. This round lets tellme reuse a
  # previously discovered tool list (a cross-invocation cache), so the common path
  # dials nothing, while a cold/stale cache still degrades to the existing bounded
  # discovery and the offered tools are exactly what a live discovery would offer.
  #
  # This layer is PM-readable business language; the executable interface truth
  # lives under specs/truth/features/cli/**.

  Rule: A repeat prompt reuses the tools discovered earlier instead of dialing the server

    Example: A second prompt makes no new contact with the server
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint answers with "done"
      And tellme has already discovered the tools of the MCP server "shop"
      When the operator starts tellme with the prompt "hello again"
      Then tellme made no further contact with the MCP server "shop"
      And the request still offered the tool "lookup_price" from the MCP server "shop"
      And tellme exits successfully

  Rule: A first prompt (no cached tools yet) discovers and remembers them

    Example: The first prompt discovers the server's tools and remembers them
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint answers with "done"
      When the operator starts tellme with the prompt "hello"
      Then tellme discovered the tools of the MCP server "shop"
      And tellme remembered the tools of the MCP server "shop"
      And the request offered the tool "lookup_price" from the MCP server "shop"
      And tellme exits successfully

  Rule: A remembered tool list that has aged is still served, then refreshed quietly

    Example: An aged tool list is served before the request and refreshed after it
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint answers with "done"
      And the remembered tools of the MCP server "shop" are older than their usable age
      When the operator starts tellme with the prompt "hello"
      Then the request still offered the tool "lookup_price" from the MCP server "shop"
      And tellme exits successfully

  Rule: A server that has gone silent does not break a turn that uses a remembered tool

    Example: A remembered tool against a silent server fails softly
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the price" and then answers with "I could not check the price."
      And tellme has already discovered the tools of the MCP server "shop"
      And the MCP server "shop" has since stopped answering tool calls
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the run continued past the failed MCP tool call
      And tellme prints the provider's answer "I could not check the price."
      And tellme exits successfully

  Rule: A fresh start does not forget the discovered tools

    Example: Starting a new session keeps the remembered tools
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint answers with "done"
      And tellme has already discovered the tools of the MCP server "shop"
      When the operator starts a new session with the prompt "hello again"
      Then tellme made no further contact with the MCP server "shop"
      And the request still offered the tool "lookup_price" from the MCP server "shop"
      And tellme exits successfully
