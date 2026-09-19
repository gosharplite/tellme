Feature: Reading a local image with the agent

  # Interface truth (CLI end, `chat` module) — round 062. tellme gains its first
  # non-text capability: the agent can read a local image and show it to a
  # vision-capable provider. The image rides the OpenAI-compatible wire as an
  # inline base64 `image_url` block on a `user` message (the family's inline path);
  # the image's kind is resolved from its CONTENT (magic bytes), never its name.
  # The `read_image` tool is offered ONLY when the selected provider declares the
  # capability (`VISION: true`); a provider without it is offered no such tool.
  # An image past the family's inline ceiling, or a file that is not a supported
  # picture, fails LOUDLY and no image reaches the request.
  #
  # Round 063 (ADR 0033) extends this feature to the SECOND family: the
  # Gemini/Vertex wire carries the same image as an `inlineData` blob on a `user`
  # turn (the round-062 loud Gemini refusal is retired). The capability key, the
  # offered-set gate, and the `read_image` tool are UNCHANGED; the inline ceiling
  # is now family-aware and enforced by the tool as a loud refusal.
  # Acceptance journeys: features/acceptance/reading-a-local-image-with-the-agent.feature
  # and features/acceptance/reading-a-local-image-with-a-gemini-provider.feature.

  Rule: The agent can show a local image to a model that can see

    Example: The model reads a picture and the request carries it
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "shot.png" that is a PNG picture
      And a configured provider "eye" that can take images and whose endpoint asks tellme to read "shot.png" and then answers with "a red square"
      When the operator starts tellme with the prompt "What is in shot.png?"
      Then the request carried the image file "shot.png"
      And the image was attached as a "image/png" picture
      And tellme prints the provider's answer "a red square"
      And tellme exits successfully

    Example: The picture's kind comes from its content, not its name
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "mystery.bin" that is a JPEG picture
      And a configured provider "eye" that can take images and whose endpoint asks tellme to read "mystery.bin" and then answers with "a cat"
      When the operator starts tellme with the prompt "What is in mystery.bin?"
      Then the request carried the image file "mystery.bin"
      And the image was attached as a "image/jpeg" picture
      And tellme exits successfully

    Example: A turn with no image is untouched
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "eye" whose endpoint answers with "plain answer"
      When the operator starts tellme with the prompt "Say hello"
      Then the request carried no image
      And tellme prints the provider's answer "plain answer"
      And tellme exits successfully

  Rule: The read-image tool is offered only when the selected model can see

    Example: A model that cannot see is offered no read-image tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "blind" that cannot take images whose endpoint reports the offered tools and then answers with "I cannot see images"
      When the operator starts tellme with the prompt "What is in shot.png?"
      Then the request offered no read-image tool
      And tellme exits successfully

    Example: A model that can see is offered the read-image tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "eye" that can take images whose endpoint reports the offered tools and then answers with "ready"
      When the operator starts tellme with the prompt "What tools can you use?"
      Then the request offered the read-image tool
      And tellme exits successfully

  Rule: An image that cannot be sent fails loudly, never silently

    Example: An image past the size limit is refused before it is sent
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "huge.png" larger than the image size limit
      And a configured provider "eye" that can take images and whose endpoint asks tellme to read "huge.png" and then answers with "ok"
      When the operator starts tellme with the prompt "What is in huge.png?"
      Then the tool result reported the image is too large
      And the request carried no image
      And tellme exits successfully

    Example: A file that is not a picture is refused, whatever it is called
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds a file "notes.png" whose content is not a supported picture
      And a configured provider "eye" that can take images and whose endpoint asks tellme to read "notes.png" and then answers with "ok"
      When the operator starts tellme with the prompt "What is in notes.png?"
      Then the tool result reported the content is not a supported picture
      And the request carried no image
      And tellme exits successfully

  Rule: A Gemini model that can see is shown a local image

    Example: The model reads a picture and the Vertex request carries it
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "shot.png" that is a PNG picture
      And a configured Gemini provider "eye" that can take images and whose endpoint asks tellme to read "shot.png" and then answers with "a red square"
      When the operator starts tellme with the prompt "What is in shot.png?"
      Then the request carried the image file "shot.png"
      And the image was attached as a "image/png" picture
      And tellme prints the provider's answer "a red square"
      And tellme exits successfully

    Example: The picture's kind comes from its content, not its name
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "mystery.bin" that is a JPEG picture
      And a configured Gemini provider "eye" that can take images and whose endpoint asks tellme to read "mystery.bin" and then answers with "a cat"
      When the operator starts tellme with the prompt "What is in mystery.bin?"
      Then the request carried the image file "mystery.bin"
      And the image was attached as a "image/jpeg" picture
      And tellme exits successfully

    Example: A turn with no image is untouched
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "eye" that can take images whose endpoint answers with "plain answer"
      When the operator starts tellme with the prompt "Say hello"
      Then the request carried no image
      And tellme prints the provider's answer "plain answer"
      And tellme exits successfully

  Rule: One declaration of the capability serves both families

    Example: A Gemini provider that cannot see is offered no read-image tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "blind" that cannot take images whose endpoint answers with "I cannot see images"
      When the operator starts tellme with the prompt "What is in shot.png?"
      Then the request offered no read-image tool
      And tellme exits successfully

    Example: A Gemini provider that can see is offered the read-image tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "eye" that can take images whose endpoint answers with "ready"
      When the operator starts tellme with the prompt "What tools can you use?"
      Then the request offered the read-image tool
      And tellme exits successfully

  Rule: An image the Gemini family cannot carry fails loudly, never silently

    Example: An image past the size limit is refused before it is sent
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "huge.png" larger than the image size limit
      And a configured Gemini provider "eye" that can take images and whose endpoint asks tellme to read "huge.png" and then answers with "ok"
      When the operator starts tellme with the prompt "What is in huge.png?"
      Then the tool result reported the image is too large
      And the request carried no image
      And tellme exits successfully

    Example: A file that is not a picture is refused, whatever it is called
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds a file "notes.png" whose content is not a supported picture
      And a configured Gemini provider "eye" that can take images and whose endpoint asks tellme to read "notes.png" and then answers with "ok"
      When the operator starts tellme with the prompt "What is in notes.png?"
      Then the tool result reported the content is not a supported picture
      And the request carried no image
      And tellme exits successfully

  Rule: The protection is part of every delivery

    # Documented narrowing (the round-059 second-Rule precedent): whether a real
    # provider would accept an image cannot be asked through the built binary
    # alone. The carriers are the Examples above (which observe the bytes the
    # fake provider recorded) plus the hermetic unit pin over the request body,
    # designed to fail if the image serialization is removed (ADR 0032
    # Verification). No `# [need clarification]` gap remains: the clarify round
    # closed all five decisions (family scope · capability key · tool surface ·
    # offered surface · size ceiling).
