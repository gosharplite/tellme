Feature: Listing the available skills

  # Interface truth (CLI end, `chat` module) — round 033: on a prompt-bearing turn the agent can list the
  # skills the workspace holds in its skills directory, by name and description, through a read-only
  # `list_skills` tool. The listing is **on-demand**: tellme adds **no** skill content to the request
  # unless the agent opens a skill with the existing `read_files` tool. A workspace with no skills lists
  # none. (No skills.sh ecosystem; no automatic injection.) Acceptance journey:
  # features/acceptance/discovering-the-available-skills.feature.

  Rule: A prompt can list the workspace's skills

    Example: The workspace holds two skills and both are listed
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds a skill "golang-patterns" described as "Idiomatic Go patterns"
      And the workspace holds a skill "golang-testing" described as "Go testing patterns"
      And a configured provider "test-model" whose endpoint asks tellme to list its skills and then answers with "done"
      When the operator starts tellme with the prompt "What skills do you have?"
      Then tellme listed the skills using its list_skills tool
      And the listing includes the skill "golang-patterns"
      And the listing includes the skill "golang-testing"
      And tellme exits successfully

  Rule: A skill's reference material is not listed as a skill

    Example: A skill that also ships a reference file
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds a skill "golang-patterns" whose folder also holds a reference file "rules/reference.md"
      And a configured provider "test-model" whose endpoint asks tellme to list its skills and then answers with "done"
      When the operator starts tellme with the prompt "What skills do you have?"
      Then the listing includes the skill "golang-patterns"
      And the listing does not include "reference"
      And tellme exits successfully

  Rule: A workspace with no skills lists none

    Example: No skills installed
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds no skills
      And a configured provider "test-model" whose endpoint asks tellme to list its skills and then answers with "I have no skills."
      When the operator starts tellme with the prompt "What skills do you have?"
      Then tellme listed the skills using its list_skills tool
      And the listing reports that no skills are available
      And tellme exits successfully

  Rule: A listed skill can be opened on demand

    Example: The agent lists the skills and then reads one
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds a skill "golang-testing" described as "Go testing patterns"
      And a configured provider "test-model" whose endpoint asks tellme to list its skills, then to read the skill "golang-testing", and then answers with "noted"
      When the operator starts tellme with the prompt "Follow the golang-testing skill."
      Then tellme listed the skills using its list_skills tool
      And tellme read the skill "golang-testing" using its read_files tool
      And tellme exits successfully

  Rule: Skill content is not added to the request unless a skill is opened

    Example: A prompt answered without opening a skill carries no skill content
      Given the operator has a runnable tellme installation
      And the runtime home is "ait-tmg"
      And the workspace holds a skill "golang-patterns" described as "UNIQUE-DESCRIPTION-TOKEN"
      And a configured provider "test-model" whose endpoint answers with "done"
      When the operator starts tellme with the prompt "Say hello."
      Then the request carried none of the text "UNIQUE-DESCRIPTION-TOKEN"
      And tellme exits successfully
