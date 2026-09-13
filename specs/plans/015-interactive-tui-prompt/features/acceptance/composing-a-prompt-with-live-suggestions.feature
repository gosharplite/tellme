Feature: Composing a prompt with live suggestions

  # Acceptance only: what the operator experiences when they open the rich
  # interactive prompt (-i) and type — recent-prompt, workspace-path, and tool
  # suggestions appear, are accepted with a keystroke, and the composed text
  # runs one reasoning turn, without fighting shell quoting.

  Rule: Typing at the interactive prompt offers matching suggestions

    Example: The operator narrows to a recent prompt and submits it
      Given the operator has a runnable tellme installation
      And the shared prompt log already contains "review the last two commits"
      And the shared prompt log already contains "deploy to staging with version 015"
      When the operator opens the interactive prompt
      Then the interactive prompt offers the recent prompts "deploy to staging with version 015" and "review the last two commits"
      When the operator types "deploy"
      Then the interactive prompt narrows its suggestions to "deploy to staging with version 015"
      When the operator accepts the suggestion
      And the operator submits the prompt
      Then tellme runs one reasoning turn with the prompt "deploy to staging with version 015"
      And tellme prints the model's answer

    Example: The operator narrows to a workspace entry and submits
      Given the operator has a runnable tellme installation
      And the working directory contains a file "notes.txt"
      When the operator opens the interactive prompt
      And the operator types "./"
      Then the interactive prompt suggests the workspace entry "notes.txt"
      When the operator accepts the suggestion
      And the operator submits the prompt
      Then tellme runs one reasoning turn with a prompt that mentions "notes.txt"

    Example: The operator discovers an available tool
      Given the operator has a runnable tellme installation
      And tellme offers a tool named "read_files"
      When the operator opens the interactive prompt
      And the operator types "read"
      Then the interactive prompt suggests the available tool "read_files"

  Rule: The interactive prompt distinguishes submit from a new line and from abort

    Example: The operator writes several lines, then submits
      Given the operator has a runnable tellme installation
      When the operator opens the interactive prompt
      And the operator types two lines of text
      Then the interactive prompt keeps the line break in the editor
      When the operator submits the prompt
      Then tellme runs one reasoning turn whose prompt contains both lines

    Example: Aborting the prompt sends nothing to the provider
      Given the operator has a runnable tellme installation
      When the operator opens the interactive prompt
      And the operator types "wipe the workspace"
      And the operator aborts the prompt
      Then tellme sends no request to the provider
      And tellme exits successfully
