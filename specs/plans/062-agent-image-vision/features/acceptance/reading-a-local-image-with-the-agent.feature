Feature: Reading a local image with the agent

  # Acceptance only (round 062, operator request). Business language only.
  #
  # The operator asked tellme to be able to read an image, like the reference
  # tool can. tellme is text-only today, so this journey is about the agent
  # being shown a picture the operator already has on disk — and about telling
  # the truth when the selected model simply cannot look at pictures.

  Rule: The agent can show a local image to a model that can see

    Example: The operator asks about a picture and the model answers
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "shot.png" that is a small PNG picture
      And a configured provider "eye" that can take images and whose endpoint records the request then answers "a red square"
      When the operator starts tellme with the prompt "What is in shot.png?"
      And the model asks to read the image file "shot.png"
      Then the request carried the image file "shot.png"
      And the image was attached as a "image/png" picture
      And tellme prints the provider's answer "a red square"
      And tellme exits successfully

    Example: The picture's kind is taken from its content, not its name
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "mystery.bin" that is a small JPEG picture
      And a configured provider "eye" that can take images and whose endpoint records the request then answers "a cat"
      When the operator starts tellme with the prompt "What is in mystery.bin?"
      And the model asks to read the image file "mystery.bin"
      Then the request carried the image file "mystery.bin"
      And the image was attached as a "image/jpeg" picture
      And tellme exits successfully

    Example: A turn with no image is untouched
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "eye" that can take images and whose endpoint records the request then answers "plain answer"
      When the operator starts tellme with the prompt "Say hello"
      Then the request carried no image
      And tellme prints the provider's answer "plain answer"
      And tellme exits successfully

  Rule: The read-image tool is offered only when the selected model can see

    Example: A model that cannot see is never offered the tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "shot.png" that is a small PNG picture
      And a configured provider "blind" that cannot take images and whose endpoint records the request then answers "I cannot see images"
      When the operator starts tellme with the prompt "What is in shot.png?"
      Then the request offered no read-image tool
      And tellme exits successfully

    Example: A model that can see is offered the tool
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And a configured provider "eye" that can take images and whose endpoint records the request then answers "ready"
      When the operator starts tellme with the prompt "What tools can you use?"
      Then the request offered the read-image tool
      And tellme exits successfully

  Rule: An image that cannot be sent fails loudly, never silently

    Example: An image past the size limit is refused before it is sent
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds an image file "huge.png" that is larger than the image size limit
      And a configured provider "eye" that can take images and whose endpoint records the request then answers "ok"
      When the operator starts tellme with the prompt "What is in huge.png?"
      And the model asks to read the image file "huge.png"
      Then the tool result reported the image is too large
      And the request carried no image
      And tellme exits successfully

    Example: A file that is not a picture is refused, whatever it is called
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds a file "notes.png" whose content is not a supported picture
      And a configured provider "eye" that can take images and whose endpoint records the request then answers "ok"
      When the operator starts tellme with the prompt "What is in notes.png?"
      And the model asks to read the image file "notes.png"
      Then the tool result reported the content is not a supported picture
      And the request carried no image
      And tellme exits successfully

  Rule: The protection is part of every delivery

    # Documented narrowing (the round-059 second-Rule precedent): whether a real
    # provider would accept an image cannot be asked through the built binary
    # alone. The carriers are therefore the Examples above (which observe the
    # bytes the fake provider recorded — the image reaching the wire, the
    # offered tool set, the loud refusals) plus the hermetic unit pin over the
    # request body, designed to fail if the image serialization is removed.
    # No `# [need clarification]` gap remains: the clarify round closed all
    # five decisions (Q1 → OpenAI-compatible family only; Q2 → an explicit
    # per-provider capability key; Q3 → read_image only; Q4 → the tool is
    # offered only for a seeing provider; Q5 → inline, 32 MiB, loud past it).
