Feature: Requiring a reason to call a tool

  # Interface truth (CLI end, `chat` module) — round 056 (ADR 0025; folds #121): *no reason, no go*.
  # A tool call — tellme's own or a remote MCP server's — whose reason is missing or blank does NOT
  # execute: the loop refuses it and folds back a recoverable result asking the model to retry with a
  # reason. The reason-presence rule is single-owned (the loop asks its injected `Lines.ReasonLine`);
  # the MCP `MCP_PAYLOAD` envelope check is the only MCP-specific addition.
  # Acceptance journeys: features/acceptance/refusing-a-tool-call-without-a-reason.feature.

  Rule: A remote tool call that states no reason is refused and never reaches the server

    Example: A reasonless remote call is turned away, then retried with a reason
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint first asks tellme to use the MCP tool "lookup_price" from the server "shop" without a reason and then with the reason "check the gadget price" and then answers with "The gadget costs $42."
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the MCP server "shop" received only the arguments its tool expects
      And the MCP server "shop" received exactly one call
      And the run asked for a reason before running a tool
      And the run reported the reason "check the gadget price" for the MCP tool "lookup_price" on the server "shop"
      And tellme exits successfully

  Rule: An MCP call that is not a valid envelope is refused without contacting the server

    # Round-056 review TD-056-1: the shape-violation half of FR-003 had only a Go
    # unit pin. A stray top-level key is refused at the adapter BEFORE CallTool, so
    # the server records no call at all.

    Example: A call carrying a stray argument is refused and the server is not contacted
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the gadget price" and a stray argument, and then answers with "done"
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the MCP server "shop" received no call
      And tellme exits successfully

  Rule: tellme's own tool call that states no reason does not run

    Example: A reasonless built-in call is turned away, then retried with a reason
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains no file "out.txt"
      And a configured provider "test-model" whose endpoint first asks tellme to create the file "out.txt" without a reason, then with the content "hello" and the reason "write the greeting", and then answers with "done"
      When the operator starts tellme with the prompt "Create out.txt."
      Then the run asked for a reason before running a tool
      And tellme created the file "out.txt"
      And the content of "out.txt" is exactly "hello"
      And tellme exits successfully
