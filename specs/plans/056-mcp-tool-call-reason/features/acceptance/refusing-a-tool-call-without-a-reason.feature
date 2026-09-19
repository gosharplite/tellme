Feature: Refusing a tool call that does not say why

  # Acceptance only (round 056). "No reason, no go": a tool call — tellme's own
  # or a remote one — that does not state a reason does not run; the assistant
  # is asked to try again with a reason. Business language only.

  Rule: A remote call that does not state a reason does not reach the server

    Example: A reasonless remote call is turned away, then retried with a reason
      Given the operator has a runnable tellme installation
      And the runtime home holds a configuration with a remote MCP server "shop" that offers a tool "lookup_price"
      And the MCP server "shop" is reachable
      And the assistant first calls the tool "lookup_price" on the server "shop" without saying why, then calls it again saying why
      When the operator asks tellme "What does the gadget cost?"
      Then the reasonless call did not reach the server "shop"
      And the assistant was asked to try again with a reason
      And the retried call used the tool "lookup_price" and tellme showed its reason
      And tellme exits successfully

  Rule: One of tellme's own tool calls without a reason does not run

    Example: A reasonless built-in call is turned away, then retried with a reason
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the working directory contains a file "notes.txt"
      And the assistant first reads "notes.txt" without saying why, then reads it again saying why
      When the operator starts tellme with the prompt "Read the notes."
      Then the reasonless read did not happen
      And the assistant was asked to try again with a reason
      And the retried read happened and tellme showed its reason
      And tellme exits successfully
