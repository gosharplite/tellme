Feature: Reading a local image with a Gemini provider

  # Acceptance only (round 063, operator request). Business language only.
  #
  # The operator asked whether the reference's Vertex Gemini could see images
  # (it can) and asked tellme to do the same. Round 062 gave tellme the
  # `read_image` capability on the OpenAI-compatible family and made the Gemini
  # family refuse an image loudly. This journey is about the same capability
  # working on a Gemini provider — one capability, one declaration, two families.

  Rule: A Gemini model that can see is shown a local image

    Example: The operator asks about a picture and the Gemini model answers
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "shot.png" that is a small PNG picture
      And a configured Gemini provider "eye" that can take images and whose endpoint records the request then answers "a red square"
      When the operator starts tellme with the prompt "What is in shot.png?"
      And the model asks to read the image file "shot.png"
      Then the request carried the image file "shot.png" to the Gemini provider
      And the image was attached as a "image/png" picture
      And tellme prints the provider's answer "a red square"
      And tellme exits successfully

    Example: The picture's kind is taken from its content, not its name
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "mystery.bin" that is a small JPEG picture
      And a configured Gemini provider "eye" that can take images and whose endpoint records the request then answers "a cat"
      When the operator starts tellme with the prompt "What is in mystery.bin?"
      And the model asks to read the image file "mystery.bin"
      Then the request carried the image file "mystery.bin" to the Gemini provider
      And the image was attached as a "image/jpeg" picture
      And tellme exits successfully

    Example: A turn with no image is untouched
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "eye" that can take images and whose endpoint records the request then answers "plain answer"
      When the operator starts tellme with the prompt "Say hello"
      Then the request carried no image
      And tellme prints the provider's answer "plain answer"
      And tellme exits successfully

  Rule: One declaration of the capability serves both families

    Example: A Gemini provider that cannot see is never offered the tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "blind" that cannot take images and whose endpoint records the request then answers "I cannot see images"
      When the operator starts tellme with the prompt "What is in shot.png?"
      Then the request offered no read-image tool
      And tellme exits successfully

    Example: A Gemini provider that can see is offered the same tool as before
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured Gemini provider "eye" that can take images and whose endpoint records the request then answers "ready"
      When the operator starts tellme with the prompt "What tools can you use?"
      Then the request offered the read-image tool
      And tellme exits successfully

  Rule: An image the Gemini family cannot carry still fails loudly, never silently

    Example: An image past the Gemini size limit is refused before it is sent
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "huge.png" that is larger than the Gemini image size limit
      And a configured Gemini provider "eye" that can take images and whose endpoint records the request then answers "ok"
      When the operator starts tellme with the prompt "What is in huge.png?"
      And the model asks to read the image file "huge.png"
      Then the tool result reported the image is too large
      And the request carried no image
      And tellme exits successfully

    Example: A file that is not a picture is refused on the Gemini path too
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds a file "notes.png" whose content is not a supported picture
      And a configured Gemini provider "eye" that can take images and whose endpoint records the request then answers "ok"
      When the operator starts tellme with the prompt "What is in notes.png?"
      And the model asks to read the image file "notes.png"
      Then the tool result reported the content is not a supported picture
      And the request carried no image
      And tellme exits successfully

  Rule: The protection is part of every delivery

    # Documented narrowing (the round-059 second-Rule precedent): whether a real
    # Vertex endpoint would accept an image cannot be asked through the built
    # binary alone. The carriers are therefore the Examples above (which observe
    # the bytes the Vertex-shaped fake recorded — the image reaching that wire,
    # the offered tool set, the loud family-aware refusals) plus the hermetic
    # unit pin over the Gemini request body, designed to fail if the `inline_data`
    # serialization is removed. No `# [need clarification]` gap remains: the
    # clarify round closed both decisions (Q1 → reuse the single `VISION` key;
    # Q2 → a family-aware inline ceiling enforced by the tool as a loud refusal).
