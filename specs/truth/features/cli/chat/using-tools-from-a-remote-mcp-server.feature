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
      And the offered tool "lookup_price" from the MCP server "shop" carries no server-side mark
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
      Then the run reported the reason "check the gadget price" for the MCP tool "lookup_price" on the server "shop"
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

  Rule: A server's argument marks never reach the provider

    # Round 061 (issue #127 / ADR 0031): a server may annotate its arguments with
    # vendor extensions (the GitHub server's `x-mcp-header`, which only tells the
    # SERVER to route the argument as an HTTP header). A provider whose
    # tool-definition reader is closed (Vertex/Gemini) rejects the WHOLE request
    # when it meets such a keyword, so the annotation must not reach any provider;
    # the declared arguments are preserved.

    Example: A Gemini turn survives an annotated server
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "github" that offers a tool "add_issue_comment" answering "commented" whose arguments carry a server-side mark
      And a configured gemini provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "comment on the issue"
      Then the offered tool "add_issue_comment" from the MCP server "github" carries no server-side mark
      And the offered tool "add_issue_comment" from the MCP server "github" carries no keyword the provider cannot read
      And tellme prints the provider's answer "done"
      And tellme exits successfully

    Example: The server's arguments survive the projection
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "github" that offers a tool "add_issue_comment" answering "commented" whose arguments carry a server-side mark
      And a configured gemini provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "comment on the issue"
      Then the offered tool "add_issue_comment" from the MCP server "github" still describes the arguments "owner,repo,body"
      And tellme exits successfully

    Example: A tolerant provider also receives no server-side mark
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "github" that offers a tool "add_issue_comment" answering "commented" whose arguments carry a server-side mark
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "comment on the issue"
      Then the offered tool "add_issue_comment" from the MCP server "github" carries no server-side mark
      And tellme exits successfully

  Rule: An offered MCP tool's declaration makes its callable name discoverable

    # Round 077 (issue #155 / ADR 0049): the wire name is namespaced
    # (`mcp_<server>_<tool>`); the offered declaration's DESCRIPTION carries tellme's
    # call-name note naming that callable name, so a model that only recalls the
    # bare upstream name (e.g. `get_me`) can read the correct name. The server's own
    # text is relayed UNCHANGED — the note is added-to, never a rewrite (ADR 0025 D1).

    Example: The offered declaration names the callable wire name
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the offered tool "lookup_price" from the MCP server "shop" names its callable wire name in the description
      And tellme exits successfully

    Example: A server that ships no description is still offered by its callable name
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" with no description answering "$42"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the offered tool "lookup_price" from the MCP server "shop" falls back to its callable name, not the bare one
      And tellme exits successfully

  Rule: A remembered tool list is reused without contacting the server

    # Round 087 (issue #180 / ADR 0058): the prelude consults the cross-invocation
    # cache at $TELL_ME_HOME/mcp-toolcache.json. A warm, fresh entry is served with
    # ZERO MCP connections — a lazy client defers the connect to the first call.

    Example: A remembered tool list is offered without contacting the server
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      And the tools of the MCP server "shop" have already been discovered
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then tellme never contacted the MCP server "shop"
      And the request offered the tool "lookup_price" from the MCP server "shop" alongside the agent tools
      And tellme exits successfully

    Example: A fresh session keeps the remembered tool list
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      And the tools of the MCP server "shop" have already been discovered
      When the operator starts a fresh session with "--new" and the prompt "Which tools can you use?"
      Then tellme never contacted the MCP server "shop"
      And the request offered the tool "lookup_price" from the MCP server "shop" alongside the agent tools
      And tellme exits successfully

  Rule: A server's tools are discovered with no memory yet, then remembered

    Example: The first prompt discovers and remembers the server's tools
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then tellme contacted the MCP server "shop"
      And tellme remembered the tools of the MCP server "shop"
      And the request offered the tool "lookup_price" from the MCP server "shop" alongside the agent tools
      And tellme exits successfully

  Rule: An aged tool list is served before the request and refreshed afterwards

    # The server is closed after its tools were cached, so a SYNCHRONOUS
    # revalidation would fail and drop the tool — the discriminating witness that
    # a stale entry is served, not re-fetched on the critical path.

    Example: An aged tool list is still offered to the model
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint reports the offered tools and then answers with "done"
      And the tools of the MCP server "shop" were discovered more than a day ago
      And the MCP server "shop" has since stopped answering
      When the operator starts tellme with the prompt "Which tools can you use?"
      Then the request offered the tool "lookup_price" from the MCP server "shop" alongside the agent tools
      And tellme reported on stderr that the MCP server "shop" could not be reached
      And tellme exits successfully

  Rule: A remembered tool that the server can no longer serve fails softly

    Example: A remembered tool against a server that fails the call
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that fails the tool call
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the price" and then answers with "I could not check the price."
      And the tools of the MCP server "shop" have already been discovered
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the request offered the tool "lookup_price" from the MCP server "shop" alongside the agent tools
      And the run continued past the failed MCP tool call
      And tellme prints the provider's answer "I could not check the price."
      And tellme exits successfully

    Example: A remembered tool against a server that has stopped
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a remote MCP server "shop" that offers a tool "lookup_price" answering "$42"
      And a configured provider "test-model" whose endpoint asks tellme to use the MCP tool "lookup_price" from the server "shop" with the reason "check the price" and then answers with "I could not check the price."
      And the tools of the MCP server "shop" have already been discovered
      And the MCP server "shop" has since stopped answering
      When the operator starts tellme with the prompt "What does the gadget cost?"
      Then the request offered the tool "lookup_price" from the MCP server "shop" alongside the agent tools
      And the run continued past the failed MCP tool call
      And tellme prints the provider's answer "I could not check the price."
      And tellme exits successfully
