Feature: Discovering the callable name of an MCP tool

  # Acceptance journey (plan-side) — round 077, anchor issue #155. A remote MCP
  # server's tool is callable only under tellme's namespaced wire name
  # (`mcp_<server>_<tool>`), but a weaker model may recall the tool by its bare
  # upstream name. tellme makes the callable name discoverable in the tool's own
  # offered description, so the model can read the correct name, and never
  # advertises an uncallable name.

  Rule: An offered MCP tool tells the model the name it must call
    When tellme offers a server's tool, the text the model reads names the exact
    wire name that call will be dispatched under, so the model does not have to
    infer the namespacing.

    Example: The offered tool names its callable wire name
      Given a remote MCP server offers a tool with a description
      When tellme offers that tool to the model
      Then the offered description names the tool's callable wire name
      And the server's own description text is preserved

  Rule: A server that ships no description is still offered by its callable name
    When a server advertises a tool without any description, tellme must not fall
    back to a text that names the tool by something the model cannot call.

    Example: A description-less tool is offered by its callable name
      Given a remote MCP server offers a tool without a description
      When tellme offers that tool to the model
      Then the offered description names the tool's callable wire name
      And the offered description does not present the bare upstream name as callable
